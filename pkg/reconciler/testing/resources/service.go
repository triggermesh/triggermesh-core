package resources

import (
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func NewServiceForBroker(namespace, name string, bh BrokerHelper) *corev1.Service {
	serviceName := name + "-" + bh.Suffix + "-broker"

	s := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      serviceName,
			Labels: map[string]string{
				"app.kubernetes.io/component":  "broker-service",
				"app.kubernetes.io/instance":   serviceName,
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
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app.kubernetes.io/component": "broker-deployment",
				"app.kubernetes.io/instance":  serviceName,
			},
			Ports: []corev1.ServicePort{
				{
					Name: "httpce",
					Port: 80,
					TargetPort: intstr.IntOrString{
						IntVal: 8080,
					},
				},
			},
			Type: corev1.ServiceTypeClusterIP,
		},
	}

	return s
}
