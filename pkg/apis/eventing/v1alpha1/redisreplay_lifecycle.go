// Copyright 2023 TriggerMesh Inc.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
)

// ReplayReplay refers to the Replay instance backing the replay adapter.

const (
	ReplayConditionReady = apis.ConditionReady
	ReplayConditionOK    = "OK"
	ReplayConditionError = "Error"
)

var replayCondSet = apis.NewLivingConditionSet(
	ReplayConditionOK,
	ReplayConditionError,
)

// GetGroupVersionKind returns GroupVersionKind for Brokers
func (t *Replay) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("Replay")
}

// GetStatus retrieves the status of the Broker. Implements the KRShaped interface.
func (t *Replay) GetStatus() *duckv1.Status {
	return &t.Status.Status
}

// GetConditionSet retrieves the condition set for this resource. Implements the KRShaped interface.
func (*Replay) GetConditionSet() apis.ConditionSet {
	return replayCondSet
}

// GetCondition returns the condition currently associated with the given type, or nil.
func (s *ReplayStatus) GetCondition(t apis.ConditionType) *apis.Condition {
	return replayCondSet.Manage(s).GetCondition(t)
}

// GetTopLevelCondition returns the top level Condition.
func (s *ReplayStatus) GetTopLevelCondition() *apis.Condition {
	return replayCondSet.Manage(s).GetTopLevelCondition()
}

// IsReady returns true if the resource is ready overall.
func (s *ReplayStatus) IsReady() bool {
	return replayCondSet.Manage(s).IsHappy()
}

// IsOk returns true if the resource is ready overall.
func (s *ReplayStatus) IsOk() bool {
	return replayCondSet.Manage(s).GetCondition(ReplayConditionOK).IsTrue()
}

// MarkOk sets the condition that the resource is ready to true.
func (s *ReplayStatus) MarkOk() {
	replayCondSet.Manage(s).MarkTrue(ReplayConditionOK)
}

// MarkError sets the condition that the resource is ready to false with the given reason and message.
func (s *ReplayStatus) MarkError(reason, message string) {
	replayCondSet.Manage(s).MarkFalse(ReplayConditionError, reason, message)
}

// IsError returns true if the resource is ready overall.
func (s *ReplayStatus) IsError() bool {
	return replayCondSet.Manage(s).GetCondition(ReplayConditionError).IsTrue()
}

// InitializeConditions sets relevant unset conditions to Unknown state.
func (s *ReplayStatus) InitializeConditions() {
	s.GetConditionSet().Manage(s).InitializeConditions()
}

func (*ReplayStatus) GetConditionSet() apis.ConditionSet {
	return replayCondSet
}

// MarkCondition sets the condition of the specified type, status, severity, reason, and message.
func (s *ReplayStatus) MarkCondition(conditionType apis.ConditionType, status v1.ConditionStatus, severity apis.ConditionSeverity, reason, message string) {
	replayCondSet.Manage(s).SetCondition(apis.Condition{
		Type:     conditionType,
		Status:   status,
		Severity: severity,
		Reason:   reason,
		Message:  message,
	})
}
