// Copyright 2022 TriggerMesh Inc.
// SPDX-License-Identifier: Apache-2.0

package memorybroker

import (
	"context"
	"strconv"

	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"

	"knative.dev/pkg/apis"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/network"
	"knative.dev/pkg/reconciler"

	eventingv1alpha1 "github.com/triggermesh/triggermesh-core/pkg/apis/eventing/v1alpha1"
	"github.com/triggermesh/triggermesh-core/pkg/reconciler/common"
)

const (
	appAnnotationValue = "memorybroker"
)

type Reconciler struct {
	kubeClientSet    kubernetes.Interface
	secretReconciler common.SecretReconciler
	// secretReconciler secretReconciler
	// memoryReconciler  memoryReconciler
	// brokerReconciler brokerReconciler
	// saReconciler     serviceAccountReconciler
}

func (r *Reconciler) ReconcileKind(ctx context.Context, rb *eventingv1alpha1.MemoryBroker) reconciler.Event {
	logging.FromContext(ctx).Infow("Reconciling", zap.Any("Broker", *rb))

	// Make sure the Memory deployment and service exists.
	_, memorySvc, err := r.memoryReconciler.reconcile(ctx, rb)
	if err != nil {
		return err
	}

	// Iterate triggers and create secret.
	secret, err := r.secretReconciler.Reconcile(ctx, rb)
	if err != nil {
		return err
	}

	// Make sure the Broker service account and roles exists.
	sa, _, err := r.saReconciler.reconcile(ctx, rb)
	if err != nil {
		return err
	}

	// Make sure the Broker deployment for Memory exists and that it points to the Memory service.
	_, brokerSvc, err := r.brokerReconciler.reconcile(ctx, rb, sa, memorySvc, secret)
	if err != nil {
		return err
	}

	// Set address to the Broker service.
	rb.Status.SetAddress(getSericeAddress(brokerSvc))

	return nil
}

func getSericeAddress(svc *v1.Service) *apis.URL {
	var port string
	if svc.Spec.Ports[0].Port != 80 {
		port = ":" + strconv.Itoa(int(svc.Spec.Ports[0].Port))
	}

	return apis.HTTP(
		network.GetServiceHostname(svc.Name, svc.Namespace) + port)
}
