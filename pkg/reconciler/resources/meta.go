// Copyright 2022 TriggerMesh Inc.
// SPDX-License-Identifier: Apache-2.0

package resources

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type MetaOption func(*metav1.ObjectMeta)

func NewMeta(ns, name string, opts ...MetaOption) *metav1.ObjectMeta {
	m := &metav1.ObjectMeta{
		Namespace: ns,
		Name:      name,
	}

	for _, opt := range opts {
		opt(m)
	}

	return m
}

func MetaAddOwner(o metav1.Object, gvk schema.GroupVersionKind) MetaOption {
	return func(m *metav1.ObjectMeta) {
		m.OwnerReferences = append(m.OwnerReferences, *metav1.NewControllerRef(o, gvk))
	}
}

func MetaAddLabel(key, value string) MetaOption {
	return func(m *metav1.ObjectMeta) {
		if m.Labels == nil {
			m.Labels = make(map[string]string, 1)
		}
		m.Labels[key] = value
	}
}
