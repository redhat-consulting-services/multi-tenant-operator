package namespaced

import (
	"context"

	tenantv1alpha1 "github.com/redhat-consulting-services/multi-tenant-operator/api/tenant/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// CreateOrUpdateConfigMap creates or updates a ConfigMap with the specified name and namespace, and sets the appropriate labels and ownership reference to the MultiTenantConfig.
func CreateOrUpdateConfigMap(ctx context.Context, cl client.Client, mtc *tenantv1alpha1.MultiTenantConfig, namespace string) error {
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "user-ca-bundle",
			Namespace: namespace,
		},
	}

	_, err := controllerutil.CreateOrUpdate(ctx, cl, configMap, func() error {
		if configMap.Labels == nil {
			configMap.Labels = map[string]string{}
		}
		configMap.Labels[managedNamespacetenantNameLabelKey] = mtc.Name
		configMap.Labels[managedByLabelKey] = managedByLabelValue
		configMap.Labels[multiTenantConfigNameLabelKey] = mtc.Name
		configMap.Labels["config.openshift.io/inject-trusted-cabundle"] = trueKeyValue

		// set ownership reference to the MultiTenantConfig
		if err := controllerutil.SetControllerReference(mtc, configMap, cl.Scheme()); err != nil {
			return err
		}
		return nil
	})
	return err
}
