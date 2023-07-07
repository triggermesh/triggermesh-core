// Copyright 2022 TriggerMesh Inc.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"context"
	"sync"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
)

// RedisBrokerRedis refers to the Redis instance backing the broker.
// RedisBrokerBroker refers to the TriggerMesh Broker that manages events on top of Redis.

const (
	RedisBrokerConditionReady                                          = apis.ConditionReady
	RedisBrokerRedisDeployment                      apis.ConditionType = "RedisDeploymentReady"
	RedisBrokerRedisService                         apis.ConditionType = "RedisServiceReady"
	RedisBrokerRedisServiceEndpointsConditionReady  apis.ConditionType = "RedisEndpointsReady"
	RedisBrokerBrokerDeployment                     apis.ConditionType = "BrokerDeploymentReady"
	RedisBrokerBrokerServiceAccount                 apis.ConditionType = "BrokerServiceAccountReady"
	RedisBrokerBrokerRoleBinding                    apis.ConditionType = "RedisBrokerBrokerRoleBinding"
	RedisBrokerBrokerService                        apis.ConditionType = "BrokerServiceReady"
	RedisBrokerBrokerServiceEndpointsConditionReady apis.ConditionType = "BrokerEndpointsReady"
	RedisBrokerConfigSecret                         apis.ConditionType = "BrokerConfigSecretReady"
	RedisBrokerConditionAddressable                 apis.ConditionType = "Addressable"
	RedisBrokerStatusConfig                         apis.ConditionType = "BrokerStatusConfigReady"

	RedisBrokerReasonUserProvided string = "ReasonUserProvidedRedis"
)

var redisBrokerCondSet = apis.NewLivingConditionSet(
	RedisBrokerRedisDeployment,
	RedisBrokerRedisService,
	RedisBrokerRedisServiceEndpointsConditionReady,
	RedisBrokerBrokerServiceAccount,
	RedisBrokerBrokerRoleBinding,
	RedisBrokerBrokerDeployment,
	RedisBrokerBrokerService,
	RedisBrokerBrokerServiceEndpointsConditionReady,
	RedisBrokerConfigSecret,
	RedisBrokerConditionAddressable,
	RedisBrokerStatusConfig,
)
var redisBrokerCondSetLock = sync.RWMutex{}

// GetGroupVersionKind returns GroupVersionKind for Brokers
func (t *RedisBroker) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("RedisBroker")
}

// GetStatus retrieves the status of the Broker. Implements the KRShaped interface.
func (t *RedisBroker) GetStatus() *duckv1.Status {
	return &t.Status.Status
}

// GetReconcilableBrokerSpec returns the all brokers common Broker spec.
func (t *RedisBroker) GetReconcilableBrokerSpec() *Broker {
	return &t.Spec.Broker
}

// GetReconcilableBrokerStatus returns a status interface that allows generic reconciler
// to manage it.
func (t *RedisBroker) GetReconcilableBrokerStatus() ReconcilableBrokerStatus {
	return &t.Status
}

// GetOwnedObjectsSuffix returns a string to be appended for created/owned objects.
func (t *RedisBroker) GetOwnedObjectsSuffix() string {
	return "rb"
}

// GetConditionSet retrieves the condition set for this resource. Implements the KRShaped interface.
func (b *RedisBroker) GetConditionSet() apis.ConditionSet {
	redisBrokerCondSetLock.RLock()
	defer redisBrokerCondSetLock.RUnlock()

	return redisBrokerCondSet
}

// IsExternalRedis returns if the Redis instance is user provided.
func (b *RedisBroker) IsUserProvidedRedis() bool {
	if b.Spec.Redis != nil && b.Spec.Redis.Connection != nil {
		return true
	}
	return false
}

// GetConditionSet retrieves the condition set for this resource.
func (bs *RedisBrokerStatus) GetConditionSet() apis.ConditionSet {
	redisBrokerCondSetLock.RLock()
	defer redisBrokerCondSetLock.RUnlock()

	return redisBrokerCondSet
}

// GetTopLevelCondition returns the top level Condition.
func (bs *RedisBrokerStatus) GetTopLevelCondition() *apis.Condition {
	return bs.GetConditionSet().Manage(bs).GetTopLevelCondition()
}

// SetAddress makes this Broker addressable by setting the URI. It also
// sets the BrokerConditionAddressable to true.
func (bs *RedisBrokerStatus) SetAddress(url *apis.URL) {
	bs.Address.URL = url
	if url != nil {
		bs.GetConditionSet().Manage(bs).MarkTrue(RedisBrokerConditionAddressable)
	} else {
		bs.GetConditionSet().Manage(bs).MarkFalse(RedisBrokerConditionAddressable, "nil URL", "URL is nil")
	}
}

// GetCondition returns the condition currently associated with the given type, or nil.
func (bs *RedisBrokerStatus) GetCondition(t apis.ConditionType) *apis.Condition {
	return bs.GetConditionSet().Manage(bs).GetCondition(t)
}

