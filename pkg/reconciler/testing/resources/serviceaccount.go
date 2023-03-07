package resources

import (
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewServiceAccountForBroker(namespace, name string, bh BrokerHelper) *corev1.ServiceAccount {
	saName := name + "-" + bh.Suffix + "-broker"

	sa := &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ServiceAccount",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      saName,
			Labels: map[string]string{
				"app.kubernetes.io/component":  "broker-serviceaccount",
				"app.kubernetes.io/instance":   saName,
				"app.kubernetes.io/managed-by": "triggermesh-core",
				"app.kubernetes.io/name":       strings.ToLower(bh.Kind),
				"app.kubernetes.io/part-of":    "triggermesh",
			},
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion:         "eventing.triggermesh.io/v1alpha1",
					Kind:               bh.Kind,
					Name:               name,
					Controller:         &TestTrue,
					BlockOwnerDeletion: &TestTrue,
				},
			},
		},
	}

	return sa
}
