package namespaced

import (
	"context"
	"errors"

	argocdv1alpha1 "github.com/argoproj/argo-cd/v3/pkg/apis/application/v1alpha1"
	tenantv1alpha1 "github.com/redhat-consulting-services/multi-tenant-operator/api/tenant/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// CreateOrUpdateArgoCDProject creates or updates an Argo CD AppProject in the Argo CD instance namespace based on the provided MultiTenantConfig and Argo CD project configuration. It returns an error if any operation fails.
func CreateOrUpdateArgoCDProject(ctx context.Context, cl client.Client, mtc *tenantv1alpha1.MultiTenantConfig, namespaces []string) error {
	if mtc.Spec.ArgoCD == nil {
		return nil
	}
	if mtc.Spec.ArgoCD.InstanceName == "" {
		return errors.New("Argo CD instance name is required when Argo CD configuration is provided")
	}
	if mtc.Spec.ArgoCD.InstanceNamespace == "" {
		return errors.New("Argo CD instance namespace is required when Argo CD configuration is provided")
	}
	if mtc.Spec.ArgoCD.Project == nil || !mtc.Spec.ArgoCD.Project.Enabled {
		return nil
	}

	return createOrUpdateArgoCDProject(ctx, cl, mtc, namespaces)
}

func createOrUpdateArgoCDProject(ctx context.Context, client client.Client, mtc *tenantv1alpha1.MultiTenantConfig, namespaces []string) error {
	argoCDProject := &argocdv1alpha1.AppProject{
		ObjectMeta: metav1.ObjectMeta{
			Name:      getGeneratedName(*mtc, mtc.Spec.ArgoCD.Project.Name),
			Namespace: mtc.Spec.ArgoCD.InstanceNamespace,
		},
	}

	_, err := controllerutil.CreateOrUpdate(ctx, client, argoCDProject, func() error {
		if argoCDProject.Labels == nil {
			argoCDProject.Labels = map[string]string{}
		}
		argoCDProject.Labels[managedNamespacetenantNameLabelKey] = mtc.Name
		argoCDProject.Labels[managedByLabelKey] = managedByLabelValue
		argoCDProject.Labels[multiTenantConfigNameLabelKey] = mtc.Name

		// set config spec fields
		argoCDProject.Spec.SourceRepos = mtc.Spec.ArgoCD.Project.SourceRepos
		argoCDProject.Spec.Destinations = mtc.Spec.ArgoCD.Project.Destinations
		argoCDProject.Spec.Description = mtc.Spec.ArgoCD.Project.Description
		argoCDProject.Spec.Roles = mtc.Spec.ArgoCD.Project.Roles
		argoCDProject.Spec.ClusterResourceWhitelist = mtc.Spec.ArgoCD.Project.ClusterResourceWhitelist
		argoCDProject.Spec.NamespaceResourceBlacklist = mtc.Spec.ArgoCD.Project.NamespaceResourceBlacklist
		argoCDProject.Spec.OrphanedResources = mtc.Spec.ArgoCD.Project.OrphanedResources
		argoCDProject.Spec.SyncWindows = mtc.Spec.ArgoCD.Project.SyncWindows
		argoCDProject.Spec.NamespaceResourceWhitelist = mtc.Spec.ArgoCD.Project.NamespaceResourceWhitelist
		argoCDProject.Spec.SignatureKeys = mtc.Spec.ArgoCD.Project.SignatureKeys
		argoCDProject.Spec.ClusterResourceBlacklist = mtc.Spec.ArgoCD.Project.ClusterResourceBlacklist
		argoCDProject.Spec.SourceNamespaces = mtc.Spec.ArgoCD.Project.SourceNamespaces
		argoCDProject.Spec.PermitOnlyProjectScopedClusters = mtc.Spec.ArgoCD.Project.PermitOnlyProjectScopedClusters
		argoCDProject.Spec.DestinationServiceAccounts = mtc.Spec.ArgoCD.Project.DestinationServiceAccounts

		// set ownership reference to the MultiTenantConfig
		if err := controllerutil.SetControllerReference(mtc, argoCDProject, client.Scheme()); err != nil {
			return err
		}
		return nil
	})
	return err
}
