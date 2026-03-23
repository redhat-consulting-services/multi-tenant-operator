package namespaced

import (
	"context"
	"testing"

	tenantv1alpha1 "github.com/redhat-consulting-services/multi-tenant-operator/api/tenant/v1alpha1"
	tenantconfigv1alpha1 "github.com/redhat-consulting-services/multi-tenant-operator/api/tenantconfig/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestCreateOrUpdateResourceQuotasSkipsWhenReferenceNotSet(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatalf("failed to add corev1 scheme: %v", err)
	}
	if err := tenantv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("failed to add tenantv1alpha1 scheme: %v", err)
	}

	cl := fake.NewClientBuilder().WithScheme(scheme).Build()
	mtc := &tenantv1alpha1.MultiTenantConfig{
		ObjectMeta: metav1.ObjectMeta{Name: tenantA},
	}
	rqSpec := &tenantconfigv1alpha1.NamespaceResourceQuota{}

	err := CreateOrUpdateResourceQuotas(context.Background(), cl, mtc, rqSpec, []string{"ns-a", "ns-b"})
	if err != nil {
		t.Fatalf("CreateOrUpdateResourceQuotas returned error: %v", err)
	}

	rqList := &corev1.ResourceQuotaList{}
	if err := cl.List(context.Background(), rqList); err != nil {
		t.Fatalf("failed to list resourcequotas: %v", err)
	}
	if len(rqList.Items) != 0 {
		t.Fatalf("expected no resourcequotas to be created, got %d", len(rqList.Items))
	}
}

func TestCreateOrUpdateResourceQuotasCreatesPerNamespace(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatalf("failed to add corev1 scheme: %v", err)
	}
	if err := tenantv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("failed to add tenantv1alpha1 scheme: %v", err)
	}

	cl := fake.NewClientBuilder().WithScheme(scheme).Build()
	mtc := &tenantv1alpha1.MultiTenantConfig{
		TypeMeta: metav1.TypeMeta{
			APIVersion: tenantv1alpha1.GroupVersion.String(),
			Kind:       mtcKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: tenantB,
			UID:  types.UID("57f7e17f-572e-49f0-802f-574dc2f6e2f2"),
		},
		Spec: tenantv1alpha1.MultiTenantConfigSpec{ResourceQuotaReference: "default-quota"},
	}
	rqSpec := &tenantconfigv1alpha1.NamespaceResourceQuota{
		Spec: tenantconfigv1alpha1.NamespaceResourceQuotaSpec{
			ResourceQuotaSpec: corev1.ResourceQuotaSpec{
				Hard: corev1.ResourceList{
					corev1.ResourceCPU:             resourceMustParse(t, "2"),
					corev1.ResourceMemory:          resourceMustParse(t, "4Gi"),
					corev1.ResourcePods:            resourceMustParse(t, "20"),
					corev1.ResourceRequestsStorage: resourceMustParse(t, "10Gi"),
				},
			},
		},
	}

	err := CreateOrUpdateResourceQuotas(context.Background(), cl, mtc, rqSpec, []string{"team-a", "team-b"})
	if err != nil {
		t.Fatalf("CreateOrUpdateResourceQuotas returned error: %v", err)
	}

	for _, namespace := range []string{"team-a", "team-b"} {
		rq := &corev1.ResourceQuota{}
		if err := cl.Get(context.Background(), client.ObjectKey{Namespace: namespace, Name: "tenant-resource-quota"}, rq); err != nil {
			t.Fatalf("failed to get resourcequota for namespace %q: %v", namespace, err)
		}

		if got := rq.Labels[managedNamespacetenantNameLabelKey]; got != tenantB {
			t.Fatalf("tenant-name label mismatch for namespace %q: got %q, want %q", namespace, got, tenantB)
		}
		if got := rq.Labels[managedByLabelKey]; got != managedByLabelValue {
			t.Fatalf("managed-by label mismatch for namespace %q: got %q, want %q", namespace, got, managedByLabelValue)
		}
		if got := rq.Labels[multiTenantConfigNameLabelKey]; got != tenantB {
			t.Fatalf("multitenantconfig label mismatch for namespace %q: got %q, want %q", namespace, got, tenantB)
		}

		if got := rq.Spec.Hard.Cpu().String(); got != "2" {
			t.Fatalf("hard cpu mismatch for namespace %q: got %q, want %q", namespace, got, "2")
		}
		if got := rq.Spec.Hard.Memory().String(); got != "4Gi" {
			t.Fatalf("hard memory mismatch for namespace %q: got %q, want %q", namespace, got, "4Gi")
		}
		if got := rq.Spec.Hard.Pods().String(); got != "20" {
			t.Fatalf("hard pods mismatch for namespace %q: got %q, want %q", namespace, got, "20")
		}

		if len(rq.OwnerReferences) != 1 {
			t.Fatalf("expected one owner reference for namespace %q, got %d", namespace, len(rq.OwnerReferences))
		}
		ownerRef := rq.OwnerReferences[0]
		if ownerRef.Kind != mtcKind || ownerRef.Name != tenantB {
			t.Fatalf("owner reference mismatch for namespace %q: got kind=%q name=%q", namespace, ownerRef.Kind, ownerRef.Name)
		}
	}
}

