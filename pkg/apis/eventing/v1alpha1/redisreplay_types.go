// Copyright 2022 TriggerMesh Inc.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
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
	// Check that the type confroms to duck Knative Resource shape.
	_ duckv1.KRShaped = (*RedisReplay)(nil)
)

type RedisReplaySpec struct {
	// Redis connection information.
	Redis      RedisConnection     `json:"redis"`
	Sink       *duckv1.Destination `json:"sink"`
	StartTime  string              `json:"startTime"`
	EndTime    string              `json:"endTime"`
	Filter     string              `json:"filter"`
	FilterKind string              `json:"filterKind"`
}

type RedisReplayStatus struct {
	duckv1.Status `json:",inline"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type RedisReplayList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `json:"metadata"`
	Items           []RedisReplay `json:"items"`
}
