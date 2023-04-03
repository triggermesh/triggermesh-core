// Copyright 2023 TriggerMesh Inc.
// SPDX-License-Identifier: Apache-2.0

package resources

import (
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type JobOption func(*batchv1.Job)

func NewJob(namespace, name string, opts ...JobOption) *batchv1.Job {
	meta := NewMeta(namespace, name)
	j := &batchv1.Job{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Job",
			APIVersion: batchv1.SchemeGroupVersion.String(),
		},
		ObjectMeta: *meta,
	}

	for _, opt := range opts {
		opt(j)
	}

	return j
}

func JobWithMetaOptions(opts ...MetaOption) JobOption {
	return func(j *batchv1.Job) {
		for _, opt := range opts {
			opt(&j.ObjectMeta)
		}
	}
}

func JobSetSelector(key, value string) JobOption {
	return func(j *batchv1.Job) {
		if j.Spec.Selector == nil {
			j.Spec.Selector = &metav1.LabelSelector{}
		}

		sl := j.Spec.Selector.MatchLabels
		if sl == nil {
			sl = make(map[string]string, 1)
			j.Spec.Selector.MatchLabels = sl
		}
		sl[key] = value
	}
}

func JobWithTemplateSpecOptions(opts ...PodTemplateSpecOption) JobOption {
	return func(j *batchv1.Job) {
		if j.Spec.Template.Spec.Containers == nil {
			j.Spec.Template.Spec.Containers = make([]corev1.Container, 0)
		}

		for _, opt := range opts {
			opt(&j.Spec.Template)
		}
	}
}

func PodTemplateSpecWithRestartPolicy(policy corev1.RestartPolicy) PodTemplateSpecOption {
	return func(pts *corev1.PodTemplateSpec) {
		pts.Spec.RestartPolicy = policy
	}
}
