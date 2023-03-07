package resources

import (
	"strings"

	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewRoleBindingForBroker(namespace, name string, bh BrokerHelper) *rbacv1.RoleBinding {
	rbName := name + "-" + bh.Suffix + "-broker"
	rb := &rbacv1.RoleBinding{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "rbac.authorization.k8s.io/v1",
			Kind:       "RoleBinding",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      rbName,
			Labels: map[string]string{
				"app.kubernetes.io/component":  "broker-rolebinding",
				"app.kubernetes.io/instance":   rbName,
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
		Subjects: []rbacv1.Subject{
			{Kind: "ServiceAccount", Name: rbName, Namespace: namespace},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     "triggermesh-broker",
		},
	}

	return rb
}
