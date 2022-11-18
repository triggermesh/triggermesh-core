package common

import (
	"context"
	"fmt"
	"strconv"

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
	k8sclient "knative.dev/pkg/client/injection/kube/client"
	"knative.dev/pkg/logging"
	pkgreconciler "knative.dev/pkg/reconciler"

	eventingv1alpha1 "github.com/triggermesh/triggermesh-core/pkg/apis/eventing/v1alpha1"
	"github.com/triggermesh/triggermesh-core/pkg/reconciler/resources"
	"github.com/triggermesh/triggermesh-core/pkg/reconciler/semantic"
)

const (
	brokerResourceSuffix           = "broker"
	brokerDeploymentComponentLabel = "broker-deployment"

	defaultBrokerServicePort = 80
	metricsServicePort       = 9090
)

type BrokerReconciler interface {
	Reconcile(ctx context.Context, rb eventingv1alpha1.ReconcilableBroker, sa *corev1.ServiceAccount, secret *corev1.Secret, do ...resources.DeploymentOption) (*appsv1.Deployment, *corev1.Service, error)
}

type brokerReconciler struct {
	client           kubernetes.Interface
	deploymentLister appsv1listers.DeploymentLister
	serviceLister    corev1listers.ServiceLister
	endpointsLister  corev1listers.EndpointsLister
	image            string
	// TODO remove when using releases
	pullPolicy corev1.PullPolicy
}

func NewBrokerReconciler(ctx context.Context,
	deploymentLister appsv1listers.DeploymentLister,
	serviceLister corev1listers.ServiceLister,
	endpointsLister corev1listers.EndpointsLister,
	image string,
	pullPolicy corev1.PullPolicy) BrokerReconciler {

	return &brokerReconciler{
		client:           k8sclient.Get(ctx),
		deploymentLister: deploymentLister,
		serviceLister:    serviceLister,
		endpointsLister:  endpointsLister,
		image:            image,
		pullPolicy:       pullPolicy,
	}
}

