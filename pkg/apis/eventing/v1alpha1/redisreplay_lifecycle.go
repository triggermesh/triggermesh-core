// Copyright 2023 TriggerMesh Inc.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
)

// RedisReplayReplay refers to the RedisReplay instance backing the replay adapter.

const (
	RedisReplayConditionReady = apis.ConditionReady
	RedisReplayConditionOK    = "OK"
	RedisReplayConditionError = "Error"
)

var redisReplayCondSet = apis.NewLivingConditionSet(
	RedisReplayConditionOK,
	RedisReplayConditionError,
)

// GetGroupVersionKind returns GroupVersionKind for Brokers
func (t *RedisReplay) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("RedisReplay")
}

// GetStatus retrieves the status of the Broker. Implements the KRShaped interface.
func (t *RedisReplay) GetStatus() *duckv1.Status {
	return &t.Status.Status
}

// GetConditionSet retrieves the condition set for this resource. Implements the KRShaped interface.
func (*RedisReplay) GetConditionSet() apis.ConditionSet {
	return redisReplayCondSet
}

// GetCondition returns the condition currently associated with the given type, or nil.
func (s *RedisReplayStatus) GetCondition(t apis.ConditionType) *apis.Condition {
	return redisReplayCondSet.Manage(s).GetCondition(t)
}

// GetTopLevelCondition returns the top level Condition.
func (s *RedisReplayStatus) GetTopLevelCondition() *apis.Condition {
	return redisReplayCondSet.Manage(s).GetTopLevelCondition()
}

// IsReady returns true if the resource is ready overall.
func (s *RedisReplayStatus) IsReady() bool {
	return redisReplayCondSet.Manage(s).IsHappy()
}

// IsOk returns true if the resource is ready overall.
func (s *RedisReplayStatus) IsOk() bool {
	return redisReplayCondSet.Manage(s).GetCondition(RedisReplayConditionOK).IsTrue()
}

// MarkOk sets the condition that the resource is ready to true.
func (s *RedisReplayStatus) MarkOk() {
	redisReplayCondSet.Manage(s).MarkTrue(RedisReplayConditionOK)
}

// MarkError sets the condition that the resource is ready to false with the given reason and message.
func (s *RedisReplayStatus) MarkError(reason, message string) {
	redisReplayCondSet.Manage(s).MarkFalse(RedisReplayConditionError, reason, message)
}

// IsError returns true if the resource is ready overall.
func (s *RedisReplayStatus) IsError() bool {
	return redisReplayCondSet.Manage(s).GetCondition(RedisReplayConditionError).IsTrue()
}

// InitializeConditions sets relevant unset conditions to Unknown state.
func (s *RedisReplayStatus) InitializeConditions() {
	s.GetConditionSet().Manage(s).InitializeConditions()
}

func (*RedisReplayStatus) GetConditionSet() apis.ConditionSet {
	return redisReplayCondSet
}

// MarkCondition sets the condition of the specified type, status, severity, reason, and message.
func (s *RedisReplayStatus) MarkCondition(conditionType apis.ConditionType, status v1.ConditionStatus, severity apis.ConditionSeverity, reason, message string) {
	redisReplayCondSet.Manage(s).SetCondition(apis.Condition{
		Type:     conditionType,
		Status:   status,
		Severity: severity,
		Reason:   reason,
		Message:  message,
	})
}
