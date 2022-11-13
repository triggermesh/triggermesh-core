package redisbroker

import (
	"context"

	"go.uber.org/zap"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	corev1listers "k8s.io/client-go/listers/core/v1"
	rbacv1listers "k8s.io/client-go/listers/rbac/v1"
	"knative.dev/pkg/logging"
	pkgreconciler "knative.dev/pkg/reconciler"

	eventingv1alpha1 "github.com/triggermesh/triggermesh-core/pkg/apis/eventing/v1alpha1"
	"github.com/triggermesh/triggermesh-core/pkg/reconciler"
	"github.com/triggermesh/triggermesh-core/pkg/reconciler/resources"
)

const (
	BrokerDeploymentRole = "triggermesh-broker"
)

type serviceAccountReconciler struct {
	client               kubernetes.Interface
	serviceAccountLister corev1listers.ServiceAccountLister
	roleBindingLister    rbacv1listers.RoleBindingLister
}

func (r *serviceAccountReconciler) reconcile(ctx context.Context, rb *eventingv1alpha1.RedisBroker) (*corev1.ServiceAccount, *rbacv1.RoleBinding, error) {
	sa, err := r.reconcileServiceAccount(ctx, rb)
	if err != nil {
		return nil, nil, err
	}

	roleb, err := r.reconcileRoleBinding(ctx, rb, sa)
	if err != nil {
		return nil, nil, err
	}

	return sa, roleb, nil
}

func buildBrokerServiceAccount(rb *eventingv1alpha1.RedisBroker) *corev1.ServiceAccount {
	return resources.NewServiceAccount(rb.Namespace, rb.Name+"-"+brokerResourceSuffix,
		resources.ServiceAccountWithMetaOptions(
			resources.MetaAddLabel(resources.AppNameLabel, appAnnotationValue),
			resources.MetaAddLabel(resources.AppComponentLabel, "broker-serviceaccount"),
			resources.MetaAddLabel(resources.AppPartOfLabel, resources.PartOf),
			resources.MetaAddLabel(resources.AppManagedByLabel, resources.ManagedBy),
			resources.MetaAddLabel(resources.AppInstanceLabel, rb.Name+"-"+brokerResourceSuffix),
			resources.MetaAddOwner(rb, rb.GetGroupVersionKind())))
}

func (r *serviceAccountReconciler) reconcileServiceAccount(ctx context.Context, rb *eventingv1alpha1.RedisBroker) (*corev1.ServiceAccount, error) {
	desired := buildBrokerServiceAccount(rb)
	current, err := r.serviceAccountLister.ServiceAccounts(desired.Namespace).Get(desired.Name)

	switch {
	case err == nil:
		// TODO compare

	case !apierrs.IsNotFound(err):
		// An error occurred retrieving current object.
		fullname := types.NamespacedName{Namespace: desired.Namespace, Name: desired.Name}
		logging.FromContext(ctx).Error("Unable to get broker ServiceAccount", zap.String("serviceAccount", fullname.String()), zap.Error(err))
		rb.Status.MarkBrokerServiceAccountFailed(reconciler.ReasonFailedServiceAccountGet, "Failed to get broker ServiceAccount")

		return nil, pkgreconciler.NewEvent(corev1.EventTypeWarning, reconciler.ReasonFailedServiceAccountGet,
			"Failed to get broker ServiceAccount %s: %w", fullname, err)

	default:
		// The ServiceAccount has not been found, create it.
		current, err = r.client.CoreV1().ServiceAccounts(desired.Namespace).Create(ctx, desired, metav1.CreateOptions{})
		if err != nil {
			fullname := types.NamespacedName{Namespace: desired.Namespace, Name: desired.Name}
			logging.FromContext(ctx).Error("Unable to create broker ServiceAccount", zap.String("serviceAccount", fullname.String()), zap.Error(err))
			rb.Status.MarkBrokerServiceAccountFailed(reconciler.ReasonFailedServiceAccountCreate, "Failed to create broker ServiceAccount")

			return nil, pkgreconciler.NewEvent(corev1.EventTypeWarning, reconciler.ReasonFailedServiceAccountCreate,
				"Failed to create broker ServiceAccount %s: %w", fullname, err)
		}
	}

	// Update status
	rb.Status.MarkBrokerServiceAccountReady()

	return current, nil
}

func buildBrokerRoleBinding(rb *eventingv1alpha1.RedisBroker, sa *corev1.ServiceAccount) *rbacv1.RoleBinding {
	return resources.NewRoleBinding(rb.Namespace, rb.Name+"-"+brokerResourceSuffix, BrokerDeploymentRole, sa.Name,
		resources.RoleBindingWithMetaOptions(
			resources.MetaAddLabel(resources.AppNameLabel, appAnnotationValue),
			resources.MetaAddLabel(resources.AppComponentLabel, "broker-rolebinding"),
			resources.MetaAddLabel(resources.AppPartOfLabel, resources.PartOf),
			resources.MetaAddLabel(resources.AppManagedByLabel, resources.ManagedBy),
			resources.MetaAddLabel(resources.AppInstanceLabel, rb.Name+"-"+brokerResourceSuffix),
			resources.MetaAddOwner(rb, rb.GetGroupVersionKind())))
}

func (r *serviceAccountReconciler) reconcileRoleBinding(ctx context.Context, rb *eventingv1alpha1.RedisBroker, sa *corev1.ServiceAccount) (*rbacv1.RoleBinding, error) {
	desired := buildBrokerRoleBinding(rb, sa)
	current, err := r.roleBindingLister.RoleBindings(desired.Namespace).Get(desired.Name)

	switch {
	case err == nil:
		// TODO compare

	case !apierrs.IsNotFound(err):
		// An error occurred retrieving current object.
		fullname := types.NamespacedName{Namespace: desired.Namespace, Name: desired.Name}
		logging.FromContext(ctx).Error("Unable to get broker RoleBinding", zap.String("roleBinding", fullname.String()), zap.Error(err))
		rb.Status.MarkBrokerRoleBindingFailed(reconciler.ReasonFailedRoleBindingGet, "Failed to get broker RoleBinding")

		return nil, pkgreconciler.NewEvent(corev1.EventTypeWarning, reconciler.ReasonFailedRoleBindingGet,
			"Failed to get broker RoleBinding %s: %w", fullname, err)

	default:
		// The RoleBinding has not been found, create it.
		current, err = r.client.RbacV1().RoleBindings(desired.Namespace).Create(ctx, desired, metav1.CreateOptions{})
		if err != nil {
			fullname := types.NamespacedName{Namespace: desired.Namespace, Name: desired.Name}
			logging.FromContext(ctx).Error("Unable to create broker RoleBinding", zap.String("roleBinding", fullname.String()), zap.Error(err))
			rb.Status.MarkBrokerRoleBindingFailed(reconciler.ReasonFailedRoleBindingCreate, "Failed to create broker RoleBinding")

			return nil, pkgreconciler.NewEvent(corev1.EventTypeWarning, reconciler.ReasonFailedRoleBindingCreate,
				"Failed to create broker RoleBinding %s: %w", fullname, err)
		}
	}

	// Update status
	rb.Status.MarkBrokerRoleBindingReady()

	return current, nil
}
