// Copyright 2022 TriggerMesh Inc.
// SPDX-License-Identifier: Apache-2.0

package resources

import (
	"testing"

	"github.com/stretchr/testify/assert"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNewMeta(t *testing.T) {
	testCases := map[string]struct {
		options  []MetaOption
		expected metav1.ObjectMeta
	}{
		"basic": {
			expected: metav1.ObjectMeta{
				Name:      tName,
				Namespace: tNamespace,
			}},
		"with owner": {
			options: []MetaOption{
				MetaAddOwner(&tPod, corev1.SchemeGroupVersion.WithKind("Pod")),
			},
			expected: metav1.ObjectMeta{
				Name:      tName,
				Namespace: tNamespace,
				OwnerReferences: []metav1.OwnerReference{
					{
						APIVersion:         "v1",
						Kind:               "Pod",
						Name:               tName,
						BlockOwnerDeletion: &tTrue,
						Controller:         &tTrue,
					},
				},
			}},
		"with labels": {
			options: []MetaOption{
				MetaAddLabel("key1", "label1"),
			},
			expected: metav1.ObjectMeta{
				Name:      tName,
				Namespace: tNamespace,
				Labels: map[string]string{
					"key1": "label1",
				},
			}},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			got := NewMeta(tNamespace, tName, tc.options...)
			assert.Equal(t, &tc.expected, got)
		})
	}
}
