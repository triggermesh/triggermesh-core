// Copyright 2023 TriggerMesh Inc.
// SPDX-License-Identifier: Apache-2.0

/*
Copyright 2019 The Knative Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package testing

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
	appsv1listers "k8s.io/client-go/listers/apps/v1"
	corev1listers "k8s.io/client-go/listers/core/v1"
	rbacv1listers "k8s.io/client-go/listers/rbac/v1"
	"k8s.io/client-go/tools/cache"

	// fakeeventingclientset "knative.dev/eventing/pkg/client/clientset/versioned/fake"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	rt "knative.dev/pkg/reconciler/testing"

	fakekubeclientset "k8s.io/client-go/kubernetes/fake"

	eventingv1alpha1 "github.com/triggermesh/triggermesh-core/pkg/apis/eventing/v1alpha1"
	fakeclient "github.com/triggermesh/triggermesh-core/pkg/client/generated/clientset/internalclientset/fake"

	// fakeeventingclientset "github.com/triggermesh/triggermesh-core/pkg/client/generated/clientset/versioned/fake"
	eventinglistersv1alpha1 "github.com/triggermesh/triggermesh-core/pkg/client/generated/listers/eventing/v1alpha1"
)

var clientSetSchemes = []func(*runtime.Scheme) error{
	fakekubeclientset.AddToScheme,
	fakeclient.AddToScheme,
	duckv1.AddToScheme,
}

// Listers returns listers and objects filtered from those listers.
type Listers struct {
	sorter rt.ObjectSorter
}

// NewScheme returns a new scheme populated with the types defined in clientSetSchemes.
func NewScheme() *runtime.Scheme {
	scheme := runtime.NewScheme()

	for _, addTo := range clientSetSchemes {
		if err := addTo(scheme); err != nil {
			panic(err)
		}
	}
	return scheme
}

// NewListers returns a new instance of Listers initialized with the given objects.
func NewListers(objs []runtime.Object) Listers {
	scheme := runtime.NewScheme()

	for _, addTo := range clientSetSchemes {
		if err := addTo(scheme); err != nil {
			panic(err)
		}

	}

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

// GetKubeObjects returns objects from Kubernetes APIs.
func (l *Listers) GetKubeObjects() []runtime.Object {
	return l.sorter.ObjectsForSchemeFunc(fakekubeclientset.AddToScheme)
}

// GetTriggerMeshObjects returns objects from TriggerMesh APIs.
func (l *Listers) GetTriggerMeshObjects() []runtime.Object {
	return l.sorter.ObjectsForSchemeFunc(fakeclient.AddToScheme)
}

// GetDeploymentLister returns a lister for Deployment objects.
func (l *Listers) GetDeploymentLister() appsv1listers.DeploymentLister {
	return appsv1listers.NewDeploymentLister(l.IndexerFor(&appsv1.Deployment{}))
}

// GetSecretLister returns a lister for Secret objects.
func (l *Listers) GetSecretLister() corev1listers.SecretLister {
	return corev1listers.NewSecretLister(l.IndexerFor(&corev1.Secret{}))
}

// GetConfigMapLister returns a lister for ConfigMap objects.
func (l *Listers) GetConfigMapLister() corev1listers.ConfigMapLister {
	return corev1listers.NewConfigMapLister(l.IndexerFor(&corev1.ConfigMap{}))
}

// GetPodLister returns a lister for Pod objects.
func (l *Listers) GetPodLister() corev1listers.PodLister {
	return corev1listers.NewPodLister(l.IndexerFor(&corev1.Pod{}))
}

// GetServiceLister returns a lister for Service objects.
func (l *Listers) GetServiceLister() corev1listers.ServiceLister {
	return corev1listers.NewServiceLister(l.IndexerFor(&corev1.Service{}))
}

// GetEndpointsLister returns a lister for Endpoint objects.
func (l *Listers) GetEndpointsLister() corev1listers.EndpointsLister {
	return corev1listers.NewEndpointsLister(l.IndexerFor(&corev1.Endpoints{}))
}

// GetServiceAccountLister returns a lister for ServiceAccount objects.
func (l *Listers) GetServiceAccountLister() corev1listers.ServiceAccountLister {
	return corev1listers.NewServiceAccountLister(l.IndexerFor(&corev1.ServiceAccount{}))
}

// GetRoleBindingLister returns a lister for RoleBinding objects.
func (l *Listers) GetRoleBindingLister() rbacv1listers.RoleBindingLister {
	return rbacv1listers.NewRoleBindingLister(l.IndexerFor(&rbacv1.RoleBinding{}))
}

// GetMemoryBrokerLister returns a Lister for MemoryBroker objects.
func (l *Listers) GetMemoryBrokerLister() eventinglistersv1alpha1.MemoryBrokerLister {
	return eventinglistersv1alpha1.NewMemoryBrokerLister(l.IndexerFor(&eventingv1alpha1.MemoryBroker{}))
}

// GetRedisBrokerLister returns a Lister for RedisBroker objects.
func (l *Listers) GetRedisBrokerLister() eventinglistersv1alpha1.RedisBrokerLister {
	return eventinglistersv1alpha1.NewRedisBrokerLister(l.IndexerFor(&eventingv1alpha1.RedisBroker{}))
}

// GetTriggerLister returns a Lister for Trigger objects.
func (l *Listers) GetTriggerLister() eventinglistersv1alpha1.TriggerLister {
	return eventinglistersv1alpha1.NewTriggerLister(l.IndexerFor(&eventingv1alpha1.Trigger{}))
}
