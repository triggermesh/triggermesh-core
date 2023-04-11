// Copyright 2023 TriggerMesh Inc.
// SPDX-License-Identifier: Apache-2.0

package redisreplay

import (
	"context"

	"github.com/kelseyhightower/envconfig"
	"k8s.io/client-go/tools/cache"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/resolver"
	"knative.dev/pkg/tracker"

	kubeclient "knative.dev/pkg/client/injection/kube/client"
	jobinformer "knative.dev/pkg/client/injection/kube/informers/batch/v1/job"

	"github.com/triggermesh/triggermesh-core/pkg/apis/eventing/v1alpha1"
	// rbinformer "github.com/triggermesh/triggermesh-core/pkg/client/generated/injection/informers/eventing/v1alpha1/redisbroker"
	rrinformer "github.com/triggermesh/triggermesh-core/pkg/client/generated/injection/informers/eventing/v1alpha1/redisreplay"
	rrreconciler "github.com/triggermesh/triggermesh-core/pkg/client/generated/injection/reconciler/eventing/v1alpha1/redisreplay"
)

// envConfig will be used to extract the required environment variables using
// github.com/kelseyhightower/envconfig. If this configuration cannot be extracted, then
// NewController will panic.
type envConfig struct {
	RedisReplayImage string `envconfig:"REDISBROKER_REPLAY_IMAGE" required:"true"`
	ImagePullPolicy  string `envconfig:"IMAGE_PULL_POLICY" default:"IfNotPresent"`
}

// NewController initializes the controller and is called by the generated code
// Registers event handlers to enqueue events
func NewController(
	ctx context.Context,
	cmw configmap.Watcher,
) *controller.Impl {

	// log that we are starting the controller
	logging.FromContext(ctx).Info("Starting RedisReplay controller")

	env := &envConfig{}
	if err := envconfig.Process("", env); err != nil {
		logging.FromContext(ctx).Panicf("unable to process RedisReplay's required environment variables: %v", err)
	}

	rrinformer := rrinformer.Get(ctx)
	jinformer := jobinformer.Get(ctx)
	// rbinformer := rbinformer.Get(ctx)

	// create the reconciler
	r := &reconciler{
		rrLister: rrinformer.Lister(),
		image:    env.RedisReplayImage,
		// rbLister:   rbinformer.Lister(),
		jobsLister: jobinformer.Get(ctx).Lister(),
		client:     kubeclient.Get(ctx),
	}

	impl := rrreconciler.NewImpl(ctx, r)
	// Create a new tracker
	t := tracker.New(impl.EnqueueKey, controller.GetTrackerLease(ctx))

	r.uriResolver = resolver.NewURIResolverFromTracker(ctx, t)

	rrinformer.Informer().AddEventHandler(controller.HandleAll(impl.Enqueue))
	jinformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: controller.FilterController(&v1alpha1.RedisReplay{}),
		Handler:    controller.HandleAll(impl.EnqueueControllerOf),
	})
	return impl
}
