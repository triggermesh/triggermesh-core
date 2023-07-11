// Copyright 2022 TriggerMesh Inc.
// SPDX-License-Identifier: Apache-2.0

package trigger

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	corev1listers "k8s.io/client-go/listers/core/v1"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"
	pkgreconciler "knative.dev/pkg/reconciler"
	"knative.dev/pkg/resolver"

	"github.com/triggermesh/brokers/pkg/status"

	eventingv1alpha1 "github.com/triggermesh/triggermesh-core/pkg/apis/eventing/v1alpha1"
	eventingv1alpha1listers "github.com/triggermesh/triggermesh-core/pkg/client/generated/listers/eventing/v1alpha1"
	"github.com/triggermesh/triggermesh-core/pkg/reconciler/common"
)

type Reconciler struct {
	// TODO duck brokers
	rbLister    eventingv1alpha1listers.RedisBrokerLister
	mbLister    eventingv1alpha1listers.MemoryBrokerLister
	cmLister    corev1listers.ConfigMapLister
	uriResolver *resolver.URIResolver
}

func (r *Reconciler) ReconcileKind(ctx context.Context, t *eventingv1alpha1.Trigger) pkgreconciler.Event {
	b, err := r.resolveBroker(ctx, t)
	if err != nil {
		return err
	}

	if err := r.resolveTarget(ctx, t); err != nil {
		return err
	}

	if err := r.resolveDLS(ctx, t); err != nil {
		return err
	}

	return r.reconcileStatusConfigMap(ctx, t, b)
}

func (r *Reconciler) resolveBroker(ctx context.Context, t *eventingv1alpha1.Trigger) (eventingv1alpha1.ReconcilableBroker, pkgreconciler.Event) {
	// TODO duck
	// TODO move to webhook
	switch {
	case t.Spec.Broker.Group == "":
		t.Spec.Broker.Group = eventingv1alpha1.SchemeGroupVersion.Group
	case t.Spec.Broker.Group != eventingv1alpha1.SchemeGroupVersion.Group:
		return nil, controller.NewPermanentError(fmt.Errorf("not supported Broker Group %q", t.Spec.Broker.Group))
	}

	var rb *eventingv1alpha1.RedisBroker
	if t.Spec.Broker.Kind == rb.GetGroupVersionKind().Kind {
		return r.resolveRedisBroker(ctx, t)
	}

	var mb *eventingv1alpha1.MemoryBroker
	if t.Spec.Broker.Kind != mb.GetGroupVersionKind().Kind {
		return nil, controller.NewPermanentError(fmt.Errorf("not supported Broker Kind %q", t.Spec.Broker.Kind))
	}

	return r.resolveMemoryBroker(ctx, t)
}

func (r *Reconciler) resolveRedisBroker(ctx context.Context, t *eventingv1alpha1.Trigger) (eventingv1alpha1.ReconcilableBroker, pkgreconciler.Event) {
	rb, err := r.rbLister.RedisBrokers(t.Namespace).Get(t.Spec.Broker.Name)
	if err != nil {
		if apierrs.IsNotFound(err) {
			logging.FromContext(ctx).Errorw(fmt.Sprintf("Trigger %s/%s references non existing broker %q", t.Namespace, t.Name, t.Spec.Broker.Name))
			t.Status.MarkBrokerFailed(common.ReasonBrokerDoesNotExist, "Broker %q does not exist", t.Spec.Broker.Name)
			// No need to requeue, we will be notified when if broker is created.
			return nil, controller.NewPermanentError(err)
		}

		t.Status.MarkBrokerFailed(common.ReasonFailedBrokerGet, "Failed to get broker %q : %s", t.Spec.Broker, err)
		return nil, pkgreconciler.NewEvent(corev1.EventTypeWarning, common.ReasonFailedBrokerGet,
			"Failed to get broker for trigger %s/%s: %w", t.Namespace, t.Name, err)
	}

	t.Status.PropagateBrokerCondition(rb.Status.GetTopLevelCondition())

	// No need to requeue, we'll get requeued when broker changes status.
	if !rb.IsReady() {
		logging.FromContext(ctx).Errorw(fmt.Sprintf("Trigger %s/%s references non ready broker %q", t.Namespace, t.Name, t.Spec.Broker.Name))
	}

	return rb, nil
}

