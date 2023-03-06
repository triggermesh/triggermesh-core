package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	knapis "knative.dev/pkg/apis"

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

func MemoryBrokerWithStatusAddress(url string) MemoryBrokerOption {
	return func(d *eventingv1alpha1.MemoryBroker) {

		pu, err := knapis.ParseURL(url)
		if err != nil {
			panic(err)
		}
		d.Status.Address.URL = pu
	}
}

func MemoryBrokerWithStatusCondition(typ string, status corev1.ConditionStatus, reason, msg string) MemoryBrokerOption {
	return func(d *eventingv1alpha1.MemoryBroker) {

		d.Status.Conditions = append(d.Status.Conditions,
			knapis.Condition{
				Type:    knapis.ConditionType(typ),
				Status:  status,
				Reason:  reason,
				Message: msg,
			},
		)
	}
}
