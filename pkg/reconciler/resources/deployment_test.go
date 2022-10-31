// Copyright 2022 TriggerMesh Inc.
// SPDX-License-Identifier: Apache-2.0

package resources

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
)

var (
	tOwner = &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "my-owner",
		},
	}
)

func TestNewDeployment(t *testing.T) {
	testCases := map[string]struct {
		options  []DeploymentOption
		expected string
	}{
		"basic": {
			options: []DeploymentOption{
				DeploymentWithMetaOptions(MetaAddLabel("app", "controller-my-app")),
				DeploymentAddSelectorForTemplate("app", "my-app"),
				DeploymentSetReplicas(1),
				DeploymentWithTemplateOptions(PodSpecAddContainer(NewContainer("container-name", "my-image"))),
			},
			expected: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-name
  namespace: test-namespace
  labels:
    app: controller-my-app
spec:
  replicas: 1
  selector:
    matchLabels:
      app: my-app
  template:
    metadata:
      labels:
        app: my-app
    spec:
      containers:
      - name: container-name
        image: my-image
`},
		"with-owner": {
			options: []DeploymentOption{
				DeploymentWithMetaOptions(
					MetaAddLabel("app", "controller-my-app"),
					MetaAddOwner(tOwner, appsv1.SchemeGroupVersion.WithKind("Deployment"))),
				DeploymentAddSelectorForTemplate("app", "my-app"),
				DeploymentSetReplicas(1),
				DeploymentWithTemplateOptions(PodSpecAddContainer(NewContainer("container-name", "my-image"))),
			},
			expected: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-name
  namespace: test-namespace
  labels:
    app: controller-my-app
  ownerReferences:
  - kind: Deployment
    name: my-owner
    apiVersion: apps/v1
    controller: true
    blockOwnerDeletion: true
spec:
  replicas: 1
  selector:
    matchLabels:
      app: my-app
  template:
    metadata:
      labels:
        app: my-app
    spec:
      containers:
      - name: container-name
        image: my-image
`},
		"with-env": {
			options: []DeploymentOption{
				DeploymentWithMetaOptions(MetaAddLabel("app", "controller-my-app")),
				DeploymentAddSelectorForTemplate("app", "my-app"),
				DeploymentSetReplicas(1),
				DeploymentWithTemplateOptions(PodSpecAddContainer(
					NewContainer("container-name", "my-image",
						ContainerAddEnvFromValue("MYENV", "env-value"),
					))),
			},
			expected: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-name
  namespace: test-namespace
  labels:
    app: controller-my-app
spec:
  replicas: 1
  selector:
    matchLabels:
      app: my-app
  template:
    metadata:
      labels:
        app: my-app
    spec:
      containers:
      - name: container-name
        image: my-image
        env:
        - name: MYENV
          value: env-value
`},
		"with-port": {
			options: []DeploymentOption{
				DeploymentWithMetaOptions(MetaAddLabel("app", "controller-my-app")),
				DeploymentAddSelectorForTemplate("app", "my-app"),
				DeploymentSetReplicas(1),
				DeploymentWithTemplateOptions(PodSpecAddContainer(
					NewContainer("container-name", "my-image",
						ContainerAddPort("myport", 12345),
					))),
			},
			expected: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-name
  namespace: test-namespace
  labels:
    app: controller-my-app
spec:
  replicas: 1
  selector:
    matchLabels:
      app: my-app
  template:
    metadata:
      labels:
        app: my-app
    spec:
      containers:
      - name: container-name
        image: my-image
        ports:
        - name: myport
          containerPort: 12345
`}}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			got := NewDeployment(tNamespace, tName, tc.options...)
			obj, _, err := scheme.Codecs.UniversalDeserializer().Decode([]byte(tc.expected), nil, nil)
			require.NoError(t, err, "internal test error when decoding deployment")

			expected := obj.(*appsv1.Deployment)

			assert.Equal(t, expected, got)
		})
	}
}

func TestNewPodSpec(t *testing.T) {
	testCases := map[string]struct {
		options  []PodSpecOption
		expected corev1.PodSpec
	}{
		"with container": {
			options: []PodSpecOption{
				PodSpecAddContainer(NewContainer(tName, tImage)),
			},
			expected: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:  tName,
						Image: tImage,
					},
				},
			}},
		"with volume": {
			options: []PodSpecOption{
				PodSpecAddVolume(NewVolume(tName)),
			},
			expected: corev1.PodSpec{
				Volumes: []corev1.Volume{
					{
						Name: tName,
					},
				},
			}},
		"with serviceAccount name": {
			options: []PodSpecOption{
				PodSpecWithServiceAccountName(tServiceAccountName),
			},
			expected: corev1.PodSpec{
				ServiceAccountName: tServiceAccountName,
			}},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			got := NewPodSpec(tc.options...)
			assert.Equal(t, tc.expected, *got)
		})
	}
}
