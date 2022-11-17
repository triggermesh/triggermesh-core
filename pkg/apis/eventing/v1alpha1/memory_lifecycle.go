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

// MemoryBrokerMemory refers to the Memory instance backing the broker.
// MemoryBrokerBroker refers to the TriggerMesh Broker that manages events on top of Memory.

const (
	MemoryBrokerConditionReady                                          = apis.ConditionReady
	MemoryBrokerBrokerDeployment                     apis.ConditionType = "BrokerDeploymentReady"
	MemoryBrokerBrokerServiceAccount                 apis.ConditionType = "BrokerServiceAccountReady"
	MemoryBrokerBrokerRoleBinding                    apis.ConditionType = "MemoryBrokerBrokerRoleBinding"
	MemoryBrokerBrokerService                        apis.ConditionType = "BrokerServiceReady"
	MemoryBrokerBrokerServiceEndpointsConditionReady apis.ConditionType = "BrokerEndpointsReady"
	MemoryBrokerConfigSecret                         apis.ConditionType = "BrokerConfigSecretReady"
	MemoryBrokerConditionAddressable                 apis.ConditionType = "Addressable"
)

var memoryBrokerCondSet = apis.NewLivingConditionSet(
	MemoryBrokerBrokerServiceAccount,
	MemoryBrokerBrokerRoleBinding,
	MemoryBrokerBrokerDeployment,
	MemoryBrokerBrokerService,
	MemoryBrokerBrokerServiceEndpointsConditionReady,
	MemoryBrokerConfigSecret,
	MemoryBrokerConditionAddressable,
)
var memoryBrokerCondSetLock = sync.RWMutex{}

// GetGroupVersionKind returns GroupVersionKind for Brokers
func (t *MemoryBroker) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("MemoryBroker")
}

// GetStatus retrieves the status of the Broker. Implements the KRShaped interface.
func (t *MemoryBroker) GetStatus() *duckv1.Status {
	return &t.Status.Status
}

// GetConditionSet retrieves the condition set for this resource. Implements the KRShaped interface.
func (b *MemoryBroker) GetConditionSet() apis.ConditionSet {
	memoryBrokerCondSetLock.RLock()
	defer memoryBrokerCondSetLock.RUnlock()

	return memoryBrokerCondSet
}

// GetConditionSet retrieves the condition set for this resource.
func (bs *MemoryBrokerStatus) GetConditionSet() apis.ConditionSet {
	memoryBrokerCondSetLock.RLock()
	defer memoryBrokerCondSetLock.RUnlock()

	return memoryBrokerCondSet
}

// GetTopLevelCondition returns the top level Condition.
func (bs *MemoryBrokerStatus) GetTopLevelCondition() *apis.Condition {
	return bs.GetConditionSet().Manage(bs).GetTopLevelCondition()
}

// SetAddress makes this Broker addressable by setting the URI. It also
// sets the BrokerConditionAddressable to true.
func (bs *MemoryBrokerStatus) SetAddress(url *apis.URL) {
	bs.Address.URL = url
	if url != nil {
		bs.GetConditionSet().Manage(bs).MarkTrue(MemoryBrokerConditionAddressable)
	} else {
		bs.GetConditionSet().Manage(bs).MarkFalse(MemoryBrokerConditionAddressable, "nil URL", "URL is nil")
	}
}

// GetCondition returns the condition currently associated with the given type, or nil.
func (bs *MemoryBrokerStatus) GetCondition(t apis.ConditionType) *apis.Condition {
	return bs.GetConditionSet().Manage(bs).GetCondition(t)
}

// IsReady returns true if the resource is ready overall and the latest spec has been observed.
func (b *MemoryBroker) IsReady() bool {
	bs := b.Status
	return bs.ObservedGeneration == b.Generation &&
		b.GetConditionSet().Manage(&bs).IsHappy()
}

// InitializeConditions sets relevant unset conditions to Unknown state.
func (bs *MemoryBrokerStatus) InitializeConditions() {
	bs.GetConditionSet().Manage(bs).InitializeConditions()
}

func (bs *MemoryBrokerStatus) MarkConfigSecretFailed(reason, messageFormat string, messageA ...interface{}) {
	memoryBrokerCondSet.Manage(bs).MarkFalse(MemoryBrokerConfigSecret, reason, messageFormat, messageA...)
}

func (bs *MemoryBrokerStatus) MarkConfigSecretUnknown(reason, messageFormat string, messageA ...interface{}) {
	memoryBrokerCondSet.Manage(bs).MarkUnknown(MemoryBrokerConfigSecret, reason, messageFormat, messageA...)
}

func (bs *MemoryBrokerStatus) MarkConfigSecretReady() {
	memoryBrokerCondSet.Manage(bs).MarkTrue(MemoryBrokerConfigSecret)
}

// Manage Memory broker service account and role binding.

