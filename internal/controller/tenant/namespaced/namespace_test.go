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

var (
	multiTenantConfigKind = (&tenantv1alpha1.MultiTenantConfig{}).GroupVersionKind().Kind
)

func TestCreateOrUpdateNamespacesCreatesNamespaces(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatalf("failed to add corev1 scheme: %v", err)
	}

	cl := fake.NewClientBuilder().WithScheme(scheme).Build()
	mtc := &tenantv1alpha1.MultiTenantConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "example-mtc",
			UID:         types.UID("f225581f-3644-40e2-a066-020f17a2f2c1"),
			Labels:      map[string]string{"team": "platform"},
			Annotations: map[string]string{"contact": "platform-team"},
		},
		Spec: tenantv1alpha1.MultiTenantConfigSpec{
			Namespaces: []tenantv1alpha1.NamespaceSpec{{Name: "tenant-a"}},
		},
	}

	if _, err := CreateOrUpdateNamespaces(context.Background(), cl, mtc); err != nil {
		t.Fatalf("CreateOrUpdateNamespaces returned error: %v", err)
	}

	created := &corev1.Namespace{}
	if err := cl.Get(context.Background(), client.ObjectKey{Name: "tenant-a"}, created); err != nil {
		t.Fatalf("failed to get created namespace: %v", err)
	}

	if got := created.Labels[managedByLabelKey]; got != managedByLabelValue {
		t.Fatalf("managed-by label mismatch: got %q, want %q", got, managedByLabelValue)
	}

	if got := created.Labels[multiTenantConfigNameLabelKey]; got != "example-mtc" {
		t.Fatalf("multitenantconfig label mismatch: got %q, want %q", got, "example-mtc")
	}

	if got := created.Labels["team"]; got != "platform" {
		t.Fatalf("custom label mismatch: got %q, want %q", got, "platform")
	}

	if got := created.Annotations["contact"]; got != "platform-team" {
		t.Fatalf("custom annotation mismatch: got %q, want %q", got, "platform-team")
	}

	if len(created.OwnerReferences) != 1 {
		t.Fatalf("expected one owner reference, got %d", len(created.OwnerReferences))
	}

	ownerRef := created.OwnerReferences[0]
	if ownerRef.Kind != multiTenantConfigKind || ownerRef.Name != "example-mtc" {
		t.Fatalf("owner reference mismatch: got kind=%q name=%q", ownerRef.Kind, ownerRef.Name)
	}
}

func TestCreateOrUpdateNamespacesUpdatesExistingNamespace(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatalf("failed to add corev1 scheme: %v", err)
	}

	existing := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "tenant-b",
			Labels: map[string]string{"custom": "kept", managedByLabelKey: "old-value"},
			Annotations: map[string]string{
				"existing": "kept",
			},
		},
	}

	cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(existing).Build()
	mtc := &tenantv1alpha1.MultiTenantConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "mtc-updated",
			UID:         types.UID("8f813bd3-174f-471b-a8d1-f7734ea260d4"),
			Labels:      map[string]string{"team": "tenant-ops"},
			Annotations: map[string]string{"sla": "gold"},
		},
		Spec: tenantv1alpha1.MultiTenantConfigSpec{
			Namespaces: []tenantv1alpha1.NamespaceSpec{{Name: "tenant-b"}},
		},
	}

	if _, err := CreateOrUpdateNamespaces(context.Background(), cl, mtc); err != nil {
		t.Fatalf("CreateOrUpdateNamespaces returned error: %v", err)
	}

	updated := &corev1.Namespace{}
	if err := cl.Get(context.Background(), client.ObjectKey{Name: "tenant-b"}, updated); err != nil {
		t.Fatalf("failed to get updated namespace: %v", err)
	}

	if got := updated.Labels["custom"]; got != "kept" {
		t.Fatalf("custom label should be preserved, got %q", got)
	}

	if got := updated.Labels[managedByLabelKey]; got != managedByLabelValue {
		t.Fatalf("managed-by label mismatch: got %q, want %q", got, managedByLabelValue)
	}

	if got := updated.Labels[multiTenantConfigNameLabelKey]; got != "mtc-updated" {
		t.Fatalf("multitenantconfig label mismatch: got %q, want %q", got, "mtc-updated")
	}

	if got := updated.Labels["team"]; got != "tenant-ops" {
		t.Fatalf("custom label mismatch: got %q, want %q", got, "tenant-ops")
	}

	if got := updated.Annotations["existing"]; got != "kept" {
		t.Fatalf("existing annotation should be preserved, got %q", got)
	}

	if got := updated.Annotations["sla"]; got != "gold" {
		t.Fatalf("custom annotation mismatch: got %q, want %q", got, "gold")
	}

	if len(updated.OwnerReferences) != 1 {
		t.Fatalf("expected one owner reference, got %d", len(updated.OwnerReferences))
	}

	ownerRef := updated.OwnerReferences[0]
	if ownerRef.Kind != multiTenantConfigKind || ownerRef.Name != "mtc-updated" {
		t.Fatalf("owner reference mismatch: got kind=%q name=%q", ownerRef.Kind, ownerRef.Name)
	}
}

