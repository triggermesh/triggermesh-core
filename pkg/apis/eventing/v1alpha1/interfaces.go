// Copyright 2023 TriggerMesh Inc.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	"knative.dev/pkg/kmeta"
)

type Broker struct {
	Port *int `json:"port,omitempty"`

	Observability *Observability `json:"observability,omitempty"`
}

type Observability struct {
	ValueFromConfigMap string `json:"valueFromConfigMap"`
}

type ReconcilableBroker interface {
	kmeta.OwnerRefable

	GetReconcilableBrokerStatus() ReconcilableBrokerStatus
	GetOwnedObjectsSuffix() string
	GetReconcilableBrokerSpec() *Broker
}

type ReconcilableBrokerStatus interface {
	// Secret as config status management.
	MarkConfigSecretFailed(reason, messageFormat string, messageA ...interface{})
	MarkConfigSecretReady()

	// Status Config management.
	MarkStatusConfigFailed(reason, messageFormat string, messageA ...interface{})
	MarkStatusConfigReady()

	// ServiceAccount status management.
	MarkBrokerServiceAccountFailed(reason, messageFormat string, messageA ...interface{})
	MarkBrokerServiceAccountReady()

	// RoleBinding status management.
	MarkBrokerRoleBindingFailed(reason, messageFormat string, messageA ...interface{})
	MarkBrokerRoleBindingReady()

	// Broker Deployment status management.
	MarkBrokerDeploymentFailed(reason, messageFormat string, messageA ...interface{})
	PropagateBrokerDeploymentAvailability(ctx context.Context, ds *appsv1.DeploymentStatus)

	// Broker Service status management
	MarkBrokerServiceFailed(reason, messageFormat string, messageA ...interface{})
	MarkBrokerServiceReady()

	// Broker Endpoints status management.
	MarkBrokerEndpointsTrue()
	MarkBrokerEndpointsUnknown(reason, messageFormat string, messageA ...interface{})
	MarkBrokerEndpointsFailed(reason, messageFormat string, messageA ...interface{})
}
