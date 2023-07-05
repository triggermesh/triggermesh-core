// Copyright 2023 TriggerMesh Inc.
// SPDX-License-Identifier: Apache-2.0

package resources

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ConfigMapOption func(*corev1.ConfigMap)

func NewConfigMap(namespace, name string, opts ...ConfigMapOption) *corev1.ConfigMap {
	meta := NewMeta(namespace, name)
	cm := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: corev1.SchemeGroupVersion.String(),
		},
		ObjectMeta: *meta,
	}

	for _, opt := range opts {
		opt(cm)
	}

	return cm
}

func ConfigMapWithMetaOptions(opts ...MetaOption) ConfigMapOption {
	return func(cm *corev1.ConfigMap) {
		for _, opt := range opts {
			opt(&cm.ObjectMeta)
		}
	}
}
