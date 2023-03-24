// Copyright 2022 TriggerMesh Inc.
// SPDX-License-Identifier: Apache-2.0

package resources

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNewRoleBinding(t *testing.T) {
	crGVK := rbacv1.SchemeGroupVersion.WithKind("ClusterRole")
	saGVK := corev1.SchemeGroupVersion.WithKind("ServiceAccount")

	testCases := map[string]struct {
		options  []RoleBindingOption
		expected rbacv1.RoleBinding
	}{
		"basic": {
			expected: rbacv1.RoleBinding{
				TypeMeta: metav1.TypeMeta{
					Kind:       "RoleBinding",
					APIVersion: rbacv1.SchemeGroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: tNamespace,
					Name:      tRoleBindingName,
				},
				RoleRef: rbacv1.RoleRef{
					APIGroup: crGVK.Group,
					Kind:     crGVK.Kind,
					Name:     tRoleName,
				},
				Subjects: []rbacv1.Subject{{
					APIGroup:  saGVK.Group,
					Kind:      saGVK.Kind,
					Namespace: tNamespace,
					Name:      tServiceAccountName,
				}},
			}},
		"with meta options": {
			options: []RoleBindingOption{
				RoleBindingWithMetaOptions(MetaAddLabel("key", "value")),
			},
			expected: rbacv1.RoleBinding{
				TypeMeta: metav1.TypeMeta{
					Kind:       "RoleBinding",
					APIVersion: rbacv1.SchemeGroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: tNamespace,
					Name:      tRoleBindingName,
					Labels: map[string]string{
						"key": "value",
					},
				},
				RoleRef: rbacv1.RoleRef{
					APIGroup: crGVK.Group,
					Kind:     crGVK.Kind,
					Name:     tRoleName,
				},
				Subjects: []rbacv1.Subject{{
					APIGroup:  saGVK.Group,
					Kind:      saGVK.Kind,
					Namespace: tNamespace,
					Name:      tServiceAccountName,
				}},
			}},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			got := NewRoleBinding(tNamespace, tRoleBindingName, tRoleName, tServiceAccountName, tc.options...)
			assert.Equal(t, &tc.expected, got)
		})
	}

}