func (r *brokerReconciler) Reconcile(ctx context.Context, rb eventingv1alpha1.ReconcilableBroker, sa *corev1.ServiceAccount, secret *corev1.Secret, deploymentOptions ...resources.DeploymentOption) (*appsv1.Deployment, *corev1.Service, error) {
	d, err := r.reconcileDeployment(ctx, rb, sa, secret, deploymentOptions)
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

func RedisDeploymentOptions(rb *eventingv1alpha1.RedisBroker, redisSvc *corev1.Service) resources.DeploymentOption {
	return func(d *appsv1.Deployment) {
		// Make sure the broker container exists before modifying it.
		if len(d.Spec.Template.Spec.Containers) == 0 {
			// Unexpected path.
			panic("The Broker Deployment to be reconciled has no containers in it.")
		}

		c := &d.Spec.Template.Spec.Containers[0]

		var stream string
		if rb.Spec.Redis != nil && rb.Spec.Redis.Stream != nil && *rb.Spec.Redis.Stream != "" {
			stream = *rb.Spec.Redis.Stream
		} else {
			stream = rb.Namespace + "." + rb.Name
		}
		resources.ContainerAddEnvFromValue("REDIS_STREAM", stream)(c)

		if rb.Spec.Redis != nil && rb.Spec.Redis.StreamMaxLen != nil && *rb.Spec.Redis.StreamMaxLen != 0 {
			resources.ContainerAddEnvFromValue("REDIS_STREAM_MAXLEN", stream)(c)
		}

		if rb.IsUserProvidedRedis() {
			resources.ContainerAddEnvFromValue("REDIS_ADDRESS", rb.Spec.Redis.Connection.URL)(c)

			if rb.Spec.Redis.Connection.Username != nil {
				resources.ContainerAddEnvVarFromSecret("REDIS_USERNAME",
					rb.Spec.Redis.Connection.Username.SecretKeyRef.Name,
					rb.Spec.Redis.Connection.Username.SecretKeyRef.Key)(c)
			}

			if rb.Spec.Redis.Connection.Password != nil {
				resources.ContainerAddEnvVarFromSecret("REDIS_PASSWORD",
					rb.Spec.Redis.Connection.Password.SecretKeyRef.Name,
					rb.Spec.Redis.Connection.Password.SecretKeyRef.Key)(c)
			}

			if rb.Spec.Redis.Connection.TLSEnabled != nil && *rb.Spec.Redis.Connection.TLSEnabled {
				resources.ContainerAddEnvFromValue("REDIS_TLS_ENABLED", "true")(c)
			}

			if rb.Spec.Redis.Connection.TLSSkipVerify != nil && *rb.Spec.Redis.Connection.TLSSkipVerify {
				resources.ContainerAddEnvFromValue("REDIS_TLS_SKIP_VERIFY", "true")(c)
			}

		} else {
			resources.ContainerAddEnvFromValue("REDIS_ADDRESS",
				fmt.Sprintf("%s:%d", redisSvc.Name, redisSvc.Spec.Ports[0].Port))(c)
		}
	}
}

// // RedisContainerOptions creates Redis environment variable options for the Broker container.
// func RedisContainerOptions(rb *eventingv1alpha1.RedisBroker, redisSvc *corev1.Service) []resources.ContainerOption {
// 	var stream string
// 	if rb.Spec.Redis != nil && rb.Spec.Redis.Stream != nil && *rb.Spec.Redis.Stream != "" {
// 		stream = *rb.Spec.Redis.Stream
// 	} else {
// 		stream = rb.Namespace + "." + rb.Name
// 	}

// 	opts := []resources.ContainerOption{
// 		resources.ContainerAddEnvFromValue("REDIS_STREAM", stream),
// 	}

// 	if rb.Spec.Redis != nil && rb.Spec.Redis.StreamMaxLen != nil && *rb.Spec.Redis.StreamMaxLen != 0 {
// 		opts = append(opts, resources.ContainerAddEnvFromValue("REDIS_STREAM_MAXLEN", stream))
// 	}

// 	if rb.IsUserProvidedRedis() {
// 		opts = append(opts, resources.ContainerAddEnvFromValue("REDIS_ADDRESS", rb.Spec.Redis.Connection.URL))

// 		if rb.Spec.Redis.Connection.Username != nil {
// 			opts = append(opts, resources.ContainerAddEnvVarFromSecret("REDIS_USERNAME",
// 				rb.Spec.Redis.Connection.Username.SecretKeyRef.Name,
// 				rb.Spec.Redis.Connection.Username.SecretKeyRef.Key))
// 		}

// 		if rb.Spec.Redis.Connection.Password != nil {
// 			opts = append(opts, resources.ContainerAddEnvVarFromSecret("REDIS_PASSWORD",
// 				rb.Spec.Redis.Connection.Password.SecretKeyRef.Name,
// 				rb.Spec.Redis.Connection.Password.SecretKeyRef.Key))
// 		}

// 		if rb.Spec.Redis.Connection.TLSEnabled != nil && *rb.Spec.Redis.Connection.TLSEnabled {
// 			opts = append(opts, resources.ContainerAddEnvFromValue("REDIS_TLS_ENABLED", "true"))
// 		}

// 		if rb.Spec.Redis.Connection.TLSSkipVerify != nil && *rb.Spec.Redis.Connection.TLSSkipVerify {
// 			opts = append(opts, resources.ContainerAddEnvFromValue("REDIS_TLS_SKIP_VERIFY", "true"))
// 		}

// 	} else {
// 		opts = append(opts, resources.ContainerAddEnvFromValue("REDIS_ADDRESS",
// 			fmt.Sprintf("%s:%d", redisSvc.Name, redisSvc.Spec.Ports[0].Port)))
// 	}

// 	return opts
// }

func buildBrokerDeployment(rb eventingv1alpha1.ReconcilableBroker, sa *corev1.ServiceAccount, secret *corev1.Secret, image string, pullPolicy corev1.PullPolicy, extraOptions ...resources.DeploymentOption) *appsv1.Deployment {
	meta := rb.GetObjectMeta()
	ns, name := meta.GetNamespace(), meta.GetName()
	bs := rb.GetReconcilableBrokerSpec()

	// var stream string
	// if rb.Spec.Redis != nil && rb.Spec.Redis.Stream != nil && *rb.Spec.Redis.Stream != "" {
	// 	stream = *rb.Spec.Redis.Stream
	// } else {
	// 	stream = rb.Namespace + "." + rb.Name
	// }

	copts := []resources.ContainerOption{
		resources.ContainerAddArgs("start"),
		resources.ContainerAddEnvFromValue("PORT", strconv.Itoa(int(defaultBrokerServicePort))),
		resources.ContainerAddEnvFromValue("BROKER_NAME", name),
		resources.ContainerAddEnvFromFieldRef("KUBERNETES_NAMESPACE", "metadata.namespace"),
		resources.ContainerAddEnvFromValue("KUBERNETES_BROKER_CONFIG_SECRET_NAME", secret.Name),
		resources.ContainerAddEnvFromValue("KUBERNETES_BROKER_CONFIG_SECRET_KEY", ConfigSecretKey),
		// resources.ContainerAddEnvFromValue("REDIS_STREAM", stream),
		resources.ContainerWithImagePullPolicy(pullPolicy),
		resources.ContainerAddPort("httpce", defaultBrokerServicePort),
		resources.ContainerAddPort("metrics", metricsServicePort),
	}

	if bs.Observability != nil && bs.Observability.ValueFromConfigMap != "" {
		copts = append(copts, resources.ContainerAddEnvFromValue("KUBERNETES_OBSERVABILITY_CONFIGMAP_NAME", bs.Observability.ValueFromConfigMap))
	}

	dn := name + "-" + rb.GetOwnedObjectsSuffix() + "-" + brokerResourceSuffix
	d := resources.NewDeployment(ns, dn,
		resources.DeploymentWithMetaOptions(
			resources.MetaAddLabel(resources.AppNameLabel, AppAnnotationValue(rb)),
			resources.MetaAddLabel(resources.AppComponentLabel, brokerDeploymentComponentLabel),
			resources.MetaAddLabel(resources.AppPartOfLabel, resources.PartOf),
			resources.MetaAddLabel(resources.AppManagedByLabel, resources.ManagedBy),
			resources.MetaAddLabel(resources.AppInstanceLabel, dn),
			resources.MetaAddOwner(meta, rb.GetGroupVersionKind())),
		resources.DeploymentAddSelectorForTemplate(resources.AppComponentLabel, brokerDeploymentComponentLabel),
		resources.DeploymentAddSelectorForTemplate(resources.AppInstanceLabel, dn),
		resources.DeploymentSetReplicas(1),
		resources.DeploymentWithTemplateSpecOptions(
			// Needed for prometheus PodMonitor.
			resources.PodTemplateSpecWithMetaOptions(
				resources.MetaAddLabel(resources.AppPartOfLabel, resources.PartOf),
				resources.MetaAddLabel(resources.AppManagedByLabel, resources.ManagedBy),
			),
			resources.PodTemplateSpecWithPodSpecOptions(
				resources.PodSpecWithServiceAccountName(sa.Name),
				resources.PodSpecAddContainer(
					resources.NewContainer("broker", image, copts...)))))

	if len(extraOptions) != 0 {
		for _, o := range extraOptions {
			o(d)
		}
	}

	return d
}

func (r *brokerReconciler) reconcileDeployment(
	ctx context.Context,
	rb eventingv1alpha1.ReconcilableBroker,
	sa *corev1.ServiceAccount,
	secret *corev1.Secret,
	deploymentOptions []resources.DeploymentOption) (*appsv1.Deployment, error) {
	//desired := buildBrokerDeployment(rb, sa, redis, secret, r.image, r.pullPolicy)
	desired := buildBrokerDeployment(rb, sa, secret, r.image, r.pullPolicy, deploymentOptions...)
	// TODO instead of using container options, use a caller provided callback
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
				rb.GetReconcilableBrokerStatus().MarkBrokerDeploymentFailed(ReasonFailedDeploymentUpdate, "Failed to update broker deployment")

				return nil, pkgreconciler.NewEvent(corev1.EventTypeWarning, ReasonFailedDeploymentUpdate,
					"Failed to get broker deployment %s: %w", fullname, err)
			}
		}

	case !apierrs.IsNotFound(err):
		// An error occurred retrieving current deployment.
		fullname := types.NamespacedName{Namespace: desired.Namespace, Name: desired.Name}
		logging.FromContext(ctx).Error("Unable to get broker deployment", zap.String("deployment", fullname.String()), zap.Error(err))
		rb.GetReconcilableBrokerStatus().MarkBrokerDeploymentFailed(ReasonFailedDeploymentGet, "Failed to get broker deployment")

		return nil, pkgreconciler.NewEvent(corev1.EventTypeWarning, ReasonFailedDeploymentGet,
			"Failed to get broker deployment %s: %w", fullname, err)

	default:
		// The deployment has not been found, create it.
		current, err = r.client.AppsV1().Deployments(desired.Namespace).Create(ctx, desired, metav1.CreateOptions{})
		if err != nil {
			fullname := types.NamespacedName{Namespace: desired.Namespace, Name: desired.Name}
			logging.FromContext(ctx).Error("Unable to create broker deployment", zap.String("deployment", fullname.String()), zap.Error(err))
			rb.GetReconcilableBrokerStatus().MarkBrokerDeploymentFailed(ReasonFailedDeploymentCreate, "Failed to create broker deployment")

			return nil, pkgreconciler.NewEvent(corev1.EventTypeWarning, ReasonFailedDeploymentCreate,
				"Failed to create broker deployment %s: %w", fullname, err)
		}
	}

	// Update status based on deployment
	rb.GetReconcilableBrokerStatus().PropagateBrokerDeploymentAvailability(ctx, &current.Status)

	return current, nil
}

