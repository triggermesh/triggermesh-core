// Copyright 2022 TriggerMesh Inc.corev1.Service
// SPDX-License-Identifier: Apache-2.0

package resources

import (
	"testing"

	"github.com/stretchr/testify/assert"

	corev1 "k8s.io/api/core/v1"
)

func TestNewVolume(t *testing.T) {
	testCases := map[string]struct {
		options  []VolumeOption
		expected corev1.Volume
	}{
		"basic": {
			expected: corev1.Volume{
				Name: tName,
			}},
		"with secret": {
			options: []VolumeOption{
				VolumeFromSecretOption(tSecretName, tSecretKey, tVolumeMountFile),
			},
			expected: corev1.Volume{
				Name: tName,
				VolumeSource: corev1.VolumeSource{
					Secret: &corev1.SecretVolumeSource{
						SecretName: tSecretName,
						Items: []corev1.KeyToPath{{
							Key:  tSecretKey,
							Path: tVolumeMountFile,
						}},
					},
				},
			}},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			got := NewVolume(tName, tc.options...)
			assert.Equal(t, &tc.expected, got)
		})
	}
}
