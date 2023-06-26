// Copyright 2023 TriggerMesh Inc.
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"context"

	"go.uber.org/zap"
	"sigs.k8s.io/yaml"

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

	"github.com/triggermesh/brokers/pkg/config/broker"

	eventingv1alpha1 "github.com/triggermesh/triggermesh-core/pkg/apis/eventing/v1alpha1"
	eventingv1alpha1listers "github.com/triggermesh/triggermesh-core/pkg/client/generated/listers/eventing/v1alpha1"
	"github.com/triggermesh/triggermesh-core/pkg/reconciler/resources"
	"github.com/triggermesh/triggermesh-core/pkg/reconciler/semantic"
)

const (
	ConfigSecretKey      = "config"
	secretResourceSuffix = "config"
)

type SecretReconciler interface {
	Reconcile(ctx context.Context, rb eventingv1alpha1.ReconcilableBroker) (*corev1.Secret, error)
}

type secretReconciler struct {
	client        kubernetes.Interface
	secretLister  corev1listers.SecretLister
	triggerLister eventingv1alpha1listers.TriggerLister
}

var _ SecretReconciler = (*secretReconciler)(nil)

func NewSecretReconciler(ctx context.Context, secretLister corev1listers.SecretLister, triggerLister eventingv1alpha1listers.TriggerLister) SecretReconciler {
	return &secretReconciler{
		client:        k8sclient.Get(ctx),
		secretLister:  secretLister,
		triggerLister: triggerLister,
	}
}

func (r *secretReconciler) Reconcile(ctx context.Context, rb eventingv1alpha1.ReconcilableBroker) (*corev1.Secret, error) {
	desired, err := r.buildConfigSecret(ctx, rb)
	if err != nil {
		rb.GetReconcilableBrokerStatus().MarkConfigSecretFailed(ReasonFailedSecretCompose, "Failed to compose secret config from broker")
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
				logging.FromContext(ctx).Error("Unable to update secret", zap.String("secret", fullname.String()), zap.Error(err))
				rb.GetReconcilableBrokerStatus().MarkConfigSecretFailed(ReasonFailedSecretUpdate, "Failed to update config from secret")

				return nil, pkgreconciler.NewEvent(corev1.EventTypeWarning, ReasonFailedSecretUpdate,
					"Failed to update config from secret %s: %w", fullname, err)
			}
		}

	case !apierrs.IsNotFound(err):
		// An error occurred retrieving current secret.
		fullname := types.NamespacedName{Namespace: desired.Namespace, Name: desired.Name}
		logging.FromContext(ctx).Error("Unable to get secret", zap.String("secret", fullname.String()), zap.Error(err))
		rb.GetReconcilableBrokerStatus().MarkConfigSecretFailed(ReasonFailedSecretGet, "Failed to get config from secret")

		return nil, pkgreconciler.NewEvent(corev1.EventTypeWarning, ReasonFailedSecretGet,
			"Failed to get config from secret %s: %w", fullname, err)

	default:
		// The secret has not been found, create it.
		current, err = r.client.CoreV1().Secrets(desired.Namespace).Create(ctx, desired, metav1.CreateOptions{})
		if err != nil {
			fullname := types.NamespacedName{Namespace: desired.Namespace, Name: desired.Name}
			logging.FromContext(ctx).Error("Unable to create secret", zap.String("secret", fullname.String()), zap.Error(err))
			rb.GetReconcilableBrokerStatus().MarkConfigSecretFailed(ReasonFailedSecretCreate, "Failed to create secret for config")

			return nil, pkgreconciler.NewEvent(corev1.EventTypeWarning, ReasonFailedSecretCreate,
				"Failed to create secret for config %s: %w", fullname, err)
		}
	}

	rb.GetReconcilableBrokerStatus().MarkConfigSecretReady()

	return current, nil
}