func buildBrokerService(rb eventingv1alpha1.ReconcilableBroker) *corev1.Service {
	meta := rb.GetObjectMeta()
	ns, name := meta.GetNamespace(), meta.GetName()
	bs := rb.GetReconcilableBrokerSpec()

	brokerPort := defaultBrokerServicePort
	if bs.Port != nil {
		brokerPort = *bs.Port
	}

	sn := name + "-" + rb.GetOwnedObjectsSuffix() + "-" + brokerResourceSuffix
	return resources.NewService(ns, sn,
		resources.ServiceWithMetaOptions(
			resources.MetaAddLabel(resources.AppNameLabel, AppAnnotationValue(rb)),
			resources.MetaAddLabel(resources.AppComponentLabel, "broker-service"),
			resources.MetaAddLabel(resources.AppPartOfLabel, resources.PartOf),
			resources.MetaAddLabel(resources.AppManagedByLabel, resources.ManagedBy),
			resources.MetaAddLabel(resources.AppInstanceLabel, sn),
			resources.MetaAddOwner(meta, rb.GetGroupVersionKind())),
		resources.ServiceSetType(corev1.ServiceTypeClusterIP),
		resources.ServiceAddSelectorLabel(resources.AppComponentLabel, brokerDeploymentComponentLabel),
		resources.ServiceAddSelectorLabel(resources.AppInstanceLabel, sn),
		resources.ServiceAddPort("httpce", int32(brokerPort), defaultBrokerServicePort))
}

