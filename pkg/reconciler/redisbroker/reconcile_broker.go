package redisbroker

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	appsv1listers "k8s.io/client-go/listers/apps/v1"
	corev1listers "k8s.io/client-go/listers/core/v1"
	"knative.dev/eventing/pkg/apis/duck"
	"knative.dev/pkg/logging"
	pkgreconciler "knative.dev/pkg/reconciler"

	eventingv1alpha1 "github.com/triggermesh/triggermesh-core/pkg/apis/eventing/v1alpha1"
	"github.com/triggermesh/triggermesh-core/pkg/reconciler"
	"github.com/triggermesh/triggermesh-core/pkg/reconciler/resources"
	"github.com/triggermesh/triggermesh-core/pkg/reconciler/semantic"
)

const (
	configSecretFile = "broker.config"
	configSecretPath = "/opt/broker"

	brokerResourceSuffix = "redisbroker-broker"
)

// var (
// 	configMountedPath = path.Join(configSecretPath, configSecretFile)
// )

type brokerReconciler struct {
	client           kubernetes.Interface
	deploymentLister appsv1listers.DeploymentLister
	serviceAccountLister    corev1listers.ServiceAccountLister
	serviceLister    corev1listers.ServiceLister
	endpointsLister  corev1listers.EndpointsLister
	image            string
	pullPolicy       corev1.PullPolicy
}

func (r *brokerReconciler) reconcile(ctx context.Context, rb *eventingv1alpha1.RedisBroker, redis *corev1.Service, secret *corev1.Secret) (*appsv1.Deployment, *corev1.Service, error) {
	sa, err := r.reconcileServiceAccount(ctx, rb)
	if err != nil {
		return nil, nil, err
	}

	d, err := r.reconcileDeployment(ctx, rb,sa,  redis, secret)
	if err != nil {
		return nil, nil, err
	}

	svc, err := r.reconcileService(ctx, rb)
	if err != nil {
		return d, nil, err
	}

	_, err = r.reconcileEndpoints(ctx, svc, rb)
	if err != nil {
		return d, nil, err
	}

	return d, svc, nil
}

func buildBrokerServiceAccount(rb *eventingv1alpha1.RedisBroker) *corev1.ServiceAccount {
	return resources.NewServiceAccount(rb.Namespace, rb.Name+"-"+brokerResourceSuffix,
		resources.ServiceAccountWithMetaOptions(
			resources.MetaAddLabel(appAnnotation, appAnnotationValue),
			resources.MetaAddLabel("component", brokerResourceSuffix),
			resources.MetaAddLabel(resourceNameAnnotation, rb.Name+"-"+brokerResourceSuffix),
			resources.MetaAddOwner(rb, rb.GetGroupVersionKind())))
}

func (r *brokerReconciler) reconcileServiceAccount(ctx context.Context, rb *eventingv1alpha1.RedisBroker) (*corev1.ServiceAccount, error) {
	desired := buildBrokerServiceAccount(rb)
	current, err := r.serviceAccountLister.ServiceAccounts(desired.Namespace).Get(desired.Name)

	switch {
	case err == nil:
		// TODO check RoleBinding

	case !apierrs.IsNotFound(err):
		// An error occurred retrieving current object.
		fullname := types.NamespacedName{Namespace: desired.Namespace, Name: desired.Name}
		logging.FromContext(ctx).Error("Unable to get broker ServiceAccount", zap.String("serviceAccount", fullname.String()), zap.Error(err))
		rb.Status.MarkBrokerServiceAccountFailed(reconciler.ReasonFailedServiceAccountGet, "Failed to get broker ServiceAccount")

		return nil, pkgreconciler.NewEvent(corev1.EventTypeWarning, reconciler.ReasonFailedServiceAccountGet,
			"Failed to get broker ServiceAccount %s: %w", fullname, err)

	default:
		// The ServiceAccount has not been found, create it.
		current, err = r.client.CoreV1().ServiceAccounts(desired.Namespace).Create(ctx, desired, metav1.CreateOptions{})
		if err != nil {
			fullname := types.NamespacedName{Namespace: desired.Namespace, Name: desired.Name}
			logging.FromContext(ctx).Error("Unable to create broker ServiceAccount", zap.String("serviceAccount", fullname.String()), zap.Error(err))
			rb.Status.MarkBrokerServiceAccountFailed(reconciler.ReasonFailedServiceAccountCreate, "Failed to create broker ServiceAccount")

			return nil, pkgreconciler.NewEvent(corev1.EventTypeWarning, reconciler.ReasonFailedServiceAccountCreate,
				"Failed to create broker ServiceAccount %s: %w", fullname, err)
		}
	}

	// Update status
	rb.Status.MarkBrokerServiceAccountReady()

	return current, nil
}

