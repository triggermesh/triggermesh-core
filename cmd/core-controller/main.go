// Copyright 2023 TriggerMesh Inc.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"

	injection "knative.dev/pkg/injection"
	"knative.dev/pkg/injection/sharedmain"
	"knative.dev/pkg/signals"

	"github.com/triggermesh/triggermesh-core/pkg/reconciler/memorybroker"
	"github.com/triggermesh/triggermesh-core/pkg/reconciler/redisbroker"
	"github.com/triggermesh/triggermesh-core/pkg/reconciler/replay"
	"github.com/triggermesh/triggermesh-core/pkg/reconciler/trigger"
)

func main() {

	ctx := signals.NewContext()

	// There is only one configuration item for the TriggerMesh Core controller.
	// Instead of creating a structure and a formal configuration retrieval
	// from environment, we use this very simplistic approach
	ns := os.Getenv("WORKING_NAMESPACE")
	if len(ns) != 0 {
		ctx = injection.WithNamespaceScope(ctx, ns)
	}

	sharedmain.MainWithContext(ctx, "core-controller",
		memorybroker.NewController,
		redisbroker.NewController,
		trigger.NewController,
		replay.NewController,
	)
}
