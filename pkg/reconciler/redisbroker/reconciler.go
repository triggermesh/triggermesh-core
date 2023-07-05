// Copyright 2022 TriggerMesh Inc.
// SPDX-License-Identifier: Apache-2.0

package redisbroker

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	"knative.dev/pkg/apis"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/network"
	knreconciler "knative.dev/pkg/reconciler"

	eventingv1alpha1 "github.com/triggermesh/triggermesh-core/pkg/apis/eventing/v1alpha1"
	"github.com/triggermesh/triggermesh-core/pkg/reconciler/common"
	"github.com/triggermesh/triggermesh-core/pkg/reconciler/resources"
)

const (
	defaultMaxLen = "1000"
)

type reconciler struct {
	secretReconciler    common.SecretReconciler
	configMapReconciler common.ConfigMapReconciler
	saReconciler        common.ServiceAccountReconciler
	brokerReconciler    common.BrokerReconciler

	redisReconciler redisReconciler
}

// options that set Broker environment variables specific for the RedisBroker.
func redisDeploymentOption(rb *eventingv1alpha1.RedisBroker, redisSvc *corev1.Service) resources.DeploymentOption {
	return func(d *appsv1.Deployment) {
		// Make sure the broker container exists before modifying it.
		if len(d.Spec.Template.Spec.Containers) == 0 {
			// Unexpected path.
			panic("The Broker Deployment to be reconciled has no containers in it.")
		}

		c := &d.Spec.Template.Spec.Containers[0]

		var stream string
		if rb.Spec.Redis != nil && rb.Spec.Redis.Stream != nil && *rb.Spec.Redis.Stream != "" {
			stream = *rb.Spec.Redis.Stream
		} else {
			stream = rb.Namespace + "." + rb.Name
		}
		resources.ContainerAddEnvFromValue("REDIS_STREAM", stream)(c)

		maxLen := defaultMaxLen
		if rb.Spec.Redis != nil && rb.Spec.Redis.StreamMaxLen != nil {
			maxLen = strconv.Itoa(*rb.Spec.Redis.StreamMaxLen)
		}
		resources.ContainerAddEnvFromValue("REDIS_STREAM_MAX_LEN", maxLen)(c)

		if rb.Spec.Redis != nil && rb.Spec.Redis.EnableTrackingID != nil && *rb.Spec.Redis.EnableTrackingID {
			resources.ContainerAddEnvFromValue("REDIS_TRACKING_ID_ENABLED", "true")(c)
		}

		if rb.IsUserProvidedRedis() {

			// Standalone connections require an address, while cluster connections require an
			// address list of each endpoint available for the initial connection.
			if rb.Spec.Redis.Connection.ClusterURLs != nil &&
				len(rb.Spec.Redis.Connection.ClusterURLs) != 0 {
				resources.ContainerAddEnvFromValue("REDIS_CLUSTER_ADDRESSES",
					strings.Join(rb.Spec.Redis.Connection.ClusterURLs, ","))(c)
			} else {
				resources.ContainerAddEnvFromValue("REDIS_ADDRESS", *rb.Spec.Redis.Connection.URL)(c)
			}

			if rb.Spec.Redis.Connection.Username != nil {
				resources.ContainerAddEnvVarFromSecret("REDIS_USERNAME",
					rb.Spec.Redis.Connection.Username.SecretKeyRef.Name,
					rb.Spec.Redis.Connection.Username.SecretKeyRef.Key)(c)
			}

			if rb.Spec.Redis.Connection.Password != nil {
				resources.ContainerAddEnvVarFromSecret("REDIS_PASSWORD",
					rb.Spec.Redis.Connection.Password.SecretKeyRef.Name,
					rb.Spec.Redis.Connection.Password.SecretKeyRef.Key)(c)
			}

			if rb.Spec.Redis.Connection.TLSCACertificate != nil {
				resources.ContainerAddEnvVarFromSecret("REDIS_TLS_CA_CERTIFICATE",
					rb.Spec.Redis.Connection.TLSCACertificate.SecretKeyRef.Name,
					rb.Spec.Redis.Connection.TLSCACertificate.SecretKeyRef.Key)(c)
			}

			if rb.Spec.Redis.Connection.TLSCertificate != nil {
				resources.ContainerAddEnvVarFromSecret("REDIS_TLS_CERTIFICATE",
					rb.Spec.Redis.Connection.TLSCertificate.SecretKeyRef.Name,
					rb.Spec.Redis.Connection.TLSCertificate.SecretKeyRef.Key)(c)
			}

			if rb.Spec.Redis.Connection.TLSKey != nil {
				resources.ContainerAddEnvVarFromSecret("REDIS_TLS_KEY",
					rb.Spec.Redis.Connection.TLSKey.SecretKeyRef.Name,
					rb.Spec.Redis.Connection.TLSKey.SecretKeyRef.Key)(c)
			}

			if rb.Spec.Redis.Connection.TLSEnabled != nil && *rb.Spec.Redis.Connection.TLSEnabled {
				resources.ContainerAddEnvFromValue("REDIS_TLS_ENABLED", "true")(c)
			}

			if rb.Spec.Redis.Connection.TLSSkipVerify != nil && *rb.Spec.Redis.Connection.TLSSkipVerify {
				tlsSkipVerifyDefault := "true"
				// TODO this should be moved to webhook
				if rb.Spec.Redis.Connection.TLSCACertificate != nil {
					tlsSkipVerifyDefault = "false"
				}
				resources.ContainerAddEnvFromValue("REDIS_TLS_SKIP_VERIFY", tlsSkipVerifyDefault)(c)
			}

		} else {
			resources.ContainerAddEnvFromValue("REDIS_ADDRESS",
				fmt.Sprintf("%s:%d", redisSvc.Name, redisSvc.Spec.Ports[0].Port))(c)
		}
	}
}

func (r *reconciler) ReconcileKind(ctx context.Context, rb *eventingv1alpha1.RedisBroker) knreconciler.Event {
	logging.FromContext(ctx).Infow("Reconciling", zap.Any("RedisBroker", *rb))

	// Make sure the Redis deployment and service exists.
	_, redisSvc, err := r.redisReconciler.reconcile(ctx, rb)
	if err != nil {
		return err
	}

	// Iterate triggers and create secret.
	secret, err := r.secretReconciler.Reconcile(ctx, rb)
	if err != nil {
		return err
	}

	// Create configMap.
	configMap, err := r.configMapReconciler.Reconcile(ctx, rb)
	if err != nil {
		return err
	}

	// Make sure the Broker service account and roles exists.
	sa, _, err := r.saReconciler.Reconcile(ctx, rb)
	if err != nil {
		return err
	}

	// Make sure the Broker deployment exists and that it points to the Redis service.
	_, brokerSvc, err := r.brokerReconciler.Reconcile(ctx, rb, sa, secret, configMap, redisDeploymentOption(rb, redisSvc))
	if err != nil {
		return err
	}

	// Set address to the Broker service.
	rb.Status.SetAddress(getServiceAddress(brokerSvc))

	return nil
}

func getServiceAddress(svc *corev1.Service) *apis.URL {
	var port string
	if svc.Spec.Ports[0].Port != 80 {
		port = ":" + strconv.Itoa(int(svc.Spec.Ports[0].Port))
	}

	return apis.HTTP(
		network.GetServiceHostname(svc.Name, svc.Namespace) + port)
}
