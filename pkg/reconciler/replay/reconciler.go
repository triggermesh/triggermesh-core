// Copyright 2023 TriggerMesh Inc.
// SPDX-License-Identifier: Apache-2.0

package replay

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"
	pkgreconciler "knative.dev/pkg/reconciler"
	"knative.dev/pkg/resolver"

	eventingv1alpha1 "github.com/triggermesh/triggermesh-core/pkg/apis/eventing/v1alpha1"
	eventingv1alpha1listers "github.com/triggermesh/triggermesh-core/pkg/client/generated/listers/eventing/v1alpha1"
	"github.com/triggermesh/triggermesh-core/pkg/reconciler/common"
)

type Reconciler struct {
	// TODO duck brokers
	rbLister    eventingv1alpha1listers.RedisBrokerLister
	mbLister    eventingv1alpha1listers.MemoryBrokerLister
	uriResolver *resolver.URIResolver
}

func (r *Reconciler) ReconcileKind(ctx context.Context, rpl *eventingv1alpha1.Replay) pkgreconciler.Event {
	err := r.resolveBroker(ctx, rpl)
	if err != nil {
		return err
	}

	err = r.resolveTarget(ctx, rpl)
	if err != nil {
		return err
	}

	return nil
}

func (r *Reconciler) resolveBroker(ctx context.Context, rpl *eventingv1alpha1.Replay) pkgreconciler.Event {
	// TODO duck
	// TODO move to webhook
	switch {
	case rpl.Spec.Broker.Group == "":
		rpl.Spec.Broker.Group = eventingv1alpha1.SchemeGroupVersion.Group
	case rpl.Spec.Broker.Group != eventingv1alpha1.SchemeGroupVersion.Group:
		return controller.NewPermanentError(fmt.Errorf("not supported Broker Group %q", rpl.Spec.Broker.Group))
	}

	var rb *eventingv1alpha1.RedisBroker
	if rpl.Spec.Broker.Kind == rb.GetGroupVersionKind().Kind {
		return r.resolveRedisBroker(ctx, rpl)
	}

	var mb *eventingv1alpha1.MemoryBroker
	if rpl.Spec.Broker.Kind != mb.GetGroupVersionKind().Kind {
		return controller.NewPermanentError(fmt.Errorf("not supported Broker Kind %q", rpl.Spec.Broker.Kind))
	}

	return r.resolveMemoryBroker(ctx, rpl)
}

func (r *Reconciler) resolveRedisBroker(ctx context.Context, rpl *eventingv1alpha1.Replay) pkgreconciler.Event {
	rb, err := r.rbLister.RedisBrokers(rpl.Namespace).Get(rpl.Spec.Broker.Name)
	if err != nil {
		if apierrs.IsNotFound(err) {
			logging.FromContext(ctx).Errorw(fmt.Sprintf("Replay %s/%s references non existing broker %q", rpl.Namespace, rpl.Name, rpl.Spec.Broker.Name))
			rpl.Status.MarkBrokerFailed(common.ReasonBrokerDoesNotExist, "Broker %q does not exist", rpl.Spec.Broker.Name)
			// No need to requeue, we will be notified when if broker is created.
			return controller.NewPermanentError(err)
		}

		rpl.Status.MarkBrokerFailed(common.ReasonFailedBrokerGet, "Failed to get broker %q : %s", rpl.Spec.Broker, err)
		return pkgreconciler.NewEvent(corev1.EventTypeWarning, common.ReasonFailedBrokerGet,
			"Failed to get broker for replay %s/%s: %w", rpl.Namespace, rpl.Name, err)
	}

	rpl.Status.PropagateBrokerCondition(rb.Status.GetTopLevelCondition())

	// No need to requeue, we'll get requeued when broker changes status.
	if !rb.IsReady() {
		logging.FromContext(ctx).Errorw(fmt.Sprintf("Replay %s/%s references non ready broker %q", rpl.Namespace, rpl.Name, rpl.Spec.Broker.Name))
	}

	return nil
}

func (r *Reconciler) resolveMemoryBroker(ctx context.Context, rpl *eventingv1alpha1.Replay) pkgreconciler.Event {
	mb, err := r.mbLister.MemoryBrokers(rpl.Namespace).Get(rpl.Spec.Broker.Name)
	if err != nil {
		if apierrs.IsNotFound(err) {
			logging.FromContext(ctx).Errorw(fmt.Sprintf("Replay %s/%s references non existing broker %q", rpl.Namespace, rpl.Name, rpl.Spec.Broker.Name))
			rpl.Status.MarkBrokerFailed(common.ReasonBrokerDoesNotExist, "Broker %q does not exist", rpl.Spec.Broker.Name)
			// No need to requeue, we will be notified when if broker is created.
			return controller.NewPermanentError(err)
		}

		rpl.Status.MarkBrokerFailed(common.ReasonFailedBrokerGet, "Failed to get broker %q : %s", rpl.Spec.Broker, err)
		return pkgreconciler.NewEvent(corev1.EventTypeWarning, common.ReasonFailedBrokerGet,
			"Failed to get broker for replay %s/%s: %w", rpl.Namespace, rpl.Name, err)
	}

	rpl.Status.PropagateBrokerCondition(mb.Status.GetTopLevelCondition())

	// No need to requeue, we'll get requeued when broker changes status.
	if !mb.IsReady() {
		logging.FromContext(ctx).Errorw(fmt.Sprintf("Replay %s/%s references non ready broker %q", rpl.Namespace, rpl.Name, rpl.Spec.Broker.Name))
	}

	return nil
}

func (r *Reconciler) resolveTarget(ctx context.Context, rpl *eventingv1alpha1.Replay) pkgreconciler.Event {
	if rpl.Spec.Target.Ref != nil && rpl.Spec.Target.Ref.Namespace == "" {
		// To call URIFromDestinationV1(ctx context.Context, dest v1.Destination, parent interface{}), dest.Ref must have a Namespace
		// If Target.Ref.Namespace is nil, We will use the Namespace of Replay as the Namespace of dest.Ref
		rpl.Spec.Target.Ref.Namespace = rpl.Namespace
	}

	targetURI, err := r.uriResolver.URIFromDestinationV1(ctx, rpl.Spec.Target, rpl)
	if err != nil {
		logging.FromContext(ctx).Errorw("Unable to get the target's URI", zap.Error(err))
		rpl.Status.MarkTargetResolvedFailed("Unable to get the target's URI", "%v", err)
		rpl.Status.TargetURI = nil
		return pkgreconciler.NewEvent(corev1.EventTypeWarning, common.ReasonFailedResolveReference,
			"Failed to get target's URI: %w", err)
	}

	rpl.Status.TargetURI = targetURI
	rpl.Status.MarkTargetResolvedSucceeded()

	return nil
}