func (r *Reconciler) resolveMemoryBroker(ctx context.Context, t *eventingv1alpha1.Trigger) (eventingv1alpha1.ReconcilableBroker, pkgreconciler.Event) {
	mb, err := r.mbLister.MemoryBrokers(t.Namespace).Get(t.Spec.Broker.Name)
	if err != nil {
		if apierrs.IsNotFound(err) {
			logging.FromContext(ctx).Errorf("Trigger %s/%s references non existing broker %q", t.Namespace, t.Name, t.Spec.Broker.Name)
			t.Status.MarkBrokerFailed(common.ReasonBrokerDoesNotExist, "Broker %q does not exist", t.Spec.Broker.Name)
			// No need to requeue, we will be notified when broker is created.
			return nil, controller.NewPermanentError(err)
		}

		t.Status.MarkBrokerFailed(common.ReasonFailedBrokerGet, "Failed to get broker %q : %s", t.Spec.Broker, err)
		return nil, pkgreconciler.NewEvent(corev1.EventTypeWarning, common.ReasonFailedBrokerGet,
			"Failed to get broker for trigger %s/%s: %w", t.Namespace, t.Name, err)
	}

	t.Status.PropagateBrokerCondition(mb.Status.GetTopLevelCondition())

	// No need to requeue, we'll get requeued when broker changes status.
	if !mb.IsReady() {
		logging.FromContext(ctx).Errorw(fmt.Sprintf("Trigger %s/%s references non ready broker %q", t.Namespace, t.Name, t.Spec.Broker.Name))
	}

	return mb, nil
}

func (r *Reconciler) resolveTarget(ctx context.Context, t *eventingv1alpha1.Trigger) pkgreconciler.Event {
	if t.Spec.Target.Ref != nil && t.Spec.Target.Ref.Namespace == "" {
		// To call URIFromDestinationV1(ctx context.Context, dest v1.Destination, parent interface{}), dest.Ref must have a Namespace
		// If Target.Ref.Namespace is nil, We will use the Namespace of Trigger as the Namespace of dest.Ref
		t.Spec.Target.Ref.Namespace = t.Namespace
	}

	targetURI, err := r.uriResolver.URIFromDestinationV1(ctx, t.Spec.Target, t)
	if err != nil {
		logging.FromContext(ctx).Errorw("Unable to get the target's URI", zap.Error(err))
		t.Status.MarkTargetResolvedFailed("Unable to get the target's URI", "%v", err)
		t.Status.TargetURI = nil
		return pkgreconciler.NewEvent(corev1.EventTypeWarning, common.ReasonFailedResolveReference,
			"Failed to get target's URI: %w", err)
	}

	t.Status.TargetURI = targetURI
	t.Status.MarkTargetResolvedSucceeded()

	return nil
}

func (r *Reconciler) resolveDLS(ctx context.Context, t *eventingv1alpha1.Trigger) pkgreconciler.Event {
	if t.Spec.Delivery == nil || t.Spec.Delivery.DeadLetterSink == nil {
		t.Status.DeadLetterSinkURI = nil
		t.Status.MarkDeadLetterSinkNotConfigured()
		return nil
	}

	if t.Spec.Delivery.DeadLetterSink.Ref != nil && t.Spec.Delivery.DeadLetterSink.Ref.Namespace == "" {
		// To call URIFromDestinationV1(ctx context.Context, dest v1.Destination, parent interface{}), dest.Ref must have a Namespace
		// If Target.Ref.Namespace is nil, We will use the Namespace of Trigger as the Namespace of dest.Ref
		t.Spec.Delivery.DeadLetterSink.Ref.Namespace = t.Namespace
	}

	dlsURI, err := r.uriResolver.URIFromDestinationV1(ctx, *t.Spec.Delivery.DeadLetterSink, t)
	if err != nil {
		logging.FromContext(ctx).Errorw("Unable to get the dead letter sink's URI", zap.Error(err))
		t.Status.MarkDeadLetterSinkResolvedFailed("Unable to get the dead letter sink's URI", "%v", err)
		t.Status.TargetURI = nil
		return pkgreconciler.NewEvent(corev1.EventTypeWarning, common.ReasonFailedResolveReference,
			"Failed to get dead letter sink's URI: %w", err)
	}

	t.Status.DeadLetterSinkURI = dlsURI
	t.Status.MarkDeadLetterSinkResolvedSucceeded()

	return nil
}

