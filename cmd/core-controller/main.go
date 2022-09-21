package main

import (
	"knative.dev/pkg/injection/sharedmain"

	"github.com/triggermesh/triggermesh-core/pkg/reconciler/redisbroker"
	"github.com/triggermesh/triggermesh-core/pkg/reconciler/trigger"
)

func main() {
	sharedmain.Main("core-controller",
		redisbroker.NewController,
		trigger.NewController,
	)
}
