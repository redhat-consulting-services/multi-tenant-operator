package namespaced

import (
	tenantv1alpha1 "github.com/redhat-consulting-services/multi-tenant-operator/api/tenant/v1alpha1"
)

const (
	managedNamespacetenantNameLabelKey = "tenant.openshift.io/name"
	managedByLabelKey                  = "app.kubernetes.io/managed-by"
	managedByLabelValue                = "multi-tenant-operator"
	multiTenantConfigNameLabelKey      = "tenant.openshift.io/multi-tenant-config"

	trueKeyValue = "true"
)

// getGeneratedNamespaceName generates a name for a resource based on the provided base name, tenant name, and configuration options for name prefix/suffix. It returns the generated name as a string.
func getGeneratedNamespaceName(mtc tenantv1alpha1.MultiTenantConfig, ns tenantv1alpha1.NamespaceSpec) string {
	config := ns.GetMergedConfigSpec(mtc.Spec.ConfigSpec)
	if config.EnableNamePrefix {
		return mtc.GetName() + "-" + ns.Name
	}
	if config.EnableNameSuffix {
		return ns.Name + "-" + mtc.GetName()
	}
	return ns.Name
}

func getGeneratedName(mtc tenantv1alpha1.MultiTenantConfig, baseName string) string {
	if mtc.Spec.ConfigSpec.EnableNamePrefix {
		return mtc.GetName() + "-" + baseName
	}
	if mtc.Spec.ConfigSpec.EnableNameSuffix {
		return baseName + "-" + mtc.GetName()
	}
	return baseName
}
