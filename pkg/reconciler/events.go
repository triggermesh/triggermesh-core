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
	ReasonDeploymentGet          = "GetDeployment"
	ReasonDeploymentUpdate       = "UpdateDeployment"
	ReasonFailedDeploymentCreate = "FailedDeploymentCreate"
	ReasonFailedDeploymentUpdate = "FailedDeploymentUpdate"

	// // ReasonServiceCreate indicates that an Service object was successfully created.
	// ReasonServiceCreate = "CreateService"
	// // ReasonServiceUpdate indicates that an Service object was successfully updated.
	// ReasonServiceUpdate = "UpdateService"
	// // ReasonFailedServiceCreate indicates that the creation of an Service object failed.
	// ReasonFailedServiceCreate = "FailedServiceCreate"
	// // ReasonFailedServiceUpdate indicates that the update of an Service object failed.
	// ReasonFailedServiceUpdate = "FailedServiceUpdate"

	// // ReasonBadSinkURI indicates that the URI of a sink can't be determined.
	// ReasonBadSinkURI = "BadSinkURI"

	// // ReasonInvalidSpec indicates that spec of a reconciled object is invalid.
	// ReasonInvalidSpec = "InvalidSpec"
)
