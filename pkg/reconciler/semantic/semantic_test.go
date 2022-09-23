// Copyright 2022 TriggerMesh Inc.
// SPDX-License-Identifier: Apache-2.0

package semantic

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	tServiceAccount = `
{
	"apiVersion": "v1",
	"kind": "ServiceAccount",
	"metadata": {
		"annotations": {
			"eks.amazonaws.com/role-arn": "arn:aws:iam::123456789012:role/test"
		},
		"creationTimestamp": "2021-11-05T15:31:03Z",
		"labels": {
			"app.kubernetes.io/component": "adapter",
			"app.kubernetes.io/managed-by": "triggermesh-controller",
			"app.kubernetes.io/name": "foosource",
			"app.kubernetes.io/part-of": "triggermesh"
		},
		"name": "foosource-sample",
		"namespace": "dev",
		"ownerReferences": [
			{
				"apiVersion": "sources.triggermesh.io/v1alpha1",
				"blockOwnerDeletion": true,
				"controller": false,
				"kind": "FooSource",
				"name": "sample",
				"uid": "6627fb5f-c220-4ef6-a7a2-edb7e1af6544"
			}
		],
		"resourceVersion": "7696258",
		"uid": "60c2e013-4c52-4364-8104-6ead0b9d4975"
	},
	"secrets": [
		{
			"name": "foosource-i-sample-token-bc5pb"
		}
	]
}
`
	tDeployment = `
{
    "apiVersion": "apps/v1",
    "kind": "Deployment",
    "metadata": {
        "annotations": {
            "deployment.kubernetes.io/revision": "1"
        },
        "creationTimestamp": "2020-04-26T13:23:03Z",
        "generation": 1,
        "labels": {
            "app.kubernetes.io/component": "adapter",
            "app.kubernetes.io/instance": "sample",
            "app.kubernetes.io/managed-by": "foo-sources-controller",
            "app.kubernetes.io/name": "foosource",
            "app.kubernetes.io/part-of": "foo-sources"
        },
        "name": "foosource-sample",
        "namespace": "dev",
        "ownerReferences": [
            {
                "apiVersion": "sources.triggermesh.io/v1alpha1",
                "blockOwnerDeletion": true,
                "controller": true,
                "kind": "FooSource",
                "name": "sample",
                "uid": "eb046145-ca7e-4f14-a208-5a70affe6dec"
            }
        ],
        "resourceVersion": "588997",
        "selfLink": "/apis/apps/v1/namespaces/dev/deployments/foosource-sample",
        "uid": "c949ce89-9953-4ad7-958e-9adaef9a5d83"
    },
    "spec": {
        "progressDeadlineSeconds": 600,
        "replicas": 1,
        "revisionHistoryLimit": 10,
        "selector": {
            "matchLabels": {
                "app.kubernetes.io/instance": "sample",
                "app.kubernetes.io/name": "foosource"
            }
        },
        "strategy": {
            "rollingUpdate": {
                "maxSurge": "25%",
                "maxUnavailable": "25%"
            },
            "type": "RollingUpdate"
        },
        "template": {
            "metadata": {
                "creationTimestamp": null,
                "labels": {
                    "app.kubernetes.io/component": "adapter",
                    "app.kubernetes.io/instance": "sample",
                    "app.kubernetes.io/managed-by": "foo-sources-controller",
                    "app.kubernetes.io/name": "foosource",
                    "app.kubernetes.io/part-of": "foo-sources"
                }
            },
            "spec": {
                "containers": [
                    {
                        "env": [
                            {
                                "name": "COMPONENT",
                                "value": "foo"
                            },
                            {
                                "name": "METRICS_PROMETHEUS_PORT",
                                "value": "9092"
                            },
                            {
                                "name": "NAMESPACE",
                                "value": "dev"
                            },
                            {
                                "name": "NAME",
                                "value": "sample"
                            },
                            {
                                "name": "K_SINK",
                                "value": "http://broker-ingress.knative-eventing.svc.cluster.local/dev/default"
                            },
                            {
                                "name": "K_LOGGING_CONFIG",
                                "value": "{\"zap-logger-config\":\"{\\n  \\\"level\\\": \\\"info\\\",\\n  \\\"development\\\": false,\\n  \\\"outputPaths\\\": [\\\"stdout\\\"],\\n  \\\"errorOutputPaths\\\": [\\\"stderr\\\"],\\n  \\\"encoding\\\": \\\"json\\\",\\n  \\\"encoderConfig\\\": {\\n    \\\"timeKey\\\": \\\"ts\\\",\\n    \\\"levelKey\\\": \\\"level\\\",\\n    \\\"nameKey\\\": \\\"logger\\\",\\n    \\\"callerKey\\\": \\\"caller\\\",\\n    \\\"messageKey\\\": \\\"msg\\\",\\n    \\\"stacktraceKey\\\": \\\"stacktrace\\\",\\n    \\\"lineEnding\\\": \\\"\\\",\\n    \\\"levelEncoder\\\": \\\"\\\",\\n    \\\"timeEncoder\\\": \\\"iso8601\\\",\\n    \\\"durationEncoder\\\": \\\"\\\",\\n    \\\"callerEncoder\\\": \\\"\\\"\\n  }\\n}\\n\"}"
                            },
                            {
                                "name": "K_METRICS_CONFIG",
                                "value": "{\"Domain\":\"triggermesh.io/source\",\"Component\":\"foosource\",\"PrometheusPort\":0,\"ConfigMap\":{}}"
                            }
                        ],
                        "image": "gcr.io/triggermesh/foosource",
                        "imagePullPolicy": "Always",
                        "name": "adapter",
                        "ports": [
                            {
                                "containerPort": 8080,
                                "name": "health",
                                "protocol": "TCP"
                            },
                            {
                                "containerPort": 9092,
                                "name": "metrics",
                                "protocol": "TCP"
                            }
                        ],
                        "readinessProbe": {
                            "failureThreshold": 3,
                            "httpGet": {
                                "path": "/health",
                                "port": "health",
                                "scheme": "HTTP"
                            },
                            "periodSeconds": 10,
                            "successThreshold": 1,
                            "timeoutSeconds": 1
                        },
                        "resources": {
                            "limits": {
                                "cpu": "1",
                                "memory": "45Mi"
                            },
                            "requests": {
                                "cpu": "90m",
                                "memory": "30Mi"
                            }
                        },
                        "terminationMessagePath": "/dev/termination-log",
                        "terminationMessagePolicy": "FallbackToLogsOnError"
                    }
                ],
                "dnsPolicy": "ClusterFirst",
                "restartPolicy": "Always",
                "schedulerName": "default-scheduler",
                "securityContext": {},
                "terminationGracePeriodSeconds": 30
            }
        }
    },
    "status": {
        "availableReplicas": 1,
        "conditions": [
            {
                "lastTransitionTime": "2020-04-26T13:23:03Z",
                "lastUpdateTime": "2020-04-26T13:23:27Z",
                "message": "ReplicaSet \"foosource-sample-5774c7984d\" has successfully progressed.",
                "reason": "NewReplicaSetAvailable",
                "status": "True",
                "type": "Progressing"
            },
            {
                "lastTransitionTime": "2020-04-26T13:24:53Z",
                "lastUpdateTime": "2020-04-26T13:24:53Z",
                "message": "Deployment has minimum availability.",
                "reason": "MinimumReplicasAvailable",
                "status": "True",
                "type": "Available"
            }
        ],
        "observedGeneration": 1,
        "readyReplicas": 1,
        "replicas": 1,
        "updatedReplicas": 1
    }
}
`
)

