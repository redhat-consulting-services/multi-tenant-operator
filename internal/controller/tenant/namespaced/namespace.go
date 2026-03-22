package namespaced

import (
	"context"
	"encoding/json"
	"fmt"

	tenantv1alpha1 "github.com/redhat-consulting-services/multi-tenant-operator/api/tenant/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func CreateOrUpdateNamespaces(ctx context.Context, client client.Client, mtc *tenantv1alpha1.MultiTenantConfig) ([]string, error) {
	var namespaceNames []string
	for _, ns := range mtc.Spec.Namespaces {
		if ns.Name == "" {
			continue
		}

		namespaceName := getGeneratedNamespaceName(*mtc, ns)
		namespace := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{Name: namespaceName},
		}
		namespaceNames = append(namespaceNames, namespaceName)

		_, err := controllerutil.CreateOrUpdate(ctx, client, namespace, func() error {
			// labels
			if namespace.Labels == nil {
				namespace.Labels = map[string]string{}
			}
			for key, value := range mtc.Labels {
				namespace.Labels[key] = value
			}
			namespace.Labels[managedNamespacetenantNameLabelKey] = mtc.Name
			namespace.Labels[managedByLabelKey] = managedByLabelValue
			namespace.Labels[multiTenantConfigNameLabelKey] = mtc.Name
			if mtc.Spec.ConfigSpec.EnableUserWorkloadMonitoring {
				namespace.Labels["openshift.io/user-monitoring"] = "true"
			}

			// annotations
			if namespace.Annotations == nil {
				namespace.Annotations = map[string]string{}
			}
			for key, value := range mtc.Annotations {
				namespace.Annotations[key] = value
			}
			if mtc.Spec.ConfigSpec.EnableAuditLogging {
				btc, err := json.Marshal(map[string]any{
					"deny":  "info",
					"allow": "info",
				})
				if err != nil {
					return fmt.Errorf("failed to marshal ACL logging config: %w", err)
				}
				namespace.Labels["k8s.ovn.org/acl-logging"] = string(btc)
			}

			// owner reference
			if err := controllerutil.SetControllerReference(mtc, namespace, client.Scheme()); err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create or update namespace %q: %w", namespaceName, err)
		}
	}
	return namespaceNames, nil
}
