package v1alpha1

import (
	"knative.dev/pkg/kmeta"
)

type ReconcilableBroker interface {
	kmeta.OwnerRefable

	GetReconcilableBrokerStatus() ReconcilableBrokerStatus
	GetOwnedObjectsSuffix() string
}

type ReconcilableBrokerStatus interface {
	// Secret status management.
	MarkConfigSecretFailed(reason, messageFormat string, messageA ...interface{})
	MarkConfigSecretReady()

	// ServiceAccount status management.
	MarkBrokerServiceAccountFailed(reason, messageFormat string, messageA ...interface{})
	MarkBrokerServiceAccountReady()

	// RoleBinding status management.
	MarkBrokerRoleBindingFailed(reason, messageFormat string, messageA ...interface{})
	MarkBrokerRoleBindingReady()
}
