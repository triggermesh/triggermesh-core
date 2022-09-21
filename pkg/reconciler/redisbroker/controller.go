// Copyright 2022 TriggerMesh Inc.
// SPDX-License-Identifier: Apache-2.0

package redisbroker

import (
	"context"

	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"

	rbinformer "github.com/triggermesh/triggermesh-core/pkg/client/generated/injection/informers/eventing/v1alpha1/redisbroker"
	rbreconciler "github.com/triggermesh/triggermesh-core/pkg/client/generated/injection/reconciler/eventing/v1alpha1/redisbroker"
)

// NewController initializes the controller and is called by the generated code
// Registers event handlers to enqueue events
func NewController(
	ctx context.Context,
	cmw configmap.Watcher,
) *controller.Impl {
	rbInformer := rbinformer.Get(ctx)

	r := &Reconciler{}

	impl := rbreconciler.NewImpl(ctx, r)

	rbInformer.Informer().AddEventHandler(controller.HandleAll(impl.Enqueue))

	return impl
}
