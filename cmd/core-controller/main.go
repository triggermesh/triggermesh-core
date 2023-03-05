// Copyright 2023 TriggerMesh Inc.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"knative.dev/pkg/injection/sharedmain"

	"github.com/triggermesh/triggermesh-core/pkg/reconciler/memorybroker"
	"github.com/triggermesh/triggermesh-core/pkg/reconciler/redisbroker"
	"github.com/triggermesh/triggermesh-core/pkg/reconciler/trigger"
)

func main() {
	sharedmain.Main("core-controller",
		memorybroker.NewController,
		redisbroker.NewController,
		trigger.NewController,
	)
}
