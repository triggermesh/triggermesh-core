package memorybroker

import (
	"context"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kt "k8s.io/client-go/testing"

	v1addr "knative.dev/pkg/client/injection/ducks/duck/v1/addressable"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	logtesting "knative.dev/pkg/logging/testing"
	knt "knative.dev/pkg/reconciler/testing"

	fakeeventingclient "github.com/triggermesh/triggermesh-core/pkg/client/generated/injection/client/fake"
	"github.com/triggermesh/triggermesh-core/pkg/client/generated/injection/reconciler/eventing/v1alpha1/memorybroker"
	"github.com/triggermesh/triggermesh-core/pkg/reconciler/common"
	"github.com/triggermesh/triggermesh-core/pkg/reconciler/resources"
	tmt "github.com/triggermesh/triggermesh-core/pkg/reconciler/testing"
	tresources "github.com/triggermesh/triggermesh-core/pkg/reconciler/testing/resources"
	tmtv1alpha1 "github.com/triggermesh/triggermesh-core/pkg/reconciler/testing/v1alpha1"
)

var (
	tKey  = tresources.TestNamespace + "/" + tresources.TestName
	tTrue = true
	tNow  = metav1.NewTime(time.Now())
)

func TestAllCases(t *testing.T) {
	bh := tresources.BrokerHelper{
		Suffix: "mb",
		Kind:   "MemoryBroker",
	}

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
				tmtv1alpha1.NewMemoryBroker(tresources.TestNamespace, tresources.TestName),
			},
			WantCreates: []runtime.Object{
				newSecretForBroker(tresources.TestNamespace, tresources.TestName),
				tresources.NewServiceAccountForBroker(tresources.TestNamespace, tresources.TestName, bh),
				tresources.NewRoleBindingForBroker(tresources.TestNamespace, tresources.TestName, bh),
				tresources.NewDeploymentForBroker(tresources.TestNamespace, tresources.TestName, bh),
				tresources.NewServiceForBroker(tresources.TestNamespace, tresources.TestName, bh),
			},
			WantStatusUpdates: []kt.UpdateActionImpl{
				{
					Object: tmtv1alpha1.NewMemoryBroker(tresources.TestNamespace, tresources.TestName,
						tmtv1alpha1.MemoryBrokerWithStatusCondition("Addressable", corev1.ConditionUnknown, "", ""),
						tmtv1alpha1.MemoryBrokerWithStatusCondition("BrokerConfigSecretReady", corev1.ConditionTrue, "", ""),
						tmtv1alpha1.MemoryBrokerWithStatusCondition("BrokerDeploymentReady", corev1.ConditionUnknown, "", ""),
						tmtv1alpha1.MemoryBrokerWithStatusCondition("BrokerEndpointsReady", corev1.ConditionFalse, "UnavailableEndpoints", "Endpoints for broker service do not exist"),
						tmtv1alpha1.MemoryBrokerWithStatusCondition("BrokerServiceAccountReady", corev1.ConditionTrue, "", ""),
						tmtv1alpha1.MemoryBrokerWithStatusCondition("BrokerServiceReady", corev1.ConditionTrue, "", ""),
						tmtv1alpha1.MemoryBrokerWithStatusCondition("MemoryBrokerBrokerRoleBinding", corev1.ConditionTrue, "", ""),
						tmtv1alpha1.MemoryBrokerWithStatusCondition("Ready", corev1.ConditionFalse, "UnavailableEndpoints", "Endpoints for broker service do not exist"),
					),
				},
			},
			WantEvents: []string{
				knt.Eventf(corev1.EventTypeWarning, "UnavailableEndpoints", `Endpoints for broker service "`+tresources.TestNamespace+`/`+tresources.TestName+`-mb-broker" do not exist`),
			},
		}, {
			Name: "update status",
			Key:  tKey,
			Objects: []runtime.Object{
				tmtv1alpha1.NewMemoryBroker(tresources.TestNamespace, tresources.TestName,
					tmtv1alpha1.MemoryBrokerWithStatusCondition("Addressable", corev1.ConditionUnknown, "", ""),
					tmtv1alpha1.MemoryBrokerWithStatusCondition("BrokerConfigSecretReady", corev1.ConditionTrue, "", ""),
					tmtv1alpha1.MemoryBrokerWithStatusCondition("BrokerDeploymentReady", corev1.ConditionUnknown, "", ""),
					tmtv1alpha1.MemoryBrokerWithStatusCondition("BrokerEndpointsReady", corev1.ConditionFalse, "UnavailableEndpoints", "Endpoints for broker service do not exist"),
					tmtv1alpha1.MemoryBrokerWithStatusCondition("BrokerServiceAccountReady", corev1.ConditionTrue, "", ""),
					tmtv1alpha1.MemoryBrokerWithStatusCondition("BrokerServiceReady", corev1.ConditionTrue, "", ""),
					tmtv1alpha1.MemoryBrokerWithStatusCondition("MemoryBrokerBrokerRoleBinding", corev1.ConditionTrue, "", ""),
					tmtv1alpha1.MemoryBrokerWithStatusCondition("Ready", corev1.ConditionFalse, "UnavailableEndpoints", "Endpoints for broker service do not exist"),
				),
				newSecretForBroker(tresources.TestNamespace, tresources.TestName),
				tresources.NewServiceAccountForBroker(tresources.TestNamespace, tresources.TestName, bh),
				tresources.NewRoleBindingForBroker(tresources.TestNamespace, tresources.TestName, bh),
				tresources.NewDeploymentForBroker(tresources.TestNamespace, tresources.TestName, bh, tresources.WithDeploymentReady()),
				tresources.NewServiceForBroker(tresources.TestNamespace, tresources.TestName, bh),
				tresources.NewEndpointForBroker(tresources.TestNamespace, tresources.TestName, bh),
			},
			WantStatusUpdates: []kt.UpdateActionImpl{
				{
					Object: tmtv1alpha1.NewMemoryBroker(tresources.TestNamespace, tresources.TestName,
						tmtv1alpha1.MemoryBrokerWithStatusCondition("Addressable", corev1.ConditionTrue, "", ""),
						tmtv1alpha1.MemoryBrokerWithStatusCondition("BrokerConfigSecretReady", corev1.ConditionTrue, "", ""),
						tmtv1alpha1.MemoryBrokerWithStatusCondition("BrokerDeploymentReady", corev1.ConditionTrue, "", ""),
						tmtv1alpha1.MemoryBrokerWithStatusCondition("BrokerEndpointsReady", corev1.ConditionTrue, "", ""),
						tmtv1alpha1.MemoryBrokerWithStatusCondition("BrokerServiceAccountReady", corev1.ConditionTrue, "", ""),
						tmtv1alpha1.MemoryBrokerWithStatusCondition("BrokerServiceReady", corev1.ConditionTrue, "", ""),
						tmtv1alpha1.MemoryBrokerWithStatusCondition("MemoryBrokerBrokerRoleBinding", corev1.ConditionTrue, "", ""),
						tmtv1alpha1.MemoryBrokerWithStatusCondition("Ready", corev1.ConditionTrue, "", ""),
						tmtv1alpha1.MemoryBrokerWithStatusAddress("http://"+tresources.TestName+"-mb-broker."+tresources.TestNamespace+".svc.cluster.local"),
					),
				},
			},
		}, {
			Name: "deleting broker",
			Key:  tKey,
			Objects: []runtime.Object{
				tmtv1alpha1.NewMemoryBroker(tresources.TestNamespace, tresources.TestName,
					tmtv1alpha1.MemoryBrokerWithMetaOptions(resources.MetaSetDeletion(&tNow))),
			},
			WantCreates: []runtime.Object{
				// Reconciliation is skipped and no objects are created.
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
				tresources.TestBrokerImage, corev1.PullAlways),
		}

		return memorybroker.NewReconciler(ctx, logger,
			fakeeventingclient.Get(ctx),
			listers.GetMemoryBrokerLister(),
			controller.GetEventRecorder(ctx),
			r,
			controller.Options{SkipStatusUpdates: false})
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
