package redisbroker

import (
	"context"

	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	appsv1listers "k8s.io/client-go/listers/apps/v1"
	k8sclient "knative.dev/pkg/client/injection/kube/client"
	"knative.dev/pkg/logging"
	pkgreconciler "knative.dev/pkg/reconciler"

	eventingv1alpha1 "github.com/triggermesh/triggermesh-core/pkg/apis/eventing/v1alpha1"
	"github.com/triggermesh/triggermesh-core/pkg/reconciler"
	"github.com/triggermesh/triggermesh-core/pkg/reconciler/resources"
)

type redisReconciler struct {
	client           kubernetes.Interface
	deploymentLister appsv1listers.DeploymentLister
}

func newRedisReconciler(ctx context.Context, deploymentLister appsv1listers.DeploymentLister) redisReconciler {
	return redisReconciler{
		client:           k8sclient.Get(ctx),
		deploymentLister: deploymentLister,
	}
}

func (r *redisReconciler) Reconcile(ctx context.Context, rb *eventingv1alpha1.RedisBroker) (*appsv1.Deployment, error) {

	desired := buildRedisDeployment(rb)
	current, err := r.deploymentLister.Deployments(desired.Namespace).Get(desired.Name)
	switch {
	case err == nil:
		// TODO compare
		// TODO if equal return
		// continue
	case apierrs.IsNotFound(err):
		// continue
	default:
		fullname := types.NamespacedName{Namespace: desired.Namespace, Name: desired.Name}
		logging.FromContext(ctx).Error("Unable to get the deployment", zap.String("deployment", fullname.String()), zap.Error(err))
		rb.Status.MarkRedisBrokerFailed("DeploymentGetFailed", "Failed to get Redis deployment")

		return nil, pkgreconciler.NewEvent(corev1.EventTypeWarning, reconciler.ReasonDeploymentGet,
			"Failed to get Redis deployment %s: %w", fullname, err)
	}

	// Create

	// TODO Update statuses

	return current, nil
}

func buildRedisDeployment(rb *eventingv1alpha1.RedisBroker) *appsv1.Deployment {
	return resources.NewDeployment(rb.Namespace, rb.Name+"-redis-server",
		resources.DeploymentWithMetaOptions(
			resources.MetaAddLabel("app", "redis-server"),
			resources.MetaAddLabel("eventing.triggermesh.io/redis-name", rb.Name+"-redis-server"),
			resources.MetaAddOwner(rb, rb.GetGroupVersionKind())),
		resources.DeploymentAddSelectorForTemplate("eventing.triggermesh.io/redis-name", rb.Name+"-redis-server"),
		resources.DeploymentSetReplicas(1),
		resources.DeploymentWithTemplateOption(
			resources.PodSpecAddContainer(
				resources.NewContainer("redis", "redis/redis-stack-server:latest",
					resources.ContainerAddEnvFromValue("REDIS_ARGS", "--appendonly yes"),
					resources.ContainerAddPort("redis", 6379)))))
}
