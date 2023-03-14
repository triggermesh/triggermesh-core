// Copyright 2022 TriggerMesh Inc.
// SPDX-License-Identifier: Apache-2.0
// Code generated by injection-gen. DO NOT EDIT.

package client

import (
	context "context"
	json "encoding/json"
	errors "errors"
	fmt "fmt"

	v1alpha1 "github.com/triggermesh/triggermesh-core/pkg/apis/eventing/v1alpha1"
	internalclientset "github.com/triggermesh/triggermesh-core/pkg/client/generated/clientset/internalclientset"
	typedeventingv1alpha1 "github.com/triggermesh/triggermesh-core/pkg/client/generated/clientset/internalclientset/typed/eventing/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	unstructured "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	runtime "k8s.io/apimachinery/pkg/runtime"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	discovery "k8s.io/client-go/discovery"
	dynamic "k8s.io/client-go/dynamic"
	rest "k8s.io/client-go/rest"
	injection "knative.dev/pkg/injection"
	dynamicclient "knative.dev/pkg/injection/clients/dynamicclient"
	logging "knative.dev/pkg/logging"
)

func init() {
	injection.Default.RegisterClient(withClientFromConfig)
	injection.Default.RegisterClientFetcher(func(ctx context.Context) interface{} {
		return Get(ctx)
	})
	injection.Dynamic.RegisterDynamicClient(withClientFromDynamic)
}

// Key is used as the key for associating information with a context.Context.
type Key struct{}

func withClientFromConfig(ctx context.Context, cfg *rest.Config) context.Context {
	return context.WithValue(ctx, Key{}, internalclientset.NewForConfigOrDie(cfg))
}

func withClientFromDynamic(ctx context.Context) context.Context {
	return context.WithValue(ctx, Key{}, &wrapClient{dyn: dynamicclient.Get(ctx)})
}

// Get extracts the internalclientset.Interface client from the context.
func Get(ctx context.Context) internalclientset.Interface {
	untyped := ctx.Value(Key{})
	if untyped == nil {
		if injection.GetConfig(ctx) == nil {
			logging.FromContext(ctx).Panic(
				"Unable to fetch github.com/triggermesh/triggermesh-core/pkg/client/generated/clientset/internalclientset.Interface from context. This context is not the application context (which is typically given to constructors via sharedmain).")
		} else {
			logging.FromContext(ctx).Panic(
				"Unable to fetch github.com/triggermesh/triggermesh-core/pkg/client/generated/clientset/internalclientset.Interface from context.")
		}
	}
	return untyped.(internalclientset.Interface)
}

type wrapClient struct {
	dyn dynamic.Interface
}

var _ internalclientset.Interface = (*wrapClient)(nil)

func (w *wrapClient) Discovery() discovery.DiscoveryInterface {
	panic("Discovery called on dynamic client!")
}

func convert(from interface{}, to runtime.Object) error {
	bs, err := json.Marshal(from)
	if err != nil {
		return fmt.Errorf("Marshal() = %w", err)
	}
	if err := json.Unmarshal(bs, to); err != nil {
		return fmt.Errorf("Unmarshal() = %w", err)
	}
	return nil
}

// EventingV1alpha1 retrieves the EventingV1alpha1Client
func (w *wrapClient) EventingV1alpha1() typedeventingv1alpha1.EventingV1alpha1Interface {
	return &wrapEventingV1alpha1{
		dyn: w.dyn,
	}
}

type wrapEventingV1alpha1 struct {
	dyn dynamic.Interface
}

func (w *wrapEventingV1alpha1) RESTClient() rest.Interface {
	panic("RESTClient called on dynamic client!")
}

func (w *wrapEventingV1alpha1) MemoryBrokers(namespace string) typedeventingv1alpha1.MemoryBrokerInterface {
	return &wrapEventingV1alpha1MemoryBrokerImpl{
		dyn: w.dyn.Resource(schema.GroupVersionResource{
			Group:    "eventing.triggermesh.io",
			Version:  "v1alpha1",
			Resource: "memorybrokers",
		}),

		namespace: namespace,
	}
}

type wrapEventingV1alpha1MemoryBrokerImpl struct {
	dyn dynamic.NamespaceableResourceInterface

	namespace string
}

var _ typedeventingv1alpha1.MemoryBrokerInterface = (*wrapEventingV1alpha1MemoryBrokerImpl)(nil)

