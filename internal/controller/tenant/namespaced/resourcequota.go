package namespaced

import (
	"context"

	tenantv1alpha1 "github.com/redhat-consulting-services/multi-tenant-operator/api/tenant/v1alpha1"
	tenantconfigv1alpha1 "github.com/redhat-consulting-services/multi-tenant-operator/api/tenantconfig/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	resourceQuotaName = "tenant-resource-quota"
)

func CreateOrUpdateResourceQuota(ctx context.Context, cl client.Client, mtc *tenantv1alpha1.MultiTenantConfig, rqSpec *tenantconfigv1alpha1.NamespaceResourceQuota, namespace string) error {
	if mtc.Spec.ResourceQuotaReference == "" || rqSpec == nil {
		// if no spec is provided, skip creating/updating the ResourceQuota
		return nil
	}

	resourceQuota := &corev1.ResourceQuota{
		ObjectMeta: metav1.ObjectMeta{
			Name:      resourceQuotaName,
			Namespace: namespace,
		},
	}

	_, err := controllerutil.CreateOrUpdate(ctx, cl, resourceQuota, func() error {
		if resourceQuota.Labels == nil {
			resourceQuota.Labels = map[string]string{}
		}
		resourceQuota.Labels[managedNamespacetenantNameLabelKey] = mtc.Name
		resourceQuota.Labels[managedByLabelKey] = managedByLabelValue
		resourceQuota.Labels[multiTenantConfigNameLabelKey] = mtc.Name

		// set config spec fields
		resourceQuota.Spec.Hard = rqSpec.Spec.Hard

		// set ownership reference to the MultiTenantConfig
		if err := controllerutil.SetControllerReference(mtc, resourceQuota, cl.Scheme()); err != nil {
			return err
		}
		return nil
	})
	return err
}