func TestCreateOrUpdateResourceQuotasUpdatesExistingResourceQuota(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatalf("failed to add corev1 scheme: %v", err)
	}
	if err := tenantv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("failed to add tenantv1alpha1 scheme: %v", err)
	}

	existing := &corev1.ResourceQuota{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "tenant-resource-quota",
			Namespace: "team-c",
			Labels: map[string]string{
				"custom":          keepMe,
				managedByLabelKey: "old-value",
			},
		},
		Spec: corev1.ResourceQuotaSpec{
			Hard: corev1.ResourceList{
				corev1.ResourceCPU: resourceMustParse(t, "1"),
			},
		},
	}

	cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(existing).Build()
	mtc := &tenantv1alpha1.MultiTenantConfig{
		TypeMeta: metav1.TypeMeta{
			APIVersion: tenantv1alpha1.GroupVersion.String(),
			Kind:       mtcKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: tenantC,
			UID:  types.UID("5cb6afce-160e-47f8-b8c0-9120353647f4"),
		},
		Spec: tenantv1alpha1.MultiTenantConfigSpec{ResourceQuotaReference: "tenant-quota"},
	}
	rqSpec := &tenantconfigv1alpha1.NamespaceResourceQuota{
		Spec: tenantconfigv1alpha1.NamespaceResourceQuotaSpec{
			ResourceQuotaSpec: corev1.ResourceQuotaSpec{
				Hard: corev1.ResourceList{
					corev1.ResourceCPU:    resourceMustParse(t, "3"),
					corev1.ResourceMemory: resourceMustParse(t, "6Gi"),
				},
			},
		},
	}

	err := CreateOrUpdateResourceQuotas(context.Background(), cl, mtc, rqSpec, []string{"team-c"})
	if err != nil {
		t.Fatalf("CreateOrUpdateResourceQuotas returned error: %v", err)
	}

	updated := &corev1.ResourceQuota{}
	if err := cl.Get(context.Background(), client.ObjectKey{Namespace: "team-c", Name: "tenant-resource-quota"}, updated); err != nil {
		t.Fatalf("failed to get updated resourcequota: %v", err)
	}

	if got := updated.Labels["custom"]; got != keepMe {
		t.Fatalf("custom label should be preserved, got %q", got)
	}
	if got := updated.Labels[managedByLabelKey]; got != managedByLabelValue {
		t.Fatalf("managed-by label mismatch: got %q, want %q", got, managedByLabelValue)
	}
	if got := updated.Labels[multiTenantConfigNameLabelKey]; got != tenantC {
		t.Fatalf("multitenantconfig label mismatch: got %q, want %q", got, tenantC)
	}

	if got := updated.Spec.Hard.Cpu().String(); got != "3" {
		t.Fatalf("hard cpu mismatch: got %q, want %q", got, "3")
	}
	if got := updated.Spec.Hard.Memory().String(); got != "6Gi" {
		t.Fatalf("hard memory mismatch: got %q, want %q", got, "6Gi")
	}

	if len(updated.OwnerReferences) != 1 {
		t.Fatalf("expected one owner reference, got %d", len(updated.OwnerReferences))
	}
	ownerRef := updated.OwnerReferences[0]
	if ownerRef.Kind != mtcKind || ownerRef.Name != tenantC {
		t.Fatalf("owner reference mismatch: got kind=%q name=%q", ownerRef.Kind, ownerRef.Name)
	}
}
