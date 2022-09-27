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
	k8sclient "knative.dev/pkg/client/injection/kube/client"
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
)

var (
	configMountedPath = path.Join(configSecretPath, configSecretFile)
)

type brokerReconciler struct {
	client           kubernetes.Interface
	deploymentLister appsv1listers.DeploymentLister
	serviceLister    corev1listers.ServiceLister
	image            string
}

func newBrokerReconciler(ctx context.Context, deploymentLister appsv1listers.DeploymentLister, serviceLister corev1listers.ServiceLister, image string) brokerReconciler {
	return brokerReconciler{
		client:           k8sclient.Get(ctx),
		deploymentLister: deploymentLister,
		serviceLister:    serviceLister,
		image:            image,
	}
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

	return d, svc, nil
}

func buildBrokerDeployment(rb *eventingv1alpha1.RedisBroker, redis *corev1.Service, secret *corev1.Secret, image string) *appsv1.Deployment {

	v := resources.NewVolume("config",
		resources.VolumeFromSecretOption(secret.Name, configSecretKey, configSecretFile))
	vm := resources.NewVolumeMount("config", configSecretPath,
		resources.VolumeMountWithReadOnlyOption(true))

	redisService := fmt.Sprintf("%s:%d", redis.Name, redis.Spec.Ports[0].Port)

	return resources.NewDeployment(rb.Namespace, rb.Name+"-redis-broker",
		resources.DeploymentWithMetaOptions(
			resources.MetaAddLabel("app", "redis-broker"),
			resources.MetaAddLabel("eventing.triggermesh.io/redis-broker-name", rb.Name+"-redis-broker"),
			resources.MetaAddOwner(rb, rb.GetGroupVersionKind())),
		resources.DeploymentAddSelectorForTemplate("eventing.triggermesh.io/redis-broker-name", rb.Name+"-redis-broker"),
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

func buildBrokerService(rb *eventingv1alpha1.RedisBroker) *corev1.Service {
	return resources.NewService(rb.Namespace, rb.Name+"-redis-broker",
		resources.ServiceWithMetaOptions(
			resources.MetaAddLabel("app", "redis-broker"),
			resources.MetaAddLabel("eventing.triggermesh.io/redis-broker-name", rb.Name+"-redis-broker"),
			resources.MetaAddOwner(rb, rb.GetGroupVersionKind())),
		resources.ServiceSetType(corev1.ServiceTypeClusterIP),
		resources.ServiceAddSelectorLabel("eventing.triggermesh.io/redis-broker", rb.Name+"-redis-broker"),
		resources.ServiceAddPort("httpce", 8080, 8080))
}

func (r *brokerReconciler) reconcileService(ctx context.Context, rb *eventingv1alpha1.RedisBroker) (*corev1.Service, error) {
	desired := buildBrokerService(rb)
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
				rb.Status.MarkRedisServiceFailed(reconciler.ReasonFailedServiceUpdate, "Failed to update Redis broker service")

				return nil, pkgreconciler.NewEvent(corev1.EventTypeWarning, reconciler.ReasonFailedServiceUpdate,
					"Failed to get Redis service %s: %w", fullname, err)
			}
		}

	case !apierrs.IsNotFound(err):
		// An error ocurred retrieving current object.
		fullname := types.NamespacedName{Namespace: desired.Namespace, Name: desired.Name}
		logging.FromContext(ctx).Error("Unable to get the service", zap.String("service", fullname.String()), zap.Error(err))
		rb.Status.MarkRedisServiceFailed(reconciler.ReasonFailedServiceGet, "Failed to get Redis broker service")

		return nil, pkgreconciler.NewEvent(corev1.EventTypeWarning, reconciler.ReasonFailedServiceGet,
			"Failed to get Redis service %s: %w", fullname, err)

	default:
		// The object has not been found, create it.
		current, err = r.client.CoreV1().Services(desired.Namespace).Create(ctx, desired, metav1.CreateOptions{})
		if err != nil {
			fullname := types.NamespacedName{Namespace: desired.Namespace, Name: desired.Name}
			logging.FromContext(ctx).Error("Unable to create the service", zap.String("service", fullname.String()), zap.Error(err))
			rb.Status.MarkRedisServiceFailed(reconciler.ReasonFailedServiceCreate, "Failed to create Redis broker service")

			return nil, pkgreconciler.NewEvent(corev1.EventTypeWarning, reconciler.ReasonFailedServiceCreate,
				"Failed to create Redis broker service %s: %w", fullname, err)
		}
	}

	// Service exists and is up to date.
	rb.Status.MarkRedisServiceReady()

	return current, nil
}
