// Copyright 2022 TriggerMesh Inc.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	"knative.dev/pkg/kmeta"
)

// +genclient
// +genreconciler
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MemoryBroker is a Memory based broker implementation that collects a pool of
// events that are consumable using Triggers. Brokers provide a well-known endpoint
// for event delivery that senders can use with minimal knowledge of the event
// routing strategy. Subscribers use Triggers to request delivery of events from a
// broker's pool to a specific URL or Addressable endpoint.
type MemoryBroker struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the desired state of the broker.
	Spec MemoryBrokerSpec `json:"spec,omitempty"`

	// Status represents the current state of the broker. This data may be out of
	// date.
	// +optional
	Status MemoryBrokerStatus `json:"status,omitempty"`
}

var (
	// Make sure this is a kubernetes object.
	_ runtime.Object = (*MemoryBroker)(nil)
	// Check that we can create OwnerReferences with this object.
	_ kmeta.OwnerRefable = (*MemoryBroker)(nil)
	// Check that the type conforms to the duck Knative Resource shape.
	_ duckv1.KRShaped = (*MemoryBroker)(nil)
)

type Memory struct {
	// Maximum number of items the stream can host.
	BufferSize *int `json:"streamMaxLen,omitempty"`
}

type MemoryBrokerSpec struct {
	Memory *Memory `json:"memory,omitempty"`

	Broker Broker `json:"broker,omitempty"`
}

// MemoryBrokerStatus represents the current state of a Memory broker.
type MemoryBrokerStatus struct {
	// inherits duck/v1 Status, which currently provides:
	// * ObservedGeneration - the 'Generation' of the Broker that was last processed by the controller.
	// * Conditions - the latest available observations of a resource's current state.
	duckv1.Status `json:",inline"`

	// Broker is Addressable. It exposes the endpoint as an URI to get events
	// delivered into the Broker mesh.
	// +optional
	Address duckv1.Addressable `json:"address,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MemoryBrokerList is a collection of Brokers.
type MemoryBrokerList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []MemoryBroker `json:"items"`
}
