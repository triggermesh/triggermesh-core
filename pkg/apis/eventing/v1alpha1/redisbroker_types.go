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

// RedisBroker is a Redis based broker implementation that collects a pool of
// events that are consumable using Triggers. Brokers provide a well-known endpoint
// for event delivery that senders can use with minimal knowledge of the event
// routing strategy. Subscribers use Triggers to request delivery of events from a
// broker's pool to a specific URL or Addressable endpoint.
type RedisBroker struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the desired state of the broker.
	Spec RedisBrokerSpec `json:"spec,omitempty"`

	// Status represents the current state of the broker. This data may be out of
	// date.
	// +optional
	Status RedisBrokerStatus `json:"status,omitempty"`
}

var (
	// Make sure this is a kubernetes object.
	_ runtime.Object = (*RedisBroker)(nil)
	// Check that we can create OwnerReferences with this object.
	_ kmeta.OwnerRefable = (*RedisBroker)(nil)
	// Check that the type conforms to the duck Knative Resource shape.
	_ duckv1.KRShaped = (*RedisBroker)(nil)
)

type RedisBrokerSpec struct {
}

// RedisBrokerStatus represents the current state of a Redis broker.
type RedisBrokerStatus struct {
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

// RedisBrokerList is a collection of Brokers.
type RedisBrokerList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []RedisBroker `json:"items"`
}
