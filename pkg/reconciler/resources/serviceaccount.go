// Copyright 2022 TriggerMesh Inc.
// SPDX-License-Identifier: Apache-2.0

package resources

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ServiceAccountOption func(*corev1.ServiceAccount)

func NewServiceAccount(namespace, name string, opts ...ServiceAccountOption) *corev1.ServiceAccount {
	meta := NewMeta(namespace, name)
	sa := &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceAccount",
			APIVersion: corev1.SchemeGroupVersion.String(),
		},
		ObjectMeta: *meta,
	}

	for _, opt := range opts {
		opt(sa)
	}

	return sa
}

func ServiceAccountWithMetaOptions(opts ...MetaOption) ServiceAccountOption {
	return func(s *corev1.ServiceAccount) {
		for _, opt := range opts {
			opt(&s.ObjectMeta)
		}
	}
}
