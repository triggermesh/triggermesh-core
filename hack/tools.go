//go:build tools
// +build tools

// Copyright 2022 TriggerMesh Inc.
// SPDX-License-Identifier: Apache-2.0

package hack

// These imports ensure build tools are included in Go modules so that we can
// `go install` them in module-aware mode.
// See https://github.com/golang/go/wiki/Modules#how-can-i-track-tool-dependencies-for-a-module
import (
	_ "k8s.io/code-generator"
)