func (w *wrapEventingV1alpha1MemoryBrokerImpl) Create(ctx context.Context, in *v1alpha1.MemoryBroker, opts v1.CreateOptions) (*v1alpha1.MemoryBroker, error) {
	in.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "eventing.triggermesh.io",
		Version: "v1alpha1",
		Kind:    "MemoryBroker",
	})
	uo := &unstructured.Unstructured{}
	if err := convert(in, uo); err != nil {
		return nil, err
	}
	uo, err := w.dyn.Namespace(w.namespace).Create(ctx, uo, opts)
	if err != nil {
		return nil, err
	}
	out := &v1alpha1.MemoryBroker{}
	if err := convert(uo, out); err != nil {
		return nil, err
	}
	return out, nil
}

func (w *wrapEventingV1alpha1MemoryBrokerImpl) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	return w.dyn.Namespace(w.namespace).Delete(ctx, name, opts)
}

func (w *wrapEventingV1alpha1MemoryBrokerImpl) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	return w.dyn.Namespace(w.namespace).DeleteCollection(ctx, opts, listOpts)
}

func (w *wrapEventingV1alpha1MemoryBrokerImpl) Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha1.MemoryBroker, error) {
	uo, err := w.dyn.Namespace(w.namespace).Get(ctx, name, opts)
	if err != nil {
		return nil, err
	}
	out := &v1alpha1.MemoryBroker{}
	if err := convert(uo, out); err != nil {
		return nil, err
	}
	return out, nil
}

func (w *wrapEventingV1alpha1MemoryBrokerImpl) List(ctx context.Context, opts v1.ListOptions) (*v1alpha1.MemoryBrokerList, error) {
	uo, err := w.dyn.Namespace(w.namespace).List(ctx, opts)
	if err != nil {
		return nil, err
	}
	out := &v1alpha1.MemoryBrokerList{}
	if err := convert(uo, out); err != nil {
		return nil, err
	}
	return out, nil
}

func (w *wrapEventingV1alpha1MemoryBrokerImpl) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.MemoryBroker, err error) {
	uo, err := w.dyn.Namespace(w.namespace).Patch(ctx, name, pt, data, opts)
	if err != nil {
		return nil, err
	}
	out := &v1alpha1.MemoryBroker{}
	if err := convert(uo, out); err != nil {
		return nil, err
	}
	return out, nil
}

func (w *wrapEventingV1alpha1MemoryBrokerImpl) Update(ctx context.Context, in *v1alpha1.MemoryBroker, opts v1.UpdateOptions) (*v1alpha1.MemoryBroker, error) {
	in.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "eventing.triggermesh.io",
		Version: "v1alpha1",
		Kind:    "MemoryBroker",
	})
	uo := &unstructured.Unstructured{}
	if err := convert(in, uo); err != nil {
		return nil, err
	}
	uo, err := w.dyn.Namespace(w.namespace).Update(ctx, uo, opts)
	if err != nil {
		return nil, err
	}
	out := &v1alpha1.MemoryBroker{}
	if err := convert(uo, out); err != nil {
		return nil, err
	}
	return out, nil
}

func (w *wrapEventingV1alpha1MemoryBrokerImpl) UpdateStatus(ctx context.Context, in *v1alpha1.MemoryBroker, opts v1.UpdateOptions) (*v1alpha1.MemoryBroker, error) {
	in.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "eventing.triggermesh.io",
		Version: "v1alpha1",
		Kind:    "MemoryBroker",
	})
	uo := &unstructured.Unstructured{}
	if err := convert(in, uo); err != nil {
		return nil, err
	}
	uo, err := w.dyn.Namespace(w.namespace).UpdateStatus(ctx, uo, opts)
	if err != nil {
		return nil, err
	}
	out := &v1alpha1.MemoryBroker{}
	if err := convert(uo, out); err != nil {
		return nil, err
	}
	return out, nil
}

func (w *wrapEventingV1alpha1MemoryBrokerImpl) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return nil, errors.New("NYI: Watch")
}

func (w *wrapEventingV1alpha1) RedisBrokers(namespace string) typedeventingv1alpha1.RedisBrokerInterface {
	return &wrapEventingV1alpha1RedisBrokerImpl{
		dyn: w.dyn.Resource(schema.GroupVersionResource{
			Group:    "eventing.triggermesh.io",
			Version:  "v1alpha1",
			Resource: "redisbrokers",
		}),

		namespace: namespace,
	}
}

