package memorybroker

import (
	"context"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	v1addr "knative.dev/pkg/client/injection/ducks/duck/v1/addressable"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	logtesting "knative.dev/pkg/logging/testing"
	knt "knative.dev/pkg/reconciler/testing"

	fakeeventingclient "github.com/triggermesh/triggermesh-core/pkg/client/generated/injection/client/fake"
	"github.com/triggermesh/triggermesh-core/pkg/client/generated/injection/reconciler/eventing/v1alpha1/memorybroker"
	"github.com/triggermesh/triggermesh-core/pkg/reconciler/common"
	tmt "github.com/triggermesh/triggermesh-core/pkg/reconciler/testing"
	tmtv1alpha1 "github.com/triggermesh/triggermesh-core/pkg/reconciler/testing/v1alpha1"
)

const (
	tBrokerImage  = "image.test:v.test"
	tNamespace    = "test-namespace"
	tName         = "test-name"
	tBrokerSuffix = "-mb-broker"
)

var (
	tKey            = tNamespace + "/" + tName
	tTrue           = true
	tReplicas int32 = 1
)

func TestAllCases(t *testing.T) {
	table := knt.TableTest{
		{
			Name: "bad workqueue key",
			// Make sure Reconcile handles bad keys.
			Key: "too/many/parts",
		}, {
			Name:    "no element found by that name",
			Key:     tKey,
			WantErr: false,
		}, {
			Name: "new broker",
			Key:  tKey,
			Objects: []runtime.Object{
				tmtv1alpha1.NewMemoryBroker(tNamespace, tName),
			},
			WantCreates: []runtime.Object{
				newSecretForBroker(tNamespace, tName),
				newServiceAccountForBroker(tNamespace, tName),
				newRoleBindingForBroker(tNamespace, tName),
				newDeploymentForBroker(tNamespace, tName),
				newServiceForBroker(tNamespace, tName),
			},
			WantEvents: []string{
				knt.Eventf(corev1.EventTypeWarning, "UnavailableEndpoints", `Endpoints for broker service "`+tNamespace+`/`+tName+`-mb-broker" do not exist`),
			},
		},
	}

	logger := logtesting.TestLogger(t)
	table.Test(t, tmt.MakeFactory(func(ctx context.Context, listers *tmt.Listers, cmw configmap.Watcher) controller.Reconciler {
		ctx = v1addr.WithDuck(ctx)
		r := &reconciler{
			secretReconciler: common.NewSecretReconciler(ctx,
				listers.GetSecretLister(),
				listers.GetTriggerLister(),
			),
			saReconciler: common.NewServiceAccountReconciler(ctx,
				listers.GetServiceAccountLister(),
				listers.GetRoleBindingLister(),
			),
			brokerReconciler: common.NewBrokerReconciler(ctx,
				listers.GetDeploymentLister(),
				listers.GetServiceLister(),
				listers.GetEndpointsLister(),
				tBrokerImage, corev1.PullAlways),
		}

		return memorybroker.NewReconciler(ctx, logger,
			fakeeventingclient.Get(ctx),
			listers.GetMemoryBrokerLister(),
			controller.GetEventRecorder(ctx),
			r,
			controller.Options{SkipStatusUpdates: true})
	}, false, logger))
}

func newSecretForBroker(namespace, name string) *corev1.Secret {
	s := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name + "-mb-config",
			Labels: map[string]string{
				"app.kubernetes.io/component":  "broker-config",
				"app.kubernetes.io/instance":   name + "-mb-config",
				"app.kubernetes.io/managed-by": "triggermesh-core",
				"app.kubernetes.io/name":       "memorybroker",
				"app.kubernetes.io/part-of":    "triggermesh",
			},
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion:         "eventing.triggermesh.io/v1alpha1",
					Kind:               "MemoryBroker",
					Name:               name,
					Controller:         &tTrue,
					BlockOwnerDeletion: &tTrue,
				},
			},
		},
		Type: corev1.SecretTypeOpaque,
		Data: map[string][]byte{
			"config": []byte("triggers: {}\n"),
		},
	}

	return s
}

