// Copyright 2022 TriggerMesh Inc.
// SPDX-License-Identifier: Apache-2.0

package redisbroker

import (
	"context"

	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"
	corev1listers "k8s.io/client-go/listers/core/v1"
	rbacv1listers "k8s.io/client-go/listers/rbac/v1"

	"knative.dev/pkg/logging"
	"knative.dev/pkg/reconciler"

	eventingv1alpha1 "github.com/triggermesh/triggermesh-core/pkg/apis/eventing/v1alpha1"
)

type Reconciler struct {
	kubeClientSet        kubernetes.Interface
	redisReconciler      redisReconciler
	serviceLister        corev1listers.ServiceLister
	serviceAccountLister corev1listers.ServiceAccountLister
	roleBindingLister    rbacv1listers.RoleBindingLister
}

func (r *Reconciler) ReconcileKind(ctx context.Context, rb *eventingv1alpha1.RedisBroker) reconciler.Event {
	logging.FromContext(ctx).Infow("Reconciling", zap.Any("Broker", *rb))

	// Clean any dangling resources

	// Iterate triggers and create secret

	// Make sure the Redis deployment exists and propagate the status to the Channel
	_, err := r.redisReconciler.Reconcile(ctx, rb)
	if err != nil {
		return err
	}

	// create service for redis

	// create deployment for broker

	// create service for broker

	return nil
}
