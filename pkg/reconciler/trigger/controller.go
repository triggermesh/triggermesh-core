// Copyright 2022 TriggerMesh Inc.
// SPDX-License-Identifier: Apache-2.0

package trigger

import (
	"context"

	"go.uber.org/zap"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/cache"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/kmeta"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/resolver"

	cfgInformer "knative.dev/pkg/client/injection/kube/informers/core/v1/configmap"

	"github.com/triggermesh/triggermesh-core/pkg/apis/eventing"
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
	cmInformer := cfgInformer.Get(ctx)

	r := &Reconciler{
		rbLister: rbInformer.Lister(),
		mbLister: mbInformer.Lister(),
		cmLister: cmInformer.Lister(),
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
			if tg.OwnerRefableMatchesBroker(accessor) {
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
			if tg.OwnerRefableMatchesBroker(accessor) {
				impl.EnqueueKey(types.NamespacedName{
					Name:      tg.Name,
					Namespace: tg.Namespace,
				})
			}
		}
	}

	filterConfigMapBroker := func(obj interface{}) bool {
		cm, ok := obj.(*corev1.ConfigMap)
		if !ok {
			return false
		}

		// Get the list of owner references and filter for those that
		// are owned by a Broker.
		obs := eventing.GetOwnerBrokers(cm)
		if len(obs) == 0 {
			return false
		}

		// Iterate all triggers at the namespace and select those that are applied
		// to the ConfigMap broker(s).
		tgs, err := tgInformer.Lister().Triggers(cm.Namespace).List(labels.Everything())
		if err != nil {
			logging.FromContext(ctx).Error("Unable to list Triggers", zap.Error(err))
			return false
		}

		// Finding one will make the filter pass.
		for i := range tgs {
			for j := range obs {
				if tgs[i].OwnerReferenceMatchesBroker(obs[j]) {
					return true
				}
			}
		}

		// No triggers that match the brokers found, do not enqueue.
		return false
	}

	enqueueFromConfigMapBroker := func(obj interface{}) {
		cm, ok := obj.(*corev1.ConfigMap)
		if !ok {
			return
		}

		// Get the list of owner references and filter for those that
		// are owned by a Broker.
		obs := eventing.GetOwnerBrokers(cm)
		if len(obs) == 0 {
			return
		}

		// Iterate all triggers at the namespace and select those that are applied
		// to the ConfigMap broker(s).
		tgs, err := tgInformer.Lister().Triggers(cm.Namespace).List(labels.Everything())
		if err != nil {
			logging.FromContext(ctx).Error("Unable to list Triggers", zap.Error(err))
			return
		}

		// Iterate all triggers at the namespace and select those that are applied
		// to the ConfigMap broker(s).
		// Finding one will make the filter pass.
		for i := range tgs {
			for j := range obs {
				if tgs[i].OwnerReferenceMatchesBroker(obs[j]) {
					impl.EnqueueKey(types.NamespacedName{
						Name:      tgs[i].Name,
						Namespace: tgs[i].Namespace,
					})
					break
				}
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

	cmInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: filterConfigMapBroker,
		Handler:    controller.HandleAll(enqueueFromConfigMapBroker),
	})

	return impl
}
