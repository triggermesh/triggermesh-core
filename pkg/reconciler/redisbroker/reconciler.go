// Copyright 2022 TriggerMesh Inc.
// SPDX-License-Identifier: Apache-2.0

package redisbroker

import (
	"context"

	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"

	"knative.dev/pkg/logging"
	"knative.dev/pkg/reconciler"

	eventingv1alpha1 "github.com/triggermesh/triggermesh-core/pkg/apis/eventing/v1alpha1"
)

type Reconciler struct {
	kubeClientSet    kubernetes.Interface
	redisReconciler  redisReconciler
	brokerReconciler brokerReconciler
}

func (r *Reconciler) ReconcileKind(ctx context.Context, rb *eventingv1alpha1.RedisBroker) reconciler.Event {
	logging.FromContext(ctx).Infow("Reconciling", zap.Any("Broker", *rb))

	// Clean any dangling resources

	// Make sure the Redis deployment and service exists.
	_, _, err := r.redisReconciler.reconcile(ctx, rb)
	if err != nil {
		return err
	}

	// Iterate triggers and create secret

	// Make sure the Broker deployment for Redis exists and that it points to the Redis service.
	_, _, err = r.brokerReconciler.reconcile(ctx, rb)
	if err != nil {
		return err
	}

	return nil
}
