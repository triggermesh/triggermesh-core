package common

import (
	"strings"

	"knative.dev/pkg/kmeta"
)

func AppAnnotationValue(or kmeta.OwnerRefable) string {
	return strings.ToLower(or.GetGroupVersionKind().Kind)
}
