package namespaced

import (
	"context"

	tenantv1alpha1 "github.com/redhat-consulting-services/multi-tenant-operator/api/tenant/v1alpha1"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func CreateOrUpdateNetworkPolicies(ctx context.Context, cl client.Client, mtc *tenantv1alpha1.MultiTenantConfig, namespaces []string) error {
	for _, ns := range namespaces {
		if err := createOrUpdateNetworkPolicy(ctx, cl, mtc, ns); err != nil {
			return err
		}
	}
	return nil
}

func createOrUpdateNetworkPolicy(ctx context.Context, cl client.Client, mtc *tenantv1alpha1.MultiTenantConfig, namespace string) error {
	networkPolicy := &netv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mtc-namespace-config",
			Namespace: namespace,
		},
		Spec: netv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{},
		},
	}

	_, err := controllerutil.CreateOrUpdate(ctx, cl, networkPolicy, func() error {
		if mtc.Spec.ConfigSpec.EnableNetworkPolicyTenantInternalAllow {
			// allow ingress and egress traffic within the tenant namespaces

			networkPolicy.Spec.PolicyTypes = []netv1.PolicyType{netv1.PolicyTypeIngress}
			networkPolicy.Spec.Ingress = []netv1.NetworkPolicyIngressRule{
				{
					From: []netv1.NetworkPolicyPeer{
						{
							NamespaceSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									managedNamespacetenantNameLabelKey: mtc.Name,
									managedByLabelKey:                  managedByLabelValue,
								},
							},
						},
					},
				},
			}
			networkPolicy.Spec.PolicyTypes = append(networkPolicy.Spec.PolicyTypes, netv1.PolicyTypeEgress)
			networkPolicy.Spec.Egress = []netv1.NetworkPolicyEgressRule{
				{
					To: []netv1.NetworkPolicyPeer{
						{
							NamespaceSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									managedNamespacetenantNameLabelKey: mtc.Name,
									managedByLabelKey:                  managedByLabelValue,
								},
							},
						},
					},
				},
			}
		} else {
			if mtc.Spec.ConfigSpec.EnableNetworkPolicyIngressDenyAll {
				networkPolicy.Spec.PolicyTypes = []netv1.PolicyType{netv1.PolicyTypeIngress}
				networkPolicy.Spec.Ingress = []netv1.NetworkPolicyIngressRule{}
			}

			if mtc.Spec.ConfigSpec.EnableNetworkPolicyEgressDenyAll {
				networkPolicy.Spec.PolicyTypes = append(networkPolicy.Spec.PolicyTypes, netv1.PolicyTypeEgress)
				networkPolicy.Spec.Egress = []netv1.NetworkPolicyEgressRule{}
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
