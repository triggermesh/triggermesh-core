// Copyright 2022 TriggerMesh Inc.
// SPDX-License-Identifier: Apache-2.0

package redisreplay

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	batchv1listers "k8s.io/client-go/listers/batch/v1"
	"knative.dev/pkg/controller"
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
	// err := r.resolveBroker(ctx, t)
	// if err != nil {
	// 	return err
	// }

	// err = r.resolveTarget(ctx, t)
	// if err != nil {
	// 	return err
	// }

	// i need access to the spec of the redis broker that is refrenced in the spec of the RedisReplay object

	// after i have the spec of the redis broker, i need to get the redis db connection info from the spec of the redis broker

	// then i need to create a redis client using the redis db connection info

	// retrieve the redis broker from the RedisBrokerLister

	l, err := r.rbLister.RedisBrokers(t.Namespace).Get(t.Spec.Broker.Name)
	if err != nil {
		return err
	}

	// create a desired job in memory.
	// TODO:: It would be nice if this function did not touch kubernetes.
	desired := r.createDesiredJob(ctx, l, t)

	// compare the desired job with existing jobs
	// if the job does not exist, create it
	// if the job exists, update it
	// if the job exists and is the same, do nothing
	current, err := r.jobsLister.Jobs(t.Namespace).Get(desired.Name)
	switch {
	case err == nil:
		// compare current object with desired, update if needed.
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

// func replayJobOption(opts *resources.JobOption) resources.JobOption {
// 	return func(j *resources.Job) {

// 	}
// }

func (r *reconciler) resolveBroker(ctx context.Context, t *eventingv1alpha1.RedisReplay) pkgreconciler.Event {
	// TODO duck
	// TODO move to webhook
	switch {
	case t.Spec.Broker.Group == "":
		t.Spec.Broker.Group = eventingv1alpha1.SchemeGroupVersion.Group
	case t.Spec.Broker.Group != eventingv1alpha1.SchemeGroupVersion.Group:
		return controller.NewPermanentError(fmt.Errorf("not supported Broker Group %q", t.Spec.Broker.Group))
	}

	// var rb *eventingv1alpha1.RedisBroker
	// if t.Spec.Broker.Kind == rb.GetGroupVersionKind().Kind {
	// 	return r.resolveRedisBroker(ctx, t)
	// }

	return controller.NewPermanentError(fmt.Errorf("not supported Broker Kind %q", t.Spec.Broker.Kind))
}

func (r *reconciler) createDesiredJob(ctx context.Context, rb *eventingv1alpha1.RedisBroker, rr *eventingv1alpha1.RedisReplay) *batchv1.Job {
	meta := rb.GetObjectMeta()

	// check if the password is not empty. if it is not empty, then add the password to the container env vars
	// rb.Spec.Redis.Connection.Password.SecretKeyRef
	ns, name := meta.GetNamespace(), meta.GetName()
	copts := []resources.ContainerOption{
		// resources.ContainerAddEnvFromValue("BROKER_NAME", *rb.Spec.Broker
		resources.ContainerAddEnvFromFieldRef("KUBERNETES_NAMESPACE", "metadata.namespace"),
		// TODO: check if rb.Spec.Redis.Connection.URL is not nil
		// if we are using the default redis backend. or if we have a custom redis provided
		resources.ContainerAddEnvFromValue("REDIS_ADDRESS", "10.152.183.171:6379"),
		resources.ContainerAddEnvFromValue("K_SINK", "localhost:8080"),
		// resources.ContainerAddEnvVarFromSecret("REDIS_PASSWORD", rb.Spec.Redis.Connection.Password.SecretKeyRef.Name, rb.Spec.Redis.Connection.Password.SecretKeyRef.Key),
		// resources.ContainerAddEnvVarFromSecret("REDIS_USER", rb.Spec.Redis.Connection.Username.SecretKeyRef.Name, rb.Spec.Redis.Connection.Username.SecretKeyRef.Key),
		// resources.ContainerAddEnvFromValue("KUBERNETES_BROKER_CONFIG_SECRET_NAME", secret.Name),
		// resources.ContainerAddEnvFromValue("KUBERNETES_BROKER_CONFIG_SECRET_KEY", ConfigSecretKey),
		// add the image pull policy to the container
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

// func (r *reconciler) resolveRedisBroker(ctx context.Context, t *eventingv1alpha1.Trigger) pkgreconciler.Event {
// 	rb, err := r.rrLister.RedisReplays(t.Namespace).Get(t.Spec.Broker.Name)
// 	if err != nil {
// 		if apierrs.IsNotFound(err) {
// 			logging.FromContext(ctx).Errorw(fmt.Sprintf("Trigger %s/%s references non existing broker %q", t.Namespace, t.Name, t.Spec.Broker.Name))
// 			t.Status.MarkBrokerFailed(common.ReasonBrokerDoesNotExist, "Broker %q does not exist", t.Spec.Broker.Name)
// 			// No need to requeue, we will be notified when if broker is created.
// 			return controller.NewPermanentError(err)
// 		}

// 		t.Status.MarkBrokerFailed(common.ReasonFailedBrokerGet, "Failed to get broker %q : %s", t.Spec.Broker, err)
// 		return pkgreconciler.NewEvent(corev1.EventTypeWarning, common.ReasonFailedBrokerGet,
// 			"Failed to get broker for trigger %s/%s: %w", t.Namespace, t.Name, err)
// 	}

// 	t.Status.PropagateBrokerCondition(rb.Status.GetTopLevelCondition())

// 	// No need to requeue, we'll get requeued when broker changes status.
// 	if !rb.IsReady() {
// 		logging.FromContext(ctx).Errorw(fmt.Sprintf("Trigger %s/%s references non ready broker %q", t.Namespace, t.Name, t.Spec.Broker.Name))
// 	}

// 	return nil
// }

// func (r *reconciler) resolveTarget(ctx context.Context, t *eventingv1alpha1.RedisReplay) pkgreconciler.Event {
// 	if t.Spec.Target.Ref != nil && t.Spec.Target.Ref.Namespace == "" {
// 		// To call URIFromDestinationV1(ctx context.Context, dest v1.Destination, parent interface{}), dest.Ref must have a Namespace
// 		// If Target.Ref.Namespace is nil, We will use the Namespace of Trigger as the Namespace of dest.Ref
// 		t.Spec.Target.Ref.Namespace = t.Namespace
// 	}

// 	targetURI, err := r.uriResolver.URIFromDestinationV1(ctx, t.Spec.Target, t)
// 	if err != nil {
// 		logging.FromContext(ctx).Errorw("Unable to get the target's URI", zap.Error(err))
// 		t.Status.MarkTargetResolvedFailed("Unable to get the target's URI", "%v", err)
// 		t.Status.TargetURI = nil
// 		return pkgreconciler.NewEvent(corev1.EventTypeWarning, common.ReasonFailedResolveReference,
// 			"Failed to get target's URI: %w", err)
// 	}

// 	t.Status.TargetURI = targetURI
// 	t.Status.MarkTargetResolvedSucceeded()

// 	return nil
// }
