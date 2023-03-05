package memorybroker

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
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
	testNamespace = "test-namespace"
	testName      = "test-name"
)

var (
	tKey  = testNamespace + "/" + testName
	tTrue = true
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
				tmtv1alpha1.NewMemoryBroker(testNamespace, testName),
			},
			WantCreates: []runtime.Object{
				newSecretForBroker(testNamespace, testName),
				newServiceAccountForBroker(testNamespace, testName),
				newRoleBindingForBroker(testNamespace, testName),
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
			Name:      name + "-mb-broker",
			Labels: map[string]string{
				"app.kubernetes.io/component":  "broker-serviceaccount",
				"app.kubernetes.io/instance":   name + "-mb-broker",
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
			Name:      name + "-mb-broker",
			Labels: map[string]string{
				"app.kubernetes.io/component":  "broker-rolebinding",
				"app.kubernetes.io/instance":   name + "mb-broker",
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
			{Kind: "ServiceAccount", Name: name + "-mb-broker", Namespace: namespace},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     "triggermesh-broker",
		},
	}

	return rb
}
