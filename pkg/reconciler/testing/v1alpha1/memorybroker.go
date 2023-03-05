package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	eventingv1alpha1 "github.com/triggermesh/triggermesh-core/pkg/apis/eventing/v1alpha1"
	"github.com/triggermesh/triggermesh-core/pkg/reconciler/resources"
)

// MemoryBrokerOption enables further configuration of a v1alpha1.MemoryBroker.
type MemoryBrokerOption func(*eventingv1alpha1.MemoryBroker)

// NewMemoryBroker creates a v1alpha1.MemoryBroker with MemoryBrokerOption .
func NewMemoryBroker(namespace, name string, opts ...MemoryBrokerOption) *eventingv1alpha1.MemoryBroker {
	b := &eventingv1alpha1.MemoryBroker{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Spec: eventingv1alpha1.MemoryBrokerSpec{},
	}

	for _, opt := range opts {
		opt(b)
	}

	return b
}

func MemoryBrokerWithMetaOptions(opts ...resources.MetaOption) MemoryBrokerOption {
	return func(d *eventingv1alpha1.MemoryBroker) {
		for _, opt := range opts {
			opt(&d.ObjectMeta)
		}
	}
}