func newServiceAccountForBroker(namespace, name string) *corev1.ServiceAccount {
	sa := &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ServiceAccount",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name + tBrokerSuffix,
			Labels: map[string]string{
				"app.kubernetes.io/component":  "broker-serviceaccount",
				"app.kubernetes.io/instance":   name + tBrokerSuffix,
				"app.kubernetes.io/managed-by": "triggermesh-core",
				"app.kubernetes.io/name":       "memorybroker",
				"app.kubernetes.io/part-of":    "triggermesh",
			},
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion:         "eventing.triggermesh.io/v1alpha1",
					Kind:               "MemoryBroker",
					Name:               name,
					Controller:         &tTrue,
					BlockOwnerDeletion: &tTrue,
				},
			},
		},
	}

	return sa
}

func newRoleBindingForBroker(namespace, name string) *rbacv1.RoleBinding {
	rb := &rbacv1.RoleBinding{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "rbac.authorization.k8s.io/v1",
			Kind:       "RoleBinding",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name + tBrokerSuffix,
			Labels: map[string]string{
				"app.kubernetes.io/component":  "broker-rolebinding",
				"app.kubernetes.io/instance":   name + tBrokerSuffix,
				"app.kubernetes.io/managed-by": "triggermesh-core",
				"app.kubernetes.io/name":       "memorybroker",
				"app.kubernetes.io/part-of":    "triggermesh",
			},
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion:         "eventing.triggermesh.io/v1alpha1",
					Kind:               "MemoryBroker",
					Name:               name,
					Controller:         &tTrue,
					BlockOwnerDeletion: &tTrue,
				},
			},
		},
		Subjects: []rbacv1.Subject{
			{Kind: "ServiceAccount", Name: name + tBrokerSuffix, Namespace: namespace},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     "triggermesh-broker",
		},
	}

	return rb
}

func newServiceForBroker(namespace, name string) *corev1.Service {
	s := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name + tBrokerSuffix,
			Labels: map[string]string{
				"app.kubernetes.io/component":  "broker-service",
				"app.kubernetes.io/instance":   name + tBrokerSuffix,
				"app.kubernetes.io/managed-by": "triggermesh-core",
				"app.kubernetes.io/name":       "memorybroker",
				"app.kubernetes.io/part-of":    "triggermesh",
			},
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion:         "eventing.triggermesh.io/v1alpha1",
					Kind:               "MemoryBroker",
					Name:               name,
					Controller:         &tTrue,
					BlockOwnerDeletion: &tTrue,
				},
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app.kubernetes.io/component": "broker-deployment",
				"app.kubernetes.io/instance":  name + "-mb-broker",
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

func newDeploymentForBroker(namespace, name string) *appsv1.Deployment {
	d := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name + tBrokerSuffix,
			Labels: map[string]string{
				"app.kubernetes.io/component":  "broker-deployment",
				"app.kubernetes.io/instance":   name + tBrokerSuffix,
				"app.kubernetes.io/managed-by": "triggermesh-core",
				"app.kubernetes.io/name":       "memorybroker",
				"app.kubernetes.io/part-of":    "triggermesh",
			},
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion:         "eventing.triggermesh.io/v1alpha1",
					Kind:               "MemoryBroker",
					Name:               name,
					Controller:         &tTrue,
					BlockOwnerDeletion: &tTrue,
				},
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &tReplicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app.kubernetes.io/component": "broker-deployment",
					"app.kubernetes.io/instance":  name + tBrokerSuffix,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app.kubernetes.io/component":  "broker-deployment",
						"app.kubernetes.io/instance":   name + tBrokerSuffix,
						"app.kubernetes.io/managed-by": "triggermesh-core",
						"app.kubernetes.io/part-of":    "triggermesh",
					},
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: name + tBrokerSuffix,
					Containers: []corev1.Container{
						{
							Name:            "broker",
							Image:           tBrokerImage,
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
								{Name: "KUBERNETES_BROKER_CONFIG_SECRET_NAME", Value: name + "-mb-config"},
								{Name: "KUBERNETES_BROKER_CONFIG_SECRET_KEY", Value: "config"},
							},
						},
					},
				},
			},
		},
	}

	return d
}
