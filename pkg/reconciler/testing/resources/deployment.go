package resources

import (
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/triggermesh/triggermesh-core/pkg/reconciler/resources"
)

func NewDeploymentForBroker(namespace, name string, bh BrokerHelper, opts ...resources.DeploymentOption) *appsv1.Deployment {
	deploymentName := name + "-" + bh.Suffix + "-broker"
	configName := name + "-" + bh.Suffix + "-config"
	statusConfigName := name + "-" + bh.Suffix + "-status"

	d := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      deploymentName,
			Labels: map[string]string{
				"app.kubernetes.io/component":  "broker-deployment",
				"app.kubernetes.io/instance":   deploymentName,
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
		Spec: appsv1.DeploymentSpec{
			Replicas: &TestReplicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app.kubernetes.io/component": "broker-deployment",
					"app.kubernetes.io/instance":  deploymentName,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app.kubernetes.io/component":  "broker-deployment",
						"app.kubernetes.io/instance":   deploymentName,
						"app.kubernetes.io/managed-by": "triggermesh-core",
						"app.kubernetes.io/part-of":    "triggermesh",
					},
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: deploymentName,
					Containers: []corev1.Container{
						{
							Name:            "broker",
							Image:           TestBrokerImage,
							Args:            []string{"start"},
							ImagePullPolicy: corev1.PullAlways,
							Ports: []corev1.ContainerPort{
								{
									Name:          "httpce",
									ContainerPort: 8080,
								},
								{
									Name:          "metrics",
									ContainerPort: 9090,
								},
							},
							Env: []corev1.EnvVar{
								{Name: "PORT", Value: "8080"},
								{Name: "BROKER_NAME", Value: name},
								{
									Name: "KUBERNETES_NAMESPACE",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.namespace",
										},
									},
								},
								{Name: "KUBERNETES_BROKER_CONFIG_SECRET_NAME", Value: configName},
								{Name: "KUBERNETES_BROKER_CONFIG_SECRET_KEY", Value: "config"},
								{Name: "KUBERNETES_STATUS_CONFIGMAP_NAME", Value: statusConfigName},
							},
						},
					},
				},
			},
		},
	}

	for _, opt := range opts {
		opt(d)
	}

	return d
}

func WithDeploymentReady() resources.DeploymentOption {
	return func(d *appsv1.Deployment) {
		d.Status.Conditions = append(d.Status.Conditions, appsv1.DeploymentCondition{
			Type:   appsv1.DeploymentAvailable,
			Status: corev1.ConditionTrue,
		})
	}
}
