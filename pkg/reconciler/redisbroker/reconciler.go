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

const (
	appAnnotation          = "app"
	appAnnotationValue     = "redisbroker"
	resourceNameAnnotation = "eventing.triggermesh.io/name"
)

type Reconciler struct {
	kubeClientSet    kubernetes.Interface
	secretReconciler secretReconciler
	redisReconciler  redisReconciler
	brokerReconciler brokerReconciler
}

func (r *Reconciler) ReconcileKind(ctx context.Context, rb *eventingv1alpha1.RedisBroker) reconciler.Event {
	logging.FromContext(ctx).Infow("Reconciling", zap.Any("Broker", *rb))

	// Make sure the Redis deployment and service exists.
	_, svc, err := r.redisReconciler.reconcile(ctx, rb)
	if err != nil {
		return err
	}

	// Iterate triggers and create secret
	secret, err := r.secretReconciler.reconcile(ctx, rb)
	if err != nil {
		return err
	}

	// Make sure the Broker deployment for Redis exists and that it points to the Redis service.
	_, _, err = r.brokerReconciler.reconcile(ctx, rb, svc, secret)
	if err != nil {
		return err
	}

	return nil
}
