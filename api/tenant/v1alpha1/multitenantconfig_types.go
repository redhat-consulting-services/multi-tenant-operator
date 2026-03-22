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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// MultiTenantConfigSpec defines the desired state of MultiTenantConfig.
type MultiTenantConfigSpec struct {
	// ArgoCD contains the configuration for Argo CD integration in the multi-tenant environment.
	ArgoCD *ArgoCDSpec `json:"argocd,omitempty"`
	// QuotaReference is the name of the NamespaceResourceQuota resource to be applied to tenant namespaces.
	QuotaReference string `json:"quotaReference,omitempty"`
	// LimitRangeReference is the name of the NamespaceLimitRange resource to be applied to tenant namespaces.
	LimitRangeReference string `json:"limitRangeReference,omitempty"`
	// ConfigSpec contains additional configuration options for the multi-tenant environment.
	// These options are applied globally to all tenant namespaces.
	ConfigSpec ConfigSpec `json:"configSpec,omitempty"`
	// RoleBindings is a list of RoleBinding specifications to be applied to all tenant namespaces.
	RoleBindings []RoleBindingSpec `json:"roleBindings,omitempty"`
	// Namespaces is a list of namespace specifications for the multi-tenant environment.
	Namespaces []NamespaceSpec `json:"namespaces,omitempty"`
}

type ConfigSpec struct {
	EnableAuditLogging                 bool `json:"enableAuditLogging,omitempty"`
	EnableUserWorkloadMonitoring       bool `json:"enableUserWorkloadMonitoring,omitempty"`
	EnableCertificateConfigMapCreation bool `json:"enableCertificateConfigMapCreation,omitempty"`
	EnableArgoCDControllerManagement   bool `json:"enableArgoCDControllerManagement,omitempty"`
	EnableNameSuffix                   bool `json:"enableNameSuffix,omitempty"`
	EnableNamePrefix                   bool `json:"enableNamePrefix,omitempty"`
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
	ConfigSpec ConfigSpec `json:"configSpec,omitempty"`
	// RoleBindings is a list of RoleBinding specifications to be applied to all tenant namespaces.
	RoleBindings []RoleBindingSpec `json:"roleBindings,omitempty"`
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
// +kubebuilder:resource:shortName=mtc
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