func (r *brokerReconciler) reconcileService(ctx context.Context, rb eventingv1alpha1.ReconcilableBroker) (*corev1.Service, error) {
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
				rb.GetReconcilableBrokerStatus().MarkBrokerServiceFailed(ReasonFailedServiceUpdate, "Failed to update broker service")

				return nil, pkgreconciler.NewEvent(corev1.EventTypeWarning, ReasonFailedServiceUpdate,
					"Failed to get broker service %s: %w", fullname, err)
			}
		}

	case !apierrs.IsNotFound(err):
		// An error occurred retrieving current object.
		fullname := types.NamespacedName{Namespace: desired.Namespace, Name: desired.Name}
		logging.FromContext(ctx).Error("Unable to get the service", zap.String("service", fullname.String()), zap.Error(err))
		rb.GetReconcilableBrokerStatus().MarkBrokerServiceFailed(ReasonFailedServiceGet, "Failed to get broker service")

		return nil, pkgreconciler.NewEvent(corev1.EventTypeWarning, ReasonFailedServiceGet,
			"Failed to get broker service %s: %w", fullname, err)

	default:
		// The object has not been found, create it.
		current, err = r.client.CoreV1().Services(desired.Namespace).Create(ctx, desired, metav1.CreateOptions{})
		if err != nil {
			fullname := types.NamespacedName{Namespace: desired.Namespace, Name: desired.Name}
			logging.FromContext(ctx).Error("Unable to create the service", zap.String("service", fullname.String()), zap.Error(err))
			rb.GetReconcilableBrokerStatus().MarkBrokerServiceFailed(ReasonFailedServiceCreate, "Failed to create broker service")

			return nil, pkgreconciler.NewEvent(corev1.EventTypeWarning, ReasonFailedServiceCreate,
				"Failed to create broker service %s: %w", fullname, err)
		}
	}

	// Service exists and is up to date.
	rb.GetReconcilableBrokerStatus().MarkBrokerServiceReady()

	return current, nil
}

