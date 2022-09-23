// Copyright 2022 TriggerMesh Inc.corev1.Service
// SPDX-License-Identifier: Apache-2.0

package resources

import (
	"testing"

	"github.com/stretchr/testify/assert"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNewSecret(t *testing.T) {
	testCases := map[string]struct {
		options  []SecretOption
		expected corev1.Secret
	}{
		"basic": {
			expected: corev1.Secret{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Secret",
					APIVersion: corev1.SchemeGroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: tNamespace,
					Name:      tName,
				},
				Type: corev1.SecretTypeOpaque,
			}},
		"with meta options": {
			options: []SecretOption{
				SecretWithMetaOptions(MetaAddLabel("key", "value")),
			},
			expected: corev1.Secret{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Secret",
					APIVersion: corev1.SchemeGroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: tNamespace,
					Name:      tName,
					Labels: map[string]string{
						"key": "value",
					},
				},
				Type: corev1.SecretTypeOpaque,
			}},
		"with with data": {
			options: []SecretOption{
				SecretSetData("key", []byte("value")),
			},
			expected: corev1.Secret{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Secret",
					APIVersion: corev1.SchemeGroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: tNamespace,
					Name:      tName,
				},
				Type: corev1.SecretTypeOpaque,
				Data: map[string][]byte{
					"key": []byte("value"),
				},
			}},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			got := NewSecret(tNamespace, tName, tc.options...)
			assert.Equal(t, &tc.expected, got)
		})
	}
}
