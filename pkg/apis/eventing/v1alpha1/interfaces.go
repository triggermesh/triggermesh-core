package v1alpha1

import (
	"knative.dev/pkg/kmeta"
)

type ReconcilableBroker interface {
	kmeta.OwnerRefable

	GetReconcilableBrokerStatus() ReconcilableBrokerStatus
	GetOwnedObjectsPrefix() string
}

type ReconcilableBrokerStatus interface {
	MarkConfigSecretFailed(reason, messageFormat string, messageA ...interface{})
	MarkConfigSecretReady()
}
