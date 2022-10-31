// Copyright 2022 TriggerMesh Inc.
// SPDX-License-Identifier: Apache-2.0

package resources

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type DeploymentOption func(*appsv1.Deployment)

func NewDeployment(namespace, name string, opts ...DeploymentOption) *appsv1.Deployment {
	meta := NewMeta(namespace, name)
	d := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: appsv1.SchemeGroupVersion.String(),
		},
		ObjectMeta: *meta,
	}

	for _, opt := range opts {
		opt(d)
	}

	return d
}

func DeploymentWithMetaOptions(opts ...MetaOption) DeploymentOption {
	return func(d *appsv1.Deployment) {
		for _, opt := range opts {
			opt(&d.ObjectMeta)
		}
	}
}

func DeploymentSetReplicas(replicas int32) DeploymentOption {
	return func(d *appsv1.Deployment) {
		d.Spec.Replicas = &replicas
	}
}

func DeploymentAddSelectorForTemplate(key, value string) DeploymentOption {
	return func(d *appsv1.Deployment) {
		if d.Spec.Selector == nil {
			d.Spec.Selector = &metav1.LabelSelector{}
		}

		sl := d.Spec.Selector.MatchLabels
		if sl == nil {
			sl = make(map[string]string, 1)
			d.Spec.Selector.MatchLabels = sl
		}
		sl[key] = value

		MetaAddLabel(key, value)(&d.Spec.Template.ObjectMeta)
	}
}

func DeploymentWithTemplateOptions(opts ...PodSpecOption) DeploymentOption {
	return func(d *appsv1.Deployment) {
		for _, opt := range opts {
			opt(&d.Spec.Template.Spec)
		}
	}
}

type PodSpecOption func(*corev1.PodSpec)

func NewPodSpec(opts ...PodSpecOption) *corev1.PodSpec {
	ps := &corev1.PodSpec{}

	for _, opt := range opts {
		opt(ps)
	}

	return ps
}

func PodSpecAddContainer(c *corev1.Container) PodSpecOption {
	return func(ps *corev1.PodSpec) {
		if ps.Containers == nil {
			ps.Containers = make([]corev1.Container, 0, 1)
		}
		ps.Containers = append(ps.Containers, *c)
	}
}

func PodSpecAddVolume(v *corev1.Volume) PodSpecOption {
	return func(ps *corev1.PodSpec) {
		if ps.Volumes == nil {
			ps.Volumes = make([]corev1.Volume, 0, 1)
		}
		ps.Volumes = append(ps.Volumes, *v)
	}
}

func PodSpecWithServiceAccountName(saName string) PodSpecOption {
	return func(ps *corev1.PodSpec) {
		ps.ServiceAccountName = saName
	}
}
