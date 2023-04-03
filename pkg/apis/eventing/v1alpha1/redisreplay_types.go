// Copyright 2022 TriggerMesh Inc.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	broker "github.com/triggermesh/brokers/pkg/config/broker"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	duckv1 "knative.dev/pkg/apis/duck/v1"
)

// +genclient
// +genreconciler
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type RedisReplay struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the desired state of the broker.
	Spec RedisReplaySpec `json:"spec,omitempty"`

	// Status represents the current state of the broker. This data may be out of
	// date.
	// +optional
	Status RedisReplayStatus `json:"status,omitempty"`
}

var (
	// Make sure this is a kubernetes object.
	_ runtime.Object = (*RedisReplay)(nil)
	// Check that the type conforms to the duck Knative Resource shape.
	_ duckv1.KRShaped = (*RedisReplay)(nil)
)

type RedisReplaySpec struct {
	// Redis connection information.
	Broker BrokerInfo `json:"broker"`
	Redis  *Redis     `json:"redis,omitempty"`
	// +optional
	StartTime *string `json:"startTime,omitempty"`
	// +optional
	EndTime *string `json:"endTime,omitempty"`
	// +optional
	Filters []broker.Filter     `json:"filters,omitempty"`
	Target  *duckv1.Destination `json:"target"`
}

type RedisReplayStatus struct {
	duckv1.Status `json:",inline"`
}

type BrokerInfo struct {
	Group string `json:"group"`
	Kind  string `json:"kind"`
	Name  string `json:"name"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type RedisReplayList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `json:"metadata"`
	Items           []RedisReplay `json:"items"`
}
