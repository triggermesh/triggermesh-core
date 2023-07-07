// Copyright 2022 TriggerMesh Inc.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"knative.dev/pkg/apis"

	duckv1 "knative.dev/pkg/apis/duck/v1"
	"knative.dev/pkg/kmeta"
)

var triggerCondSet = apis.NewLivingConditionSet(TriggerConditionBroker, TriggerConditionTargetResolved, TriggerConditionDeadLetterSinkResolved)

const (
	// TriggerConditionReady has status True when all subconditions below have been set to True.
	TriggerConditionReady = apis.ConditionReady

	TriggerConditionBroker apis.ConditionType = "BrokerReady"

	TriggerConditionTargetResolved apis.ConditionType = "TargetResolved"

	TriggerConditionDeadLetterSinkResolved apis.ConditionType = "DeadLetterSinkResolved"

	// TriggerAnyFilter Constant to represent that we should allow anything.
	TriggerAnyFilter = ""
)

// GetStatus retrieves the status of the Trigger. Implements the KRShaped interface.
func (t *Trigger) GetStatus() *duckv1.Status {
	return &t.Status.Status
}

// GetConditionSet retrieves the condition set for this resource. Implements the KRShaped interface.
func (*Trigger) GetConditionSet() apis.ConditionSet {
	return triggerCondSet
}

// GetGroupVersionKind returns GroupVersionKind for Triggers
func (t *Trigger) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("Trigger")
}

// GetUntypedSpec returns the spec of the Trigger.
func (t *Trigger) GetUntypedSpec() interface{} {
	return t.Spec
}

// GetCondition returns the condition currently associated with the given type, or nil.
func (ts *TriggerStatus) GetCondition(t apis.ConditionType) *apis.Condition {
	return triggerCondSet.Manage(ts).GetCondition(t)
}

// GetTopLevelCondition returns the top level Condition.
func (ts *TriggerStatus) GetTopLevelCondition() *apis.Condition {
	return triggerCondSet.Manage(ts).GetTopLevelCondition()
}

// IsReady returns true if the resource is ready overall.
func (ts *TriggerStatus) IsReady() bool {
	return triggerCondSet.Manage(ts).IsHappy()
}

// InitializeConditions sets relevant unset conditions to Unknown state.
func (ts *TriggerStatus) InitializeConditions() {
	triggerCondSet.Manage(ts).InitializeConditions()
}

func (ts *TriggerStatus) PropagateBrokerCondition(bc *apis.Condition) {
	if bc == nil {
		ts.MarkBrokerNotConfigured()
		return
	}

	switch {
	case bc.Status == corev1.ConditionUnknown:
		ts.MarkBrokerUnknown(bc.Reason, bc.Message)
	case bc.Status == corev1.ConditionTrue:
		triggerCondSet.Manage(ts).MarkTrue(TriggerConditionBroker)
	case bc.Status == corev1.ConditionFalse:
		ts.MarkBrokerFailed(bc.Reason, bc.Message)
	default:
		ts.MarkBrokerUnknown("BrokerUnknown", "The status of Broker is invalid: %v", bc.Status)
	}
}

func (ts *TriggerStatus) MarkBrokerFailed(reason, messageFormat string, messageA ...interface{}) {
	triggerCondSet.Manage(ts).MarkFalse(TriggerConditionBroker, reason, messageFormat, messageA...)
}

func (ts *TriggerStatus) MarkBrokerUnknown(reason, messageFormat string, messageA ...interface{}) {
	triggerCondSet.Manage(ts).MarkUnknown(TriggerConditionBroker, reason, messageFormat, messageA...)
}

func (ts *TriggerStatus) MarkBrokerNotConfigured() {
	triggerCondSet.Manage(ts).MarkUnknown(TriggerConditionBroker,
		"BrokerNotConfigured", "Broker has not yet been reconciled.")
}

func (ts *TriggerStatus) MarkTargetResolvedSucceeded() {
	triggerCondSet.Manage(ts).MarkTrue(TriggerConditionTargetResolved)
}

func (ts *TriggerStatus) MarkTargetResolvedFailed(reason, messageFormat string, messageA ...interface{}) {
	triggerCondSet.Manage(ts).MarkFalse(TriggerConditionTargetResolved, reason, messageFormat, messageA...)
}

func (ts *TriggerStatus) MarkTargetResolvedUnknown(reason, messageFormat string, messageA ...interface{}) {
	triggerCondSet.Manage(ts).MarkUnknown(TriggerConditionTargetResolved, reason, messageFormat, messageA...)
}

func (ts *TriggerStatus) MarkDeadLetterSinkResolvedSucceeded() {
	triggerCondSet.Manage(ts).MarkTrue(TriggerConditionDeadLetterSinkResolved)
}

func (ts *TriggerStatus) MarkDeadLetterSinkNotConfigured() {
	triggerCondSet.Manage(ts).MarkTrueWithReason(TriggerConditionDeadLetterSinkResolved, "DeadLetterSinkNotConfigured", "No dead letter sink is configured.")
}

func (ts *TriggerStatus) MarkDeadLetterSinkResolvedFailed(reason, messageFormat string, messageA ...interface{}) {
	triggerCondSet.Manage(ts).MarkFalse(TriggerConditionDeadLetterSinkResolved, reason, messageFormat, messageA...)
}

func (t *Trigger) OwnerRefableMatchesBroker(broker kmeta.OwnerRefable) bool {
	gvk := broker.GetGroupVersionKind()

	// Require same namespace for Trigger and Broker.
	if t.Spec.Broker.Namespace != "" &&
		t.Spec.Broker.Namespace != broker.GetObjectMeta().GetNamespace() {
		return false
	}

	// If APIVersion is informed it should match the Broker's.
	if t.Spec.Broker.APIVersion != "" {
		if t.Spec.Broker.APIVersion != gvk.GroupVersion().String() {
			return false
		}
	} else if t.Spec.Broker.Group != gvk.Group {
		return false
	}

	return t.Spec.Broker.Name == broker.GetObjectMeta().GetName() &&
		t.Spec.Broker.Kind == gvk.Kind
}

func (t *Trigger) OwnerReferenceMatchesBroker(broker metav1.OwnerReference) bool {
	if t.Spec.Broker.APIVersion != "" && t.Spec.Broker.APIVersion != broker.APIVersion {
		return false
	}

	if t.Spec.Broker.Group != "" {
		gv, err := schema.ParseGroupVersion(broker.APIVersion)
		if err != nil {
			return false
		}
		if t.Spec.Broker.Group != gv.Group {
			return false
		}
	}

	return t.Spec.Broker.Name == broker.Name &&
		t.Spec.Broker.Kind == broker.Kind
}
