// Copyright 2023 TriggerMesh Inc.
// SPDX-License-Identifier: Apache-2.0

package redisreplay

import (
	"context"

	"github.com/kelseyhightower/envconfig"

	apierrs "k8s.io/apimachinery/pkg/api/errors"

	redisreplayclient "github.com/triggermesh/triggermesh-core/pkg/client/generated/injection/client"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"
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

	r := &Reconciler{
		kubeClientSet:     kubeclient.Get(ctx),
		redisReplayLister: rrInformer.Lister(),
		redisReplayClient: redisreplayclient.Get(ctx),
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

	impl := controller.NewImpl(r, r.Logger, "RedisReplay", reconciler.MustNewStatsReporter("RedisReplay", r.Logger))

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

// Reconciler implements controller.Reconciler for RedisReplay resources.
type Reconciler struct {
	// kubeClientSet allows us to talk to the k8s for core APIs
	kubeClientSet kubernetes.Interface

	// redisReplayLister index properties about RedisReplay
	redisReplayLister listers.RedisReplayLister

	// redisReplayClient allows us to configure RedisReplay objects
	redisReplayClient redisreplayclientset.Interface

	// secretLister index properties about Secret
	secretLister listers.SecretLister

	// deploymentLister index properties about Deployment
	deploymentLister listers.DeploymentLister

	// serviceLister index properties about Service
	serviceLister listers.ServiceLister

	// endpointsLister index properties about Endpoints
	endpointsLister listers.EndpointsLister

	// roleBindingLister index properties about RoleBinding
	roleBindingLister listers.RoleBindingLister

	// serviceAccountLister index properties about ServiceAccount
	serviceAccount listers.ServiceAccountLister

	// sinkURI is the URI messages will be forwarded on to.
	sink string

	// redisReplayImage is the image used to run the RedisReplay
	redisReplayImage string

	// redisAddress is the address of the Redis server
	redisAddress string

	// redisUser is the username to connect to the Redis server
	redisUser string

	// redisPassword is the password to connect to the Redis server
	redisPassword string

	// redisDatabase is the database to connect to the Redis server
	redisDatabase string

	// startTime is the start time of the RedisReplay
	startTime string

	// endTime is the end time of the RedisReplay
	endTime string

	// filter is the filter used to filter the RedisReplay
	filter string

	// filterKind is the filter kind used to filter the RedisReplay
	filterKind string

	// logger for logging
	logger *zap.SugaredLogger
}

// Reconcile compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the RedisReplay resource
// with the current status of the resource.
func (r *Reconciler) Reconcile(ctx context.Context, key string) error {
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

func (r *Reconciler) reconcile(ctx context.Context, rr *v1alpha1.RedisReplay) error {
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

func (r *Reconciler) reconcileRedisReplay(ctx context.Context, rr *v1alpha1.RedisReplay) error {
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

func (r *Reconciler) reconcileRedisReplay(ctx context.Context, rr *v1alpha1.RedisReplay) error {
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

func (r *Reconciler) reconcileRedisReplay(ctx context.Context, rr *v1alpha1.RedisReplay) error {
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
