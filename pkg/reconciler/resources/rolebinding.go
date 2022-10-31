// Copyright 2022 TriggerMesh Inc.
// SPDX-License-Identifier: Apache-2.0

package resources

import (
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type RoleBindingOption func(*rbacv1.RoleBinding)

func NewRoleBinding(namespace, name, roleName, subjectName string, opts ...RoleBindingOption) *rbacv1.RoleBinding {
	crGVK := rbacv1.SchemeGroupVersion.WithKind("ClusterRole")
	saGVK := corev1.SchemeGroupVersion.WithKind("ServiceAccount")

	meta := NewMeta(namespace, name)
	rb := &rbacv1.RoleBinding{
		TypeMeta: metav1.TypeMeta{
			Kind:       "RoleBinding",
			APIVersion: rbacv1.SchemeGroupVersion.String(),
		},
		ObjectMeta: *meta,
		RoleRef: rbacv1.RoleRef{
			APIGroup: crGVK.Group,
			Kind:     crGVK.Kind,
			Name:     roleName,
		},
		Subjects: []rbacv1.Subject{{
			APIGroup:  saGVK.Group,
			Kind:      saGVK.Kind,
			Namespace: namespace,
			Name:      subjectName,
		}},
	}

	for _, opt := range opts {
		opt(rb)
	}

	return rb
}

func RoleBindingWithMetaOptions(opts ...MetaOption) RoleBindingOption {
	return func(s *rbacv1.RoleBinding) {
		for _, opt := range opts {
			opt(&s.ObjectMeta)
		}
	}
}
