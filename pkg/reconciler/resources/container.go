// Copyright 2022 TriggerMesh Inc.
// SPDX-License-Identifier: Apache-2.0

package resources

import (
	"strings"

	corev1 "k8s.io/api/core/v1"
)

type ContainerOption func(*corev1.Container)

func NewContainer(name, image string, opts ...ContainerOption) *corev1.Container {
	c := &corev1.Container{
		Name:  name,
		Image: image,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

func ContainerAddEnvFromValue(name, value string) ContainerOption {
	return func(c *corev1.Container) {
		if c.Env == nil {
			c.Env = make([]corev1.EnvVar, 0, 1)
		}
		c.Env = append(c.Env, corev1.EnvVar{
			Name:  name,
			Value: value,
		})
	}
}

func ContainerAddEnvVarFromSecret(name, secretName, secretKey string) ContainerOption {
	return func(c *corev1.Container) {
		if c.Env == nil {
			c.Env = make([]corev1.EnvVar, 0, 1)
		}
		c.Env = append(c.Env, corev1.EnvVar{
			Name: name,
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: secretName,
					},
					Key: secretKey,
				},
			},
		})
	}
}

func ContainerAddEnvFromFieldRef(name, path string) ContainerOption {
	return func(c *corev1.Container) {
		if c.Env == nil {
			c.Env = make([]corev1.EnvVar, 0, 1)
		}
		c.Env = append(c.Env, corev1.EnvVar{
			Name: name,
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: path,
				},
			},
		})
	}
}

func ContainerAddArgs(s string) ContainerOption {
	return func(c *corev1.Container) {
		args := strings.Split(s, " ")
		if c.Args == nil {
			c.Args = make([]string, 0, len(args))
		}

		c.Args = append(c.Args, args...)
	}
}

func ContainerAddPort(name string, containerPort int32) ContainerOption {
	return func(c *corev1.Container) {
		if c.Ports == nil {
			c.Ports = make([]corev1.ContainerPort, 0, 1)
		}
		c.Ports = append(c.Ports, corev1.ContainerPort{
			Name:          name,
			ContainerPort: containerPort,
		})
	}
}

func ContainerAddVolumeMount(vm *corev1.VolumeMount) ContainerOption {
	return func(c *corev1.Container) {
		if c.VolumeMounts == nil {
			c.VolumeMounts = make([]corev1.VolumeMount, 0, 1)
		}
		c.VolumeMounts = append(c.VolumeMounts, *vm)
	}
}

func ContainerWithImagePullPolicy(policy corev1.PullPolicy) ContainerOption {
	return func(c *corev1.Container) {
		c.ImagePullPolicy = policy
	}
}