func TestDeploymentEqual(t *testing.T) {
	current := &appsv1.Deployment{}
	loadFixture(t, tDeployment, current)

	require.GreaterOrEqual(t, len(current.Labels), 2,
		"Test suite requires a reference object with at least 2 labels to run properly")
	require.True(t, len(current.Spec.Template.Spec.Containers) > 0 &&
		len(current.Spec.Template.Spec.Containers[0].Env) > 0 &&
		current.Spec.Template.Spec.Containers[0].Env[0].Value != "",
		"Test suite requires a reference object with a Container that has at least 1 EnvVar to run properly")

	assert.True(t, deploymentEqual(nil, nil), "Two nil elements should be equal")

	testCases := map[string]struct {
		prep   func() *appsv1.Deployment
		expect bool
	}{
		"not equal when one element is nil": {
			func() *appsv1.Deployment {
				return nil
			},
			false,
		},
		// counter intuitive but expected result for deep derivative comparisons
		"equal when all desired attributes are empty": {
			func() *appsv1.Deployment {
				return &appsv1.Deployment{}
			},
			true,
		},
		"not equal when some existing attribute differs": {
			func() *appsv1.Deployment {
				desired := current.DeepCopy()
				for k := range desired.Labels {
					desired.Labels[k] += "test"
					break // changing one is enough
				}
				return desired
			},
			false,
		},
		"equal when current has more attributes than desired": {
			func() *appsv1.Deployment {
				desired := current.DeepCopy()
				for k := range desired.Labels {
					delete(desired.Labels, k)
					break // deleting one is enough
				}
				return desired
			},
			true,
		},
		"not equal when desired has more attributes than current": {
			func() *appsv1.Deployment {
				desired := current.DeepCopy()
				for k := range desired.Labels {
					desired.Labels[k+"test"] = "test"
					break // adding one is enough
				}
				return desired
			},
			false,
		},
		"not equal when EnvVar desired value is empty": {
			func() *appsv1.Deployment {
				desired := current.DeepCopy()
				desired.Spec.Template.Spec.Containers[0].Env[0].Value = ""
				return desired
			},
			false,
		},
	}

	for name, tc := range testCases {
		//nolint:scopelint
		t.Run(name, func(t *testing.T) {
			desired := tc.prep()
			switch tc.expect {
			case true:
				assert.True(t, deploymentEqual(desired, current))
			case false:
				assert.False(t, deploymentEqual(desired, current))
			}
		})
	}
}

