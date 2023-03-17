// Copyright 2023 TriggerMesh Inc.
// SPDX-License-Identifier: Apache-2.0

package redisreplay

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/kelseyhightower/envconfig"

	apierrs "k8s.io/apimachinery/pkg/api/errors"
	appsv1listers "k8s.io/client-go/listers/apps/v1"
	"k8s.io/client-go/tools/cache"

	corev1 "k8s.io/api/core/v1"
	corev1listers "k8s.io/client-go/listers/core/v1"

	"knative.dev/pkg/client/injection/kube/informers/apps/v1/deployment"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"

	"github.com/triggermesh/triggermesh-core/pkg/apis/eventing/v1alpha1"
	"github.com/triggermesh/triggermesh-core/pkg/client/generated/injection/client"
	"github.com/triggermesh/triggermesh-core/pkg/client/generated/injection/informers/eventing/v1alpha1/redisreplay"
	rrinformer "github.com/triggermesh/triggermesh-core/pkg/client/generated/injection/informers/eventing/v1alpha1/redisreplay"
)

// envConfig will be used to extract the required environment variables using
// github.com/kelseyhightower/envconfig. If this configuration cannot be extracted, then
// NewController will panic.
type envConfig struct {
	RedisReplayImage       string `envconfig:"REDIS_REPLAY_IMAGE" required:"true"`
	RedisReplayIPullPolicy string `envconfig:"REDIS_REPLAY_IMAGE_PULL_POLICY" default:"IfNotPresent"`
	RedisAddress           string `envconfig:"REDIS_ADDRESS" required:"true"`
	RedisUser              string `envconfig:"REDIS_USER" required:"false"`
	RedisPassword          string `envconfig:"REDIS_PASSWORD" required:"false"`
	RedisDatabase          string `envconfig:"REDIS_DATABASE" required:"false" default:"0"`
	Sink                   string `envconfig:"K_SINK" required:"true"`
	StartTime              string `envconfig:"START_TIME" required:"false"`
	EndTime                string `envconfig:"END_TIME" required:"false"`
	Filter                 string `envconfig:"FILTER" required:"false"`
	Filter_Kind            string `envconfig:"FILTER_KIND" required:"false"`
}

// NewController initializes the controller and is called by the generated code
// Registers event handlers to enqueue events
func NewController(
	ctx context.Context,
	cmw configmap.Watcher,
) *controller.Impl {

	env := &envConfig{}
	if err := envconfig.Process("", env); err != nil {
		logging.FromContext(ctx).Panicf("unable to process RedisReplay's required environment variables: %v", err)
	}

	rrInformer := rrinformer.Get(ctx)
	// secretInformer := secret.Get(ctx)
	deploymentInformer := deployment.Get(ctx)
	// serviceInformer := service.Get(ctx)
	// endpointsInformer := endpointsinformer.Get(ctx)
	// rolebindingInformer := rolebindingsinformer.Get(ctx)
	// serviceaccountInformer := serviceaccount.Get(ctx)
	r := &reconciler{
		// redisReplayClient: redisreplayclient.Get(ctx),
		// secretLister:      secretInformer.Lister(),
		// serviceLister:     serviceInformer.Lister(),
		// endpointsLister:   endpointsInformer.Lister(),
		// roleBindingLister: rolebindingInformer.Lister(),
		// serviceAccount:    serviceaccountInformer.Lister(),
		deploymentLister: deploymentInformer.Lister(),
		image:            env.RedisReplayImage,
		sink:             env.Sink,
		redisAddress:     env.RedisAddress,
		redisUser:        env.RedisUser,
		redisPassword:    env.RedisPassword,
		redisDatabase:    env.RedisDatabase,
		startTime:        env.StartTime,
		endTime:          env.EndTime,
		filter:           env.Filter,
		filterKind:       env.Filter_Kind,
	}

	impl := reconciler.NewImpl(r)

	rrInformer.Informer().AddEventHandler(controller.HandleAll(impl.Enqueue))

	deploymentInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: controller.FilterControllerGK(v1alpha1.Kind("RedisReplay")),
		Handler:    controller.HandleAll(impl.EnqueueControllerOf),
	})

	return impl
}

