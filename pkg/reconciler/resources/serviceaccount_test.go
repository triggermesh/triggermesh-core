// Copyright 2022 TriggerMesh Inc.
// SPDX-License-Identifier: Apache-2.0

package resources

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNewServiceAccount(t *testing.T) {
	testCases := map[string]struct {
		options  []ServiceAccountOption
		expected corev1.ServiceAccount
	}{
		"basic": {
			expected: corev1.ServiceAccount{
				TypeMeta: metav1.TypeMeta{
					Kind:       "ServiceAccount",
					APIVersion: corev1.SchemeGroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: tNamespace,
					Name:      tServiceAccountName,
				},
			}},
		"with meta options": {
			options: []ServiceAccountOption{
				ServiceAccountWithMetaOptions(MetaAddLabel("key", "value")),
			},
			expected: corev1.ServiceAccount{
				TypeMeta: metav1.TypeMeta{
					Kind:       "ServiceAccount",
					APIVersion: corev1.SchemeGroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: tNamespace,
					Name:      tServiceAccountName,
					Labels: map[string]string{
						"key": "value",
					},
				},
			}},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			got := NewServiceAccount(tNamespace, tServiceAccountName, tc.options...)
			assert.Equal(t, &tc.expected, got)
		})
	}

}
