// Copyright 2022 TriggerMesh Inc.corev1.Service
// SPDX-License-Identifier: Apache-2.0

package resources

import (
	"testing"

	"github.com/stretchr/testify/assert"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestNewService(t *testing.T) {
	testCases := map[string]struct {
		options  []ServiceOption
		expected corev1.Service
	}{
		"basic": {
			expected: corev1.Service{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Service",
					APIVersion: corev1.SchemeGroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: tNamespace,
					Name:      tName,
				},
			}},
		"with meta options": {
			options: []ServiceOption{
				ServiceWithMetaOptions(MetaAddLabel("key", "value")),
			},
			expected: corev1.Service{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Service",
					APIVersion: corev1.SchemeGroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: tNamespace,
					Name:      tName,
					Labels: map[string]string{
						"key": "value",
					},
				},
			}},
		"with selector label": {
			options: []ServiceOption{
				ServiceAddSelectorLabel("key", "value"),
			},
			expected: corev1.Service{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Service",
					APIVersion: corev1.SchemeGroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: tNamespace,
					Name:      tName,
				},
				Spec: corev1.ServiceSpec{
					Selector: map[string]string{
						"key": "value",
					},
				},
			}},
		"with port": {
			options: []ServiceOption{
				ServiceAddPort("TEST", 8888, 80),
			},
			expected: corev1.Service{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Service",
					APIVersion: corev1.SchemeGroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: tNamespace,
					Name:      tName,
				},
				Spec: corev1.ServiceSpec{
					Ports: []corev1.ServicePort{
						{
							Name:       "TEST",
							Port:       8888,
							TargetPort: intstr.FromInt(80),
						},
					},
				},
			}},
		"with service type": {
			options: []ServiceOption{
				ServiceSetType(corev1.ServiceTypeLoadBalancer),
			},
			expected: corev1.Service{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Service",
					APIVersion: corev1.SchemeGroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: tNamespace,
					Name:      tName,
				},
				Spec: corev1.ServiceSpec{
					Type: corev1.ServiceTypeLoadBalancer,
				},
			}},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			got := NewService(tNamespace, tName, tc.options...)
			assert.Equal(t, &tc.expected, got)
		})
	}
}