func NewRedisReplayReconciler(
	deploymentLister appsv1listers.DeploymentLister,
	image string,
	pullPolicy corev1.PullPolicy,
	sink string,
	redisAddress string,
	redisUser string,
	redisPassword string,
	redisDatabase string,
	startTime string,
	endTime string,
	filter string,
	filterKind string,
) *reconciler {
	return &reconciler{
		deploymentLister: deploymentLister,
		image:            image,
		pullPolicy:       pullPolicy,
		sink:             sink,
		redisAddress:     redisAddress,
		redisUser:        redisUser,
		redisPassword:    redisPassword,
		redisDatabase:    redisDatabase,
		startTime:        startTime,
		endTime:          endTime,
		filter:           filter,
		filterKind:       filterKind,
	}
}

func NewBrokerReconciler(ctx context.Context,
	deploymentLister appsv1listers.DeploymentLister,
	serviceLister corev1listers.ServiceLister,
	endpointsLister corev1listers.EndpointsLister,
	image string,
	pullPolicy corev1.PullPolicy) reconciler {

	return &reconciler{
		client:           k8sclient.Get(ctx),
		deploymentLister: deploymentLister,
		image:            image,
		pullPolicy:       pullPolicy,
	}
}

func (r *rRreconciler) Reconcile(ctx context.Context, key string) error {
	// Convert the namespace/name string into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		r.logger.Errorf("invalid resource key: %s", key)
		return nil
	}

	// Get the RedisReplay resource with this namespace/name
	original, err := r.redisReplayLister.RedisReplays(namespace).Get(name)
	if apierrs.IsNotFound(err) {
		// The RedisReplay resource may no longer exist, in which case we stop processing.
		logging.FromContext(ctx).Errorf("RedisReplay %q in work queue no longer exists", key)
		return nil
	}
	if err != nil {
		r.logger.Errorf("Error fetching RedisReplay %q: %v", key, err)
		return err
	}

	// Don't modify the informers copy
	rr := original.DeepCopy()

	// Reconcile this copy of the RedisReplay and then write back any status
	// updates regardless of whether the reconciliation errored out.
	reconcileErr := r.reconcile(ctx, rr)
	if reconcileErr != nil {
		r.logger.Errorf("Error reconciling RedisReplay: %v", reconcileErr)
	} else {
		r.logger.Infof("RedisReplay reconciled: %s", key)
	}
	if _, updateStatusErr := r.updateStatus(ctx, rr.DeepCopy()); updateStatusErr != nil {
		r.logger.Warnf("Failed to update RedisReplay status: %v", updateStatusErr)
		return updateStatusErr
	}

	// Requeue if the resource is not ready:
	return reconcileErr
}

func (r *reconciler) reconcile(ctx context.Context, rr *v1alpha1.RedisReplay) error {
	rr.Status.InitializeConditions()

	// See if the source has been deleted.
	if rr.DeletionTimestamp != nil {
		return r.delete(ctx, rr)
	}

	// Reconcile this copy of the RedisReplay and then write back any status
	// updates regardless of whether the reconciliation errored out.
	reconcileErr := r.reconcileRedisReplay(ctx, rr)
	if reconcileErr != nil {
		r.logger.Errorf("Error reconciling RedisReplay: %v", reconcileErr)
	} else {
		r.logger.Infof("RedisReplay reconciled: %s", rr.Name)
	}
	if _, updateStatusErr := r.updateStatus(ctx, rr.DeepCopy()); updateStatusErr != nil {
		r.logger.Warnf("Failed to update RedisReplay status: %v", updateStatusErr)
		return updateStatusErr
	}

	// Requeue if the resource is not ready:
	return reconcileErr
}

func (r *reconciler) reconcileRedisReplay(ctx context.Context, rr *v1alpha1.RedisReplay) error {
	rr.Status.InitializeConditions()

	// See if the source has been deleted.
	if rr.DeletionTimestamp != nil {
		return r.delete(ctx, rr)
	}

	// Reconcile this copy of the RedisReplay and then write back any status
	// updates regardless of whether the reconciliation errored out.
	reconcileErr := r.reconcileRedisReplay(ctx, rr)
	if reconcileErr != nil {
		r.logger.Errorf("Error reconciling RedisReplay: %v", reconcileErr)
	} else {
		r.logger.Infof("RedisReplay reconciled: %s", rr.Name)
	}
	if _, updateStatusErr := r.updateStatus(ctx, rr.DeepCopy()); updateStatusErr != nil {
		r.logger.Warnf("Failed to update RedisReplay status: %v", updateStatusErr)
		return updateStatusErr
	}

	// Requeue if the resource is not ready:
	return reconcileErr
}