func (r *brokerReconciler) reconcileEndpoints(ctx context.Context, service *corev1.Service, rb eventingv1alpha1.ReconcilableBroker) (*corev1.Endpoints, error) {
	ep, err := r.endpointsLister.Endpoints(service.Namespace).Get(service.Name)
	switch {
	case err == nil:
		if duck.EndpointsAreAvailable(ep) {
			rb.GetReconcilableBrokerStatus().MarkBrokerEndpointsTrue()
			return ep, nil
		}

		rb.GetReconcilableBrokerStatus().MarkBrokerEndpointsFailed(ReasonUnavailableEndpoints, "Endpoints for broker service are not available")
		return nil, pkgreconciler.NewEvent(corev1.EventTypeWarning, ReasonUnavailableEndpoints,
			"Endpoints for broker service are not available %s",
			types.NamespacedName{Namespace: ep.Namespace, Name: ep.Name})

	case apierrs.IsNotFound(err):
		rb.GetReconcilableBrokerStatus().MarkBrokerEndpointsFailed(ReasonUnavailableEndpoints, "Endpoints for broker service do not exist")
		return nil, pkgreconciler.NewEvent(corev1.EventTypeWarning, ReasonUnavailableEndpoints,
			"Endpoints for broker service do not exist %s",
			types.NamespacedName{Namespace: service.Namespace, Name: service.Name})
	}

	fullname := types.NamespacedName{Namespace: service.Namespace, Name: service.Name}
	rb.GetReconcilableBrokerStatus().MarkBrokerEndpointsUnknown(ReasonFailedEndpointsGet, "Could not retrieve endpoints for broker service")
	logging.FromContext(ctx).Error("Unable to get the broker service endpoints", zap.String("endpoint", fullname.String()), zap.Error(err))
	return nil, pkgreconciler.NewEvent(corev1.EventTypeWarning, ReasonFailedEndpointsGet,
		"Failed to get broker service ednpoints %s: %w", fullname, err)
}