// IsReady returns true if the resource is ready overall and the latest spec has been observed.
func (b *RedisBroker) IsReady() bool {
	bs := b.Status
	return bs.ObservedGeneration == b.Generation &&
		b.GetConditionSet().Manage(&bs).IsHappy()
}

// InitializeConditions sets relevant unset conditions to Unknown state.
func (bs *RedisBrokerStatus) InitializeConditions() {
	bs.GetConditionSet().Manage(bs).InitializeConditions()
}

func (bs *RedisBrokerStatus) MarkConfigSecretFailed(reason, messageFormat string, messageA ...interface{}) {
	redisBrokerCondSet.Manage(bs).MarkFalse(RedisBrokerConfigSecret, reason, messageFormat, messageA...)
}

func (bs *RedisBrokerStatus) MarkConfigSecretUnknown(reason, messageFormat string, messageA ...interface{}) {
	redisBrokerCondSet.Manage(bs).MarkUnknown(RedisBrokerConfigSecret, reason, messageFormat, messageA...)
}

func (bs *RedisBrokerStatus) MarkConfigSecretReady() {
	redisBrokerCondSet.Manage(bs).MarkTrue(RedisBrokerConfigSecret)
}

func (bs *RedisBrokerStatus) MarkStatusConfigFailed(reason, messageFormat string, messageA ...interface{}) {
	redisBrokerCondSet.Manage(bs).MarkFalse(RedisBrokerStatusConfig, reason, messageFormat, messageA...)
}

func (bs *RedisBrokerStatus) MarkStatusConfigUnknown(reason, messageFormat string, messageA ...interface{}) {
	redisBrokerCondSet.Manage(bs).MarkUnknown(RedisBrokerStatusConfig, reason, messageFormat, messageA...)
}

func (bs *RedisBrokerStatus) MarkStatusConfigReady() {
	redisBrokerCondSet.Manage(bs).MarkTrue(RedisBrokerStatusConfig)
}

// Manage Broker's service account and rolebinding.

func (bs *RedisBrokerStatus) MarkBrokerServiceAccountFailed(reason, messageFormat string, messageA ...interface{}) {
	redisBrokerCondSet.Manage(bs).MarkFalse(RedisBrokerBrokerServiceAccount, reason, messageFormat, messageA...)
}

func (bs *RedisBrokerStatus) MarkBrokerServiceAccountUnknown(reason, messageFormat string, messageA ...interface{}) {
	redisBrokerCondSet.Manage(bs).MarkUnknown(RedisBrokerBrokerServiceAccount, reason, messageFormat, messageA...)
}

func (bs *RedisBrokerStatus) MarkBrokerServiceAccountReady() {
	redisBrokerCondSet.Manage(bs).MarkTrue(RedisBrokerBrokerServiceAccount)
}

func (bs *RedisBrokerStatus) MarkBrokerRoleBindingFailed(reason, messageFormat string, messageA ...interface{}) {
	redisBrokerCondSet.Manage(bs).MarkFalse(RedisBrokerBrokerRoleBinding, reason, messageFormat, messageA...)
}

func (bs *RedisBrokerStatus) MarkBrokerRoleBindingUnknown(reason, messageFormat string, messageA ...interface{}) {
	redisBrokerCondSet.Manage(bs).MarkUnknown(RedisBrokerBrokerRoleBinding, reason, messageFormat, messageA...)
}

func (bs *RedisBrokerStatus) MarkBrokerRoleBindingReady() {
	redisBrokerCondSet.Manage(bs).MarkTrue(RedisBrokerBrokerRoleBinding)
}

// Manage Redis server state for both
// Service and Deployment

func (bs *RedisBrokerStatus) MarkRedisDeploymentFailed(reason, messageFormat string, messageA ...interface{}) {
	redisBrokerCondSet.Manage(bs).MarkFalse(RedisBrokerRedisDeployment, reason, messageFormat, messageA...)
}

func (bs *RedisBrokerStatus) MarkRedisDeploymentUnknown(reason, messageFormat string, messageA ...interface{}) {
	redisBrokerCondSet.Manage(bs).MarkUnknown(RedisBrokerRedisDeployment, reason, messageFormat, messageA...)
}

func (bs *RedisBrokerStatus) PropagateRedisDeploymentAvailability(ctx context.Context, ds *appsv1.DeploymentStatus) {
	for _, cond := range ds.Conditions {

		if cond.Type == appsv1.DeploymentAvailable {
			switch cond.Status {
			case corev1.ConditionTrue:
				redisBrokerCondSet.Manage(bs).MarkTrue(RedisBrokerRedisDeployment)
			case corev1.ConditionFalse:
				bs.MarkRedisDeploymentFailed("RedisDeploymentFalse", "The status of Redis Deployment is False: %s : %s", cond.Reason, cond.Message)
			default:
				// expected corev1.ConditionUnknown
				bs.MarkRedisDeploymentUnknown("RedisDeploymentUnknown", "The status of Redis Deployment is Unknown: %s : %s", cond.Reason, cond.Message)
			}
		}
	}
}

func (bs *RedisBrokerStatus) MarkRedisServiceFailed(reason, messageFormat string, messageA ...interface{}) {
	redisBrokerCondSet.Manage(bs).MarkFalse(RedisBrokerRedisService, reason, messageFormat, messageA...)
}

