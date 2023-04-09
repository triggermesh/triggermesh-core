// Copyright 2022 TriggerMesh Inc.
// SPDX-License-Identifier: Apache-2.0

package resources

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Kubernetes recommended labels
// https://kubernetes.io/docs/concepts/overview/working-with-objects/common-labels/
const (
	// appNameLabel is the name of the application.
	AppNameLabel = "app.kubernetes.io/name"
	// appInstanceLabel is a unique name identifying the instance of an application.
	AppInstanceLabel = "app.kubernetes.io/instance"
	// appComponentLabel is the component within the architecture.
	AppComponentLabel = "app.kubernetes.io/component"
	// appPartOfLabel is the name of a higher level application this one is part of.
	AppPartOfLabel = "app.kubernetes.io/part-of"
	// appManagedByLabel is the tool being used to manage the operation of an application.
	AppManagedByLabel = "app.kubernetes.io/managed-by"
)

// Common label values
const (
	PartOf    = "triggermesh"
	ManagedBy = "triggermesh-core"
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

func MetaSetDeletion(t *metav1.Time) MetaOption {
	return func(m *metav1.ObjectMeta) {
		m.DeletionTimestamp = t
	}
}

func MetaAddOwnerReferences(ownerReference metav1.OwnerReference) MetaOption {
	return func(m *metav1.ObjectMeta) {
		m.OwnerReferences = append(m.OwnerReferences, ownerReference)
	}
}
