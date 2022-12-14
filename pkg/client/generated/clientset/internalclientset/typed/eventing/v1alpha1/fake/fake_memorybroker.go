// Copyright 2022 TriggerMesh Inc.
// SPDX-License-Identifier: Apache-2.0
// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	"context"

	v1alpha1 "github.com/triggermesh/triggermesh-core/pkg/apis/eventing/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeMemoryBrokers implements MemoryBrokerInterface
type FakeMemoryBrokers struct {
	Fake *FakeEventingV1alpha1
	ns   string
}

var memorybrokersResource = schema.GroupVersionResource{Group: "eventing.triggermesh.io", Version: "v1alpha1", Resource: "memorybrokers"}

var memorybrokersKind = schema.GroupVersionKind{Group: "eventing.triggermesh.io", Version: "v1alpha1", Kind: "MemoryBroker"}

// Get takes name of the memoryBroker, and returns the corresponding memoryBroker object, and an error if there is any.
func (c *FakeMemoryBrokers) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.MemoryBroker, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(memorybrokersResource, c.ns, name), &v1alpha1.MemoryBroker{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.MemoryBroker), err
}

// List takes label and field selectors, and returns the list of MemoryBrokers that match those selectors.
func (c *FakeMemoryBrokers) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.MemoryBrokerList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(memorybrokersResource, memorybrokersKind, c.ns, opts), &v1alpha1.MemoryBrokerList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.MemoryBrokerList{ListMeta: obj.(*v1alpha1.MemoryBrokerList).ListMeta}
	for _, item := range obj.(*v1alpha1.MemoryBrokerList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested memoryBrokers.
func (c *FakeMemoryBrokers) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(memorybrokersResource, c.ns, opts))

}

// Create takes the representation of a memoryBroker and creates it.  Returns the server's representation of the memoryBroker, and an error, if there is any.
func (c *FakeMemoryBrokers) Create(ctx context.Context, memoryBroker *v1alpha1.MemoryBroker, opts v1.CreateOptions) (result *v1alpha1.MemoryBroker, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(memorybrokersResource, c.ns, memoryBroker), &v1alpha1.MemoryBroker{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.MemoryBroker), err
}

// Update takes the representation of a memoryBroker and updates it. Returns the server's representation of the memoryBroker, and an error, if there is any.
func (c *FakeMemoryBrokers) Update(ctx context.Context, memoryBroker *v1alpha1.MemoryBroker, opts v1.UpdateOptions) (result *v1alpha1.MemoryBroker, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(memorybrokersResource, c.ns, memoryBroker), &v1alpha1.MemoryBroker{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.MemoryBroker), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeMemoryBrokers) UpdateStatus(ctx context.Context, memoryBroker *v1alpha1.MemoryBroker, opts v1.UpdateOptions) (*v1alpha1.MemoryBroker, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(memorybrokersResource, "status", c.ns, memoryBroker), &v1alpha1.MemoryBroker{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.MemoryBroker), err
}

// Delete takes name of the memoryBroker and deletes it. Returns an error if one occurs.
func (c *FakeMemoryBrokers) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(memorybrokersResource, c.ns, name, opts), &v1alpha1.MemoryBroker{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeMemoryBrokers) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(memorybrokersResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.MemoryBrokerList{})
	return err
}

// Patch applies the patch and returns the patched memoryBroker.
func (c *FakeMemoryBrokers) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.MemoryBroker, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(memorybrokersResource, c.ns, name, pt, data, subresources...), &v1alpha1.MemoryBroker{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.MemoryBroker), err
}
