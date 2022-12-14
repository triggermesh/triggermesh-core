// Copyright 2022 TriggerMesh Inc.
// SPDX-License-Identifier: Apache-2.0
// Code generated by lister-gen. DO NOT EDIT.

package v1alpha1

import (
	v1alpha1 "github.com/triggermesh/triggermesh-core/pkg/apis/eventing/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// MemoryBrokerLister helps list MemoryBrokers.
// All objects returned here must be treated as read-only.
type MemoryBrokerLister interface {
	// List lists all MemoryBrokers in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.MemoryBroker, err error)
	// MemoryBrokers returns an object that can list and get MemoryBrokers.
	MemoryBrokers(namespace string) MemoryBrokerNamespaceLister
	MemoryBrokerListerExpansion
}

// memoryBrokerLister implements the MemoryBrokerLister interface.
type memoryBrokerLister struct {
	indexer cache.Indexer
}

// NewMemoryBrokerLister returns a new MemoryBrokerLister.
func NewMemoryBrokerLister(indexer cache.Indexer) MemoryBrokerLister {
	return &memoryBrokerLister{indexer: indexer}
}

// List lists all MemoryBrokers in the indexer.
func (s *memoryBrokerLister) List(selector labels.Selector) (ret []*v1alpha1.MemoryBroker, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.MemoryBroker))
	})
	return ret, err
}

// MemoryBrokers returns an object that can list and get MemoryBrokers.
func (s *memoryBrokerLister) MemoryBrokers(namespace string) MemoryBrokerNamespaceLister {
	return memoryBrokerNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// MemoryBrokerNamespaceLister helps list and get MemoryBrokers.
// All objects returned here must be treated as read-only.
type MemoryBrokerNamespaceLister interface {
	// List lists all MemoryBrokers in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.MemoryBroker, err error)
	// Get retrieves the MemoryBroker from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1alpha1.MemoryBroker, error)
	MemoryBrokerNamespaceListerExpansion
}

// memoryBrokerNamespaceLister implements the MemoryBrokerNamespaceLister
// interface.
type memoryBrokerNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all MemoryBrokers in the indexer for a given namespace.
func (s memoryBrokerNamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.MemoryBroker, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.MemoryBroker))
	})
	return ret, err
}

// Get retrieves the MemoryBroker from the indexer for a given namespace and name.
func (s memoryBrokerNamespaceLister) Get(name string) (*v1alpha1.MemoryBroker, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("memorybroker"), name)
	}
	return obj.(*v1alpha1.MemoryBroker), nil
}
