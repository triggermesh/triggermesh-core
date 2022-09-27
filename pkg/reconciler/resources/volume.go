// Copyright 2022 TriggerMesh Inc.
// SPDX-License-Identifier: Apache-2.0

package resources

import (
	corev1 "k8s.io/api/core/v1"
)

type VolumeOption func(*corev1.Volume)

func NewVolume(name string, opts ...VolumeOption) *corev1.Volume {

	v := &corev1.Volume{
		Name: name,
	}

	for _, opt := range opts {
		opt(v)
	}

	return v
}

func VolumeFromSecretOption(secretName, secretKey, mountFile string) VolumeOption {
	return func(v *corev1.Volume) {
		v.Secret = &corev1.SecretVolumeSource{
			SecretName: secretName,
			Items: []corev1.KeyToPath{{
				Key:  secretKey,
				Path: mountFile,
			}},
		}
	}
}
