package namespaced

import (
	"context"

	tenantv1alpha1 "github.com/redhat-consulting-services/multi-tenant-operator/api/tenant/v1alpha1"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	networkPolicyNameTenantInternal      = "mtc-tenant-internal-allow"
	networkPolicyNameDenyAll             = "mtc-deny-all"
	networkPolicyNameNamespaceLocalAllow = "mtc-namespace-local-allow"
)

func CreateOrUpdateNetworkPolicyTenantInternalAllow(ctx context.Context, cl client.Client, mtc *tenantv1alpha1.MultiTenantConfig, namespace tenantv1alpha1.NamespaceSpec) error {
	if !namespace.ConfigSpec.EnableNetworkPolicyTenantInternalAllow {
		return nil
	}

	networkPolicy := &netv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      networkPolicyNameTenantInternal,
			Namespace: namespace.Name,
		},
		Spec: netv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{},
		},
	}

	_, err := controllerutil.CreateOrUpdate(ctx, cl, networkPolicy, func() error {
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
		if err := controllerutil.SetControllerReference(mtc, networkPolicy, cl.Scheme()); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func CreateOrUpdateNetworkPolicyAllDeny(ctx context.Context, cl client.Client, mtc *tenantv1alpha1.MultiTenantConfig, namespace tenantv1alpha1.NamespaceSpec) error {
	if !namespace.ConfigSpec.EnableNetworkPolicyEgressDenyAll && !namespace.ConfigSpec.EnableNetworkPolicyIngressDenyAll {
		return nil
	}

	networkPolicy := &netv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      networkPolicyNameDenyAll,
			Namespace: namespace.Name,
		},
		Spec: netv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{},
		},
	}

	_, err := controllerutil.CreateOrUpdate(ctx, cl, networkPolicy, func() error {
		// allow ingress and egress traffic within the tenant namespaces
		networkPolicy.Spec.PolicyTypes = []netv1.PolicyType{}
		if namespace.ConfigSpec.EnableNetworkPolicyIngressDenyAll {
			networkPolicy.Spec.PolicyTypes = append(networkPolicy.Spec.PolicyTypes, netv1.PolicyTypeIngress)
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
		}
		if namespace.ConfigSpec.EnableNetworkPolicyEgressDenyAll {
			networkPolicy.Spec.PolicyTypes = append(networkPolicy.Spec.PolicyTypes, netv1.PolicyTypeEgress)
			networkPolicy.Spec.Egress = []netv1.NetworkPolicyEgressRule{
				{
					To: []netv1.NetworkPolicyPeer{
						{
							NamespaceSelector: &metav1.LabelSelector{},
						},
						{
							IPBlock: &netv1.IPBlock{
								CIDR: "0.0.0.0/0",
							},
						},
						{
							IPBlock: &netv1.IPBlock{
								CIDR: "::/0",
							},
						},
					},
				},
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	if err := controllerutil.SetControllerReference(mtc, networkPolicy, cl.Scheme()); err != nil {
		return err
	}
	return nil
}

func CreateOrUpdateNetworkPolicyNamespaceLocalAllow(ctx context.Context, cl client.Client, mtc *tenantv1alpha1.MultiTenantConfig, namespace tenantv1alpha1.NamespaceSpec) error {
	if !namespace.ConfigSpec.EnableNetworkPolicyNamespaceLocalAllow {
		return nil
	}

	networkPolicy := &netv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      networkPolicyNameNamespaceLocalAllow,
			Namespace: namespace.Name,
		},
		Spec: netv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{},
		},
	}

	_, err := controllerutil.CreateOrUpdate(ctx, cl, networkPolicy, func() error {
		// allow ingress and egress traffic within the tenant namespaces (namespace-local)
		networkPolicy.Spec.PolicyTypes = []netv1.PolicyType{netv1.PolicyTypeIngress, netv1.PolicyTypeEgress}
		networkPolicy.Spec.Ingress = []netv1.NetworkPolicyIngressRule{
			{
				From: []netv1.NetworkPolicyPeer{
					{
						PodSelector: &metav1.LabelSelector{},
					},
				},
			},
		}
		networkPolicy.Spec.Egress = []netv1.NetworkPolicyEgressRule{
			{
				To: []netv1.NetworkPolicyPeer{
					{
						PodSelector: &metav1.LabelSelector{},
					},
				},
			},
		}
		if err := controllerutil.SetControllerReference(mtc, networkPolicy, cl.Scheme()); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
