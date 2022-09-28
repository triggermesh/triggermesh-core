package redisbroker

import (
	"context"
	"fmt"
	"path"

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
	configSecretFile = "config"
	configSecretPath = "/opt/broker"

	brokerResourceSuffix = "redisbroker-broker"
)

var (
	configMountedPath = path.Join(configSecretPath, configSecretFile)
)

type brokerReconciler struct {
	client           kubernetes.Interface
	deploymentLister appsv1listers.DeploymentLister
	serviceLister    corev1listers.ServiceLister
	endpointsLister  corev1listers.EndpointsLister
	image            string
}

func (r *brokerReconciler) reconcile(ctx context.Context, rb *eventingv1alpha1.RedisBroker, redis *corev1.Service, secret *corev1.Secret) (*appsv1.Deployment, *corev1.Service, error) {
	d, err := r.reconcileDeployment(ctx, rb, redis, secret)
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

func buildBrokerDeployment(rb *eventingv1alpha1.RedisBroker, redis *corev1.Service, secret *corev1.Secret, image string) *appsv1.Deployment {

	v := resources.NewVolume("config",
		resources.VolumeFromSecretOption(secret.Name, configSecretKey, configSecretFile))
	vm := resources.NewVolumeMount("config", configSecretPath,
		resources.VolumeMountWithReadOnlyOption(true))

	redisService := fmt.Sprintf("%s:%d", redis.Name, redis.Spec.Ports[0].Port)

	return resources.NewDeployment(rb.Namespace, rb.Name+"-"+brokerResourceSuffix,
		resources.DeploymentWithMetaOptions(
			resources.MetaAddLabel(appAnnotation, appAnnotationValue),
			resources.MetaAddLabel("component", brokerResourceSuffix),
			resources.MetaAddLabel(resourceNameAnnotation, rb.Name+"-"+brokerResourceSuffix),
			resources.MetaAddOwner(rb, rb.GetGroupVersionKind())),
		resources.DeploymentAddSelectorForTemplate(resourceNameAnnotation, rb.Name+"-"+brokerResourceSuffix),
		resources.DeploymentSetReplicas(1),
		resources.DeploymentWithTemplateOptions(
			resources.PodSpecAddVolume(v),
			resources.PodSpecAddContainer(
				resources.NewContainer("broker", image,
					resources.ContainerAddArgs("start --redis.address "+redisService+" --config-path "+configMountedPath),
					resources.ContainerAddVolumeMount(vm),
				),
			),
		))
}

func (r *brokerReconciler) reconcileDeployment(ctx context.Context, rb *eventingv1alpha1.RedisBroker, redis *corev1.Service, secret *corev1.Secret) (*appsv1.Deployment, error) {
	desired := buildBrokerDeployment(rb, redis, secret, r.image)
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
		// An error ocurred retrieving current deployment.
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
		// An error ocurred retrieving current object.
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
