// Copyright 2023 TriggerMesh Inc.
// SPDX-License-Identifier: Apache-2.0

package redisreplay

import (
	"context"

	"github.com/kelseyhightower/envconfig"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"

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

	// create the reconciler
	r := &reconciler{
		rrLister: rrinformer.Lister(),
	}

	impl := rrreconciler.NewImpl(ctx, r)

	rrinformer.Informer().AddEventHandler(controller.HandleAll(impl.Enqueue))

	// rbInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
	// 	FilterFunc:
	// 	Handler:    controller.HandleAll(enqueueFromBroker),
	// })

	return impl
}