type wrapEventingV1alpha1RedisBrokerImpl struct {
	dyn dynamic.NamespaceableResourceInterface

	namespace string
}

var _ typedeventingv1alpha1.RedisBrokerInterface = (*wrapEventingV1alpha1RedisBrokerImpl)(nil)

func (w *wrapEventingV1alpha1RedisBrokerImpl) Create(ctx context.Context, in *v1alpha1.RedisBroker, opts v1.CreateOptions) (*v1alpha1.RedisBroker, error) {
	in.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "eventing.triggermesh.io",
		Version: "v1alpha1",
		Kind:    "RedisBroker",
	})
	uo := &unstructured.Unstructured{}
	if err := convert(in, uo); err != nil {
		return nil, err
	}
	uo, err := w.dyn.Namespace(w.namespace).Create(ctx, uo, opts)
	if err != nil {
		return nil, err
	}
	out := &v1alpha1.RedisBroker{}
	if err := convert(uo, out); err != nil {
		return nil, err
	}
	return out, nil
}

func (w *wrapEventingV1alpha1RedisBrokerImpl) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	return w.dyn.Namespace(w.namespace).Delete(ctx, name, opts)
}

func (w *wrapEventingV1alpha1RedisBrokerImpl) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	return w.dyn.Namespace(w.namespace).DeleteCollection(ctx, opts, listOpts)
}

func (w *wrapEventingV1alpha1RedisBrokerImpl) Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha1.RedisBroker, error) {
	uo, err := w.dyn.Namespace(w.namespace).Get(ctx, name, opts)
	if err != nil {
		return nil, err
	}
	out := &v1alpha1.RedisBroker{}
	if err := convert(uo, out); err != nil {
		return nil, err
	}
	return out, nil
}

func (w *wrapEventingV1alpha1RedisBrokerImpl) List(ctx context.Context, opts v1.ListOptions) (*v1alpha1.RedisBrokerList, error) {
	uo, err := w.dyn.Namespace(w.namespace).List(ctx, opts)
	if err != nil {
		return nil, err
	}
	out := &v1alpha1.RedisBrokerList{}
	if err := convert(uo, out); err != nil {
		return nil, err
	}
	return out, nil
}

func (w *wrapEventingV1alpha1RedisBrokerImpl) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.RedisBroker, err error) {
	uo, err := w.dyn.Namespace(w.namespace).Patch(ctx, name, pt, data, opts)
	if err != nil {
		return nil, err
	}
	out := &v1alpha1.RedisBroker{}
	if err := convert(uo, out); err != nil {
		return nil, err
	}
	return out, nil
}

func (w *wrapEventingV1alpha1RedisBrokerImpl) Update(ctx context.Context, in *v1alpha1.RedisBroker, opts v1.UpdateOptions) (*v1alpha1.RedisBroker, error) {
	in.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "eventing.triggermesh.io",
		Version: "v1alpha1",
		Kind:    "RedisBroker",
	})
	uo := &unstructured.Unstructured{}
	if err := convert(in, uo); err != nil {
		return nil, err
	}
	uo, err := w.dyn.Namespace(w.namespace).Update(ctx, uo, opts)
	if err != nil {
		return nil, err
	}
	out := &v1alpha1.RedisBroker{}
	if err := convert(uo, out); err != nil {
		return nil, err
	}
	return out, nil
}

func (w *wrapEventingV1alpha1RedisBrokerImpl) UpdateStatus(ctx context.Context, in *v1alpha1.RedisBroker, opts v1.UpdateOptions) (*v1alpha1.RedisBroker, error) {
	in.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "eventing.triggermesh.io",
		Version: "v1alpha1",
		Kind:    "RedisBroker",
	})
	uo := &unstructured.Unstructured{}
	if err := convert(in, uo); err != nil {
		return nil, err
	}
	uo, err := w.dyn.Namespace(w.namespace).UpdateStatus(ctx, uo, opts)
	if err != nil {
		return nil, err
	}
	out := &v1alpha1.RedisBroker{}
	if err := convert(uo, out); err != nil {
		return nil, err
	}
	return out, nil
}

func (w *wrapEventingV1alpha1RedisBrokerImpl) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return nil, errors.New("NYI: Watch")
}

func (w *wrapEventingV1alpha1) RedisReplays(namespace string) typedeventingv1alpha1.RedisReplayInterface {
	return &wrapEventingV1alpha1RedisReplayImpl{
		dyn: w.dyn.Resource(schema.GroupVersionResource{
			Group:    "eventing.triggermesh.io",
			Version:  "v1alpha1",
			Resource: "redisreplays",
		}),

		namespace: namespace,
	}
}