func (r *reconciler) reconcileRedisReplay(ctx context.Context, rr *v1alpha1.RedisReplay) error {
	rr.Status.InitializeConditions()

	// See if the source has been deleted.
	if rr.DeletionTimestamp != nil {
		return r.delete(ctx, rr)
	}

	// Reconcile this copy of the RedisReplay and then write back any status
	// updates regardless of whether the reconciliation errored out.
	reconcileErr := r.reconcileRedisReplay(ctx, rr)
	if reconcileErr != nil {
		r.logger.Errorf("Error reconciling RedisReplay: %v", reconcileErr)
	} else {
		r.logger.Infof("RedisReplay reconciled: %s", rr.Name)
	}
	if _, updateStatusErr := r.updateStatus(ctx, rr.DeepCopy()); updateStatusErr != nil {
		r.logger.Warnf("Failed to update RedisReplay status: %v", updateStatusErr)
		return updateStatusErr
	}

	// Requeue if the resource is not ready:
	return reconcileErr
}

func (r *reconciler) reconcileRedisReplay(ctx context.Context, rr *v1alpha1.RedisReplay) error {
	rr.Status.InitializeConditions()

	// See if the source has been deleted.
	if rr.DeletionTimestamp != nil {
		return r.delete(ctx, rr)
	}

	// Reconcile this copy of the RedisReplay and then write back any status
	// updates regardless of whether the reconciliation errored out.
	reconcileErr := r.reconcileRedisReplay(ctx, rr)
	if reconcileErr != nil {
		r.logger.Errorf("Error reconciling RedisReplay: %v", reconcileErr)
	} else {
		r.logger.Infof("RedisReplay reconciled: %s", rr.Name)
	}
	if _, updateStatusErr := r.updateStatus(ctx, rr.DeepCopy()); updateStatusErr != nil {
		r.logger.Warnf("Failed to update RedisReplay status: %v", updateStatusErr)
		return updateStatusErr
	}

	// Requeue if the resource is not ready:
	return reconcileErr
}

const (
	defaultControllerAgentName = "redisreplay-controller"
	defaultFinalizerName       = "redisreplay.eventing.triggermesh.io"
)

// NewImpl returns a controller.Impl that handles queuing and feeding work from
// the queue through an implementation of controller.Reconciler, delegating to
// the provided Interface and optional Finalizer methods. OptionsFn is used to return
// controller.ControllerOptions to be used by the internal reconciler.
func NewImpl(ctx context.Context, r Interface, logger *zap.SugaredLogger, name string, optsFns ...controller.OptionsFn) *controller.Impl {
	logger := logging.FromContext(ctx)

	// check the options function input. it should be 0 or 1.
	if len(optsFns) > 1 {
		logger.Fatal("Up to one options function is supported, found: ", len(optionsFns))
	}

	redisreplayInformer := redisreplay.Get(ctx)
	redisreplayLister := redisreplayInformer.Lister()
	var promoteFilterFunc func(obj interface{}) bool

	rec := &reconcilerImpl{
		LeaderAwareFuncs: reconciler.LeaderAwareFuncs{
			PromoteFunc: func(bkt reconciler.Bucket, enq func(reconciler.Bucket, types.NamespacedName)) error {
				all, err := redisreplayLister.List(labels.Everything())
				if err != nil {
					return err
				}
				for _, rr := range all {
					if rr.Status.IsReady() {
						enq(bkt, rr)
					}
				}
				enq(bkt, types.NamespacedName{Name: "redisreplay-controller-leader"})
				return nil
			},
			FilterFunc: func(obj interface{}) bool {
				if promoteFilterFunc != nil {
					return promoteFilterFunc(obj)
				}
				return true
			},
		},
		reconciler: r,
		Client:     client.Get(ctx),
	}

	ctrType := reflect.TypeOf(r).Elem()
	ctrTypeName := fmt.Sprintf("%s.%s", ctrType.PkgPath(), ctrType.Name())
	ctrTypeName = strings.ReplaceAll(ctrTypeName, "/", ".")

	logger = logger.With(
		zap.String(logkey.ControllerType, ctrTypeName),
		zap.String(logkey.Kind, "eventing.triggermesh.io.redisreplay"),
	)

	impl := controller.NewContext(ctx, rec, controller.ControllerOptions{WorkQueueName: ctrTypeName, Logger: logger})
	agentName := defaultControllerAgentName

}
