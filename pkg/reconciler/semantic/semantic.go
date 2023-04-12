// Copyright 2022 TriggerMesh Inc.
// SPDX-License-Identifier: Apache-2.0

package semantic

import (
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/conversion"
)

// Semantic can do semantic deep equality checks for Kubernetes API objects.
//
// For a given comparison function
//
//	comp(a, b interface{})
//
// 'a' should always be the desired state, and 'b' the current state for
// DeepDerivative comparisons to work as expected.
var Semantic = conversion.EqualitiesOrDie(
	deploymentEqual,
	serviceAccountEqual,
	serviceEqual,
	secretEqual,
	jobEqual,
)

// eq is an instance of Equalities for internal deep derivative comparisons
// of API objects. Adapted from "k8s.io/apimachinery/pkg/api/equality".Semantic.
var eq = conversion.EqualitiesOrDie(
	func(a, b resource.Quantity) bool {
		if a.IsZero() {
			return true
		}
		return a.Cmp(b) == 0
	},
	func(a, b metav1.Time) bool { // e.g. metadata.creationTimestamp
		if a.IsZero() {
			return true
		}
		return a.UTC() == b.UTC()
	},
	func(a, b int64) bool { // e.g. metadata.generation
		if a == 0 {
			return true
		}
		return a == b
	},
	// Needed because DeepDerivative compares int values directly, which
	// doesn't yield the expected result with defaulted int32 probe fields.
	func(a, b *corev1.Probe) bool {
		if a == nil {
			return true
		}
		if b == nil {
			return false
		}

		if a.InitialDelaySeconds != 0 && a.InitialDelaySeconds != b.InitialDelaySeconds {
			return false
		}
		if a.TimeoutSeconds != 0 && a.TimeoutSeconds != b.TimeoutSeconds {
			return false
		}
		if a.PeriodSeconds != 0 && a.PeriodSeconds != b.PeriodSeconds {
			return false
		}
		if a.SuccessThreshold != 0 && a.SuccessThreshold != b.SuccessThreshold {
			return false
		}
		if a.FailureThreshold != 0 && a.FailureThreshold != b.FailureThreshold {
			return false
		}

		return (conversion.Equalities{}).DeepDerivative(a.ProbeHandler, b.ProbeHandler)
	},
	// Needed because DeepDerivative compares EnvVar.Value string fields
	// without considering EnvVar as a whole. If an EnvVar is specified, we
	// consider its value to be intentional and force the comparison.
	func(a, b corev1.EnvVar) bool {
		if a.Name != b.Name || a.Value != b.Value {
			return false
		}
		return (conversion.Equalities{}).DeepDerivative(a.ValueFrom, b.ValueFrom)
	},
)

// deploymentEqual returns whether two Deployments are semantically equivalent.
func deploymentEqual(a, b *appsv1.Deployment) bool {
	if a == b {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	if !eq.DeepDerivative(&a.ObjectMeta, &b.ObjectMeta) {
		return false
	}

	if !eq.DeepDerivative(&a.Spec, &b.Spec) {
		return false
	}

	return true
}

// serviceEqual returns whether two Services are semantically equivalent.
func serviceEqual(a, b *corev1.Service) bool {
	if a == b {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	if !eq.DeepDerivative(&a.ObjectMeta, &b.ObjectMeta) {
		return false
	}

	if !eq.DeepDerivative(&a.Spec, &b.Spec) {
		return false
	}

	return true
}

// secretEqual returns whether two Secrets are semantically equivalent.
func secretEqual(a, b *corev1.Secret) bool {
	if a == b {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	if !eq.DeepDerivative(&a.ObjectMeta, &b.ObjectMeta) {
		return false
	}

	if !eq.DeepEqual(&a.Data, &b.Data) {
		return false
	}

	return true
}

// serviceAccountEqual returns whether two ServiceAccounts are semantically equivalent.
func serviceAccountEqual(a, b *corev1.ServiceAccount) bool {
	if a == b {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	if !eq.DeepDerivative(&a.ObjectMeta, &b.ObjectMeta) {
		return false
	}

	if !eq.DeepDerivative(&a.Secrets, &b.Secrets) {
		return false
	}
	if !eq.DeepDerivative(&a.ImagePullSecrets, &b.ImagePullSecrets) {
		return false
	}
	if !eq.DeepDerivative(&a.AutomountServiceAccountToken, &b.AutomountServiceAccountToken) {
		return false
	}

	return true
}

func jobEqual(a, b *batchv1.Job) bool {
	if a == b {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	if !eq.DeepDerivative(&a.ObjectMeta, &b.ObjectMeta) {
		return false
	}

	if !eq.DeepDerivative(&a.Spec, &b.Spec) {
		return false
	}

	return true
}
