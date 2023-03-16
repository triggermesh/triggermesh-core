// Copyright 2023 TriggerMesh Inc.
// SPDX-License-Identifier: Apache-2.0

package redisreplay

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"

	"knative.dev/pkg/logging"
	knreconciler "knative.dev/pkg/reconciler"

	eventingv1alpha1 "github.com/triggermesh/triggermesh-core/pkg/apis/eventing/v1alpha1"
	"github.com/triggermesh/triggermesh-core/pkg/reconciler/common"
	"github.com/triggermesh/triggermesh-core/pkg/reconciler/resources"
	"github.com/triggermesh/triggermesh-core/pkg/reconciler/semantic"
)

type reconciler struct {
	secretReconciler common.SecretReconciler
	saReconciler     common.ServiceAccountReconciler
	startTime        string
	endTime          string
	filterKind       string
	filter           string
	sink             string
	image            string
}

// options that set the required environment variables for the RedisReplay.
func redisDeploymentOption(rr *eventingv1alpha1.RedisReplay, redisSvc *corev1.Service) resources.DeploymentOption {
	return func(d *appsv1.Deployment) {
		c := &d.Spec.Template.Spec.Containers[0]

		if rr.Spec.EndTime != nil {
			resources.ContainerAddEnvFromValue("END_TIME", *rr.Spec.EndTime)(c)
		}

		if rr.Spec.StartTime != nil {
			resources.ContainerAddEnvFromValue("START_TIME", rr.Spec.StartTime.String())(c)
		}

		if rr.Spec.FilterKind != nil {
			resources.ContainerAddEnvFromValue("FILTER_KIND", *rr.Spec.FilterKind)(c)
		}

		if rr.Spec.Filter != nil {
			resources.ContainerAddEnvFromValue("FILTER", *rr.Spec.Filter)(c)
		}

		resources.ContainerAddEnvFromValue("SINK", *rr.Spec.Sink)(c)
		resources.ContainerAddEnvFromValue("REDIS_ADDRESS", redisSvc.Name+"."+redisSvc.Namespace+".svc.cluster.local:6379")(c)
	}
}

// ReconcileKind implements Interface.ReconcileKind.
func (r *reconciler) ReconcileKind(ctx context.Context, rr *eventingv1alpha1.RedisReplay) knreconciler.Event {
	logging.FromContext(ctx).Info("Reconciling", "kind", "RedisReplay")
	d, err := r.reconcileDeployment(ctx, rb, sa, secret, deploymentOptions)
	if err != nil {
		return nil, nil, err
	}
}

// Reconcile implements Interface.Reconcile.
func (r *reconciler) Reconcile(ctx context.Context, rr *eventingv1alpha1.RedisReplay) (*appsv1.Deployment, error) {
	d, err := r.reconcileDeployment(ctx, rr)
	if err != nil {
		return nil, err
	}
	return d, nil
}

