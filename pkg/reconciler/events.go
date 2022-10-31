// Copyright 2022 TriggerMesh Inc.
// SPDX-License-Identifier: Apache-2.0

package reconciler

// Reasons for API Events
const (
	// // ReasonRBACCreate indicates that an RBAC object was successfully created.
	// ReasonRBACCreate = "CreateRBAC"
	// // ReasonRBACUpdate indicates that an RBAC object was successfully updated.
	// ReasonRBACUpdate = "UpdateRBAC"
	// // ReasonFailedRBACCreate indicates that the creation of an RBAC object failed.
	// ReasonFailedRBACCreate = "FailedRBACCreate"
	// // ReasonFailedRBACUpdate indicates that the update of an RBAC object failed.
	// ReasonFailedRBACUpdate = "FailedRBACUpdate"

	ReasonDeploymentCreate       = "CreateDeployment"
	ReasonDeploymentUpdate       = "UpdateDeployment"
	ReasonFailedDeploymentGet    = "FailedDeploymentGet"
	ReasonFailedDeploymentCreate = "FailedDeploymentCreate"
	ReasonFailedDeploymentUpdate = "FailedDeploymentUpdate"

	ReasonFailedServiceAccountGet    = "FailedServiceAccountGet"
	ReasonFailedServiceAccountCreate = "FailedServiceAccountCreate"
	ReasonFailedRoleBindingGet       = "FailedRoleBindingGet"
	ReasonFailedRoleBindingCreate    = "FailedRoleBindingCreate"

	ReasonServiceCreate       = "CreateService"
	ReasonServiceUpdate       = "UpdateService"
	ReasonFailedServiceGet    = "FailedServiceGet"
	ReasonFailedServiceCreate = "FailedServiceCreate"
	ReasonFailedServiceUpdate = "FailedServiceUpdate"

	ReasonFailedTriggerList     = "FailedTriggerList"
	ReasonFailedConfigSerialize = "FailedConfigSerialize"

	ReasonUnavailableEndpoints = "UnavailableEndpoints"
	ReasonFailedEndpointsGet   = "FailedEndpointsGet"

	ReasonBrokerDoesNotExist = "BrokerDoesNotExist"
	ReasonFailedBrokerGet    = "FailedBrokerGet"

	ReasonFailedResolveReference = "FailedResolveReference"
)
