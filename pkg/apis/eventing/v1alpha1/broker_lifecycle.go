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
	BrokerConditionReady                          = apis.ConditionReady
	BrokerConditionIngress     apis.ConditionType = "IngressReady"
	BrokerConditionAddressable apis.ConditionType = "Addressable"
)

var brokerCondSet = apis.NewLivingConditionSet(
	BrokerConditionIngress,
	BrokerConditionAddressable,
)
var brokerCondSetLock = sync.RWMutex{}

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
	brokerCondSetLock.Lock()
	defer brokerCondSetLock.Unlock()

	brokerCondSet = conditionSet
}

// GetConditionSet retrieves the condition set for this resource. Implements the KRShaped interface.
func (b *RedisBroker) GetConditionSet() apis.ConditionSet {
	brokerCondSetLock.RLock()
	defer brokerCondSetLock.RUnlock()

	return brokerCondSet
}

// GetConditionSet retrieves the condition set for this resource.
func (bs *RedisBrokerStatus) GetConditionSet() apis.ConditionSet {
	brokerCondSetLock.RLock()
	defer brokerCondSetLock.RUnlock()

	return brokerCondSet
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
		bs.GetConditionSet().Manage(bs).MarkTrue(BrokerConditionAddressable)
	} else {
		bs.GetConditionSet().Manage(bs).MarkFalse(BrokerConditionAddressable, "nil URL", "URL is nil")
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
