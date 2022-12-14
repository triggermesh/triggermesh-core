// Copyright 2022 TriggerMesh Inc.
// SPDX-License-Identifier: Apache-2.0
// Code generated by injection-gen. DO NOT EDIT.

package fake

import (
	context "context"

	memorybroker "github.com/triggermesh/triggermesh-core/pkg/client/generated/injection/informers/eventing/v1alpha1/memorybroker"
	fake "github.com/triggermesh/triggermesh-core/pkg/client/generated/injection/informers/factory/fake"
	controller "knative.dev/pkg/controller"
	injection "knative.dev/pkg/injection"
)

var Get = memorybroker.Get

func init() {
	injection.Fake.RegisterInformer(withInformer)
}

func withInformer(ctx context.Context) (context.Context, controller.Informer) {
	f := fake.Get(ctx)
	inf := f.Eventing().V1alpha1().MemoryBrokers()
	return context.WithValue(ctx, memorybroker.Key{}, inf), inf.Informer()
}