func buildRedisReplayDeployment(rr *eventingv1alpha1.RedisReplay, image string) *appsv1.Deployment {
	return resources.NewDeployment(rr.Namespace, rr.Name+"-"+redisResourceSuffix,
		resources.DeploymentWithMetaOptions(
			resources.MetaAddLabel(resources.AppNameLabel, common.AppAnnotationValue(rb)),
			resources.MetaAddLabel(resources.AppComponentLabel, "redis-replay-deployment"),
			resources.MetaAddLabel(resources.AppPartOfLabel, resources.PartOf),
			resources.MetaAddLabel(resources.AppManagedByLabel, resources.ManagedBy),
			resources.MetaAddLabel(resources.AppInstanceLabel, rb.Name+"-"+redisResourceSuffix),
			resources.MetaAddOwner(rb, rb.GetGroupVersionKind())),
		resources.DeploymentAddSelectorForTemplate(resources.AppComponentLabel, "redis-replay-deployment"),
		resources.DeploymentAddSelectorForTemplate(resources.AppInstanceLabel, rb.Name+"-"+redisResourceSuffix),
		resources.DeploymentSetReplicas(1),
		resources.DeploymentWithTemplateSpecOptions(
			resources.PodTemplateSpecWithMetaOptions(
				resources.MetaAddLabel(resources.AppPartOfLabel, resources.PartOf),
				resources.MetaAddLabel(resources.AppManagedByLabel, resources.ManagedBy),
			),
			resources.PodTemplateSpecWithPodSpecOptions(
				resources.PodSpecAddContainer(
					resources.NewContainer("redisreplay", image,
						resources.ContainerAddEnvFromValue("REDIS_ADDRESS", *rr.Spec.Redis.URL),
						// resources.ContainerAddEnvFromValue("REDIS_PASSWORD", rr.Spec.Redis.Password),
						// resources.ContainerAddEnvFromValue("REDIS_USER", rr.Spec.Redis.User),
						resources.ContainerAddEnvFromValue("START_TIME", *rr.Spec.StartTime),
						resources.ContainerAddEnvFromValue("END_TIME", *rr.Spec.EndTime),
						resources.ContainerAddEnvFromValue("FILTER_KIND", *rr.Spec.FilterKind),
						resources.ContainerAddEnvFromValue("FILTER", *rr.Spec.Filter),
						resources.ContainerAddEnvFromValue("SINK", rr.Spec.Sink))))),
	)
}

func (r *reconciler) reconcileDeployment(ctx context.Context, rb *eventingv1alpha1.RedisBroker) (*appsv1.Deployment, error) {
	desired := buildRedisDeployment(rb, r.image)
	current, err := r.deploymentLister.Deployments(desired.Namespace).Get(desired.Name)
	switch {
	case err == nil:
		// Compare current object with desired, update if needed.
		if !semantic.Semantic.DeepEqual(desired, current) {
			desired.Status = current.Status
			desired.ResourceVersion = current.ResourceVersion

			current, err = r.client.AppsV1().Deployments(desired.Namespace).Update(ctx, desired, metav1.UpdateOptions{})
			if err != nil {
				fullname := types.NamespacedName{Namespace: desired.Namespace, Name: desired.Name}
				logging.FromContext(ctx).Error("Unable to update the deployment", zap.String("deployment", fullname.String()), zap.Error(err))
				rb.Status.MarkRedisDeploymentFailed(common.ReasonFailedDeploymentUpdate, "Failed to update Redis deployment")

				return nil, pkgreconciler.NewEvent(corev1.EventTypeWarning, common.ReasonFailedDeploymentUpdate,
					"Failed to get Redis deployment %s: %w", fullname, err)
			}
		}

	case !apierrs.IsNotFound(err):
		// An error occurred retrieving current deployment.
		fullname := types.NamespacedName{Namespace: desired.Namespace, Name: desired.Name}
		logging.FromContext(ctx).Error("Unable to get the deployment", zap.String("deployment", fullname.String()), zap.Error(err))
		rb.Status.MarkRedisDeploymentFailed(common.ReasonFailedDeploymentGet, "Failed to get Redis deployment")

		return nil, pkgreconciler.NewEvent(corev1.EventTypeWarning, common.ReasonFailedDeploymentGet,
			"Failed to get Redis deployment %s: %w", fullname, err)

	default:
		// The deployment has not been found, create it.
		current, err = r.client.AppsV1().Deployments(desired.Namespace).Create(ctx, desired, metav1.CreateOptions{})
		if err != nil {
			fullname := types.NamespacedName{Namespace: desired.Namespace, Name: desired.Name}
			logging.FromContext(ctx).Error("Unable to create the deployment", zap.String("deployment", fullname.String()), zap.Error(err))
			rb.Status.MarkRedisDeploymentFailed(common.ReasonFailedDeploymentCreate, "Failed to create Redis deployment")

			return nil, pkgreconciler.NewEvent(corev1.EventTypeWarning, common.ReasonFailedDeploymentCreate,
				"Failed to create Redis deployment %s: %w", fullname, err)
		}
	}

	// Update status based on deployment
	rb.Status.PropagateRedisDeploymentAvailability(ctx, &current.Status)

	return current, nil
}

