// Copyright 2022 TriggerMesh Inc.
// SPDX-License-Identifier: Apache-2.0
// Code generated by injection-gen. DO NOT EDIT.

package trigger

import (
	context "context"

	v1alpha1 "github.com/triggermesh/triggermesh-core/pkg/client/generated/informers/externalversions/eventing/v1alpha1"
	factory "github.com/triggermesh/triggermesh-core/pkg/client/generated/injection/informers/factory"
	controller "knative.dev/pkg/controller"
	injection "knative.dev/pkg/injection"
	logging "knative.dev/pkg/logging"
)

func init() {
	injection.Default.RegisterInformer(withInformer)
}

// Key is used for associating the Informer inside the context.Context.
type Key struct{}

func withInformer(ctx context.Context) (context.Context, controller.Informer) {
	f := factory.Get(ctx)
	inf := f.Eventing().V1alpha1().Triggers()
	return context.WithValue(ctx, Key{}, inf), inf.Informer()
}

// Get extracts the typed informer from the context.
func Get(ctx context.Context) v1alpha1.TriggerInformer {
	untyped := ctx.Value(Key{})
	if untyped == nil {
		logging.FromContext(ctx).Panic(
			"Unable to fetch github.com/triggermesh/triggermesh-core/pkg/client/generated/informers/externalversions/eventing/v1alpha1.TriggerInformer from context.")
	}
	return untyped.(v1alpha1.TriggerInformer)
}
