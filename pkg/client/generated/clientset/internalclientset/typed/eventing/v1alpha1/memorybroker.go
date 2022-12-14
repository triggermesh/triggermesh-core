// Copyright 2022 TriggerMesh Inc.
// SPDX-License-Identifier: Apache-2.0
// Code generated by client-gen. DO NOT EDIT.

package v1alpha1

import (
	"context"
	"time"

	v1alpha1 "github.com/triggermesh/triggermesh-core/pkg/apis/eventing/v1alpha1"
	scheme "github.com/triggermesh/triggermesh-core/pkg/client/generated/clientset/internalclientset/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// MemoryBrokersGetter has a method to return a MemoryBrokerInterface.
// A group's client should implement this interface.
type MemoryBrokersGetter interface {
	MemoryBrokers(namespace string) MemoryBrokerInterface
}

// MemoryBrokerInterface has methods to work with MemoryBroker resources.
type MemoryBrokerInterface interface {
	Create(ctx context.Context, memoryBroker *v1alpha1.MemoryBroker, opts v1.CreateOptions) (*v1alpha1.MemoryBroker, error)
	Update(ctx context.Context, memoryBroker *v1alpha1.MemoryBroker, opts v1.UpdateOptions) (*v1alpha1.MemoryBroker, error)
	UpdateStatus(ctx context.Context, memoryBroker *v1alpha1.MemoryBroker, opts v1.UpdateOptions) (*v1alpha1.MemoryBroker, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha1.MemoryBroker, error)
	List(ctx context.Context, opts v1.ListOptions) (*v1alpha1.MemoryBrokerList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.MemoryBroker, err error)
	MemoryBrokerExpansion
}

// memoryBrokers implements MemoryBrokerInterface
type memoryBrokers struct {
	client rest.Interface
	ns     string
}

// newMemoryBrokers returns a MemoryBrokers
func newMemoryBrokers(c *EventingV1alpha1Client, namespace string) *memoryBrokers {
	return &memoryBrokers{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the memoryBroker, and returns the corresponding memoryBroker object, and an error if there is any.
func (c *memoryBrokers) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.MemoryBroker, err error) {
	result = &v1alpha1.MemoryBroker{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("memorybrokers").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of MemoryBrokers that match those selectors.
func (c *memoryBrokers) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.MemoryBrokerList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1alpha1.MemoryBrokerList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("memorybrokers").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested memoryBrokers.
func (c *memoryBrokers) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("memorybrokers").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

// Create takes the representation of a memoryBroker and creates it.  Returns the server's representation of the memoryBroker, and an error, if there is any.
func (c *memoryBrokers) Create(ctx context.Context, memoryBroker *v1alpha1.MemoryBroker, opts v1.CreateOptions) (result *v1alpha1.MemoryBroker, err error) {
	result = &v1alpha1.MemoryBroker{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("memorybrokers").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(memoryBroker).
		Do(ctx).
		Into(result)
	return
}

// Update takes the representation of a memoryBroker and updates it. Returns the server's representation of the memoryBroker, and an error, if there is any.
func (c *memoryBrokers) Update(ctx context.Context, memoryBroker *v1alpha1.MemoryBroker, opts v1.UpdateOptions) (result *v1alpha1.MemoryBroker, err error) {
	result = &v1alpha1.MemoryBroker{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("memorybrokers").
		Name(memoryBroker.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(memoryBroker).
		Do(ctx).
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *memoryBrokers) UpdateStatus(ctx context.Context, memoryBroker *v1alpha1.MemoryBroker, opts v1.UpdateOptions) (result *v1alpha1.MemoryBroker, err error) {
	result = &v1alpha1.MemoryBroker{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("memorybrokers").
		Name(memoryBroker.Name).
		SubResource("status").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(memoryBroker).
		Do(ctx).
		Into(result)
	return
}

// Delete takes name of the memoryBroker and deletes it. Returns an error if one occurs.
func (c *memoryBrokers) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("memorybrokers").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *memoryBrokers) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("memorybrokers").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

// Patch applies the patch and returns the patched memoryBroker.
func (c *memoryBrokers) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.MemoryBroker, err error) {
	result = &v1alpha1.MemoryBroker{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("memorybrokers").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}
