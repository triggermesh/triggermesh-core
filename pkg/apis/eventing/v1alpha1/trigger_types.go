// Copyright 2022 TriggerMesh Inc.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	eventingduckv1 "knative.dev/eventing/pkg/apis/duck/v1"
	eventingv1 "knative.dev/eventing/pkg/apis/eventing/v1"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	"knative.dev/pkg/kmeta"
)

// +genclient
// +genreconciler
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Trigger represents a request to have events delivered to a target from a
// Broker's event pool.
type Trigger struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the desired state of the Trigger.
	Spec TriggerSpec `json:"spec,omitempty"`

	// Status represents the current state of the Trigger. This data may be out of
	// date.
	// +optional
	Status TriggerStatus `json:"status,omitempty"`
}

var (
	// Make sure this is a kubernetes object.
	_ runtime.Object = (*Trigger)(nil)
	// Check that we can create OwnerReferences with this object.
	_ kmeta.OwnerRefable = (*Trigger)(nil)
	// Check that the type conforms to the duck Knative Resource shape.
	_ duckv1.KRShaped = (*Trigger)(nil)
)

// TriggerSpec defines the desired state of Trigger
type TriggerSpec struct {
	// Broker is the broker that this trigger receives events from.
	Broker duckv1.KReference `json:"broker,omitempty"`

	// Filters is an experimental field that conforms to the CNCF CloudEvents Subscriptions
	// API. It's an array of filter expressions that evaluate to true or false.
	// If any filter expression in the array evaluates to false, the event MUST
	// NOT be sent to the target. If all the filter expressions in the array
	// evaluate to true, the event MUST be attempted to be delivered. Absence of
	// a filter or empty array implies a value of true. In the event of users
	// specifying both Filter and Filters, then the latter will override the former.
	// This will allow users to try out the effect of the new Filters field
	// without compromising the existing attribute-based Filter and try it out on existing
	// Trigger objects.
	//
	// +optional
	Filters []eventingv1.SubscriptionsAPIFilter `json:"filters,omitempty"`

	// Target is the addressable that receives events from the Broker that pass
	// the Filter. It is required.
	Target duckv1.Destination `json:"target,omitempty"`

	// Delivery contains the delivery spec for this specific trigger.
	// +optional
	Delivery *eventingduckv1.DeliverySpec `json:"delivery,omitempty"`
}

// TriggerStatus represents the current state of a Trigger.
type TriggerStatus struct {
	// inherits duck/v1 Status, which currently provides:
	// * ObservedGeneration - the 'Generation' of the Trigger that was last processed by the controller.
	// * Conditions - the latest available observations of a resource's current state.
	duckv1.Status `json:",inline"`

	// TargetURI is the resolved URI of the receiver for this Trigger.
	// +optional
	TargetURI *apis.URL `json:"targetUri,omitempty"`

	// DeliveryStatus contains a resolved URL to the dead letter sink address, and any other
	// resolved delivery options.
	eventingduckv1.DeliveryStatus `json:",inline"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// TriggerList is a collection of Triggers.
type TriggerList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Trigger `json:"items"`
}
