// Copyright 2022 TriggerMesh Inc.
// SPDX-License-Identifier: Apache-2.0

package resources

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type ServiceOption func(*corev1.Service)

func NewService(namespace, name string, opts ...ServiceOption) *corev1.Service {
	meta := NewMeta(namespace, name)
	d := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: corev1.SchemeGroupVersion.String(),
		},
		ObjectMeta: *meta,
	}

	for _, opt := range opts {
		opt(d)
	}

	return d
}

func ServiceWithMetaOptions(opts ...MetaOption) ServiceOption {
	return func(s *corev1.Service) {
		for _, opt := range opts {
			opt(&s.ObjectMeta)
		}
	}
}

func ServiceAddSelectorLabel(key, value string) ServiceOption {
	return func(s *corev1.Service) {
		if s.Spec.Selector == nil {
			s.Spec.Selector = make(map[string]string, 1)
		}

		s.Spec.Selector[key] = value
	}
}

func ServiceAddPort(name string, port int32, targetPort int32) ServiceOption {
	return func(s *corev1.Service) {
		if s.Spec.Ports == nil {
			s.Spec.Ports = make([]corev1.ServicePort, 0, 1)
		}

		s.Spec.Ports = append(s.Spec.Ports, corev1.ServicePort{
			Name:       name,
			Port:       port,
			TargetPort: intstr.FromInt(int(targetPort)),
		})
	}
}

func ServiceSetType(st corev1.ServiceType) ServiceOption {
	return func(s *corev1.Service) {
		s.Spec.Type = st
	}
}