func TestServiceAccountEqual(t *testing.T) {
	current := &corev1.ServiceAccount{}
	loadFixture(t, tServiceAccount, current)

	require.GreaterOrEqual(t, len(current.Labels), 2,
		"Test suite requires a reference object with at least 2 labels to run properly")
	require.Nil(t, current.AutomountServiceAccountToken,
		"Test suite requires a reference object with a nil automountServiceAccountTokent attribute to run properly")

	assert.True(t, serviceAccountEqual(nil, nil), "Two nil elements should be equal")

	testCases := map[string]struct {
		prep   func() *corev1.ServiceAccount
		expect bool
	}{
		"not equal when one element is nil": {
			func() *corev1.ServiceAccount {
				return nil
			},
			false,
		},
		// counter intuitive but expected result for deep derivative comparisons
		"equal when all desired attributes are empty": {
			func() *corev1.ServiceAccount {
				return &corev1.ServiceAccount{}
			},
			true,
		},
		"not equal when some existing attribute differs": {
			func() *corev1.ServiceAccount {
				desired := current.DeepCopy()
				for k := range desired.Labels {
					desired.Labels[k] += "test"
					break // changing one is enough
				}
				return desired
			},
			false,
		},
		"equal when some attribute is set in current but not in desired": {
			func() *corev1.ServiceAccount {
				desired := current.DeepCopy()
				true := true
				desired.AutomountServiceAccountToken = &true
				return desired
			},
			false,
		},
		"equal when current has more attributes than desired": {
			func() *corev1.ServiceAccount {
				desired := current.DeepCopy()
				for k := range desired.Labels {
					delete(desired.Labels, k)
					break // deleting one is enough
				}
				return desired
			},
			true,
		},
		"not equal when desired has more attributes than current": {
			func() *corev1.ServiceAccount {
				desired := current.DeepCopy()
				for k := range desired.Labels {
					desired.Labels[k+"test"] = "test"
					break // adding one is enough
				}
				return desired
			},
			false,
		},
	}

	for name, tc := range testCases {
		//nolint:scopelint
		t.Run(name, func(t *testing.T) {
			desired := tc.prep()
			switch tc.expect {
			case true:
				assert.True(t, serviceAccountEqual(desired, current))
			case false:
				assert.False(t, serviceAccountEqual(desired, current))
			}
		})
	}
}

func loadFixture(t *testing.T, contents string, obj runtime.Object) {
	t.Helper()

	if err := json.Unmarshal([]byte(contents), obj); err != nil {
		t.Fatalf("Error deserializing fixture object: %s", err)
	}
}