func buildBrokerDeployment(rb *eventingv1alpha1.RedisBroker,sa *corev1.ServiceAccount, redis *corev1.Service, secret *corev1.Secret, image string, pullPolicy corev1.PullPolicy) *appsv1.Deployment {

	// v := resources.NewVolume("config",
	// 	resources.VolumeFromSecretOption(secret.Name, configSecretKey, configSecretFile))
	// vm := resources.NewVolumeMount("config", configSecretPath,
	// 	resources.VolumeMountWithReadOnlyOption(true))

	var stream string
	if rb.Spec.Redis != nil && rb.Spec.Redis.Stream != nil && *rb.Spec.Redis.Stream != "" {
		stream = *rb.Spec.Redis.Stream
	} else {
		stream = rb.Namespace + "." + rb.Name
	}

	opts := []resources.ContainerOption{
		resources.ContainerAddArgs("start"),
		// resources.ContainerAddVolumeMount(vm),
		// resources.ContainerAddEnvFromValue("BROKER_CONFIG_PATH", configMountedPath),
		resources.ContainerAddEnvFromFieldRef("KUBERNETES_NAMESPACE", "metadata.namespace"),
		// resources.ContainerAddEnvFromValue("KUBERNETES_NAMESPACE", rb.Namespace),
		resources.ContainerAddEnvFromValue("BROKER_CONFIG_KUBERNETES_SECRET_NAME", secret.Name),
		resources.ContainerAddEnvFromValue("BROKER_CONFIG_KUBERNETES_SECRET_KEY", configSecretKey),

		resources.ContainerAddEnvFromValue("REDIS_STREAM", stream),
		resources.ContainerWithImagePullPolicy(pullPolicy),
	}

	if rb.Spec.Redis != nil && rb.Spec.Redis.StreamMaxLen != nil && *rb.Spec.Redis.StreamMaxLen != 0 {
		opts = append(opts, resources.ContainerAddEnvFromValue("REDIS_STREAM_MAXLEN", stream))
	}

	if rb.IsUserProvidedRedis() {
		opts = append(opts, resources.ContainerAddEnvFromValue("REDIS_ADDRESS", rb.Spec.Redis.Connection.URL))

		if rb.Spec.Redis.Connection.Username != nil {
			opts = append(opts, resources.ContainerAddEnvVarFromSecret("REDIS_USERNAME",
				rb.Spec.Redis.Connection.Username.SecretKeyRef.Name,
				rb.Spec.Redis.Connection.Username.SecretKeyRef.Key))
		}

		if rb.Spec.Redis.Connection.Password != nil {
			opts = append(opts, resources.ContainerAddEnvVarFromSecret("REDIS_PASSWORD",
				rb.Spec.Redis.Connection.Password.SecretKeyRef.Name,
				rb.Spec.Redis.Connection.Password.SecretKeyRef.Key))
		}

		if rb.Spec.Redis.Connection.TLSEnabled != nil && *rb.Spec.Redis.Connection.TLSEnabled {
			opts = append(opts, resources.ContainerAddEnvFromValue("REDIS_TLS_ENABLED", "true"))
		}

		if rb.Spec.Redis.Connection.TLSSkipVerify != nil && *rb.Spec.Redis.Connection.TLSSkipVerify {
			opts = append(opts, resources.ContainerAddEnvFromValue("REDIS_TLS_SKIP_VERIFY", "true"))
		}

	} else {
		opts = append(opts, resources.ContainerAddEnvFromValue("REDIS_ADDRESS",
			fmt.Sprintf("%s:%d", redis.Name, redis.Spec.Ports[0].Port)))
	}

	return resources.NewDeployment(rb.Namespace, rb.Name+"-"+brokerResourceSuffix,
		resources.DeploymentWithMetaOptions(
			resources.MetaAddLabel(appAnnotation, appAnnotationValue),
			resources.MetaAddLabel("component", brokerResourceSuffix),
			resources.MetaAddLabel(resourceNameAnnotation, rb.Name+"-"+brokerResourceSuffix),
			resources.MetaAddOwner(rb, rb.GetGroupVersionKind())),
		resources.DeploymentAddSelectorForTemplate(resourceNameAnnotation, rb.Name+"-"+brokerResourceSuffix),
		resources.DeploymentSetReplicas(1),
		resources.DeploymentWithTemplateOptions(
			// resources.PodSpecAddVolume(v),
			resources.PodSpecAddContainer(
				resources.NewContainer("broker", image, opts...))))
}