func (r *Reconciler) reconcileStatusConfigMap(ctx context.Context, t *eventingv1alpha1.Trigger, b eventingv1alpha1.ReconcilableBroker) pkgreconciler.Event {
	configMapName := common.GetBrokerConfigMapName(b)

	cm, err := r.cmLister.ConfigMaps(t.Namespace).Get(configMapName)
	if err != nil {
		if apierrs.IsNotFound(err) {
			logging.FromContext(ctx).Errorf("Trigger %s/%s could not find the Status ConfigMap for the referenced broker %q", t.Namespace, t.Name, configMapName)
			t.Status.MarkStatusConfigMapFailed(common.ReasonStatusConfigMapDoesNotExist, "Status ConfigMap %q does not exist", configMapName)
			// No need to requeue, we will be notified when the status ConfigMap is created.
			return controller.NewPermanentError(err)
		}

		t.Status.MarkStatusConfigMapFailed(common.ReasonStatusConfigMapGetFailed, "Failed to get ConfigMap for broker %q : %s", configMapName, err)
		return pkgreconciler.NewEvent(corev1.EventTypeWarning, common.ReasonStatusConfigMapGetFailed,
			"Failed to get ConfigMap for broker %s: %w", configMapName, err)
	}

	cmst, ok := cm.Data[common.ConfigMapStatusKey]
	if !ok {
		errmsg := fmt.Sprintf("ConfigMap %q does not contain key %q", configMapName, common.ConfigMapStatusKey)
		t.Status.MarkStatusConfigMapFailed(common.ReasonStatusConfigMapReadFailed, errmsg)
		// No need to requeue, we will be notified when the status ConfigMap is updated.
		return controller.NewPermanentError(errors.New(errmsg))
	}

	sts := map[string]status.Status{}
	if err := json.Unmarshal([]byte(cmst), &sts); err != nil {
		errmsg := fmt.Sprintf("ConfigMap %s/%s could not be unmarshalled as a status: %v", configMapName, common.ConfigMapStatusKey, err)
		t.Status.MarkStatusConfigMapFailed(common.ReasonStatusConfigMapReadFailed, errmsg)
		// No need to requeue, we will be notified when the status ConfigMap is updated.
		return controller.NewPermanentError(errors.New(errmsg))
	}

	return r.summarizeStatus(t, sts)
}

func (r *Reconciler) summarizeStatus(t *eventingv1alpha1.Trigger, sts map[string]status.Status) pkgreconciler.Event {
	// Iterate all nodes and take note of the status for this trigger
	var temp status.SubscriptionStatusChoice
	for instance, st := range sts {
		subs, ok := st.Subscriptions[t.Name]
		if !ok {
			continue
		}

		switch subs.Status {
		case status.SubscriptionStatusFailed:
			// If one instance reports failure, consider the trigger failed.
			errmsg := fmt.Sprintf("subscription failure reported by %s", instance)
			t.Status.MarkStatusConfigMapFailed(common.ReasonStatusSubscriptionFailed, errmsg)
			return controller.NewPermanentError(errors.New(errmsg))

		case status.SubscriptionStatusComplete:
			// If one instance reports complete, consider the trigger completed.
			// Note: this is eventually consistent, some nodes might be still sending events!
			t.Status.MarkStatusConfigMapSucceeded(common.ReasonStatusSubscriptionCompleted, fmt.Sprintf("subscription failure reported by %s", instance))
			return nil

		case status.SubscriptionStatusReady:
			// Running state takes precedence over ready state.
			if temp != status.SubscriptionStatusRunning {
				temp = status.SubscriptionStatusReady
			}

		case status.SubscriptionStatusRunning:
			if temp != status.SubscriptionStatusRunning {
				temp = status.SubscriptionStatusRunning
			}
		}
	}

	switch temp {
	case status.SubscriptionStatusReady:
		t.Status.MarkStatusConfigMapSucceeded(common.ReasonStatusSubscriptionReady, "subscription ready to dispatch events")

	case status.SubscriptionStatusRunning:
		t.Status.MarkStatusConfigMapSucceeded(common.ReasonStatusSubscriptionRunning, "subscription running")

	default:
		t.Status.MarkStatusConfigMapSucceeded(common.ReasonStatusSubscriptionUnknown, "no subscription status information")
	}

	return nil
}