func (bs *MemoryBrokerStatus) MarkBrokerServiceAccountFailed(reason, messageFormat string, messageA ...interface{}) {
	memoryBrokerCondSet.Manage(bs).MarkFalse(MemoryBrokerBrokerServiceAccount, reason, messageFormat, messageA...)
}

func (bs *MemoryBrokerStatus) MarkBrokerServiceAccountUnknown(reason, messageFormat string, messageA ...interface{}) {
	memoryBrokerCondSet.Manage(bs).MarkUnknown(MemoryBrokerBrokerServiceAccount, reason, messageFormat, messageA...)
}

func (bs *MemoryBrokerStatus) MarkBrokerServiceAccountReady() {
	memoryBrokerCondSet.Manage(bs).MarkTrue(MemoryBrokerBrokerServiceAccount)
}

func (bs *MemoryBrokerStatus) MarkBrokerRoleBindingFailed(reason, messageFormat string, messageA ...interface{}) {
	memoryBrokerCondSet.Manage(bs).MarkFalse(MemoryBrokerBrokerRoleBinding, reason, messageFormat, messageA...)
}

func (bs *MemoryBrokerStatus) MarkBrokerRoleBindingUnknown(reason, messageFormat string, messageA ...interface{}) {
	memoryBrokerCondSet.Manage(bs).MarkUnknown(MemoryBrokerBrokerRoleBinding, reason, messageFormat, messageA...)
}

func (bs *MemoryBrokerStatus) MarkBrokerRoleBindingReady() {
	memoryBrokerCondSet.Manage(bs).MarkTrue(MemoryBrokerBrokerRoleBinding)
}

// Manage Memory broker state for
// Deployment, Service and Endpoint

func (bs *MemoryBrokerStatus) MarkBrokerDeploymentFailed(reason, messageFormat string, messageA ...interface{}) {
	memoryBrokerCondSet.Manage(bs).MarkFalse(MemoryBrokerBrokerDeployment, reason, messageFormat, messageA...)
}

func (bs *MemoryBrokerStatus) MarkBrokerDeploymentUnknown(reason, messageFormat string, messageA ...interface{}) {
	memoryBrokerCondSet.Manage(bs).MarkUnknown(MemoryBrokerBrokerDeployment, reason, messageFormat, messageA...)
}

func (bs *MemoryBrokerStatus) PropagateBrokerDeploymentAvailability(ctx context.Context, ds *appsv1.DeploymentStatus) {
	for _, cond := range ds.Conditions {

		if cond.Type == appsv1.DeploymentAvailable {
			switch cond.Status {
			case corev1.ConditionTrue:
				memoryBrokerCondSet.Manage(bs).MarkTrue(MemoryBrokerBrokerDeployment)
			case corev1.ConditionFalse:
				bs.MarkBrokerDeploymentFailed("BrokerDeploymentFalse", "The status of Broker Deployment is False: %s : %s", cond.Reason, cond.Message)
			default:
				// expected corev1.ConditionUnknown
				bs.MarkBrokerDeploymentUnknown("BrokerDeploymentUnknown", "The status of Broker Deployment is Unknown: %s : %s", cond.Reason, cond.Message)
			}
		}
	}
}

func (bs *MemoryBrokerStatus) MarkBrokerServiceFailed(reason, messageFormat string, messageA ...interface{}) {
	memoryBrokerCondSet.Manage(bs).MarkFalse(MemoryBrokerBrokerService, reason, messageFormat, messageA...)
}

func (bs *MemoryBrokerStatus) MarkBrokerServiceUnknown(reason, messageFormat string, messageA ...interface{}) {
	memoryBrokerCondSet.Manage(bs).MarkUnknown(MemoryBrokerBrokerService, reason, messageFormat, messageA...)
}

func (bs *MemoryBrokerStatus) MarkBrokerServiceReady() {
	memoryBrokerCondSet.Manage(bs).MarkTrue(MemoryBrokerBrokerService)
}

func (bs *MemoryBrokerStatus) MarkBrokerEndpointsFailed(reason, messageFormat string, messageA ...interface{}) {
	memoryBrokerCondSet.Manage(bs).MarkFalse(MemoryBrokerBrokerServiceEndpointsConditionReady, reason, messageFormat, messageA...)
}

func (bs *MemoryBrokerStatus) MarkBrokerEndpointsUnknown(reason, messageFormat string, messageA ...interface{}) {
	memoryBrokerCondSet.Manage(bs).MarkUnknown(MemoryBrokerBrokerServiceEndpointsConditionReady, reason, messageFormat, messageA...)
}

func (bs *MemoryBrokerStatus) MarkBrokerEndpointsTrue() {
	memoryBrokerCondSet.Manage(bs).MarkTrue(MemoryBrokerBrokerServiceEndpointsConditionReady)
}