func (r *brokerReconciler) reconcileDeployment(ctx context.Context, rb *eventingv1alpha1.RedisBroker,sa *corev1.ServiceAccount, redis *corev1.Service, secret *corev1.Secret) (*appsv1.Deployment, error) {
	desired := buildBrokerDeployment(rb,sa, redis, secret, r.image, r.pullPolicy)
	current, err := r.deploymentLister.Deployments(desired.Namespace).Get(desired.Name)
	switch {
	case err == nil:
		// Compare current object with desired, update if needed.
		if !semantic.Semantic.DeepEqual(desired, current) {
			desired.Status = current.Status
			desired.ResourceVersion = current.ResourceVersion

			current, err = r.client.AppsV1().Deployments(desired.Namespace).Update(ctx, desired, metav1.UpdateOptions{})
			if err != nil {
				fullname := types.NamespacedName{Namespace: desired.Namespace, Name: desired.Name}
				logging.FromContext(ctx).Error("Unable to update broker deployment", zap.String("deployment", fullname.String()), zap.Error(err))
				rb.Status.MarkBrokerDeploymentFailed(reconciler.ReasonFailedDeploymentUpdate, "Failed to update broker deployment")

				return nil, pkgreconciler.NewEvent(corev1.EventTypeWarning, reconciler.ReasonFailedDeploymentUpdate,
					"Failed to get broker deployment %s: %w", fullname, err)
			}
		}

	case !apierrs.IsNotFound(err):
		// An error occurred retrieving current deployment.
		fullname := types.NamespacedName{Namespace: desired.Namespace, Name: desired.Name}
		logging.FromContext(ctx).Error("Unable to get broker deployment", zap.String("deployment", fullname.String()), zap.Error(err))
		rb.Status.MarkBrokerDeploymentFailed(reconciler.ReasonFailedDeploymentGet, "Failed to get broker deployment")

		return nil, pkgreconciler.NewEvent(corev1.EventTypeWarning, reconciler.ReasonFailedDeploymentGet,
			"Failed to get broker deployment %s: %w", fullname, err)

	default:
		// The deployment has not been found, create it.
		current, err = r.client.AppsV1().Deployments(desired.Namespace).Create(ctx, desired, metav1.CreateOptions{})
		if err != nil {
			fullname := types.NamespacedName{Namespace: desired.Namespace, Name: desired.Name}
			logging.FromContext(ctx).Error("Unable to create broker deployment", zap.String("deployment", fullname.String()), zap.Error(err))
			rb.Status.MarkBrokerDeploymentFailed(reconciler.ReasonFailedDeploymentCreate, "Failed to create broker deployment")

			return nil, pkgreconciler.NewEvent(corev1.EventTypeWarning, reconciler.ReasonFailedDeploymentCreate,
				"Failed to create broker deployment %s: %w", fullname, err)
		}
	}

	// Update status based on deployment
	rb.Status.PropagateBrokerDeploymentAvailability(ctx, &current.Status)

	return current, nil
}

func buildBrokerService(rb *eventingv1alpha1.RedisBroker) *corev1.Service {
	return resources.NewService(rb.Namespace, rb.Name+"-"+brokerResourceSuffix,
		resources.ServiceWithMetaOptions(
			resources.MetaAddLabel(appAnnotation, appAnnotationValue),
			resources.MetaAddLabel("component", brokerResourceSuffix),
			resources.MetaAddLabel(resourceNameAnnotation, rb.Name+"-"+brokerResourceSuffix),
			resources.MetaAddOwner(rb, rb.GetGroupVersionKind())),
		resources.ServiceSetType(corev1.ServiceTypeClusterIP),
		resources.ServiceAddSelectorLabel(resourceNameAnnotation, rb.Name+"-"+brokerResourceSuffix),
		resources.ServiceAddPort("httpce", 8080, 8080))
}

