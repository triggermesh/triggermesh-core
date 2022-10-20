package redisbroker

import (
	"context"

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
	redisResourceSuffix = "redisbroker-redis"
)

type redisReconciler struct {
	client           kubernetes.Interface
	deploymentLister appsv1listers.DeploymentLister
	serviceLister    corev1listers.ServiceLister
	endpointsLister  corev1listers.EndpointsLister
	image            string
}

func (r *redisReconciler) reconcile(ctx context.Context, rb *eventingv1alpha1.RedisBroker) (*appsv1.Deployment, *corev1.Service, error) {
	if rb.IsUserProvidedRedis() {
		// Nothing to do but mark the status for each of the elements reconciled.
		rb.Status.MarkRedisUserProvided()
		return nil, nil, nil
	}

	d, err := r.reconcileDeployment(ctx, rb)
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

func buildRedisDeployment(rb *eventingv1alpha1.RedisBroker, image string) *appsv1.Deployment {
	return resources.NewDeployment(rb.Namespace, rb.Name+"-"+redisResourceSuffix,
		resources.DeploymentWithMetaOptions(
			resources.MetaAddLabel(appAnnotation, appAnnotationValue),
			resources.MetaAddLabel("component", redisResourceSuffix),
			resources.MetaAddLabel(resourceNameAnnotation, rb.Name+"-"+redisResourceSuffix),
			resources.MetaAddOwner(rb, rb.GetGroupVersionKind())),
		resources.DeploymentAddSelectorForTemplate(resourceNameAnnotation, rb.Name+"-"+redisResourceSuffix),
		resources.DeploymentSetReplicas(1),
		resources.DeploymentWithTemplateOptions(
			resources.PodSpecAddContainer(
				resources.NewContainer("redis", image,
					resources.ContainerAddEnvFromValue("REDIS_ARGS", "--appendonly yes"),
					resources.ContainerAddPort("redis", 6379)))))
}

func (r *redisReconciler) reconcileDeployment(ctx context.Context, rb *eventingv1alpha1.RedisBroker) (*appsv1.Deployment, error) {
	desired := buildRedisDeployment(rb, r.image)
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
				logging.FromContext(ctx).Error("Unable to update the deployment", zap.String("deployment", fullname.String()), zap.Error(err))
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
		current, err = r.client.AppsV1().Deployments(desired.Namespace).Create(ctx, desired, metav1.CreateOptions{})
		if err != nil {
			fullname := types.NamespacedName{Namespace: desired.Namespace, Name: desired.Name}
			logging.FromContext(ctx).Error("Unable to create the deployment", zap.String("deployment", fullname.String()), zap.Error(err))
			rb.Status.MarkRedisDeploymentFailed(reconciler.ReasonFailedDeploymentCreate, "Failed to create Redis deployment")

			return nil, pkgreconciler.NewEvent(corev1.EventTypeWarning, reconciler.ReasonFailedDeploymentCreate,
				"Failed to create Redis deployment %s: %w", fullname, err)
		}
	}

	// Update status based on deployment
	rb.Status.PropagateRedisDeploymentAvailability(ctx, &current.Status)

	return current, nil
}

func buildRedisService(rb *eventingv1alpha1.RedisBroker) *corev1.Service {
	return resources.NewService(rb.Namespace, rb.Name+"-"+redisResourceSuffix,
		resources.ServiceWithMetaOptions(
			resources.MetaAddLabel(appAnnotation, appAnnotationValue),
			resources.MetaAddLabel("component", redisResourceSuffix),
			resources.MetaAddLabel(resourceNameAnnotation, rb.Name+"-"+redisResourceSuffix),
			resources.MetaAddOwner(rb, rb.GetGroupVersionKind())),
		resources.ServiceSetType(corev1.ServiceTypeClusterIP),
		resources.ServiceAddSelectorLabel(resourceNameAnnotation, rb.Name+"-"+redisResourceSuffix),
		resources.ServiceAddPort("redis", 6379, 6379))
}

