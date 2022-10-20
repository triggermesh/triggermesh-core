// Copyright 2022 TriggerMesh Inc.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
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

type RedisConnection struct {
	// Redis URL.
	URL string `json:"url"`

	// Redis username.
	Username *SecretValueFromSource `json:"username,omitempty"`

	// Redis password.
	Password *SecretValueFromSource `json:"password,omitempty"`

	// Use TLS enctrypted connection.
	TLSEnabled *bool `json:"tlsEnabled,omitempty"`

	// Skip TLS certificate verification.
	TLSSkipVerify *bool `json:"tlsSkipVerify,omitempty"`
}

type Redis struct {
	// Redis connection data.
	Connection *RedisConnection `json:"connection,omitempty"`

	// Stream name used by the broker.
	Stream *string `json:"stream,omitempty"`

	// Maximum number of items the stream can host.
	StreamMaxLen *int `json:"streamMaxLen,omitempty"`
}

// SecretValueFromSource represents the source of a secret value
type SecretValueFromSource struct {
	// The Secret key to select from.
	SecretKeyRef corev1.SecretKeySelector `json:"secretKeyRef,omitempty"`
}

type Broker struct {
	Port *int `json:"port,omitempty"`
}

type RedisBrokerSpec struct {
	Redis *Redis `json:"redis,omitempty"`

	Broker Broker `json:"broker,omitempty"`
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
