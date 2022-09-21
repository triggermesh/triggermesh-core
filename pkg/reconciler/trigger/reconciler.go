// Copyright 2022 TriggerMesh Inc.
// SPDX-License-Identifier: Apache-2.0

package trigger

import (
	"context"

	"knative.dev/pkg/reconciler"

	eventingv1alpha1 "github.com/triggermesh/triggermesh-core/pkg/apis/eventing/v1alpha1"
)

type Reconciler struct {
}

func (r *Reconciler) ReconcileKind(ctx context.Context, rb *eventingv1alpha1.Trigger) reconciler.Event {
	return nil
}
