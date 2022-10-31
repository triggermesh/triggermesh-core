// Copyright 2022 TriggerMesh Inc.
// SPDX-License-Identifier: Apache-2.0

package redisbroker

import (
	"context"

	"github.com/kelseyhightower/envconfig"
	"go.uber.org/zap"

	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/cache"

	kubeclient "knative.dev/pkg/client/injection/kube/client"
	"knative.dev/pkg/client/injection/kube/informers/apps/v1/deployment"
	endpointsinformer "knative.dev/pkg/client/injection/kube/informers/core/v1/endpoints"
	"knative.dev/pkg/client/injection/kube/informers/core/v1/secret"
	"knative.dev/pkg/client/injection/kube/informers/core/v1/service"
	"knative.dev/pkg/client/injection/kube/informers/core/v1/serviceaccount"
	"knative.dev/pkg/client/injection/kube/informers/rbac/v1/rolebinding"
	rolebindingsinformer "knative.dev/pkg/client/injection/kube/informers/rbac/v1/rolebinding"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"

	eventingv1alpha1 "github.com/triggermesh/triggermesh-core/pkg/apis/eventing/v1alpha1"
	rbinformer "github.com/triggermesh/triggermesh-core/pkg/client/generated/injection/informers/eventing/v1alpha1/redisbroker"
	trginformer "github.com/triggermesh/triggermesh-core/pkg/client/generated/injection/informers/eventing/v1alpha1/trigger"
	rbreconciler "github.com/triggermesh/triggermesh-core/pkg/client/generated/injection/reconciler/eventing/v1alpha1/redisbroker"
)

// envConfig will be used to extract the required environment variables using
// github.com/kelseyhightower/envconfig. If this configuration cannot be extracted, then
// NewController will panic.
type envConfig struct {
	RedisImage            string `envconfig:"REDISBROKER_REDIS_IMAGE" required:"true"`
	BrokerImage           string `envconfig:"REDISBROKER_BROKER_IMAGE" required:"true"`
	BrokerImagePullPolicy string `envconfig:"REDISBROKER_BROKER_IMAGE_PULL_POLICY" default:"IfNotPresent"`
}

// NewController initializes the controller and is called by the generated code
// Registers event handlers to enqueue events
func NewController(
	ctx context.Context,
	cmw configmap.Watcher,
) *controller.Impl {

	env := &envConfig{}
	if err := envconfig.Process("", env); err != nil {
		logging.FromContext(ctx).Panicf("unable to process RedisBroker's required environment variables: %v", err)
	}

	rbInformer := rbinformer.Get(ctx)
	trgInformer := trginformer.Get(ctx)
	secretInformer := secret.Get(ctx)
	deploymentInformer := deployment.Get(ctx)
	serviceInformer := service.Get(ctx)
	endpointsInformer := endpointsinformer.Get(ctx)
	serviceAccountInformer := serviceaccount.Get(ctx)
	roleBindingsInformer := rolebindingsinformer.Get(ctx)

	_ = rolebinding.Get(ctx)

	r := &Reconciler{
		kubeClientSet:    kubeclient.Get(ctx),
		secretReconciler: newSecretReconciler(ctx, secretInformer.Lister(), trgInformer.Lister()),
		redisReconciler: redisReconciler{
			client:           kubeclient.Get(ctx),
			deploymentLister: deploymentInformer.Lister(),
			serviceLister:    serviceInformer.Lister(),
			endpointsLister:  endpointsInformer.Lister(),
			image:            env.RedisImage,
		},


		saReconciler: serviceAccountReconciler{
			client:           kubeclient.Get(ctx),
			serviceAccountLister: serviceAccountInformer.Lister(),
			roleBindingLister: roleBindingsInformer.Lister(),
		},

		brokerReconciler: brokerReconciler{
			client:           kubeclient.Get(ctx),
			deploymentLister: deploymentInformer.Lister(),
			serviceLister:    serviceInformer.Lister(),
			endpointsLister:  endpointsInformer.Lister(),
			image:            env.BrokerImage,
			pullPolicy:       corev1.PullPolicy(env.BrokerImagePullPolicy),
		},
	}

	impl := rbreconciler.NewImpl(ctx, r)

	rb := &eventingv1alpha1.RedisBroker{}
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
			if !ok || ep.Labels != nil || ep.Labels[appAnnotation] == appAnnotationValue {
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
		Handler: controller.HandleAll(impl.EnqueueControllerOf),
	})

	// Filter Triggers that reference a Redis broker.
	filterTriggerForRedisBroker := func(obj interface{}) bool {
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
		_, err := rbInformer.Lister().RedisBrokers(t.Namespace).Get(t.Spec.Broker.Name)
		switch {
		case err == nil:
			return true
		case !apierrs.IsNotFound(err):
			logging.FromContext(ctx).Error("Unable to get Redis Broker", zap.Any("broker", t.Spec.Broker), zap.Error(err))
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
		FilterFunc: filterTriggerForRedisBroker,
		Handler:    controller.HandleAll(enqueueFromTrigger),
	})

	return impl
}