func (r *redisReconciler) reconcileService(ctx context.Context, rb *eventingv1alpha1.RedisBroker) (*corev1.Service, error) {
	desired := buildRedisService(rb)
	current, err := r.serviceLister.Services(desired.Namespace).Get(desired.Name)
	switch {
	case err == nil:
		// Compare current object with desired, update if needed.
		if !semantic.Semantic.DeepEqual(desired, current) {
			desired.Status = current.Status
			desired.ResourceVersion = current.ResourceVersion

			current, err = r.client.CoreV1().Services(desired.Namespace).Update(ctx, desired, metav1.UpdateOptions{})
			if err != nil {
				fullname := types.NamespacedName{Namespace: desired.Namespace, Name: desired.Name}
				logging.FromContext(ctx).Error("Unable to update the service", zap.String("service", fullname.String()), zap.Error(err))
				rb.Status.MarkRedisServiceFailed(reconciler.ReasonFailedServiceUpdate, "Failed to update Redis service")

				return nil, pkgreconciler.NewEvent(corev1.EventTypeWarning, reconciler.ReasonFailedServiceUpdate,
					"Failed to get Redis service %s: %w", fullname, err)
			}
		}

	case !apierrs.IsNotFound(err):
		// An error ocurred retrieving current object.
		fullname := types.NamespacedName{Namespace: desired.Namespace, Name: desired.Name}
		logging.FromContext(ctx).Error("Unable to get the service", zap.String("service", fullname.String()), zap.Error(err))
		rb.Status.MarkRedisServiceFailed(reconciler.ReasonFailedServiceGet, "Failed to get Redis service")

		return nil, pkgreconciler.NewEvent(corev1.EventTypeWarning, reconciler.ReasonFailedServiceGet,
			"Failed to get Redis service %s: %w", fullname, err)

	default:
		// The object has not been found, create it.
		current, err = r.client.CoreV1().Services(desired.Namespace).Create(ctx, desired, metav1.CreateOptions{})
		if err != nil {
			fullname := types.NamespacedName{Namespace: desired.Namespace, Name: desired.Name}
			logging.FromContext(ctx).Error("Unable to create the service", zap.String("service", fullname.String()), zap.Error(err))
			rb.Status.MarkRedisServiceFailed(reconciler.ReasonFailedServiceCreate, "Failed to create Redis service")

			return nil, pkgreconciler.NewEvent(corev1.EventTypeWarning, reconciler.ReasonFailedServiceCreate,
				"Failed to create Redis service %s: %w", fullname, err)
		}
	}

	// Service exists and is up to date.
	rb.Status.MarkRedisServiceReady()

	return current, nil
}

func (r *redisReconciler) reconcileEndpoints(ctx context.Context, service *corev1.Service, rb *eventingv1alpha1.RedisBroker) (*corev1.Endpoints, error) {
	ep, err := r.endpointsLister.Endpoints(service.Namespace).Get(service.Name)
	switch {
	case err == nil:
		if duck.EndpointsAreAvailable(ep) {
			rb.Status.MarkRedisEndpointsTrue()
			return ep, nil
		}

		rb.Status.MarkRedisEndpointsFailed(reconciler.ReasonUnavailableEndpoints, "Endpoints for redis service are not available")
		return nil, pkgreconciler.NewEvent(corev1.EventTypeWarning, reconciler.ReasonUnavailableEndpoints,
			"Endpoints for redis service are not available %s",
			types.NamespacedName{Namespace: ep.Namespace, Name: ep.Name})

	case apierrs.IsNotFound(err):
		rb.Status.MarkRedisEndpointsFailed(reconciler.ReasonUnavailableEndpoints, "Endpoints for redis service do not exist")
		return nil, pkgreconciler.NewEvent(corev1.EventTypeWarning, reconciler.ReasonUnavailableEndpoints,
			"Endpoints for redis service do not exist %s",
			types.NamespacedName{Namespace: service.Namespace, Name: service.Name})
	}

	fullname := types.NamespacedName{Namespace: service.Namespace, Name: service.Name}
	rb.Status.MarkRedisEndpointsUnknown(reconciler.ReasonFailedEndpointsGet, "Could not retrieve endpoints for redis service")
	logging.FromContext(ctx).Error("Unable to get the redis service endpoints", zap.String("endpoint", fullname.String()), zap.Error(err))
	return nil, pkgreconciler.NewEvent(corev1.EventTypeWarning, reconciler.ReasonFailedEndpointsGet,
		"Failed to get redis service ednpoints %s: %w", fullname, err)
}
