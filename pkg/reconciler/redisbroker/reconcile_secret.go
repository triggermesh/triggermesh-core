package redisbroker

import (
	"context"

	"go.uber.org/zap"
	"sigs.k8s.io/yaml"

	"github.com/triggermesh/brokers/pkg/config"
	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	corev1listers "k8s.io/client-go/listers/core/v1"
	duckv1 "knative.dev/eventing/pkg/apis/duck/v1"
	k8sclient "knative.dev/pkg/client/injection/kube/client"
	"knative.dev/pkg/logging"
	pkgreconciler "knative.dev/pkg/reconciler"

	eventingv1alpha1 "github.com/triggermesh/triggermesh-core/pkg/apis/eventing/v1alpha1"
	eventingv1alpha1listers "github.com/triggermesh/triggermesh-core/pkg/client/generated/listers/eventing/v1alpha1"
	"github.com/triggermesh/triggermesh-core/pkg/reconciler"
	"github.com/triggermesh/triggermesh-core/pkg/reconciler/resources"
	"github.com/triggermesh/triggermesh-core/pkg/reconciler/semantic"
)

const configSecretKey = "config"

type secretReconciler struct {
	client        kubernetes.Interface
	secretLister  corev1listers.SecretLister
	triggerLister eventingv1alpha1listers.TriggerLister
}

func newSecretReconciler(ctx context.Context, secretLister corev1listers.SecretLister, triggerLister eventingv1alpha1listers.TriggerLister) secretReconciler {
	return secretReconciler{
		client:        k8sclient.Get(ctx),
		secretLister:  secretLister,
		triggerLister: triggerLister,
	}
}

func (r *secretReconciler) reconcile(ctx context.Context, rb *eventingv1alpha1.RedisBroker) (*corev1.Secret, error) {
	desired, err := r.buildConfigSecret(ctx, rb)
	if err != nil {
		return nil, err
	}

	current, err := r.secretLister.Secrets(desired.Namespace).Get(desired.Name)
	switch {
	case err == nil:
		// Compare current object with desired, update if needed.
		if !semantic.Semantic.DeepEqual(desired, current) {
			desired.ResourceVersion = current.ResourceVersion

			current, err = r.client.CoreV1().Secrets(desired.Namespace).Update(ctx, desired, metav1.UpdateOptions{})
			if err != nil {
				fullname := types.NamespacedName{Namespace: desired.Namespace, Name: desired.Name}
				logging.FromContext(ctx).Error("Unable to update the secret", zap.String("secret", fullname.String()), zap.Error(err))
				rb.Status.MarkRedisDeploymentFailed(reconciler.ReasonFailedDeploymentUpdate, "Failed to update Redis deployment")

				return nil, pkgreconciler.NewEvent(corev1.EventTypeWarning, reconciler.ReasonFailedDeploymentUpdate,
					"Failed to get Redis deployment %s: %w", fullname, err)
			}
		}

	case !apierrs.IsNotFound(err):
		// An error ocurred retrieving current deployment.
		fullname := types.NamespacedName{Namespace: desired.Namespace, Name: desired.Name}
		logging.FromContext(ctx).Error("Unable to get the deployment", zap.String("deployment", fullname.String()), zap.Error(err))
		rb.Status.MarkRedisDeploymentFailed(reconciler.ReasonFailedDeploymentGet, "Failed to get Redis deployment")

		return nil, pkgreconciler.NewEvent(corev1.EventTypeWarning, reconciler.ReasonFailedDeploymentGet,
			"Failed to get Redis deployment %s: %w", fullname, err)

	default:
		// The deployment has not been found, create it.
		current, err = r.client.CoreV1().Secrets(desired.Namespace).Create(ctx, desired, metav1.CreateOptions{})
		if err != nil {
			fullname := types.NamespacedName{Namespace: desired.Namespace, Name: desired.Name}
			logging.FromContext(ctx).Error("Unable to create the deployment", zap.String("deployment", fullname.String()), zap.Error(err))
			rb.Status.MarkRedisDeploymentFailed(reconciler.ReasonFailedDeploymentCreate, "Failed to create Redis deployment")

			return nil, pkgreconciler.NewEvent(corev1.EventTypeWarning, reconciler.ReasonFailedDeploymentCreate,
				"Failed to create Redis deployment %s: %w", fullname, err)
		}
	}

	rb.Status.MarkConfigSecretReady()

	return current, nil
}

func (r *secretReconciler) buildConfigSecret(ctx context.Context, rb *eventingv1alpha1.RedisBroker) (*corev1.Secret, error) {
	triggers, err := r.triggerLister.Triggers(rb.Namespace).List(labels.Everything())
	if err != nil {
		logging.FromContext(ctx).Error("Unable to list triggers at namespace", zap.Error(err))
		rb.Status.MarkConfigSecretFailed(reconciler.ReasonFailedTriggerList, "Failed to list triggers")

		return nil, pkgreconciler.NewEvent(corev1.EventTypeWarning, reconciler.ReasonFailedTriggerList,
			"Failed to list triggers: %w", err)
	}

	// TODO triggers must have resolved the URI, check status.

	cfg := &config.Config{}
	for _, t := range triggers {
		if !t.ReferencesBroker(rb) {
			continue
		}

		target := ""
		if t.Spec.Target.URI != nil {
			target = t.Spec.Target.URI.String()
			// } else {
			// TODO resolve target to URL
		}

		// TODO convert DeliveryOptions
		do := &config.DeliveryOptions{}
		if t.Spec.Delivery != nil {
			do.Retry = t.Spec.Delivery.Retry
			do.BackoffDelay = t.Spec.Delivery.BackoffDelay

			if t.Spec.Delivery.BackoffPolicy != nil {
				switch *t.Spec.Delivery.BackoffPolicy {
				case duckv1.BackoffPolicyLinear:
					*do.BackoffPolicy = config.BackoffPolicyLinear

				case duckv1.BackoffPolicyExponential:
					*do.BackoffPolicy = config.BackoffPolicyExponential
				}
			}

			// TODO resolve reference to URI
			uri := t.Spec.Delivery.DeadLetterSink.URI.String()
			do.DeadLetterURL = &uri
		}

		trg := config.Trigger{
			Name:    t.Name,
			Filters: t.Spec.Filters,
			Targets: []config.Target{
				{
					URL:             target,
					DeliveryOptions: do,
				},
			},
		}

		// Add Trigger data to config
		cfg.Triggers = append(cfg.Triggers, trg)
	}

	// add user/password

	b, err := yaml.Marshal(cfg)
	if err != nil {
		logging.FromContext(ctx).Error("Unable to marshal configuration into YAML", zap.Error(err))
		rb.Status.MarkConfigSecretFailed(reconciler.ReasonFailedConfigSerialize, "Failed to serialize configuration")

		return nil, pkgreconciler.NewEvent(corev1.EventTypeWarning, reconciler.ReasonFailedConfigSerialize,
			"Failed to serialize configuration: %w", err)
	}

	return resources.NewSecret(rb.Namespace, rb.Name,
		resources.SecretWithMetaOptions(
			resources.MetaAddLabel(appAnnotation, redisResourceSuffix),
			resources.MetaAddLabel(resourceNameAnnotation, rb.Name+"-"+"redisbroker-config"),
			resources.MetaAddOwner(rb, rb.GetGroupVersionKind())),
		resources.SecretSetData(configSecretKey, b)), nil
}
