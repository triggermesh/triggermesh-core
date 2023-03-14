// Copyright 2023 TriggerMesh Inc.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"sync"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
)

// RedisReplayReplay refers to the RedisReplay instance backing the replay adapter.

const (
	RedisReplayConditionReady                                          = apis.ConditionReady
	RedisReplayReplayDeployment                     apis.ConditionType = "ReplayDeploymentReady"
	RedisReplayReplayServiceAccount                 apis.ConditionType = "ReplayServiceAccountReady"
	RedisReplayReplayRoleBinding                    apis.ConditionType = "RedisReplayReplayRoleBinding"
	RedisReplayReplayService                        apis.ConditionType = "ReplayServiceReady"
	RedisReplayReplayServiceEndpointsConditionReady apis.ConditionType = "ReplayEndpointsReady"
	RedisReplayConfigSecret                         apis.ConditionType = "ReplayConfigSecretReady"
	RedisReplayConditionAddressable                 apis.ConditionType = "Addressable"
)

var redisReplayCondSet = apis.NewLivingConditionSet(
	RedisReplayReplayServiceAccount,
	RedisReplayReplayRoleBinding,
	RedisReplayReplayDeployment,
	RedisReplayReplayService,
	RedisReplayReplayServiceEndpointsConditionReady,
	RedisReplayConfigSecret,
	RedisReplayConditionAddressable,
)

var redisReplayCondSetLock = sync.RWMutex{}

// GetGroupVersionKind returns GroupVersionKind for Brokers
func (t *RedisReplay) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("RedisReplay")
}

// GetStatus retrieves the status of the Broker. Implements the KRShaped interface.
func (t *RedisReplay) GetStatus() *duckv1.Status {
	return &t.Status.Status
}

// GetReconcilableBrokerSpec returns the all brokers common Broker spec.
func (t *RedisReplay) GetReconcilableBrokerSpec() *Broker {
	return &t.Spec.Broker
}

