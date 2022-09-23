// Copyright 2022 TriggerMesh Inc.
// SPDX-License-Identifier: Apache-2.0

package resources

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type SecretOption func(*corev1.Secret)

func NewSecret(namespace, name string, opts ...SecretOption) *corev1.Secret {
	meta := NewMeta(namespace, name)
	s := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: corev1.SchemeGroupVersion.String(),
		},
		ObjectMeta: *meta,
		Type:       corev1.SecretTypeOpaque,
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

func SecretWithMetaOptions(opts ...MetaOption) SecretOption {
	return func(s *corev1.Secret) {
		for _, opt := range opts {
			opt(&s.ObjectMeta)
		}
	}
}

func SecretSetData(key string, value []byte) SecretOption {
	return func(s *corev1.Secret) {
		if s.Data == nil {
			s.Data = make(map[string][]byte)
		}
		s.Data[key] = value
	}
}