func (r *brokerReconciler) reconcileService(ctx context.Context, rb *eventingv1alpha1.RedisBroker) (*corev1.Service, error) {
	desired := buildBrokerService(rb)
	current, err := r.serviceLister.Services(desired.Namespace).Get(desired.Name)
	switch {
	case err == nil:
		// Set Status
		// Compare current object with desired, update if needed.
		if !semantic.Semantic.DeepEqual(desired, current) {
			desired.Status = current.Status
			desired.ResourceVersion = current.ResourceVersion

			current, err = r.client.CoreV1().Services(desired.Namespace).Update(ctx, desired, metav1.UpdateOptions{})
			if err != nil {
				fullname := types.NamespacedName{Namespace: desired.Namespace, Name: desired.Name}
				logging.FromContext(ctx).Error("Unable to update broker service", zap.String("service", fullname.String()), zap.Error(err))
				rb.Status.MarkBrokerServiceFailed(reconciler.ReasonFailedServiceUpdate, "Failed to update broker service")

				return nil, pkgreconciler.NewEvent(corev1.EventTypeWarning, reconciler.ReasonFailedServiceUpdate,
					"Failed to get broker service %s: %w", fullname, err)
			}
		}

	case !apierrs.IsNotFound(err):
		// An error occurred retrieving current object.
		fullname := types.NamespacedName{Namespace: desired.Namespace, Name: desired.Name}
		logging.FromContext(ctx).Error("Unable to get the service", zap.String("service", fullname.String()), zap.Error(err))
		rb.Status.MarkBrokerServiceFailed(reconciler.ReasonFailedServiceGet, "Failed to get broker service")

		return nil, pkgreconciler.NewEvent(corev1.EventTypeWarning, reconciler.ReasonFailedServiceGet,
			"Failed to get broker service %s: %w", fullname, err)

	default:
		// The object has not been found, create it.
		current, err = r.client.CoreV1().Services(desired.Namespace).Create(ctx, desired, metav1.CreateOptions{})
		if err != nil {
			fullname := types.NamespacedName{Namespace: desired.Namespace, Name: desired.Name}
			logging.FromContext(ctx).Error("Unable to create the service", zap.String("service", fullname.String()), zap.Error(err))
			rb.Status.MarkBrokerServiceFailed(reconciler.ReasonFailedServiceCreate, "Failed to create broker service")

			return nil, pkgreconciler.NewEvent(corev1.EventTypeWarning, reconciler.ReasonFailedServiceCreate,
				"Failed to create broker service %s: %w", fullname, err)
		}
	}

	// Service exists and is up to date.
	rb.Status.MarkBrokerServiceReady()

	return current, nil
}

func (r *brokerReconciler) reconcileEndpoints(ctx context.Context, service *corev1.Service, rb *eventingv1alpha1.RedisBroker) (*corev1.Endpoints, error) {
	ep, err := r.endpointsLister.Endpoints(service.Namespace).Get(service.Name)
	switch {
	case err == nil:
		if duck.EndpointsAreAvailable(ep) {
			rb.Status.MarkBrokerEndpointsTrue()
			return ep, nil
		}

		rb.Status.MarkBrokerEndpointsFailed(reconciler.ReasonUnavailableEndpoints, "Endpoints for broker service are not available")
		return nil, pkgreconciler.NewEvent(corev1.EventTypeWarning, reconciler.ReasonUnavailableEndpoints,
			"Endpoints for broker service are not available %s",
			types.NamespacedName{Namespace: ep.Namespace, Name: ep.Name})

	case apierrs.IsNotFound(err):
		rb.Status.MarkBrokerEndpointsFailed(reconciler.ReasonUnavailableEndpoints, "Endpoints for broker service do not exist")
		return nil, pkgreconciler.NewEvent(corev1.EventTypeWarning, reconciler.ReasonUnavailableEndpoints,
			"Endpoints for broker service do not exist %s",
			types.NamespacedName{Namespace: service.Namespace, Name: service.Name})
	}

	fullname := types.NamespacedName{Namespace: service.Namespace, Name: service.Name}
	rb.Status.MarkBrokerEndpointsUnknown(reconciler.ReasonFailedEndpointsGet, "Could not retrieve endpoints for broker service")
	logging.FromContext(ctx).Error("Unable to get the broker service endpoints", zap.String("endpoint", fullname.String()), zap.Error(err))
	return nil, pkgreconciler.NewEvent(corev1.EventTypeWarning, reconciler.ReasonFailedEndpointsGet,
		"Failed to get broker service ednpoints %s: %w", fullname, err)
}
