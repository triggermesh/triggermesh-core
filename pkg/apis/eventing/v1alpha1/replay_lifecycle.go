// Copyright 2023 ReplayMesh Inc.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	"knative.dev/pkg/kmeta"
)

var replayCondSet = apis.NewLivingConditionSet(ReplayConditionBroker, ReplayConditionTargetResolved)

const (
	// ReplayConditionReady has status True when all subconditions below have been set to True.
	ReplayConditionReady = apis.ConditionReady

	ReplayConditionBroker apis.ConditionType = "BrokerReady"

	ReplayConditionTargetResolved apis.ConditionType = "TargetResolved"

	// ReplayAnyFilter Constant to represent that we should allow anything.
	ReplayAnyFilter = ""
)

// GetStatus retrieves the status of the Replay. Implements the KRShaped interface.
func (r *Replay) GetStatus() *duckv1.Status {
	return &r.Status.Status
}

// GetConditionSet retrieves the condition set for this resource. Implements the KRShaped interface.
func (*Replay) GetConditionSet() apis.ConditionSet {
	return replayCondSet
}

// GetGroupVersionKind returns GroupVersionKind for Replays
func (r *Replay) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("Replay")
}

// GetUntypedSpec returns the spec of the Replay.
func (r *Replay) GetUntypedSpec() interface{} {
	return r.Spec
}

// GetCondition returns the condition currently associated with the given type, or nil.
func (rs *ReplayStatus) GetCondition(t apis.ConditionType) *apis.Condition {
	return replayCondSet.Manage(rs).GetCondition(t)
}

// GetTopLevelCondition returns the top level Condition.
func (rs *ReplayStatus) GetTopLevelCondition() *apis.Condition {
	return replayCondSet.Manage(rs).GetTopLevelCondition()
}

// IsReady returns true if the resource is ready overall.
func (rs *ReplayStatus) IsReady() bool {
	return replayCondSet.Manage(rs).IsHappy()
}

// InitializeConditions sets relevant unset conditions to Unknown state.
func (rs *ReplayStatus) InitializeConditions() {
	replayCondSet.Manage(rs).InitializeConditions()
}

func (rs *ReplayStatus) PropagateBrokerCondition(bc *apis.Condition) {
	if bc == nil {
		rs.MarkBrokerNotConfigured()
		return
	}

	switch {
	case bc.Status == corev1.ConditionUnknown:
		rs.MarkBrokerUnknown(bc.Reason, bc.Message)
	case bc.Status == corev1.ConditionTrue:
		replayCondSet.Manage(rs).MarkTrue(ReplayConditionBroker)
	case bc.Status == corev1.ConditionFalse:
		rs.MarkBrokerFailed(bc.Reason, bc.Message)
	default:
		rs.MarkBrokerUnknown("BrokerUnknown", "The status of Broker is invalid: %v", bc.Status)
	}
}

func (rs *ReplayStatus) MarkBrokerFailed(reason, messageFormat string, messageA ...interface{}) {
	replayCondSet.Manage(rs).MarkFalse(ReplayConditionBroker, reason, messageFormat, messageA...)
}

func (rs *ReplayStatus) MarkBrokerUnknown(reason, messageFormat string, messageA ...interface{}) {
	replayCondSet.Manage(rs).MarkUnknown(ReplayConditionBroker, reason, messageFormat, messageA...)
}

func (rs *ReplayStatus) MarkBrokerNotConfigured() {
	replayCondSet.Manage(rs).MarkUnknown(ReplayConditionBroker,
		"BrokerNotConfigured", "Broker has not yet been reconciled.")
}

func (rs *ReplayStatus) MarkTargetResolvedSucceeded() {
	replayCondSet.Manage(rs).MarkTrue(ReplayConditionTargetResolved)
}

func (rs *ReplayStatus) MarkTargetResolvedFailed(reason, messageFormat string, messageA ...interface{}) {
	replayCondSet.Manage(rs).MarkFalse(ReplayConditionTargetResolved, reason, messageFormat, messageA...)
}

func (rs *ReplayStatus) MarkTargetResolvedUnknown(reason, messageFormat string, messageA ...interface{}) {
	replayCondSet.Manage(rs).MarkUnknown(ReplayConditionTargetResolved, reason, messageFormat, messageA...)
}

func (r *Replay) ReferencesBroker(broker kmeta.OwnerRefable) bool {
	gvk := broker.GetGroupVersionKind()

	// Require same namespace for Replay and Broker.
	if r.Spec.Broker.Namespace != "" &&
		r.Spec.Broker.Namespace != broker.GetObjectMeta().GetNamespace() {
		return false
	}

	// If APIVersion is informed it should match the Broker's.
	if r.Spec.Broker.APIVersion != "" {
		if r.Spec.Broker.APIVersion != gvk.GroupVersion().String() {
			return false
		}
	} else if r.Spec.Broker.Group != gvk.Group {
		return false
	}

	return r.Spec.Broker.Name == broker.GetObjectMeta().GetName() &&
		r.Spec.Broker.Kind == gvk.Kind
}
