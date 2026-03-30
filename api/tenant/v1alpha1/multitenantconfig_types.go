/*
Copyright 2026.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	argocdv1alpha1 "github.com/argoproj/argo-cd/v3/pkg/apis/application/v1alpha1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MultiTenantConfigSpec defines the desired state of MultiTenantConfig.
type MultiTenantConfigSpec struct {
	// ArgoCD contains the configuration for Argo CD integration in the multi-tenant environment.
	ArgoCD *ArgoCDSpec `json:"argocd,omitempty"`
	// ResourceQuotaReference is the name of the NamespaceResourceQuota resource to be applied to tenant namespaces.
	// This should reference a NamespaceResourceQuota resource.
	// If specified, the operator will apply the referenced NamespaceResourceQuota to all tenant namespaces.
	// +kubebuilder:validation:Optional
	ResourceQuotaReference string `json:"resourceQuotaReference,omitempty"`
	// LimitRangeReference is the name of the NamespaceLimitRange resource to be applied to tenant namespaces.
	// This should reference a NamespaceLimitRange resource.
	// If specified, the operator will apply the referenced NamespaceLimitRange to all tenant namespaces.
	// +kubebuilder:validation:Optional
	LimitRangeReference string `json:"limitRangeReference,omitempty"`
	// ConfigSpec contains additional configuration options for the multi-tenant environment.
	// These options are applied globally to all tenant namespaces.
	ConfigSpec ConfigSpec `json:"configSpec,omitempty"`
	// RoleBindings is a list of RoleBinding specifications to be applied to all tenant namespaces.
	// If specified, the operator will create or update the specified RoleBindings in all tenant namespaces.
	// +kubebuilder:validation:Optional
	RoleBindings []RoleBindingSpec `json:"roleBindings,omitempty"`
	// Namespaces is a list of namespace specifications for the multi-tenant environment.
	// +kubebuilder:validation:required
	Namespaces []NamespaceSpec `json:"namespaces,omitempty"`
}

type ConfigSpec struct {
	// EnableAuditLogging indicates whether OVN audit logging should be enabled for tenant namespaces. If not specified, it defaults to false.
	// When enabled, the operator will annotate the Namespace for each tenant with the necessary configuration to enable OVN audit logging.
	// +kubebuilder:default:=false
	EnableAuditLogging bool `json:"enableAuditLogging,omitempty"`
	// EnableUserWorkloadMonitoring indicates whether user workload monitoring should be enabled for tenant namespaces. If not specified, it defaults to false.
	// When enabled, the operator will label the Namespace for each tenant with the necessary configuration to enable user workload monitoring.
	// +kubebuilder:default:=false
	EnableUserWorkloadMonitoring bool `json:"enableUserWorkloadMonitoring,omitempty"`
	// EnableCertificateConfigMapCreation indicates whether the operator should create a ConfigMap containing the OpenShift TLS certificate defined inside the proxy resource for each tenant namespaces.
	// If not specified, it defaults to false.
	// +kubebuilder:default:=false
	EnableCertificateConfigMapCreation bool `json:"enableCertificateConfigMapCreation,omitempty"`
	// EnableArgoCDControllerManagement indicates whether the ArgoCD operator receives management permissions for the tenant namespaces.
	// If not specified, it defaults to false.
	// +kubebuilder:default:=false
	EnableArgoCDControllerManagement bool `json:"enableArgoCDControllerManagement,omitempty"`
	// EnableNameSuffix indicates whether the operator should append a suffix to resource names for tenant namespaces. If not specified, it defaults to false.
	// +kubebuilder:default:=false
	EnableNameSuffix bool `json:"enableNameSuffix,omitempty"`
	// EnableNamePrefix indicates whether the operator should prepend a prefix to resource names for tenant namespaces. If not specified, it defaults to false.
	// +kubebuilder:default:=false
	EnableNamePrefix bool `json:"enableNamePrefix,omitempty"`
	// EnableNetworkPolicyIngressDenyAll determines whether the operator should create a default NetworkPolicy that denies all ingress traffic to pods in tenant namespaces.
	// If not specified, it defaults to false.
	// +kubebuilder:default:=false
	EnableNetworkPolicyIngressDenyAll bool `json:"enableNetworkPolicyIngressDenyAll,omitempty"`
	// EnableNetworkPolicyEgressDenyAll determines whether the operator should create a default NetworkPolicy that denies all egress traffic from pods in tenant namespaces.
	// If not specified, it defaults to false.
	// +kubebuilder:default:=false
	EnableNetworkPolicyEgressDenyAll bool `json:"enableNetworkPolicyEgressDenyAll,omitempty"`
	// EnableNetworkPolicyTenantInternalAllow determines whether the operator should create a NetworkPolicy that allows all traffic between pods part of the same tenant (across namespaces).
	// If not specified, it defaults to false.
	// +kubebuilder:default:=false
	EnableNetworkPolicyTenantInternalAllow bool `json:"enableNetworkPolicyTenantInternalAllow,omitempty"`
}

type RoleBindingSpec struct {
	Name     string           `json:"name,omitempty"`
	RoleRef  rbacv1.RoleRef   `json:"roleRef,omitempty"`
	Subjects []rbacv1.Subject `json:"subjects,omitempty"`
}

type ArgoCDSpec struct {
	// InstanceName is the name of the Argo CD instance. If not specified, it defaults to "openshift-gitops".
	// +kubebuilder:default:=openshift-gitops
	InstanceName string `json:"instanceName,omitempty"`
	// InstanceNamespace is the namespace where the Argo CD instance is installed. If not specified, it defaults to "openshift-gitops".
	// +kubebuilder:default:=openshift-gitops
	InstanceNamespace string `json:"instanceNamespace,omitempty"`
	// Project contains the Argo CD project configuration to be applied to tenant namespaces. If not specified, no Argo CD project will be created.
	// +kubebuilder:validation:Optional
	Project *ArgoCDProjectSpec `json:"project,omitempty"`
}

type ArgoCDProjectSpec struct {
	// Enabled indicates whether the Argo CD project should be created for tenant namespaces. If not specified, it defaults to false.
	// +kubebuilder:default:=false
	Enabled bool `json:"enabled,omitempty"`

	// Name is the name of the Argo CD project to be created for tenant namespaces. This field is required if Enabled is true.
	// The name of the Argo CD project will be used as a base name for the project created for each tenant namespace.
	// For example, if the project name is "tenant-project" and EnableNameSuffix is true, the Argo CD project for a tenant named "tenant1" will be named "tenant1-tenant-project".
	// If EnableNameSuffix and EnableNamePrefix are both false, the Argo CD project will be named exactly as specified in this field.
	// +kubebuilder:validation:Required
	Name string `json:"name,omitempty"`

	argocdv1alpha1.AppProjectSpec `json:",inline"`
}

type NamespaceSpec struct {
	// Name is the name of the namespace to be created for this tenant.
	// If EnableNameSuffix or EnableNamePrefix is true, the name of the tenant will be used as a base name for the namespaces.
	// For example, if the tenant name is "tenant1" and EnableNameSuffix is true, the namespace will be named "tenant1-<name>".
	Name string `json:"name,omitempty"`
	// ConfigSpec contains additional configuration options for the multi-tenant environment.
	// These options are applied globally to all tenant namespaces.
	// If a field is set to true in both the NamespaceSpec.ConfigSpec and the MultiTenantConfigSpec.ConfigSpec, the value from NamespaceSpec.ConfigSpec takes precedence for that namespace.
	// This allows for per-namespace overrides of the global configuration options defined in MultiTenantConfigSpec.ConfigSpec.
	// For example, if EnableAuditLogging is set to true in MultiTenantConfigSpec.ConfigSpec and set to false in a specific NamespaceSpec.ConfigSpec, audit logging will be enabled for all namespaces except the one with the override, where it will be disabled.
	// If a field is set to true in NamespaceSpec.ConfigSpec but not set or set to false in MultiTenantConfigSpec.ConfigSpec, the value from NamespaceSpec.ConfigSpec will enable that feature for the specific namespace, while it remains disabled for namespaces that do not have it enabled in their NamespaceSpec.ConfigSpec.
	// This design allows for flexible configuration of tenant namespaces, enabling global defaults while also supporting specific overrides on a per-namespace basis.
	// +kubebuilder:validation:Optional
	ConfigSpec *ConfigSpec `json:"configSpec,omitempty"`
	// RoleBindings is a list of RoleBinding specifications to be applied to all tenant namespaces.
	RoleBindings []RoleBindingSpec `json:"roleBindings,omitempty"`
}

// GetMergedConfigSpec returns a ConfigSpec that merges the global configuration from the MultiTenantConfigSpec with any overrides specified in the NamespaceSpec.
// The merging logic is as follows:
// - If a field is set to true in both the NamespaceSpec.ConfigSpec and the MultiTenantConfigSpec.ConfigSpec, the value from NamespaceSpec.ConfigSpec takes precedence for that namespace.
// - If a field is set to true in NamespaceSpec.ConfigSpec but not set or set to false in MultiTenantConfigSpec.ConfigSpec, the value from NamespaceSpec.ConfigSpec will enable that feature for the specific namespace, while it remains disabled for namespaces that do not have it enabled in their NamespaceSpec.ConfigSpec.
// This design allows for flexible configuration of tenant namespaces, enabling global defaults while also supporting specific overrides on a per-namespace basis.
func (ns *NamespaceSpec) GetMergedConfigSpec(global ConfigSpec) ConfigSpec {
	merged := global
	if ns.ConfigSpec != nil {
		if ns.ConfigSpec.EnableAuditLogging != global.EnableAuditLogging {
			merged.EnableAuditLogging = ns.ConfigSpec.EnableAuditLogging
		}
		if ns.ConfigSpec.EnableUserWorkloadMonitoring != global.EnableUserWorkloadMonitoring {
			merged.EnableUserWorkloadMonitoring = ns.ConfigSpec.EnableUserWorkloadMonitoring
		}
		if ns.ConfigSpec.EnableCertificateConfigMapCreation != global.EnableCertificateConfigMapCreation {
			merged.EnableCertificateConfigMapCreation = ns.ConfigSpec.EnableCertificateConfigMapCreation
		}
		if ns.ConfigSpec.EnableArgoCDControllerManagement != global.EnableArgoCDControllerManagement {
			merged.EnableArgoCDControllerManagement = ns.ConfigSpec.EnableArgoCDControllerManagement
		}
		if ns.ConfigSpec.EnableNameSuffix != global.EnableNameSuffix {
			merged.EnableNameSuffix = ns.ConfigSpec.EnableNameSuffix
		}
		if ns.ConfigSpec.EnableNamePrefix != global.EnableNamePrefix {
			merged.EnableNamePrefix = ns.ConfigSpec.EnableNamePrefix
		}
	}
	return merged
}

// MultiTenantConfigStatus defines the observed state of MultiTenantConfig.
type MultiTenantConfigStatus struct {
	// ManagedNamespaceCount is the number of namespaces currently managed by this MultiTenantConfig.
	ManagedNamespaceCount int `json:"managedNamespaceCount,omitempty"`
	// QuotaReference is the name of the NamespaceResourceQuota resource currently applied to tenant namespaces.
	QuotaReference string `json:"quotaReference,omitempty"`
	// LimitRangeReference is the name of the NamespaceLimitRange resource currently applied to tenant namespaces.
	LimitRangeReference string `json:"limitRangeReference,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,shortName=mtc
// +kubebuilder:printcolumn:name="Managed Namespaces",type="integer",JSONPath=".status.managedNamespaceCount"
// +kubebuilder:printcolumn:name="Quota Reference",type="string",JSONPath=".status.quotaReference"
// +kubebuilder:printcolumn:name="LimitRange Reference",type="string",JSONPath=".status.limitRangeReference"

// MultiTenantConfig is the Schema for the multitenantconfigs API.
type MultiTenantConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MultiTenantConfigSpec   `json:"spec,omitempty"`
	Status MultiTenantConfigStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// MultiTenantConfigList contains a list of MultiTenantConfig.
type MultiTenantConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MultiTenantConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MultiTenantConfig{}, &MultiTenantConfigList{})
}
