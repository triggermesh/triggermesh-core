// Copyright 2022 TriggerMesh Inc.
// SPDX-License-Identifier: Apache-2.0

package memorybroker

import (
	"context"

	"github.com/kelseyhightower/envconfig"
	"go.uber.org/zap"

	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/cache"

	"knative.dev/pkg/client/injection/kube/informers/apps/v1/deployment"
	endpointsinformer "knative.dev/pkg/client/injection/kube/informers/core/v1/endpoints"
	"knative.dev/pkg/client/injection/kube/informers/core/v1/secret"
	"knative.dev/pkg/client/injection/kube/informers/core/v1/service"
	"knative.dev/pkg/client/injection/kube/informers/core/v1/serviceaccount"
	rolebindingsinformer "knative.dev/pkg/client/injection/kube/informers/rbac/v1/rolebinding"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"

	eventingv1alpha1 "github.com/triggermesh/triggermesh-core/pkg/apis/eventing/v1alpha1"
	rbinformer "github.com/triggermesh/triggermesh-core/pkg/client/generated/injection/informers/eventing/v1alpha1/memorybroker"
	rplinformer "github.com/triggermesh/triggermesh-core/pkg/client/generated/injection/informers/eventing/v1alpha1/replay"
	trginformer "github.com/triggermesh/triggermesh-core/pkg/client/generated/injection/informers/eventing/v1alpha1/trigger"
	rbreconciler "github.com/triggermesh/triggermesh-core/pkg/client/generated/injection/reconciler/eventing/v1alpha1/memorybroker"
	"github.com/triggermesh/triggermesh-core/pkg/reconciler/common"
	"github.com/triggermesh/triggermesh-core/pkg/reconciler/resources"
)

// envConfig will be used to extract the required environment variables using
// github.com/kelseyhightower/envconfig. If this configuration cannot be extracted, then
// NewController will panic.
type envConfig struct {
	BrokerImage           string `envconfig:"MEMORYBROKER_BROKER_IMAGE" required:"true"`
	BrokerImagePullPolicy string `envconfig:"MEMORYBROKER_BROKER_IMAGE_PULL_POLICY" default:"IfNotPresent"`
}

// NewController initializes the controller and is called by the generated code
// Registers event handlers to enqueue events
func NewController(
	ctx context.Context,
	cmw configmap.Watcher,
) *controller.Impl {

	env := &envConfig{}
	if err := envconfig.Process("", env); err != nil {
		logging.FromContext(ctx).Panicf("unable to process MemoryBroker's required environment variables: %v", err)
	}

	rbInformer := rbinformer.Get(ctx)
	trgInformer := trginformer.Get(ctx)
	rplInformer := rplinformer.Get(ctx)
	secretInformer := secret.Get(ctx)
	deploymentInformer := deployment.Get(ctx)
	serviceInformer := service.Get(ctx)
	endpointsInformer := endpointsinformer.Get(ctx)
	serviceAccountInformer := serviceaccount.Get(ctx)
	roleBindingsInformer := rolebindingsinformer.Get(ctx)

	r := &reconciler{
		secretReconciler: common.NewSecretReconciler(ctx, secretInformer.Lister(), trgInformer.Lister(), rplInformer.Lister()),
		saReconciler:     common.NewServiceAccountReconciler(ctx, serviceAccountInformer.Lister(), roleBindingsInformer.Lister()),
		brokerReconciler: common.NewBrokerReconciler(ctx, deploymentInformer.Lister(), serviceInformer.Lister(), endpointsInformer.Lister(),
			env.BrokerImage, corev1.PullPolicy(env.BrokerImagePullPolicy)),
	}

	impl := rbreconciler.NewImpl(ctx, r)

	rb := &eventingv1alpha1.MemoryBroker{}
	gvk := rb.GetGroupVersionKind()

	rbInformer.Informer().AddEventHandler(controller.HandleAll(impl.Enqueue))

	secretInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: controller.FilterController(rb),
		Handler:    controller.HandleAll(impl.EnqueueControllerOf),
	})
	deploymentInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: controller.FilterController(rb),
		Handler:    controller.HandleAll(impl.EnqueueControllerOf),
	})
	serviceInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: controller.FilterController(rb),
		Handler:    controller.HandleAll(impl.EnqueueControllerOf),
	})
	endpointsInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: func(obj interface{}) bool {
			ep, ok := obj.(*corev1.Endpoints)
			if !ok || ep.Labels != nil || ep.Labels[resources.AppNameLabel] == common.AppAnnotationValue(rb) {
				return false
			}

			return true
		},
		Handler: controller.HandleAll(func(obj interface{}) {
			ep, ok := obj.(*corev1.Endpoints)
			if !ok {
				return
			}

			svc, err := serviceInformer.Lister().Services(ep.Namespace).Get(ep.Name)
			if err != nil {
				// no matter the error, if we cannot retrieve the service we cannot
				// read the owner and enqueue the key.
				return
			}

			impl.EnqueueControllerOf(svc)
		}),
	})
	serviceAccountInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: controller.FilterController(rb),
		Handler:    controller.HandleAll(impl.EnqueueControllerOf),
	})
	roleBindingsInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: controller.FilterController(rb),
		Handler:    controller.HandleAll(impl.EnqueueControllerOf),
	})

	// Filter Triggers that reference a Memory broker.
	filterTriggerForMemoryBroker := func(obj interface{}) bool {
		t, ok := obj.(*eventingv1alpha1.Trigger)
		if !ok {
			return false
		}

		// TODO replace with defaulting when webhook is implemented
		if !(t.Spec.Broker.Group == gvk.Group || t.Spec.Broker.Group == "") ||
			t.Spec.Broker.Kind != gvk.Kind {
			return false
		}

		// TODO replace with broker namespace when webhook defaulting is implemented
		_, err := rbInformer.Lister().MemoryBrokers(t.Namespace).Get(t.Spec.Broker.Name)
		switch {
		case err == nil:
			return true
		case !apierrs.IsNotFound(err):
			logging.FromContext(ctx).Error("Unable to get Memory Broker", zap.Any("broker", t.Spec.Broker), zap.Error(err))
		}

		return false
	}

	enqueueFromTrigger := func(obj interface{}) {
		t, ok := obj.(*eventingv1alpha1.Trigger)
		if !ok {
			return
		}

		impl.EnqueueKey(types.NamespacedName{
			Name:      t.Spec.Broker.Name,
			Namespace: t.Namespace,
		})
	}

	trgInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: filterTriggerForMemoryBroker,
		Handler:    controller.HandleAll(enqueueFromTrigger),
	})

	return impl
}
