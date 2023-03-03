// Copyright 2022 TriggerMesh Inc.
// SPDX-License-Identifier: Apache-2.0

package testing

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	fakek8sclient "k8s.io/client-go/kubernetes/fake"
	appslistersv1 "k8s.io/client-go/listers/apps/v1"
	corelistersv1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"

	eventingv1alpha1 "github.com/triggermesh/triggermesh-core/pkg/apis/eventing/v1alpha1"
	fakeclient "github.com/triggermesh/triggermesh-core/pkg/client/generated/clientset/internalclientset/fake"
	eventinglistersv1alpha1 "github.com/triggermesh/triggermesh-core/pkg/client/generated/listers/eventing/v1alpha1"
	fakeeventingclientset "knative.dev/eventing/pkg/client/clientset/versioned/fake"
	rt "knative.dev/pkg/reconciler/testing"
)

var clientSetSchemes = []func(*runtime.Scheme) error{
	fakeclient.AddToScheme,
	fakek8sclient.AddToScheme,
	// although our reconcilers do not handle eventing objects directly, we
	// do need to register the eventing Scheme so that sink URI resolvers
	// can recognize the Broker objects we use in tests
	fakeeventingclientset.AddToScheme,
}

// NewScheme returns a new scheme populated with the types defined in clientSetSchemes.
func NewScheme() *runtime.Scheme {
	scheme := runtime.NewScheme()

	sb := runtime.NewSchemeBuilder(clientSetSchemes...)
	if err := sb.AddToScheme(scheme); err != nil {
		panic(fmt.Errorf("error building Scheme: %s", err))
	}

	return scheme
}

// Listers returns listers and objects filtered from those listers.
type Listers struct {
	sorter rt.ObjectSorter
}

// NewListers returns a new instance of Listers initialized with the given objects.
func NewListers(scheme *runtime.Scheme, objs []runtime.Object) Listers {
	ls := Listers{
		sorter: rt.NewObjectSorter(scheme),
	}

	ls.sorter.AddObjects(objs...)

	return ls
}

// IndexerFor returns the indexer for the given object.
func (l *Listers) IndexerFor(obj runtime.Object) cache.Indexer {
	return l.sorter.IndexerForObjectType(obj)
}

// GetTriggerMeshObjects returns objects from TriggerMesh APIs.
func (l *Listers) GetTriggerMeshObjects() []runtime.Object {
	return l.sorter.ObjectsForSchemeFunc(fakeclient.AddToScheme)
}

// GetKubeObjects returns objects from Kubernetes APIs.
func (l *Listers) GetKubeObjects() []runtime.Object {
	return l.sorter.ObjectsForSchemeFunc(fakek8sclient.AddToScheme)
}

// GetDeploymentLister returns a lister for Deployment objects.
func (l *Listers) GetDeploymentLister() appslistersv1.DeploymentLister {
	return appslistersv1.NewDeploymentLister(l.IndexerFor(&appsv1.Deployment{}))
}

// GetPodLister returns a lister for Pod objects.
func (l *Listers) GetPodLister() corelistersv1.PodLister {
	return corelistersv1.NewPodLister(l.IndexerFor(&corev1.Pod{}))
}

// GetMemoryBrokerLister returns a Lister for MemoryBroker objects.
func (l *Listers) GetMemoryBrokerLister() eventinglistersv1alpha1.MemoryBrokerLister {
	return eventinglistersv1alpha1.NewMemoryBrokerLister(l.IndexerFor(&eventingv1alpha1.MemoryBroker{}))
}

// GetRedisBrokerLister returns a Lister for RedisBroker objects.
func (l *Listers) GetRedisBrokerLister() eventinglistersv1alpha1.RedisBrokerLister {
	return eventinglistersv1alpha1.NewRedisBrokerLister(l.IndexerFor(&eventingv1alpha1.RedisBroker{}))
}
