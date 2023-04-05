// Copyright 2022 TriggerMesh Inc.
// SPDX-License-Identifier: Apache-2.0

package redisreplay

import (
	"context"

	"go.uber.org/zap"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	batchv1listers "k8s.io/client-go/listers/batch/v1"
	"knative.dev/pkg/apis"
	"knative.dev/pkg/logging"
	pkgreconciler "knative.dev/pkg/reconciler"
	"knative.dev/pkg/resolver"

	eventingv1alpha1 "github.com/triggermesh/triggermesh-core/pkg/apis/eventing/v1alpha1"
	eventingv1alpha1listers "github.com/triggermesh/triggermesh-core/pkg/client/generated/listers/eventing/v1alpha1"
	"github.com/triggermesh/triggermesh-core/pkg/reconciler/resources"
	"github.com/triggermesh/triggermesh-core/pkg/reconciler/semantic"
)

type reconciler struct {
	// TODO duck brokers
	client      kubernetes.Interface
	rrLister    eventingv1alpha1listers.RedisReplayLister
	jobsLister  batchv1listers.JobLister
	rbLister    eventingv1alpha1listers.RedisBrokerLister
	uriResolver *resolver.URIResolver

	image      string
	pullPolicy string
}

func (r *reconciler) ReconcileKind(ctx context.Context, t *eventingv1alpha1.RedisReplay) pkgreconciler.Event {
	log := logging.FromContext(ctx)
	log.Info("Reconciling")

	l, err := r.rbLister.RedisBrokers(t.Namespace).Get(t.Spec.Broker.Name)
	if err != nil {
		return err
	}

	// create a desired job in memory.
	// TODO:: It would be nice if this function did not touch kubernetes.
	desired := r.createDesiredJob(ctx, l, t)

	current, err := r.jobsLister.Jobs(t.Namespace).Get(desired.Name)
	switch {
	case err == nil:
		if !semantic.Semantic.DeepEqual(desired, current) {
			desired.Status = current.Status
			desired.ResourceVersion = current.ResourceVersion

			_, err = r.client.BatchV1().Jobs(t.Namespace).Update(ctx, desired, metav1.UpdateOptions{})
			if err != nil {
				fullname := types.NamespacedName{Namespace: desired.Namespace, Name: desired.Name}
				logging.FromContext(ctx).Errorw("Failed to update Job", zap.String("job", fullname.String()), zap.Error(err))
				// rr. .MarkJobNotReady("JobUpdateFailed", "Failed to update Job %q: %v", fullname, err)
				return pkgreconciler.NewEvent(corev1.EventTypeWarning, "InternalError", "Failed to update Job %q: %v", fullname, err)
			}
		}
	case !apierrs.IsNotFound(err):
		fullname := types.NamespacedName{Namespace: desired.Namespace, Name: desired.Name}
		logging.FromContext(ctx).Errorw("Failed to get Job", zap.String("job", fullname.String()), zap.Error(err))
		// rr. .MarkJobNotReady("JobGetFailed", "Failed to get Job %q: %v", fullname, err)
		return pkgreconciler.NewEvent(corev1.EventTypeWarning, "InternalError", "Failed to get Job %q: %v", fullname, err)
	default:
		_, err = r.client.BatchV1().Jobs(t.Namespace).Create(ctx, desired, metav1.CreateOptions{})
		if err != nil {
			fullname := types.NamespacedName{Namespace: desired.Namespace, Name: desired.Name}
			logging.FromContext(ctx).Errorw("Failed to create Job", zap.String("job", fullname.String()), zap.Error(err))
			// rr. .MarkJobNotReady("JobCreationFailed", "Failed to create Job %q: %v", fullname, err)
			return pkgreconciler.NewEvent(corev1.EventTypeWarning, "InternalError", "Failed to create Job %q: %v", fullname, err)
		}
	}

	// rr. .MarkJobReady()
	return nil
}

func (r *reconciler) createDesiredJob(ctx context.Context, rb *eventingv1alpha1.RedisBroker, rr *eventingv1alpha1.RedisReplay) *batchv1.Job {
	meta := rb.GetObjectMeta()
	var targetURI *apis.URL
	// var err error
	// targetURI, err = r.uriResolver.URIFromDestinationV1(ctx, *rr.Spec.Target, rr)
	// if err != nil {
	// 	logging.FromContext(ctx).Errorw("Failed to resolve target URI", zap.Error(err))
	// 	return nil
	// }

	if targetURI == nil {
		// point to the broker's address
		targetURI = rb.Status.Address.URL
	}
	var redisPasswordName, redisPasswordKey, redisUsernameName, redisUsernameKey, redisAddress string
	// Check if SecretKeyRef fields exist and set default values if they don't
	if rb.Spec.Redis.Connection != nil {
		redisPasswordName = rb.Spec.Redis.Connection.Password.SecretKeyRef.Name
		redisPasswordKey = rb.Spec.Redis.Connection.Password.SecretKeyRef.Key
		redisUsernameName = rb.Spec.Redis.Connection.Username.SecretKeyRef.Name
		redisUsernameKey = rb.Spec.Redis.Connection.Username.SecretKeyRef.Key
		redisAddress = *rb.Spec.Redis.Connection.URL
	} else {
		redisPasswordName = ""
		redisPasswordKey = ""
		redisUsernameName = ""
		redisUsernameKey = ""
		redisAddress = "demo-rb-redis:6379"
	}

	ns, name := meta.GetNamespace(), meta.GetName()
	copts := []resources.ContainerOption{
		resources.ContainerAddEnvFromFieldRef("KUBERNETES_NAMESPACE", "metadata.namespace"),
		resources.ContainerAddEnvFromValue("REDIS_ADDRESS", redisAddress),
		resources.ContainerAddEnvFromValue("K_SINK", targetURI.String()),
		resources.ContainerAddEnvVarFromSecret("REDIS_PASSWORD", redisPasswordName, redisPasswordKey),
		resources.ContainerAddEnvVarFromSecret("REDIS_USER", redisUsernameName, redisUsernameKey),
		resources.ContainerWithImagePullPolicy(v1.PullAlways),
	}

	jobopt := []resources.JobOption{
		resources.JobWithTemplateSpecOptions(
			resources.PodTemplateSpecWithMetaOptions(
				resources.MetaAddLabel(resources.AppPartOfLabel, resources.PartOf),
				resources.MetaAddLabel(resources.AppManagedByLabel, resources.ManagedBy),
			),
			resources.PodTemplateSpecWithRestartPolicy(v1.RestartPolicyNever),
			resources.PodTemplateSpecWithPodSpecOptions(
				resources.PodSpecAddContainer(
					resources.NewContainer("replay", r.image, copts...))),
		),
	}

	return resources.NewJob(ns, name, jobopt...)
}