// GetReconcilableBrokerStatus returns a status interface that allows generic reconciler
// to manage it.
func (t *RedisReplay) GetReconcilableBrokerStatus() ReconcilableBrokerStatus {
	return &t.Status
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

// InitializeConditions sets relevant unset conditions to Unknown state.
func (s *RedisReplayStatus) InitializeConditions() {
	s.GetConditionSet().Manage(s).InitializeConditions()
}

// MarkReplayServiceAccountNotOwned changes the "ReplayServiceAccountReady" condition to false to reflect that the
// Replay ServiceAccount is not owned by the RedisReplay.
func (s *RedisReplayStatus) MarkReplayServiceAccountNotOwned() {
	redisReplayCondSet.Manage(s).MarkFalse(RedisReplayReplayServiceAccount, "NotOwned",
		"There is an existing ServiceAccount %q that we do not own.", s.ServiceAccountName)
}

// MarkReplayServiceAccountFailed changes the "ReplayServiceAccountReady" condition to false to reflect that the
// Replay ServiceAccount failed to be created.
func (s *RedisReplayStatus) MarkReplayServiceAccountFailed(reason, messageFormat string, messageA ...interface{}) {
	redisReplayCondSet.Manage(s).MarkFalse(RedisReplayReplayServiceAccount, reason, messageFormat, messageA...)
}

// MarkReplayServiceAccountReady changes the "ReplayServiceAccountReady" condition to true to reflect that the
// Replay ServiceAccount was created successfully.
func (s *RedisReplayStatus) MarkReplayServiceAccountReady() {
	redisReplayCondSet.Manage(s).MarkTrue(RedisReplayReplayServiceAccount)
}

// MarkReplayRoleBindingNotOwned changes the "ReplayRoleBindingReady" condition to false to reflect that the
// Replay RoleBinding is not owned by the RedisReplay. This is a transient failure.
func (s *RedisReplayStatus) MarkReplayRoleBindingNotOwned() {
	redisReplayCondSet.Manage(s).MarkFalse(RedisReplayReplayRoleBinding, "NotOwned",
		"There is an existing RoleBinding %q that we do not own.", s.RoleBindingName)
}

// MarkReplayRoleBindingFailed changes the "ReplayRoleBindingReady" condition to false to reflect that the
// Replay RoleBinding failed to be created.
func (s *RedisReplayStatus) MarkReplayRoleBindingFailed(reason, messageFormat string, messageA ...interface{}) {
	redisReplayCondSet.Manage(s).MarkFalse(RedisReplayReplayRoleBinding, reason, messageFormat, messageA...)
}

// MarkReplayRoleBindingReady changes the "ReplayRoleBindingReady" condition to true to reflect that the
// Replay RoleBinding was created successfully.
func (s *RedisReplayStatus) MarkReplayRoleBindingReady() {
	redisReplayCondSet.Manage(s).MarkTrue(RedisReplayReplayRoleBinding)
}

// MarkReplayDeploymentNotOwned changes the "ReplayDeploymentReady" condition to false to reflect that the
// Replay Deployment is not owned by the RedisReplay. This is a transient failure.
func (s *RedisReplayStatus) MarkReplayDeploymentNotOwned() {
	redisReplayCondSet.Manage(s).MarkFalse(RedisReplayReplayDeployment, "NotOwned",
		"There is an existing Deployment %q that we do not own.", s.DeploymentName)
}

// MarkReplayDeploymentFailed changes the "ReplayDeploymentReady" condition to false to reflect that the
// Replay Deployment failed to be created.
func (s *RedisReplayStatus) MarkReplayDeploymentFailed(reason, messageFormat string, messageA ...interface{}) {
	redisReplayCondSet.Manage(s).MarkFalse(RedisReplayReplayDeployment, reason, messageFormat, messageA...)
}

// MarkReplayDeploymentReady changes the "ReplayDeploymentReady" condition to true to reflect that the
// Replay Deployment was created successfully.
func (s *RedisReplayStatus) MarkReplayDeploymentReady() {
	redisReplayCondSet.Manage(s).MarkTrue(RedisReplayReplayDeployment)
}

// MarkReplayServiceNotOwned changes the "ReplayServiceReady" condition to false to reflect that the
// Replay Service is not owned by the RedisReplay. This is a transient failure.
func (s *RedisReplayStatus) MarkReplayServiceNotOwned() {
	redisReplayCondSet.Manage(s).MarkFalse(RedisReplayReplayService, "NotOwned",
		"There is an existing Service %q that we do not own.", s.ServiceName)
}

// MarkReplayServiceFailed changes the "ReplayServiceReady" condition to false to reflect that the
// Replay Service failed to be created.
func (s *RedisReplayStatus) MarkReplayServiceFailed(reason, messageFormat string, messageA ...interface{}) {
	redisReplayCondSet.Manage(s).MarkFalse(RedisReplayReplayService, reason, messageFormat, messageA...)
}

// MarkReplayServiceReady changes the "ReplayServiceReady" condition to true to reflect that the
// Replay Service was created successfully.
func (s *RedisReplayStatus) MarkReplayServiceReady() {
	redisReplayCondSet.Manage(s).MarkTrue(RedisReplayReplayService)
}

// MarkReplayAddressNotReady changes the "ReplayAddressReady" condition to false to reflect that the
// Replay Address is not ready.
func (s *RedisReplayStatus) MarkReplayAddressNotReady(reason, messageFormat string, messageA ...interface{}) {
	redisReplayCondSet.Manage(s).MarkFalse(RedisReplayReplayAddress, reason, messageFormat, messageA...)
}

// MarkReplayAddressReady changes the "ReplayAddressReady" condition to true to reflect that the
// Replay Address is ready.
func (s *RedisReplayStatus) MarkReplayAddressReady() {
	redisReplayCondSet.Manage(s).MarkTrue(RedisReplayReplayAddress)
}

// MarkReplayAddressNotOwned changes the "ReplayAddressReady" condition to false to reflect that the
// Replay Address is not owned by the RedisReplay. This is a transient failure.
func (s *RedisReplayStatus) MarkReplayAddressNotOwned() {
	redisReplayCondSet.Manage(s).MarkFalse(RedisReplayReplayAddress, "NotOwned",
		"There is an existing Addressable %q that we do not own.", s.AddressName)
}

// MarkReplayAddressFailed changes the "ReplayAddressReady" condition to false to reflect that the
// Replay Address failed to be created.
func (s *RedisReplayStatus) MarkReplayAddressFailed(reason, messageFormat string, messageA ...interface{}) {
	redisReplayCondSet.Manage(s).MarkFalse(RedisReplayReplayAddress, reason, messageFormat, messageA...)
}

// MarkReplayAddressReady changes the "ReplayAddressReady" condition to true to reflect that the
// Replay Address is ready.
func (s *RedisReplayStatus) MarkReplayAddressReady() {
	redisReplayCondSet.Manage(s).MarkTrue(RedisReplayReplayAddress)
}

// MarkReplayAddressNotOwned changes the "ReplayAddressReady" condition to false to reflect that the
// Replay Address is not owned by the RedisReplay. This is a transient failure.
func (s *RedisReplayStatus) MarkReplayAddressNotOwned() {
	redisReplayCondSet.Manage(s).MarkFalse(RedisReplayReplayAddress, "NotOwned",
		"There is an existing Addressable %q that we do not own.", s.AddressName)
}

// MarkReplayAddressFailed changes the "ReplayAddressReady" condition to false to reflect that the
// Replay Address failed to be created.
func (s *RedisReplayStatus) MarkReplayAddressFailed(reason, messageFormat string, messageA ...interface{}) {
	redisReplayCondSet.Manage(s).MarkFalse(RedisReplayReplayAddress, reason, messageFormat, messageA...)
}

// MarkReplayAddressReady changes the "ReplayAddressReady" condition to true to reflect that the
// Replay Address is ready.
func (s *RedisReplayStatus) MarkReplayAddressReady() {
	redisReplayCondSet.Manage(s).MarkTrue(RedisReplayReplayAddress)
}

// MarkReplayAddressNotOwned changes the "ReplayAddressReady" condition to false to reflect that the
// Replay Address is not owned by the RedisReplay. This is a transient failure.
func (s *RedisReplayStatus) MarkReplayAddressNotOwned() {
	redisReplayCondSet.Manage(s).MarkFalse(RedisReplayReplayAddress, "NotOwned",
		"There is an existing Addressable %q that we do not own.", s.AddressName)
}
