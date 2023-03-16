// Copyright 2023 TriggerMesh Inc.
// SPDX-License-Identifier: Apache-2.0

package redisreplay

import (
	"context"

	"github.com/kelseyhightower/envconfig"

	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/cache"

	kubeclient "knative.dev/pkg/client/injection/kube/client"
	"knative.dev/pkg/client/injection/kube/informers/apps/v1/deployment"
	endpointsinformer "knative.dev/pkg/client/injection/kube/informers/core/v1/endpoints"
	"knative.dev/pkg/client/injection/kube/informers/core/v1/secret"
	"knative.dev/pkg/client/injection/kube/informers/core/v1/service"
	"knative.dev/pkg/client/injection/kube/informers/core/v1/serviceaccount"
	rolebindingsinformer "knative.dev/pkg/client/injection/kube/informers/rbac/v1/rolebinding"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"

	"github.com/triggermesh/triggermesh-core/pkg/apis/eventing/v1alpha1"
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
	secretInformer := secret.Get(ctx)
	deploymentInformer := deployment.Get(ctx)
	serviceInformer := service.Get(ctx)
	endpointsInformer := endpointsinformer.Get(ctx)
	rolebindingInformer := rolebindingsinformer.Get(ctx)
	serviceaccountInformer := serviceaccount.Get(ctx)

	r := &RedisReplayReconciler{
		kubeClientSet:     kubeclient.Get(ctx),
		redisReplayLister: rrInformer.Lister(),
		// redisReplayClient: redisreplayclient.Get(ctx),
		secretLister:      secretInformer.Lister(),
		deploymentLister:  deploymentInformer.Lister(),
		serviceLister:     serviceInformer.Lister(),
		endpointsLister:   endpointsInformer.Lister(),
		roleBindingLister: rolebindingInformer.Lister(),
		serviceAccount:    serviceaccountInformer.Lister(),
		redisReplayImage:  env.RedisReplayImage,
		sink:              env.Sink,
		redisAddress:      env.RedisAddress,
		redisUser:         env.RedisUser,
		redisPassword:     env.RedisPassword,
		redisDatabase:     env.RedisDatabase,
		startTime:         env.StartTime,
		endTime:           env.EndTime,
		filter:            env.Filter,
		filterKind:        env.Filter_Kind,
	}

	impl := controller.NewImpl(r, r.Logger, "RedisReplay", RedisReplayReconciler.MustNewStatsReporter("RedisReplay", r.Logger))

	r.Logger.Info("Setting up event handlers")

	rrInformer.Informer().AddEventHandler(controller.HandleAll(impl.Enqueue))

	// Set up watches for RedisReplay resources
	secretInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: controller.FilterControllerGK(v1alpha1.Kind("RedisReplay")),
		Handler:    controller.HandleAll(impl.EnqueueControllerOf),
	})

	deploymentInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: controller.FilterControllerGK(v1alpha1.Kind("RedisReplay")),
		Handler:    controller.HandleAll(impl.EnqueueControllerOf),
	})

	serviceInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: controller.FilterControllerGK(v1alpha1.Kind("RedisReplay")),
		Handler:    controller.HandleAll(impl.EnqueueControllerOf),
	})

	endpointsInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: controller.FilterControllerGK(v1alpha1.Kind("RedisReplay")),
		Handler:    controller.HandleAll(impl.EnqueueControllerOf),
	})

	rolebindingInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: controller.FilterControllerGK(v1alpha1.Kind("RedisReplay")),
		Handler:    controller.HandleAll(impl.EnqueueControllerOf),
	})

	serviceaccountInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: controller.FilterControllerGK(v1alpha1.Kind("RedisReplay")),
		Handler:    controller.HandleAll(impl.EnqueueControllerOf),
	})

	return impl
}

// Reconcile compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the RedisReplay resource
// with the current status of the resource.
func (r *RedisReplayReconciler) Reconcile(ctx context.Context, key string) error {
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

func (r *RedisReplayReconciler) reconcile(ctx context.Context, rr *v1alpha1.RedisReplay) error {
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

func (r *RedisReplayReconciler) reconcileRedisReplay(ctx context.Context, rr *v1alpha1.RedisReplay) error {
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

func (r *RedisReplayReconciler) reconcileRedisReplay(ctx context.Context, rr *v1alpha1.RedisReplay) error {
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

func (r *RedisReplayReconciler) reconcileRedisReplay(ctx context.Context, rr *v1alpha1.RedisReplay) error {
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
