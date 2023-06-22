// Copyright 2022 TriggerMesh Inc.
// SPDX-License-Identifier: Apache-2.0
package v1alpha1

import (
	"testing"

	"github.com/stretchr/testify/assert"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	duckv1 "knative.dev/pkg/apis/duck/v1"
)

const (
	tBrokerName = "testbroker"
	tNamespace  = "testnamespace"
)

func TestDoesTriggerRefBroker(t *testing.T) {
	rb := &RedisBroker{
		TypeMeta: metav1.TypeMeta{
			APIVersion: SchemeGroupVersion.String(),
			Kind:       "RedisBroker",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      tBrokerName,
			Namespace: tNamespace,
		},
	}

	testCases := map[string]struct {
		trigger  *Trigger
		broker   *RedisBroker
		expected bool
	}{
		"matching GK and name, using group": {
			trigger: &Trigger{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: tNamespace,
				},
				Spec: TriggerSpecBounded{
					TriggerSpec: TriggerSpec{
						Broker: duckv1.KReference{
							Group: "eventing.triggermesh.io",
							Kind:  "RedisBroker",
							Name:  tBrokerName,
						},
					},
				},
			},
			broker:   rb,
			expected: true,
		},
		"matching GVK and name, using APIVersion": {
			trigger: &Trigger{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: tNamespace,
				},
				Spec: TriggerSpecBounded{
					TriggerSpec: TriggerSpec{
						Broker: duckv1.KReference{
							APIVersion: "eventing.triggermesh.io/v1alpha1",
							Kind:       "RedisBroker",
							Name:       tBrokerName,
						},
					},
				},
			},
			broker:   rb,
			expected: true,
		},
		"non matching version, using APIVersion": {
			trigger: &Trigger{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: tNamespace,
				},
				Spec: TriggerSpecBounded{
					TriggerSpec: TriggerSpec{
						Broker: duckv1.KReference{
							APIVersion: "eventing.triggermesh.io/v1alpha2",
							Kind:       "RedisBroker",
							Name:       tBrokerName,
						},
					},
				},
			},
			broker:   rb,
			expected: false,
		},
		"non matching group, using APIVersion": {
			trigger: &Trigger{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: tNamespace,
				},
				Spec: TriggerSpecBounded{
					TriggerSpec: TriggerSpec{
						Broker: duckv1.KReference{
							APIVersion: "test.triggermesh.io/v1alpha1",
							Kind:       "RedisBroker",
							Name:       tBrokerName,
						},
					},
				},
			},
			broker:   rb,
			expected: false,
		},
		"non matching group, using group": {
			trigger: &Trigger{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: tNamespace,
				},
				Spec: TriggerSpecBounded{
					TriggerSpec: TriggerSpec{
						Broker: duckv1.KReference{
							Group: "test.triggermesh.io",
							Kind:  "RedisBroker",
							Name:  tBrokerName,
						},
					},
				},
			},
			broker:   rb,
			expected: false,
		},
		"missing group and APIVersion": {
			trigger: &Trigger{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: tNamespace,
				},
				Spec: TriggerSpecBounded{
					TriggerSpec: TriggerSpec{
						Broker: duckv1.KReference{
							Kind: "RedisBroker",
							Name: tBrokerName,
						},
					},
				},
			},
			broker:   rb,
			expected: false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			got := tc.trigger.ReferencesBroker(tc.broker)
			assert.Equal(t, tc.expected, got)
		})
	}
}
