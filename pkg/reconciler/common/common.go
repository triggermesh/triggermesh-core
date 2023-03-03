// Copyright 2023 TriggerMesh Inc.
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"strings"

	"knative.dev/pkg/kmeta"
)

func AppAnnotationValue(or kmeta.OwnerRefable) string {
	return strings.ToLower(or.GetGroupVersionKind().Kind)
}
