package namespaced

import (
	"context"

	tenantv1alpha1 "github.com/redhat-consulting-services/multi-tenant-operator/api/tenant/v1alpha1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func CreateOrUpdateRoleBindings(ctx context.Context, client client.Client, mtc *tenantv1alpha1.MultiTenantConfig, namespaces []string) error {
	if len(mtc.Spec.RoleBindings) < 1 {
		return nil
	}
	for _, ns := range namespaces {
		if err := createOrUpdateRoleBinding(ctx, client, mtc, ns); err != nil {
			return err
		}
	}
	return nil
}

func createOrUpdateRoleBinding(ctx context.Context, client client.Client, mtc *tenantv1alpha1.MultiTenantConfig, namespace string) error {
	for _, rbSpec := range mtc.Spec.RoleBindings {
		roleBinding := &rbacv1.RoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name:      rbSpec.Name,
				Namespace: namespace,
			},
		}

		_, err := controllerutil.CreateOrUpdate(ctx, client, roleBinding, func() error {
			if roleBinding.Labels == nil {
				roleBinding.Labels = map[string]string{}
			}
			roleBinding.Labels[managedNamespacetenantNameLabelKey] = mtc.Name
			roleBinding.Labels[managedByLabelKey] = managedByLabelValue
			roleBinding.Labels[multiTenantConfigNameLabelKey] = mtc.Name

			roleBinding.RoleRef = rbSpec.RoleRef
			roleBinding.Subjects = rbSpec.Subjects

			// set ownership reference to the MultiTenantConfig
			if err := controllerutil.SetControllerReference(mtc, roleBinding, client.Scheme()); err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}