func TestCreateOrUpdateNamespacesSkipsEmptyNamespaceNames(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatalf("failed to add corev1 scheme: %v", err)
	}

	cl := fake.NewClientBuilder().WithScheme(scheme).Build()
	mtc := &tenantv1alpha1.MultiTenantConfig{
		ObjectMeta: metav1.ObjectMeta{Name: "example-mtc"},
		Spec: tenantv1alpha1.MultiTenantConfigSpec{
			Namespaces: []tenantv1alpha1.NamespaceSpec{{Name: ""}},
		},
	}

	if _, err := CreateOrUpdateNamespaces(context.Background(), cl, mtc); err != nil {
		t.Fatalf("CreateOrUpdateNamespaces returned error: %v", err)
	}

	nsList := &corev1.NamespaceList{}
	if err := cl.List(context.Background(), nsList); err != nil {
		t.Fatalf("failed to list namespaces: %v", err)
	}

	if len(nsList.Items) != 0 {
		t.Fatalf("expected no namespaces to be created, got %d", len(nsList.Items))
	}
}

func TestCreateOrUpdateNamespacesAddsPrefixWhenEnabled(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatalf("failed to add corev1 scheme: %v", err)
	}

	cl := fake.NewClientBuilder().WithScheme(scheme).Build()
	mtc := &tenantv1alpha1.MultiTenantConfig{
		ObjectMeta: metav1.ObjectMeta{Name: "tenant-x"},
		Spec: tenantv1alpha1.MultiTenantConfigSpec{
			ConfigSpec: tenantv1alpha1.ConfigSpec{EnableNamePrefix: true},
			Namespaces: []tenantv1alpha1.NamespaceSpec{{Name: "app"}},
		},
	}

	if _, err := CreateOrUpdateNamespaces(context.Background(), cl, mtc); err != nil {
		t.Fatalf("CreateOrUpdateNamespaces returned error: %v", err)
	}

	created := &corev1.Namespace{}
	if err := cl.Get(context.Background(), client.ObjectKey{Name: "tenant-x-app"}, created); err != nil {
		t.Fatalf("failed to get prefixed namespace: %v", err)
	}
}

func TestCreateOrUpdateNamespacesAddsSuffixWhenEnabled(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatalf("failed to add corev1 scheme: %v", err)
	}

	cl := fake.NewClientBuilder().WithScheme(scheme).Build()
	mtc := &tenantv1alpha1.MultiTenantConfig{
		ObjectMeta: metav1.ObjectMeta{Name: "tenant-y"},
		Spec: tenantv1alpha1.MultiTenantConfigSpec{
			ConfigSpec: tenantv1alpha1.ConfigSpec{EnableNameSuffix: true},
			Namespaces: []tenantv1alpha1.NamespaceSpec{{Name: "app"}},
		},
	}

	if _, err := CreateOrUpdateNamespaces(context.Background(), cl, mtc); err != nil {
		t.Fatalf("CreateOrUpdateNamespaces returned error: %v", err)
	}

	created := &corev1.Namespace{}
	if err := cl.Get(context.Background(), client.ObjectKey{Name: "app-tenant-y"}, created); err != nil {
		t.Fatalf("failed to get suffixed namespace: %v", err)
	}
}
