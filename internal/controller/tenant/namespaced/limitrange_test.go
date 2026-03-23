package namespaced

import (
	"context"
	"testing"

	tenantv1alpha1 "github.com/redhat-consulting-services/multi-tenant-operator/api/tenant/v1alpha1"
	tenantconfigv1alpha1 "github.com/redhat-consulting-services/multi-tenant-operator/api/tenantconfig/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestCreateOrUpdateLimitRangesSkipsWhenReferenceNotSet(t *testing.T) {
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
	limitRangeSpec := &tenantconfigv1alpha1.NamespaceLimitRange{}

	err := CreateOrUpdateLimitRanges(context.Background(), cl, mtc, limitRangeSpec, []string{"ns-a", "ns-b"})
	if err != nil {
		t.Fatalf("CreateOrUpdateLimitRanges returned error: %v", err)
	}

	limitRangeList := &corev1.LimitRangeList{}
	if err := cl.List(context.Background(), limitRangeList); err != nil {
		t.Fatalf("failed to list limitranges: %v", err)
	}
	if len(limitRangeList.Items) != 0 {
		t.Fatalf("expected no limitranges to be created, got %d", len(limitRangeList.Items))
	}
}

func TestCreateOrUpdateLimitRangesCreatesPerNamespace(t *testing.T) {
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
			UID:  types.UID("ffb55fd7-a01b-4fef-ae4e-100ced7e4f60"),
		},
		Spec: tenantv1alpha1.MultiTenantConfigSpec{LimitRangeReference: "default-limits"},
	}
	limitRangeSpec := &tenantconfigv1alpha1.NamespaceLimitRange{
		Spec: tenantconfigv1alpha1.NamespaceLimitRangeSpec{
			Limits: []corev1.LimitRangeItem{{
				Type:                 corev1.LimitTypeContainer,
				DefaultRequest:       corev1.ResourceList{corev1.ResourceCPU: resourceMustParse(t, "50m")},
				Default:              corev1.ResourceList{corev1.ResourceCPU: resourceMustParse(t, "100m")},
				Max:                  corev1.ResourceList{corev1.ResourceMemory: resourceMustParse(t, "1Gi")},
				Min:                  corev1.ResourceList{corev1.ResourceMemory: resourceMustParse(t, "128Mi")},
				MaxLimitRequestRatio: corev1.ResourceList{corev1.ResourceCPU: resourceMustParse(t, "2")},
			}},
		},
	}

	err := CreateOrUpdateLimitRanges(context.Background(), cl, mtc, limitRangeSpec, []string{"team-a", "team-b"})
	if err != nil {
		t.Fatalf("CreateOrUpdateLimitRanges returned error: %v", err)
	}

	for _, namespace := range []string{"team-a", "team-b"} {
		lr := &corev1.LimitRange{}
		if err := cl.Get(context.Background(), client.ObjectKey{Namespace: namespace, Name: "tenant-limit-range"}, lr); err != nil {
			t.Fatalf("failed to get limitrange for namespace %q: %v", namespace, err)
		}

		if got := lr.Labels[managedNamespacetenantNameLabelKey]; got != tenantB {
			t.Fatalf("tenant-name label mismatch for namespace %q: got %q, want %q", namespace, got, tenantB)
		}
		if got := lr.Labels[managedByLabelKey]; got != managedByLabelValue {
			t.Fatalf("managed-by label mismatch for namespace %q: got %q, want %q", namespace, got, managedByLabelValue)
		}
		if got := lr.Labels[multiTenantConfigNameLabelKey]; got != tenantB {
			t.Fatalf("multitenantconfig label mismatch for namespace %q: got %q, want %q", namespace, got, tenantB)
		}

		if len(lr.Spec.Limits) != 1 {
			t.Fatalf("expected one limit item for namespace %q, got %d", namespace, len(lr.Spec.Limits))
		}
		if lr.Spec.Limits[0].Type != corev1.LimitTypeContainer {
			t.Fatalf("limit type mismatch for namespace %q: got %q", namespace, lr.Spec.Limits[0].Type)
		}
		if got := lr.Spec.Limits[0].Default.Cpu().String(); got != "100m" {
			t.Fatalf("default cpu mismatch for namespace %q: got %q, want %q", namespace, got, "100m")
		}
		if got := lr.Spec.Limits[0].Min.Memory().String(); got != "128Mi" {
			t.Fatalf("min memory mismatch for namespace %q: got %q, want %q", namespace, got, "128Mi")
		}

		if len(lr.OwnerReferences) != 1 {
			t.Fatalf("expected one owner reference for namespace %q, got %d", namespace, len(lr.OwnerReferences))
		}
		ownerRef := lr.OwnerReferences[0]
		if ownerRef.Kind != mtcKind || ownerRef.Name != tenantB {
			t.Fatalf("owner reference mismatch for namespace %q: got kind=%q name=%q", namespace, ownerRef.Kind, ownerRef.Name)
		}
	}
}

