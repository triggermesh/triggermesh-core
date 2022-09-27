// Copyright 2022 TriggerMesh Inc.corev1.Service
// SPDX-License-Identifier: Apache-2.0

package resources

import (
	"testing"

	"github.com/stretchr/testify/assert"

	corev1 "k8s.io/api/core/v1"
)

func TestNewVolumeMount(t *testing.T) {
	testCases := map[string]struct {
		options  []VolumeMountOption
		expected corev1.VolumeMount
	}{
		"basic": {
			expected: corev1.VolumeMount{
				Name:      tName,
				MountPath: tVolumeMountFile,
			}},
		"with read only": {
			options: []VolumeMountOption{
				VolumeMountWithReadOnlyOption(true),
			},

			expected: corev1.VolumeMount{
				Name:      tName,
				MountPath: tVolumeMountFile,
				ReadOnly:  true,
			}},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			got := NewVolumeMount(tName, tVolumeMountFile, tc.options...)
			assert.Equal(t, &tc.expected, got)
		})
	}
}