func (bs *RedisBrokerStatus) MarkRedisServiceUnknown(reason, messageFormat string, messageA ...interface{}) {
	redisBrokerCondSet.Manage(bs).MarkUnknown(RedisBrokerRedisService, reason, messageFormat, messageA...)
}

func (bs *RedisBrokerStatus) MarkRedisServiceReady() {
	redisBrokerCondSet.Manage(bs).MarkTrue(RedisBrokerRedisService)
}

func (bs *RedisBrokerStatus) MarkRedisEndpointsFailed(reason, messageFormat string, messageA ...interface{}) {
	redisBrokerCondSet.Manage(bs).MarkFalse(RedisBrokerRedisServiceEndpointsConditionReady, reason, messageFormat, messageA...)
}

func (bs *RedisBrokerStatus) MarkRedisEndpointsUnknown(reason, messageFormat string, messageA ...interface{}) {
	redisBrokerCondSet.Manage(bs).MarkUnknown(RedisBrokerRedisServiceEndpointsConditionReady, reason, messageFormat, messageA...)
}

func (bs *RedisBrokerStatus) MarkRedisEndpointsTrue() {
	redisBrokerCondSet.Manage(bs).MarkTrue(RedisBrokerRedisServiceEndpointsConditionReady)
}

// Manage Redis broker state for
// Deployment, Service and Endpoint
func (bs *RedisBrokerStatus) MarkBrokerDeploymentFailed(reason, messageFormat string, messageA ...interface{}) {
	redisBrokerCondSet.Manage(bs).MarkFalse(RedisBrokerBrokerDeployment, reason, messageFormat, messageA...)
}

func (bs *RedisBrokerStatus) MarkBrokerDeploymentUnknown(reason, messageFormat string, messageA ...interface{}) {
	redisBrokerCondSet.Manage(bs).MarkUnknown(RedisBrokerBrokerDeployment, reason, messageFormat, messageA...)
}

func (bs *RedisBrokerStatus) PropagateBrokerDeploymentAvailability(ctx context.Context, ds *appsv1.DeploymentStatus) {
	for _, cond := range ds.Conditions {

		if cond.Type == appsv1.DeploymentAvailable {
			switch cond.Status {
			case corev1.ConditionTrue:
				redisBrokerCondSet.Manage(bs).MarkTrue(RedisBrokerBrokerDeployment)
			case corev1.ConditionFalse:
				bs.MarkBrokerDeploymentFailed("BrokerDeploymentFalse", "The status of Broker Deployment is False: %s : %s", cond.Reason, cond.Message)
			default:
				// expected corev1.ConditionUnknown
				bs.MarkBrokerDeploymentUnknown("BrokerDeploymentUnknown", "The status of Broker Deployment is Unknown: %s : %s", cond.Reason, cond.Message)
			}
		}
	}
}

func (bs *RedisBrokerStatus) MarkBrokerServiceFailed(reason, messageFormat string, messageA ...interface{}) {
	redisBrokerCondSet.Manage(bs).MarkFalse(RedisBrokerBrokerService, reason, messageFormat, messageA...)
}

func (bs *RedisBrokerStatus) MarkBrokerServiceUnknown(reason, messageFormat string, messageA ...interface{}) {
	redisBrokerCondSet.Manage(bs).MarkUnknown(RedisBrokerBrokerService, reason, messageFormat, messageA...)
}

func (bs *RedisBrokerStatus) MarkBrokerServiceReady() {
	redisBrokerCondSet.Manage(bs).MarkTrue(RedisBrokerBrokerService)
}

func (bs *RedisBrokerStatus) MarkBrokerEndpointsFailed(reason, messageFormat string, messageA ...interface{}) {
	redisBrokerCondSet.Manage(bs).MarkFalse(RedisBrokerBrokerServiceEndpointsConditionReady, reason, messageFormat, messageA...)
}

func (bs *RedisBrokerStatus) MarkBrokerEndpointsUnknown(reason, messageFormat string, messageA ...interface{}) {
	redisBrokerCondSet.Manage(bs).MarkUnknown(RedisBrokerBrokerServiceEndpointsConditionReady, reason, messageFormat, messageA...)
}

func (bs *RedisBrokerStatus) MarkBrokerEndpointsTrue() {
	redisBrokerCondSet.Manage(bs).MarkTrue(RedisBrokerBrokerServiceEndpointsConditionReady)
}

func (bs *RedisBrokerStatus) MarkRedisUserProvided() {
	redisBrokerCondSet.Manage(bs).MarkTrueWithReason(RedisBrokerRedisDeployment, RedisBrokerReasonUserProvided, "Redis instance is externally provided")
	redisBrokerCondSet.Manage(bs).MarkTrueWithReason(RedisBrokerRedisService, RedisBrokerReasonUserProvided, "Redis instance is externally provided")
	redisBrokerCondSet.Manage(bs).MarkTrueWithReason(RedisBrokerRedisServiceEndpointsConditionReady, RedisBrokerReasonUserProvided, "Redis instance is externally provided")
}
