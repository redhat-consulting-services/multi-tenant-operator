package namespaced

import (
	"context"
	"fmt"

	tenantv1alpha1 "github.com/redhat-consulting-services/multi-tenant-operator/api/tenant/v1alpha1"
	tenantconfigv1alpha1 "github.com/redhat-consulting-services/multi-tenant-operator/api/tenantconfig/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// CreateOrUpdateResourceQuotas creates or updates ResourceQuotas in the specified namespaces based on the provided MultiTenantConfig and NamespaceResourceQuota spec. It returns an error if any operation fails.
func CreateOrUpdateResourceQuotas(ctx context.Context, client client.Client, mtc *tenantv1alpha1.MultiTenantConfig, rqSpec *tenantconfigv1alpha1.NamespaceResourceQuota, namespaces []string) error {
	if mtc.Spec.ResourceQuotaReference == "" {
		return nil
	}
	for _, ns := range namespaces {
		if err := createOrUpdateResourceQuota(ctx, client, mtc, rqSpec, ns); err != nil {
			return fmt.Errorf("failed to create or update ResourceQuota in namespace %s: %w", ns, err)
		}
	}
	return nil
}

func createOrUpdateResourceQuota(ctx context.Context, client client.Client, mtc *tenantv1alpha1.MultiTenantConfig, rqSpec *tenantconfigv1alpha1.NamespaceResourceQuota, namespace string) error {
	resourceQuota := &corev1.ResourceQuota{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "tenant-resource-quota",
			Namespace: namespace,
		},
	}

	_, err := controllerutil.CreateOrUpdate(ctx, client, resourceQuota, func() error {
		if resourceQuota.Labels == nil {
			resourceQuota.Labels = map[string]string{}
		}
		resourceQuota.Labels[managedNamespacetenantNameLabelKey] = mtc.Name
		resourceQuota.Labels[managedByLabelKey] = managedByLabelValue
		resourceQuota.Labels[multiTenantConfigNameLabelKey] = mtc.Name

		// set config spec fields
		resourceQuota.Spec.Hard = rqSpec.Spec.Hard

		// set ownership reference to the MultiTenantConfig
		if err := controllerutil.SetControllerReference(mtc, resourceQuota, client.Scheme()); err != nil {
			return err
		}
		return nil
	})
	return err
}
