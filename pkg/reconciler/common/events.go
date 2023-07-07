// Copyright 2022 TriggerMesh Inc.
// SPDX-License-Identifier: Apache-2.0

package common

// Reasons for API Events
const (
	ReasonDeploymentCreate       = "CreateDeployment"
	ReasonDeploymentUpdate       = "UpdateDeployment"
	ReasonFailedDeploymentGet    = "FailedDeploymentGet"
	ReasonFailedDeploymentCreate = "FailedDeploymentCreate"
	ReasonFailedDeploymentUpdate = "FailedDeploymentUpdate"

	ReasonFailedSecretCompose = "FailedSecretCompose"
	ReasonFailedSecretGet     = "FailedSecretGet"
	ReasonFailedSecretCreate  = "FailedSecretCreate"
	ReasonFailedSecretUpdate  = "FailedSecretUpdate"

	ReasonFailedStatusConfigMapCompose = "FailedConfigMapCompose"
	ReasonFailedStatusConfigMapGet     = "FailedConfigMapGet"
	ReasonFailedStatusConfigMapCreate  = "FailedConfigMapCreate"
	ReasonFailedStatusConfigMapUpdate  = "FailedConfigMapUpdate"

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
