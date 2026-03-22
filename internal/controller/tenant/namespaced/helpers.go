package namespaced

import (
	tenantv1alpha1 "github.com/redhat-consulting-services/multi-tenant-operator/api/tenant/v1alpha1"
)

// getGeneratedName generates a name for a resource based on the provided base name, tenant name, and configuration options for name prefix/suffix. It returns the generated name as a string.
func getGeneratedName(mtc tenantv1alpha1.MultiTenantConfig, name string) string {
	if mtc.Spec.ConfigSpec.EnableNamePrefix {
		return mtc.Name + "-" + name
	}
	if mtc.Spec.ConfigSpec.EnableNameSuffix {
		return name + "-" + mtc.Name
	}

	return name
}
