// Copyright 2022 TriggerMesh Inc.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"sync"

	"k8s.io/apimachinery/pkg/runtime/schema"

	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
)

const (
	RedisBrokerConditionReady                          = apis.ConditionReady
	RedisBrokerRedisDeployment      apis.ConditionType = "RedisDeploymentReady"
	RedisBrokerRedisService         apis.ConditionType = "RedisServiceReady"
	RedisBrokerBrokerDeployment     apis.ConditionType = "BrokerDeploymentReady"
	RedisBrokerBrokerService        apis.ConditionType = "BrokerServiceReady"
	RedisBrokerConditionAddressable apis.ConditionType = "Addressable"
)

var redisBrokerCondSet = apis.NewLivingConditionSet(
	RedisBrokerRedisDeployment,
	RedisBrokerRedisService,
	RedisBrokerBrokerDeployment,
	RedisBrokerBrokerService,
	RedisBrokerConditionAddressable,
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

// RegisterAlternateBrokerConditionSet register a apis.ConditionSet for the given broker class.
func RegisterAlternateBrokerConditionSet(conditionSet apis.ConditionSet) {
	redisBrokerCondSetLock.Lock()
	defer redisBrokerCondSetLock.Unlock()

	redisBrokerCondSet = conditionSet
}

// GetConditionSet retrieves the condition set for this resource. Implements the KRShaped interface.
func (b *RedisBroker) GetConditionSet() apis.ConditionSet {
	redisBrokerCondSetLock.RLock()
	defer redisBrokerCondSetLock.RUnlock()

	return redisBrokerCondSet
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

func (bs *RedisBrokerStatus) MarkRedisBrokerFailed(reason, messageFormat string, messageA ...interface{}) {
	redisBrokerCondSet.Manage(bs).MarkFalse(RedisBrokerRedisDeployment, reason, messageFormat, messageA...)
}
