// Copyright 2022 TriggerMesh Inc.
// SPDX-License-Identifier: Apache-2.0

package trigger

import (
	"context"

	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/cache"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/kmeta"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/resolver"

	eventingv1alpha1 "github.com/triggermesh/triggermesh-core/pkg/apis/eventing/v1alpha1"
	mbinformer "github.com/triggermesh/triggermesh-core/pkg/client/generated/injection/informers/eventing/v1alpha1/memorybroker"
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
	tgInformer := tginformer.Get(ctx)
	rbInformer := rbinformer.Get(ctx)
	mbInformer := mbinformer.Get(ctx)

	r := &Reconciler{
		rbLister: rbInformer.Lister(),
		mbLister: mbInformer.Lister(),
	}

	impl := tgreconciler.NewImpl(ctx, r)

	r.uriResolver = resolver.NewURIResolverFromTracker(ctx, impl.Tracker)

	tgInformer.Informer().AddEventHandler(controller.HandleAll(impl.Enqueue))

	// Filter brokers that are referenced by triggers.
	filterBroker := func(obj interface{}) bool {
		// TODO duck
		var accessor kmeta.OwnerRefableAccessor
		rb, ok := obj.(*eventingv1alpha1.RedisBroker)
		if !ok {
			mb, ok := obj.(*eventingv1alpha1.MemoryBroker)
			if !ok {
				return false
			}
			accessor = kmeta.OwnerRefableAccessor(mb)
		} else {
			accessor = kmeta.OwnerRefableAccessor(rb)
		}

		tgl, err := tgInformer.Lister().Triggers(accessor.GetNamespace()).List(labels.Everything())
		if err != nil {
			logging.FromContext(ctx).Error("Unable to list Triggers", zap.Error(err))
			return false
		}

		for _, tg := range tgl {
			if tg.ReferencesBroker(accessor) {
				return true
			}
		}

		return false
	}

	enqueueFromBroker := func(obj interface{}) {
		// TODO duck
		var accessor kmeta.OwnerRefableAccessor
		rb, ok := obj.(*eventingv1alpha1.RedisBroker)
		if !ok {
			mb, ok := obj.(*eventingv1alpha1.MemoryBroker)
			if !ok {
				return
			}
			accessor = kmeta.OwnerRefableAccessor(mb)
		} else {
			accessor = kmeta.OwnerRefableAccessor(rb)
		}

		tgl, err := tgInformer.Lister().Triggers(accessor.GetNamespace()).List(labels.Everything())
		if err != nil {
			logging.FromContext(ctx).Error("Unable to list Triggers", zap.Error(err))
			return
		}

		for _, tg := range tgl {
			if tg.ReferencesBroker(accessor) {
				impl.EnqueueKey(types.NamespacedName{
					Name:      tg.Name,
					Namespace: tg.Namespace,
				})
			}
		}
	}

	rbInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: filterBroker,
		Handler:    controller.HandleAll(enqueueFromBroker),
	})

	mbInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: filterBroker,
		Handler:    controller.HandleAll(enqueueFromBroker),
	})

	return impl
}
