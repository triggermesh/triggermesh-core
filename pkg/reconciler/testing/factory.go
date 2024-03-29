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
	"context"
	"encoding/json"
	"testing"

	"go.uber.org/zap"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ktesting "k8s.io/client-go/testing"
	"k8s.io/client-go/tools/record"

	fakekubeclient "knative.dev/pkg/client/injection/kube/client/fake"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	fakedynamicclient "knative.dev/pkg/injection/clients/dynamicclient/fake"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/reconciler"
	rt "knative.dev/pkg/reconciler/testing"

	fakeinjectionclient "github.com/triggermesh/triggermesh-core/pkg/client/generated/injection/client/fake"
)

const (
	// maxEventBufferSize is the estimated max number of event notifications that
	// can be buffered during reconciliation.
	maxEventBufferSize = 10
)

// Ctor functions create a k8s controller with given params.
type Ctor func(context.Context, *Listers, configmap.Watcher) controller.Reconciler

// MakeFactory creates a testing factory for our controller.Reconciler, and
// initializes a Reconciler using the given Ctor as part of the process.
func MakeFactory(ctor Ctor, unstructured bool, logger *zap.SugaredLogger) rt.Factory {
	return func(t *testing.T, tr *rt.TableRow) (controller.Reconciler, rt.ActionRecorderList, rt.EventList) {
		ls := NewListers(tr.Objects)

		// enable values injected by the test case (TableRow) to be consumed in ctor
		ctx := tr.Ctx
		if ctx == nil {
			ctx = context.Background()
		}
		ctx = logging.WithLogger(ctx, logger)

		ctx, kubeClient := fakekubeclient.With(ctx, ls.GetKubeObjects()...)
		ctx, client := fakeinjectionclient.With(ctx, ls.GetTriggerMeshObjects()...)
		ctx, dynamicClient := fakedynamicclient.With(ctx,
			NewScheme(), ToUnstructured(t, tr.Objects)...)

		// The dynamic client's support for patching is BS.  Implement it
		// here via PrependReactor (this can be overridden below by the
		// provided reactors).
		dynamicClient.PrependReactor("patch", "*",
			func(action ktesting.Action) (bool, runtime.Object, error) {
				return true, nil, nil
			})

		eventRecorder := record.NewFakeRecorder(maxEventBufferSize)
		ctx = controller.WithEventRecorder(ctx, eventRecorder)

		// Check the config maps in objects and add them to the fake cm watcher
		var cms []*corev1.ConfigMap
		for _, obj := range tr.Objects {
			if cm, ok := obj.(*corev1.ConfigMap); ok {
				cms = append(cms, cm)
			}
		}
		configMapWatcher := configmap.NewStaticWatcher(cms...)

		// Set up our Controller from the fakes.
		c := ctor(ctx, &ls, configMapWatcher)

		// If the reconcilers is leader aware, then promote it.
		if la, ok := c.(reconciler.LeaderAware); ok {
			if err := la.Promote(reconciler.UniversalBucket(), func(reconciler.Bucket, types.NamespacedName) {}); err != nil {
				panic(err)
			}
		}

		// Validate all Create operations through the eventing client.
		client.PrependReactor("create", "*", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
			return rt.ValidateCreates(ctx, action)
		})
		client.PrependReactor("update", "*", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
			return rt.ValidateUpdates(ctx, action)
		})

		for _, reactor := range tr.WithReactors {
			kubeClient.PrependReactor("*", "*", reactor)
			client.PrependReactor("*", "*", reactor)
			dynamicClient.PrependReactor("*", "*", reactor)
		}

		actionRecorderList := rt.ActionRecorderList{dynamicClient, client, kubeClient}
		eventList := rt.EventList{Recorder: eventRecorder}

		return c, actionRecorderList, eventList
	}
}

// ToUnstructured takes a list of k8s resources and converts them to
// Unstructured objects.
// We must pass objects as Unstructured to the dynamic client fake, or it
// won't handle them properly.
func ToUnstructured(t *testing.T, objs []runtime.Object) (us []runtime.Object) {
	sch := NewScheme()
	for _, obj := range objs {
		obj = obj.DeepCopyObject() // Don't mess with the primary copy
		// Determine and set the TypeMeta for this object based on our test scheme.
		gvks, _, err := sch.ObjectKinds(obj)
		if err != nil {
			t.Fatal("Unable to determine kind for type:", err)
		}
		apiv, k := gvks[0].ToAPIVersionAndKind()
		ta, err := meta.TypeAccessor(obj)
		if err != nil {
			t.Fatal("Unable to create type accessor:", err)
		}
		ta.SetAPIVersion(apiv)
		ta.SetKind(k)

		b, err := json.Marshal(obj)
		if err != nil {
			t.Fatal("Unable to marshal:", err)
		}
		u := &unstructured.Unstructured{}
		if err := json.Unmarshal(b, u); err != nil {
			t.Fatal("Unable to unmarshal:", err)
		}
		us = append(us, u)
	}
	return
}