type wrapEventingV1alpha1RedisReplayImpl struct {
	dyn dynamic.NamespaceableResourceInterface

	namespace string
}

var _ typedeventingv1alpha1.RedisReplayInterface = (*wrapEventingV1alpha1RedisReplayImpl)(nil)

func (w *wrapEventingV1alpha1RedisReplayImpl) Create(ctx context.Context, in *v1alpha1.RedisReplay, opts v1.CreateOptions) (*v1alpha1.RedisReplay, error) {
	in.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "eventing.triggermesh.io",
		Version: "v1alpha1",
		Kind:    "RedisReplay",
	})
	uo := &unstructured.Unstructured{}
	if err := convert(in, uo); err != nil {
		return nil, err
	}
	uo, err := w.dyn.Namespace(w.namespace).Create(ctx, uo, opts)
	if err != nil {
		return nil, err
	}
	out := &v1alpha1.RedisReplay{}
	if err := convert(uo, out); err != nil {
		return nil, err
	}
	return out, nil
}

func (w *wrapEventingV1alpha1RedisReplayImpl) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	return w.dyn.Namespace(w.namespace).Delete(ctx, name, opts)
}

func (w *wrapEventingV1alpha1RedisReplayImpl) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	return w.dyn.Namespace(w.namespace).DeleteCollection(ctx, opts, listOpts)
}

func (w *wrapEventingV1alpha1RedisReplayImpl) Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha1.RedisReplay, error) {
	uo, err := w.dyn.Namespace(w.namespace).Get(ctx, name, opts)
	if err != nil {
		return nil, err
	}
	out := &v1alpha1.RedisReplay{}
	if err := convert(uo, out); err != nil {
		return nil, err
	}
	return out, nil
}

func (w *wrapEventingV1alpha1RedisReplayImpl) List(ctx context.Context, opts v1.ListOptions) (*v1alpha1.RedisReplayList, error) {
	uo, err := w.dyn.Namespace(w.namespace).List(ctx, opts)
	if err != nil {
		return nil, err
	}
	out := &v1alpha1.RedisReplayList{}
	if err := convert(uo, out); err != nil {
		return nil, err
	}
	return out, nil
}

func (w *wrapEventingV1alpha1RedisReplayImpl) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.RedisReplay, err error) {
	uo, err := w.dyn.Namespace(w.namespace).Patch(ctx, name, pt, data, opts)
	if err != nil {
		return nil, err
	}
	out := &v1alpha1.RedisReplay{}
	if err := convert(uo, out); err != nil {
		return nil, err
	}
	return out, nil
}

func (w *wrapEventingV1alpha1RedisReplayImpl) Update(ctx context.Context, in *v1alpha1.RedisReplay, opts v1.UpdateOptions) (*v1alpha1.RedisReplay, error) {
	in.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "eventing.triggermesh.io",
		Version: "v1alpha1",
		Kind:    "RedisReplay",
	})
	uo := &unstructured.Unstructured{}
	if err := convert(in, uo); err != nil {
		return nil, err
	}
	uo, err := w.dyn.Namespace(w.namespace).Update(ctx, uo, opts)
	if err != nil {
		return nil, err
	}
	out := &v1alpha1.RedisReplay{}
	if err := convert(uo, out); err != nil {
		return nil, err
	}
	return out, nil
}

func (w *wrapEventingV1alpha1RedisReplayImpl) UpdateStatus(ctx context.Context, in *v1alpha1.RedisReplay, opts v1.UpdateOptions) (*v1alpha1.RedisReplay, error) {
	in.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "eventing.triggermesh.io",
		Version: "v1alpha1",
		Kind:    "RedisReplay",
	})
	uo := &unstructured.Unstructured{}
	if err := convert(in, uo); err != nil {
		return nil, err
	}
	uo, err := w.dyn.Namespace(w.namespace).UpdateStatus(ctx, uo, opts)
	if err != nil {
		return nil, err
	}
	out := &v1alpha1.RedisReplay{}
	if err := convert(uo, out); err != nil {
		return nil, err
	}
	return out, nil
}

func (w *wrapEventingV1alpha1RedisReplayImpl) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return nil, errors.New("NYI: Watch")
}

