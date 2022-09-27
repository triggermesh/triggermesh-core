// Copyright 2022 TriggerMesh Inc.
// SPDX-License-Identifier: Apache-2.0

package resources

import (
	corev1 "k8s.io/api/core/v1"
)

type VolumeMountOption func(*corev1.VolumeMount)

func NewVolumeMount(name, mountPath string, opts ...VolumeMountOption) *corev1.VolumeMount {

	vm := &corev1.VolumeMount{
		Name:      name,
		MountPath: mountPath,
	}

	for _, opt := range opts {
		opt(vm)
	}

	return vm
}

func VolumeMountWithReadOnlyOption(b bool) VolumeMountOption {
	return func(vm *corev1.VolumeMount) {
		vm.ReadOnly = b
	}
}
