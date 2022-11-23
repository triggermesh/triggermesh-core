// Copyright 2022 TriggerMesh Inc.
// SPDX-License-Identifier: Apache-2.0

package memorybroker

import (
	"context"
	"strconv"

	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	"knative.dev/pkg/apis"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/network"
	knreconciler "knative.dev/pkg/reconciler"

	eventingv1alpha1 "github.com/triggermesh/triggermesh-core/pkg/apis/eventing/v1alpha1"
	"github.com/triggermesh/triggermesh-core/pkg/reconciler/common"
	"github.com/triggermesh/triggermesh-core/pkg/reconciler/resources"
)

type reconciler struct {
	secretReconciler common.SecretReconciler
	saReconciler     common.ServiceAccountReconciler
	brokerReconciler common.BrokerReconciler
}

// options that set Broker environment variables specific for the RedisBroker.
func memoryDeploymentOption(mb *eventingv1alpha1.MemoryBroker) resources.DeploymentOption {
	return func(d *appsv1.Deployment) {
		// Make sure the broker container exists before modifying it.
		if len(d.Spec.Template.Spec.Containers) == 0 {
			// Unexpected path.
			panic("The Broker Deployment to be reconciled has no containers in it.")
		}

		c := &d.Spec.Template.Spec.Containers[0]

		if mb.Spec.Memory != nil && mb.Spec.Memory.BufferSize != nil {
			resources.ContainerAddEnvFromValue("REDIS_STREAM", strconv.Itoa(*mb.Spec.Memory.BufferSize))(c)
		}
	}
}

func (r *reconciler) ReconcileKind(ctx context.Context, mb *eventingv1alpha1.MemoryBroker) knreconciler.Event {
	logging.FromContext(ctx).Infow("Reconciling", zap.Any("MemoryBroker", *mb))

	// Iterate triggers and create secret.
	secret, err := r.secretReconciler.Reconcile(ctx, mb)
	if err != nil {
		return err
	}

	// Make sure the Broker service account and roles exists.
	sa, _, err := r.saReconciler.Reconcile(ctx, mb)
	if err != nil {
		return err
	}

	// Make sure the Broker deployment exists.
	_, brokerSvc, err := r.brokerReconciler.Reconcile(ctx, mb, sa, secret, memoryDeploymentOption(mb))
	if err != nil {
		return err
	}

	// Set address to the Broker service.
	mb.Status.SetAddress(getServiceAddress(brokerSvc))

	return nil
}

func getServiceAddress(svc *corev1.Service) *apis.URL {
	var port string
	if svc.Spec.Ports[0].Port != 80 {
		port = ":" + strconv.Itoa(int(svc.Spec.Ports[0].Port))
	}

	return apis.HTTP(
		network.GetServiceHostname(svc.Name, svc.Namespace) + port)
}