func TestCreateOrUpdateLimitRangesUpdatesExistingLimitRange(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatalf("failed to add corev1 scheme: %v", err)
	}
	if err := tenantv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("failed to add tenantv1alpha1 scheme: %v", err)
	}

	existing := &corev1.LimitRange{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "tenant-limit-range",
			Namespace: "team-c",
			Labels: map[string]string{
				"custom":          keepMe,
				managedByLabelKey: "old-value",
			},
		},
		Spec: corev1.LimitRangeSpec{
			Limits: []corev1.LimitRangeItem{{
				Type: corev1.LimitTypeContainer,
				Min:  corev1.ResourceList{corev1.ResourceCPU: resourceMustParse(t, "25m")},
			}},
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
			UID:  types.UID("8de05f15-be4f-4ae1-9a84-114bc48fb019"),
		},
		Spec: tenantv1alpha1.MultiTenantConfigSpec{LimitRangeReference: "tenant-limits"},
	}
	limitRangeSpec := &tenantconfigv1alpha1.NamespaceLimitRange{
		Spec: tenantconfigv1alpha1.NamespaceLimitRangeSpec{
			Limits: []corev1.LimitRangeItem{{
				Type: corev1.LimitTypeContainer,
				Min:  corev1.ResourceList{corev1.ResourceCPU: resourceMustParse(t, "100m")},
			}},
		},
	}

	err := CreateOrUpdateLimitRanges(context.Background(), cl, mtc, limitRangeSpec, []string{"team-c"})
	if err != nil {
		t.Fatalf("CreateOrUpdateLimitRanges returned error: %v", err)
	}

	updated := &corev1.LimitRange{}
	if err := cl.Get(context.Background(), client.ObjectKey{Namespace: "team-c", Name: "tenant-limit-range"}, updated); err != nil {
		t.Fatalf("failed to get updated limitrange: %v", err)
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

	if len(updated.Spec.Limits) != 1 {
		t.Fatalf("expected one limit item, got %d", len(updated.Spec.Limits))
	}
	if got := updated.Spec.Limits[0].Min.Cpu().String(); got != "100m" {
		t.Fatalf("min cpu mismatch: got %q, want %q", got, "100m")
	}

	if len(updated.OwnerReferences) != 1 {
		t.Fatalf("expected one owner reference, got %d", len(updated.OwnerReferences))
	}
	ownerRef := updated.OwnerReferences[0]
	if ownerRef.Kind != mtcKind || ownerRef.Name != tenantC {
		t.Fatalf("owner reference mismatch: got kind=%q name=%q", ownerRef.Kind, ownerRef.Name)
	}
}

func resourceMustParse(t *testing.T, value string) resource.Quantity {
	t.Helper()
	q, err := resource.ParseQuantity(value)
	if err != nil {
		t.Fatalf("failed to parse quantity %q: %v", value, err)
	}
	return q
}
