// Copyright 2022 TriggerMesh Inc.
// SPDX-License-Identifier: Apache-2.0
package resources

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	tName      = "test-name"
	tNamespace = "test-namespace"
	tImage     = "triggermesh/test:v1"

	tSecretName      = "test-secret"
	tSecretKey       = "test-key"
	tVolumeMountFile = "myfile"
)

var (
	tTrue = true

	tPod = corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: corev1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      tName,
			Namespace: tNamespace,
		},
	}
)