func (w *wrapEventingV1alpha1) Triggers(namespace string) typedeventingv1alpha1.TriggerInterface {
	return &wrapEventingV1alpha1TriggerImpl{
		dyn: w.dyn.Resource(schema.GroupVersionResource{
			Group:    "eventing.triggermesh.io",
			Version:  "v1alpha1",
			Resource: "triggers",
		}),

		namespace: namespace,
	}
}

type wrapEventingV1alpha1TriggerImpl struct {
	dyn dynamic.NamespaceableResourceInterface

	namespace string
}

var _ typedeventingv1alpha1.TriggerInterface = (*wrapEventingV1alpha1TriggerImpl)(nil)

func (w *wrapEventingV1alpha1TriggerImpl) Create(ctx context.Context, in *v1alpha1.Trigger, opts v1.CreateOptions) (*v1alpha1.Trigger, error) {
	in.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "eventing.triggermesh.io",
		Version: "v1alpha1",
		Kind:    "Trigger",
	})
	uo := &unstructured.Unstructured{}
	if err := convert(in, uo); err != nil {
		return nil, err
	}
	uo, err := w.dyn.Namespace(w.namespace).Create(ctx, uo, opts)
	if err != nil {
		return nil, err
	}
	out := &v1alpha1.Trigger{}
	if err := convert(uo, out); err != nil {
		return nil, err
	}
	return out, nil
}

func (w *wrapEventingV1alpha1TriggerImpl) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	return w.dyn.Namespace(w.namespace).Delete(ctx, name, opts)
}

func (w *wrapEventingV1alpha1TriggerImpl) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	return w.dyn.Namespace(w.namespace).DeleteCollection(ctx, opts, listOpts)
}

func (w *wrapEventingV1alpha1TriggerImpl) Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha1.Trigger, error) {
	uo, err := w.dyn.Namespace(w.namespace).Get(ctx, name, opts)
	if err != nil {
		return nil, err
	}
	out := &v1alpha1.Trigger{}
	if err := convert(uo, out); err != nil {
		return nil, err
	}
	return out, nil
}

func (w *wrapEventingV1alpha1TriggerImpl) List(ctx context.Context, opts v1.ListOptions) (*v1alpha1.TriggerList, error) {
	uo, err := w.dyn.Namespace(w.namespace).List(ctx, opts)
	if err != nil {
		return nil, err
	}
	out := &v1alpha1.TriggerList{}
	if err := convert(uo, out); err != nil {
		return nil, err
	}
	return out, nil
}

func (w *wrapEventingV1alpha1TriggerImpl) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.Trigger, err error) {
	uo, err := w.dyn.Namespace(w.namespace).Patch(ctx, name, pt, data, opts)
	if err != nil {
		return nil, err
	}
	out := &v1alpha1.Trigger{}
	if err := convert(uo, out); err != nil {
		return nil, err
	}
	return out, nil
}

func (w *wrapEventingV1alpha1TriggerImpl) Update(ctx context.Context, in *v1alpha1.Trigger, opts v1.UpdateOptions) (*v1alpha1.Trigger, error) {
	in.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "eventing.triggermesh.io",
		Version: "v1alpha1",
		Kind:    "Trigger",
	})
	uo := &unstructured.Unstructured{}
	if err := convert(in, uo); err != nil {
		return nil, err
	}
	uo, err := w.dyn.Namespace(w.namespace).Update(ctx, uo, opts)
	if err != nil {
		return nil, err
	}
	out := &v1alpha1.Trigger{}
	if err := convert(uo, out); err != nil {
		return nil, err
	}
	return out, nil
}

func (w *wrapEventingV1alpha1TriggerImpl) UpdateStatus(ctx context.Context, in *v1alpha1.Trigger, opts v1.UpdateOptions) (*v1alpha1.Trigger, error) {
	in.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "eventing.triggermesh.io",
		Version: "v1alpha1",
		Kind:    "Trigger",
	})
	uo := &unstructured.Unstructured{}
	if err := convert(in, uo); err != nil {
		return nil, err
	}
	uo, err := w.dyn.Namespace(w.namespace).UpdateStatus(ctx, uo, opts)
	if err != nil {
		return nil, err
	}
	out := &v1alpha1.Trigger{}
	if err := convert(uo, out); err != nil {
		return nil, err
	}
	return out, nil
}

func (w *wrapEventingV1alpha1TriggerImpl) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return nil, errors.New("NYI: Watch")
}
