// Copyright 2023 TriggerMesh Inc.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	eventingduckv1 "knative.dev/eventing/pkg/apis/duck/v1"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	"knative.dev/pkg/kmeta"
)

// +genclient
// +genreconciler
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// TODO add description
type Replay struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the desired state of the Replay.
	Spec ReplaySpec `json:"spec,omitempty"`

	// Status represents the current state of the Replay. This data may be out of
	// date.
	// +optional
	Status ReplayStatus `json:"status,omitempty"`
}

var (
	// Make sure this is a kubernetes object.
	_ runtime.Object = (*Replay)(nil)
	// Check that we can create OwnerReferences with this object.
	_ kmeta.OwnerRefable = (*Replay)(nil)
	// Check that the type conforms to the duck Knative Resource shape.
	_ duckv1.KRShaped = (*Replay)(nil)
)

// ReplaySpec defines the desired state of Replay
type ReplaySpec struct {
	TriggerSpec `json:",inline"`

	// Bounds for the receiving events
	Bounds TriggerBounds `json:",omitempty"`
}

// ReplayStatus represents the current state of a Replay.
type ReplayStatus struct {
	// inherits duck/v1 Status, which currently provides:
	// * ObservedGeneration - the 'Generation' of the Replay that was last processed by the controller.
	// * Conditions - the latest available observations of a resource's current state.
	duckv1.Status `json:",inline"`

	// TargetURI is the resolved URI of the receiver for this Replay.
	// +optional
	TargetURI *apis.URL `json:"targetUri,omitempty"`

	// DeliveryStatus contains a resolved URL to the dead letter sink address, and any other
	// resolved delivery options.
	eventingduckv1.DeliveryStatus `json:",inline"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ReplayList is a collection of Replays.
type ReplayList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Replay `json:"items"`
}
