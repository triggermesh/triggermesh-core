// Copyright 2022 TriggerMesh Inc.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"context"
	"regexp"

	"knative.dev/pkg/apis"
)

var (
	// Only allow lowercase alphanumeric, starting with letters.
	validAttributeName = regexp.MustCompile(`^[a-z][a-z0-9]*$`)
)

// Validate the Trigger.
func (t *Trigger) Validate(ctx context.Context) *apis.FieldError {
	errs := t.Spec.Validate(apis.WithinSpec(ctx)).ViaField("spec")
	return errs
}

// Validate the TriggerSpec.
func (ts *TriggerSpec) Validate(ctx context.Context) (errs *apis.FieldError) {

	errs = ts.Broker.Validate(ctx).ViaField("broker")

	return errs.Also(
		ValidateSubscriptionAPIFiltersList(ctx, ts.Filters).ViaField("filters"),
	).Also(
		ts.Target.Validate(ctx).ViaField("target"),
	).Also(
		ts.Delivery.Validate(ctx).ViaField("delivery"),
	)
}

func ValidateSubscriptionAPIFiltersList(ctx context.Context, filters []Filter) (errs *apis.FieldError) {
	for i, f := range filters {
		f := f
		errs = errs.Also(ValidateSubscriptionAPIFilter(ctx, &f)).ViaIndex(i)
	}
	return errs
}

func ValidateAttributesNames(attrs map[string]string) (errs *apis.FieldError) {
	for attr := range attrs {
		if !validAttributeName.MatchString(attr) {
			errs = errs.Also(apis.ErrInvalidKeyName(attr, apis.CurrentField, "Attribute name must start with a letter and can only contain lowercase alphanumeric").ViaKey(attr))
		}
	}
	return errs
}

func ValidateOneOf(filter *Filter) (err *apis.FieldError) {
	if filter != nil && hasMultipleDialects(filter) {
		return apis.ErrGeneric("multiple dialects found, filters can have only one dialect set")
	}
	return nil
}

func hasMultipleDialects(filter *Filter) bool {
	dialectFound := false
	if len(filter.Exact) > 0 {
		dialectFound = true
	}
	if len(filter.Prefix) > 0 {
		if dialectFound {
			return true
		} else {
			dialectFound = true
		}
	}
	if len(filter.Suffix) > 0 {
		if dialectFound {
			return true
		} else {
			dialectFound = true
		}
	}
	if len(filter.All) > 0 {
		if dialectFound {
			return true
		} else {
			dialectFound = true
		}
	}
	if len(filter.Any) > 0 {
		if dialectFound {
			return true
		} else {
			dialectFound = true
		}
	}
	if filter.Not != nil && dialectFound {
		return true
	}
	return false
}

func ValidateSubscriptionAPIFilter(ctx context.Context, filter *Filter) (errs *apis.FieldError) {
	if filter == nil {
		return nil
	}
	errs = errs.Also(
		ValidateOneOf(filter),
	).Also(
		ValidateAttributesNames(filter.Exact).ViaField("exact"),
	).Also(
		ValidateAttributesNames(filter.Prefix).ViaField("prefix"),
	).Also(
		ValidateAttributesNames(filter.Suffix).ViaField("suffix"),
	).Also(
		ValidateSubscriptionAPIFiltersList(ctx, filter.All).ViaField("all"),
	).Also(
		ValidateSubscriptionAPIFiltersList(ctx, filter.Any).ViaField("any"),
	).Also(
		ValidateSubscriptionAPIFilter(ctx, filter.Not).ViaField("not"),
	)
	return errs
}
