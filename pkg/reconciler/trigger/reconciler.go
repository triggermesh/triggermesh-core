// Copyright 2022 TriggerMesh Inc.
// SPDX-License-Identifier: Apache-2.0

package trigger

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"knative.dev/pkg/logging"
	pkgreconciler "knative.dev/pkg/reconciler"
	"knative.dev/pkg/resolver"

	eventingv1alpha1 "github.com/triggermesh/triggermesh-core/pkg/apis/eventing/v1alpha1"
	eventingv1alpha1listers "github.com/triggermesh/triggermesh-core/pkg/client/generated/listers/eventing/v1alpha1"
	"github.com/triggermesh/triggermesh-core/pkg/reconciler"
)

type Reconciler struct {
	rbLister    eventingv1alpha1listers.RedisBrokerLister
	uriResolver *resolver.URIResolver
}

func (r *Reconciler) ReconcileKind(ctx context.Context, t *eventingv1alpha1.Trigger) pkgreconciler.Event {
	// TODO, use any broker, not RedisBrokers
	rb, err := r.rbLister.RedisBrokers(t.Namespace).Get(t.Spec.Broker.Name)
	if err != nil {
		if apierrs.IsNotFound(err) {
			logging.FromContext(ctx).Errorw(fmt.Sprintf("Trigger %s/%s references non existing broker %q", t.Namespace, t.Name, t.Spec.Broker.Name))
			t.Status.MarkBrokerFailed(reconciler.ReasonBrokerDoesNotExist, "Broker %q does not exist", t.Spec.Broker.Name)
			return nil
		}

		t.Status.MarkBrokerFailed(reconciler.ReasonFailedBrokerGet, "Failed to get broker %q : %s", t.Spec.Broker, err)
		return err
	}

	t.Status.PropagateBrokerCondition(rb.Status.GetTopLevelCondition())
	// If Broker is not ready, we're done, but once it becomes ready, we'll get requeued.
	if !rb.IsReady() {
		logging.FromContext(ctx).Errorw("Broker is not ready", zap.Any("Broker", *rb))
		return nil
	}

	if t.Spec.Target.Ref != nil && t.Spec.Target.Ref.Namespace == "" {
		// To call URIFromDestinationV1(ctx context.Context, dest v1.Destination, parent interface{}), dest.Ref must have a Namespace
		// If Target.Ref.Namespace is nil, We will use the Namespace of Trigger as the Namespace of dest.Ref
		t.Spec.Target.Ref.Namespace = t.GetNamespace()
	}

	targetURI, err := r.uriResolver.URIFromDestinationV1(ctx, t.Spec.Target, rb)
	if err != nil {
		logging.FromContext(ctx).Errorw("Unable to get the target's URI", zap.Error(err))
		t.Status.MarkTargetResolvedFailed("Unable to get the target's URI", "%v", err)
		t.Status.TargetURI = nil
		return err
	}

	t.Status.TargetURI = targetURI
	t.Status.MarkTargetResolvedSucceeded()

	return nil
}
