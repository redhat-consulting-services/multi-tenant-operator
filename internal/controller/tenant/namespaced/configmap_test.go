package namespaced

import (
	"context"
	"testing"

	tenantv1alpha1 "github.com/redhat-consulting-services/multi-tenant-operator/api/tenant/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestCreateOrUpdateConfigMapsSkipsWhenDisabled(t *testing.T) {
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
		Spec: tenantv1alpha1.MultiTenantConfigSpec{
			ConfigSpec: tenantv1alpha1.ConfigSpec{EnableCertificateConfigMapCreation: false},
		},
	}

	err := CreateOrUpdateConfigMaps(context.Background(), cl, mtc, []string{"ns-a", "ns-b"})
	if err != nil {
		t.Fatalf("CreateOrUpdateConfigMaps returned error: %v", err)
	}

	cmList := &corev1.ConfigMapList{}
	if err := cl.List(context.Background(), cmList); err != nil {
		t.Fatalf("failed to list configmaps: %v", err)
	}
	if len(cmList.Items) != 0 {
		t.Fatalf("expected no configmaps to be created, got %d", len(cmList.Items))
	}
}

func TestCreateOrUpdateConfigMapsCreatesPerNamespace(t *testing.T) {
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
			UID:  types.UID("2a8fb8ad-5e02-4d66-8cae-f0f379f4fdd9"),
		},
		Spec: tenantv1alpha1.MultiTenantConfigSpec{
			ConfigSpec: tenantv1alpha1.ConfigSpec{EnableCertificateConfigMapCreation: true},
		},
	}

	err := CreateOrUpdateConfigMaps(context.Background(), cl, mtc, []string{"team-a", "team-b"})
	if err != nil {
		t.Fatalf("CreateOrUpdateConfigMaps returned error: %v", err)
	}

	for _, namespace := range []string{"team-a", "team-b"} {
		cm := &corev1.ConfigMap{}
		if err := cl.Get(context.Background(), client.ObjectKey{Namespace: namespace, Name: "user-ca-bundle"}, cm); err != nil {
			t.Fatalf("failed to get configmap for namespace %q: %v", namespace, err)
		}

		if got := cm.Labels[managedNamespacetenantNameLabelKey]; got != tenantB {
			t.Fatalf("tenant-name label mismatch for namespace %q: got %q, want %q", namespace, got, tenantB)
		}
		if got := cm.Labels[managedByLabelKey]; got != managedByLabelValue {
			t.Fatalf("managed-by label mismatch for namespace %q: got %q, want %q", namespace, got, managedByLabelValue)
		}
		if got := cm.Labels[multiTenantConfigNameLabelKey]; got != tenantB {
			t.Fatalf("multitenantconfig label mismatch for namespace %q: got %q, want %q", namespace, got, tenantB)
		}
		if got := cm.Labels["config.openshift.io/inject-trusted-cabundle"]; got != trueKeyValue {
			t.Fatalf("inject-trusted-cabundle label mismatch for namespace %q: got %q, want %q", namespace, got, trueKeyValue)
		}

		if len(cm.OwnerReferences) != 1 {
			t.Fatalf("expected one owner reference for namespace %q, got %d", namespace, len(cm.OwnerReferences))
		}
		ownerRef := cm.OwnerReferences[0]
		if ownerRef.Kind != mtcKind || ownerRef.Name != tenantB {
			t.Fatalf("owner reference mismatch for namespace %q: got kind=%q name=%q", namespace, ownerRef.Kind, ownerRef.Name)
		}
	}
}

func TestCreateOrUpdateConfigMapsUpdatesExistingConfigMap(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatalf("failed to add corev1 scheme: %v", err)
	}
	if err := tenantv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("failed to add tenantv1alpha1 scheme: %v", err)
	}

	existing := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "user-ca-bundle",
			Namespace: "team-c",
			Labels: map[string]string{
				"custom":             keepMe,
				managedByLabelKey:    "old-value",
				"obsolete-managedby": "still-present",
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
			UID:  types.UID("4ab31f20-d693-4e48-84cb-bff8cc8d7d29"),
		},
		Spec: tenantv1alpha1.MultiTenantConfigSpec{
			ConfigSpec: tenantv1alpha1.ConfigSpec{EnableCertificateConfigMapCreation: true},
		},
	}

	err := CreateOrUpdateConfigMaps(context.Background(), cl, mtc, []string{"team-c"})
	if err != nil {
		t.Fatalf("CreateOrUpdateConfigMaps returned error: %v", err)
	}

	updated := &corev1.ConfigMap{}
	if err := cl.Get(context.Background(), client.ObjectKey{Namespace: "team-c", Name: "user-ca-bundle"}, updated); err != nil {
		t.Fatalf("failed to get updated configmap: %v", err)
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
	if got := updated.Labels["config.openshift.io/inject-trusted-cabundle"]; got != trueKeyValue {
		t.Fatalf("inject-trusted-cabundle label mismatch: got %q, want %q", got, trueKeyValue)
	}

	if len(updated.OwnerReferences) != 1 {
		t.Fatalf("expected one owner reference, got %d", len(updated.OwnerReferences))
	}
	ownerRef := updated.OwnerReferences[0]
	if ownerRef.Kind != mtcKind || ownerRef.Name != tenantC {
		t.Fatalf("owner reference mismatch: got kind=%q name=%q", ownerRef.Kind, ownerRef.Name)
	}
}
