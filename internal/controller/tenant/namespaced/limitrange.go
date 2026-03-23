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

// CreateOrUpdateLimitRanges creates or updates LimitRanges in the specified namespaces based on the provided MultiTenantConfig and NamespaceLimitRange spec. It returns an error if any operation fails.
func CreateOrUpdateLimitRanges(ctx context.Context, cl client.Client, mtc *tenantv1alpha1.MultiTenantConfig, limitRangeSpec *tenantconfigv1alpha1.NamespaceLimitRange, namespaces []string) error {
	if mtc.Spec.LimitRangeReference == "" {
		return nil
	}
	for _, namespace := range namespaces {
		if err := createOrUpdateLimitRange(ctx, cl, mtc, limitRangeSpec, namespace); err != nil {
			return err
		}
	}
	return nil
}

func createOrUpdateLimitRange(ctx context.Context, cl client.Client, mtc *tenantv1alpha1.MultiTenantConfig, limitRangeSpec *tenantconfigv1alpha1.NamespaceLimitRange, namespace string) error {
	limitRange := &corev1.LimitRange{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "tenant-limit-range",
			Namespace: namespace,
		},
	}

	_, err := controllerutil.CreateOrUpdate(ctx, cl, limitRange, func() error {
		if limitRange.Labels == nil {
			limitRange.Labels = map[string]string{}
		}
		limitRange.Labels[managedNamespacetenantNameLabelKey] = mtc.Name
		limitRange.Labels[managedByLabelKey] = managedByLabelValue
		limitRange.Labels[multiTenantConfigNameLabelKey] = mtc.Name

		// set config spec fields
		limitRange.Spec.Limits = limitRangeSpec.Spec.Limits

		// set ownership reference to the MultiTenantConfig
		if err := controllerutil.SetControllerReference(mtc, limitRange, cl.Scheme()); err != nil {
			return err
		}
		return nil
	})
	return err
}