func (r *secretReconciler) buildConfigSecret(ctx context.Context, rb eventingv1alpha1.ReconcilableBroker) (*corev1.Secret, error) {
	meta := rb.GetObjectMeta()
	ns, name := meta.GetNamespace(), meta.GetName()

	triggers, err := r.triggerLister.Triggers(ns).List(labels.Everything())
	if err != nil {
		logging.FromContext(ctx).Error("Unable to list triggers at namespace", zap.Error(err))
		rb.GetReconcilableBrokerStatus().MarkConfigSecretFailed(ReasonFailedTriggerList, "Failed to list triggers")

		return nil, pkgreconciler.NewEvent(corev1.EventTypeWarning, ReasonFailedTriggerList,
			"Failed to list triggers: %w", err)
	}

	cfg := &broker.Config{
		Triggers: make(map[string]broker.Trigger),
	}
	for _, t := range triggers {
		// Generate secret even if the trigger is not ready, as long as one of the URIs for target
		// or DLS exist.
		if !t.ReferencesBroker(rb) || (t.Status.TargetURI == nil && t.Status.DeadLetterSinkURI == nil) {
			continue
		}

		targetURI := ""
		if t.Status.TargetURI != nil {
			targetURI = t.Status.TargetURI.String()
		} else {
			// Configure empty URL so that all requests go to DLS when the target is
			// not ready.
			targetURI = ""
		}

		do := &broker.DeliveryOptions{}
		if t.Spec.Delivery != nil {
			do.Retry = t.Spec.Delivery.Retry
			do.BackoffDelay = t.Spec.Delivery.BackoffDelay

			if t.Spec.Delivery.BackoffPolicy != nil {
				var bop broker.BackoffPolicyType
				switch *t.Spec.Delivery.BackoffPolicy {
				case duckv1.BackoffPolicyLinear:
					bop = broker.BackoffPolicyLinear

				case duckv1.BackoffPolicyExponential:
					bop = broker.BackoffPolicyLinear
				}
				do.BackoffPolicy = &bop
			}

			if t.Status.DeadLetterSinkURI != nil {
				uri := t.Status.DeadLetterSinkURI.String()
				do.DeadLetterURL = &uri
			}
		}

		trg := broker.Trigger{
			Filters: t.Spec.Filters,
			Target: broker.Target{
				URL:             &targetURI,
				DeliveryOptions: do,
			},
		}

		if t.Spec.Bounds != nil {
			trg.Bounds = &broker.TriggerBounds{}
			if t.Spec.Bounds.ByDate != nil {
				trg.Bounds.ByDate = &broker.Bounds{
					Start: t.Spec.Bounds.ByDate.Start,
					End:   t.Spec.Bounds.ByDate.End,
				}
			}

			if t.Spec.Bounds.ById != nil {
				trg.Bounds.ByID = &broker.Bounds{
					Start: t.Spec.Bounds.ById.Start,
					End:   t.Spec.Bounds.ById.End,
				}
			}
		}

		// Add Trigger data to config
		cfg.Triggers[t.Name] = trg
	}

	// TODO add user/password

	b, err := yaml.Marshal(cfg)
	if err != nil {
		logging.FromContext(ctx).Error("Unable to marshal configuration into YAML", zap.Error(err))
		rb.GetReconcilableBrokerStatus().MarkConfigSecretFailed(ReasonFailedConfigSerialize, "Failed to serialize configuration")

		return nil, pkgreconciler.NewEvent(corev1.EventTypeWarning, ReasonFailedConfigSerialize,
			"Failed to serialize configuration: %w", err)
	}

	sn := name + "-" + rb.GetOwnedObjectsSuffix() + "-" + secretResourceSuffix

	return resources.NewSecret(ns, sn,
		resources.SecretWithMetaOptions(
			resources.MetaAddLabel(resources.AppNameLabel, AppAnnotationValue(rb)),
			resources.MetaAddLabel(resources.AppComponentLabel, "broker-config"),
			resources.MetaAddLabel(resources.AppPartOfLabel, resources.PartOf),
			resources.MetaAddLabel(resources.AppManagedByLabel, resources.ManagedBy),
			resources.MetaAddLabel(resources.AppInstanceLabel, sn),
			resources.MetaAddOwner(meta, rb.GetGroupVersionKind())),
		resources.SecretSetData(ConfigSecretKey, b)), nil
}