// func buildRedisReplayService(rrp eventingv1alpha1.RedisReplay) *corev1.Service {
// 	return resources.NewService(rrp.Namespace, rrp.Name+"-"+redisResourceSuffix,
// 		resources.ServiceWithMetaOptions(
// 			resources.MetaAddLabel(resources.AppNameLabel, common.AppAnnotationValue(rb)),
// 			resources.MetaAddLabel(resources.AppComponentLabel, "redis-replay-service"),
// 			resources.MetaAddLabel(resources.AppPartOfLabel, resources.PartOf),
// 			resources.MetaAddLabel(resources.AppManagedByLabel, resources.ManagedBy),
// 			resources.MetaAddLabel(resources.AppInstanceLabel, rb.Name+"-"+redisResourceSuffix),
// 			resources.MetaAddOwner(rb, rb.GetGroupVersionKind())),
// 		resources.ServiceAddSelectorForService(resources.AppComponentLabel, "redis-replay-service"),
// 		resources.ServiceAddSelectorForService(resources.AppInstanceLabel, rb.Name+"-"+redisResourceSuffix),
// 		resources.ServiceAddPort("http", 8080, 8080),
// 	)
// }

// func (r *reconciler) reconcileService(ctx context.Context, rb *eventingv1alpha1.RedisBroker) (*corev1.Service, error) {
// 	desired := buildRedisService(rb)
// 	current, err := r.serviceLister.Services(desired.Namespace).Get(desired.Name)
// 	switch {
// 	case err == nil:
// 		// Compare current object with desired, update if needed.
// 		if !semantic.Semantic.DeepEqual(desired, current) {
// 			desired.Status = current.Status
// 			desired.ResourceVersion = current.ResourceVersion

// 			current, err = r.client.CoreV1().Services(desired.Namespace).Update(ctx, desired, metav1.UpdateOptions{})
// 			if err != nil {
// 				fullname := types.NamespacedName{Namespace: desired.Namespace, Name: desired.Name}
// 				logging.FromContext(ctx).Error("Unable to update the service", zap.String("service", fullname.String()), zap.Error(err))
// 				rb.Status.MarkRedisServiceFailed(common.ReasonFailedServiceUpdate, "Failed to update RedisReplay service")

// 				return nil, pkgreconciler.NewEvent(corev1.EventTypeWarning, common.ReasonFailedServiceUpdate,
// 					"Failed to get Redis service %s: %w", fullname, err)
// 			}
// 		}

// 	case !apierrs.IsNotFound(err):
// 		// An error occurred retrieving current service.
// 		fullname := types.NamespacedName{Namespace: desired.Namespace, Name: desired.Name}
// 		logging.FromContext(ctx).Error("Unable to get the service", zap.String("service", fullname.String()), zap.Error(err))
// 		rb.Status.MarkRedisServiceFailed(common.ReasonFailedServiceGet, "Failed to get RedisReplay service")

// 		return nil, pkgreconciler.NewEvent(corev1.EventTypeWarning, common.ReasonFailedServiceGet,
// 			"Failed to get RedisReplay service %s: %w", fullname, err)

// 	default:
// 		// The service has not been found, create it.
// 		current, err = r.client.CoreV1().Services(desired.Namespace).Create(ctx, desired, metav1.CreateOptions{})
// 		if err != nil {
// 			fullname := types.NamespacedName{Namespace: desired.Namespace, Name: desired.Name}
// 			logging.FromContext(ctx).Error("Unable to create the service", zap.String("service", fullname.String()), zap.Error(err))
// 			rb.Status.MarkRedisServiceFailed(common.ReasonFailedServiceCreate, "Failed to create RedisReplay service")

// 			return nil, pkgreconciler.NewEvent(corev1.EventTypeWarning, common.ReasonFailedServiceCreate,
// 				"Failed to create RedisReplay service %s: %w", fullname, err)
// 		}
// 	}

// 	// Update status based on service
// 	rb.Status.PropagateRedisServiceAvailability(ctx, &current.Status)

// 	return current, nil
// }
