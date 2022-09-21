package main

import (
	"knative.dev/pkg/injection/sharedmain"

	"github.com/triggermesh/triggermesh-core/pkg/reconciler/redisbroker"
)

func main() {
	sharedmain.Main("core-controller", redisbroker.NewController)
}
