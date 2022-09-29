// Copyright 2022 TriggerMesh Inc.
// SPDX-License-Identifier: Apache-2.0

package trigger

import (
	"context"

	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/resolver"

	rbinformer "github.com/triggermesh/triggermesh-core/pkg/client/generated/injection/informers/eventing/v1alpha1/redisbroker"
	tginformer "github.com/triggermesh/triggermesh-core/pkg/client/generated/injection/informers/eventing/v1alpha1/trigger"
	tgreconciler "github.com/triggermesh/triggermesh-core/pkg/client/generated/injection/reconciler/eventing/v1alpha1/trigger"
)

// NewController initializes the controller and is called by the generated code
// Registers event handlers to enqueue events
func NewController(
	ctx context.Context,
	cmw configmap.Watcher,
) *controller.Impl {
	tginformer := tginformer.Get(ctx)

	r := &Reconciler{
		rbLister: rbinformer.Get(ctx).Lister(),
	}

	impl := tgreconciler.NewImpl(ctx, r)

	r.uriResolver = resolver.NewURIResolverFromTracker(ctx, impl.Tracker)

	// TODO add brokers event handlers and filter for triggers that use them.
	tginformer.Informer().AddEventHandler(controller.HandleAll(impl.Enqueue))

	return impl
}
