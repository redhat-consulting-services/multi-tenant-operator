package namespaced

import (
	"context"
	"reflect"
	"testing"

	tenantv1alpha1 "github.com/redhat-consulting-services/multi-tenant-operator/api/tenant/v1alpha1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestCreateOrUpdateRoleBindingsSkipsWhenNoRoleBindingsConfigured(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := rbacv1.AddToScheme(scheme); err != nil {
		t.Fatalf("failed to add rbacv1 scheme: %v", err)
	}
	if err := tenantv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("failed to add tenantv1alpha1 scheme: %v", err)
	}

	cl := fake.NewClientBuilder().WithScheme(scheme).Build()
	mtc := &tenantv1alpha1.MultiTenantConfig{
		ObjectMeta: metav1.ObjectMeta{Name: "tenant-a"},
	}

	err := CreateOrUpdateRoleBindings(context.Background(), cl, mtc, []string{"ns-a", "ns-b"})
	if err != nil {
		t.Fatalf("CreateOrUpdateRoleBindings returned error: %v", err)
	}

	rbList := &rbacv1.RoleBindingList{}
	if err := cl.List(context.Background(), rbList); err != nil {
		t.Fatalf("failed to list rolebindings: %v", err)
	}
	if len(rbList.Items) != 0 {
		t.Fatalf("expected no rolebindings to be created, got %d", len(rbList.Items))
	}
}

func TestCreateOrUpdateRoleBindingsCreatesPerNamespace(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := rbacv1.AddToScheme(scheme); err != nil {
		t.Fatalf("failed to add rbacv1 scheme: %v", err)
	}
	if err := tenantv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("failed to add tenantv1alpha1 scheme: %v", err)
	}

	cl := fake.NewClientBuilder().WithScheme(scheme).Build()
	mtc := &tenantv1alpha1.MultiTenantConfig{
		TypeMeta: metav1.TypeMeta{
			APIVersion: tenantv1alpha1.GroupVersion.String(),
			Kind:       "MultiTenantConfig",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "tenant-b",
			UID:  types.UID("0dccff11-f8f8-43e9-a7a5-d2892e15f39d"),
		},
		Spec: tenantv1alpha1.MultiTenantConfigSpec{
			RoleBindings: []tenantv1alpha1.RoleBindingSpec{
				{
					Name: "viewers",
					RoleRef: rbacv1.RoleRef{
						APIGroup: rbacv1.GroupName,
						Kind:     "ClusterRole",
						Name:     "view",
					},
					Subjects: []rbacv1.Subject{
						{Kind: rbacv1.GroupKind, Name: "team-viewers", APIGroup: rbacv1.GroupName},
					},
				},
				{
					Name: "editors",
					RoleRef: rbacv1.RoleRef{
						APIGroup: rbacv1.GroupName,
						Kind:     "ClusterRole",
						Name:     "edit",
					},
					Subjects: []rbacv1.Subject{
						{Kind: rbacv1.UserKind, Name: "alice", APIGroup: rbacv1.GroupName},
					},
				},
			},
		},
	}

	err := CreateOrUpdateRoleBindings(context.Background(), cl, mtc, []string{"team-a", "team-b"})
	if err != nil {
		t.Fatalf("CreateOrUpdateRoleBindings returned error: %v", err)
	}

	expected := map[string]tenantv1alpha1.RoleBindingSpec{}
	for _, rb := range mtc.Spec.RoleBindings {
		expected[rb.Name] = rb
	}

	for _, namespace := range []string{"team-a", "team-b"} {
		for name, wantSpec := range expected {
			rb := &rbacv1.RoleBinding{}
			if err := cl.Get(context.Background(), client.ObjectKey{Namespace: namespace, Name: name}, rb); err != nil {
				t.Fatalf("failed to get rolebinding %q for namespace %q: %v", name, namespace, err)
			}

			if got := rb.Labels[managedNamespacetenantNameLabelKey]; got != "tenant-b" {
				t.Fatalf("tenant-name label mismatch for namespace %q and rolebinding %q: got %q, want %q", namespace, name, got, "tenant-b")
			}
			if got := rb.Labels[managedByLabelKey]; got != managedByLabelValue {
				t.Fatalf("managed-by label mismatch for namespace %q and rolebinding %q: got %q, want %q", namespace, name, got, managedByLabelValue)
			}
			if got := rb.Labels[multiTenantConfigNameLabelKey]; got != "tenant-b" {
				t.Fatalf("multitenantconfig label mismatch for namespace %q and rolebinding %q: got %q, want %q", namespace, name, got, "tenant-b")
			}

			if !reflect.DeepEqual(rb.RoleRef, wantSpec.RoleRef) {
				t.Fatalf("roleref mismatch for namespace %q and rolebinding %q: got %#v, want %#v", namespace, name, rb.RoleRef, wantSpec.RoleRef)
			}
			if !reflect.DeepEqual(rb.Subjects, wantSpec.Subjects) {
				t.Fatalf("subjects mismatch for namespace %q and rolebinding %q: got %#v, want %#v", namespace, name, rb.Subjects, wantSpec.Subjects)
			}

			if len(rb.OwnerReferences) != 1 {
				t.Fatalf("expected one owner reference for namespace %q and rolebinding %q, got %d", namespace, name, len(rb.OwnerReferences))
			}
			ownerRef := rb.OwnerReferences[0]
			if ownerRef.Kind != "MultiTenantConfig" || ownerRef.Name != "tenant-b" {
				t.Fatalf("owner reference mismatch for namespace %q and rolebinding %q: got kind=%q name=%q", namespace, name, ownerRef.Kind, ownerRef.Name)
			}
		}
	}
}

func TestCreateOrUpdateRoleBindingsUpdatesExistingRoleBinding(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := rbacv1.AddToScheme(scheme); err != nil {
		t.Fatalf("failed to add rbacv1 scheme: %v", err)
	}
	if err := tenantv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("failed to add tenantv1alpha1 scheme: %v", err)
	}

	existing := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "access",
			Namespace: "team-c",
			Labels: map[string]string{
				"custom":          "keep-me",
				managedByLabelKey: "old-value",
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "ClusterRole",
			Name:     "view",
		},
		Subjects: []rbacv1.Subject{
			{Kind: rbacv1.GroupKind, Name: "old-group", APIGroup: rbacv1.GroupName},
		},
	}

	cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(existing).Build()
	mtc := &tenantv1alpha1.MultiTenantConfig{
		TypeMeta: metav1.TypeMeta{
			APIVersion: tenantv1alpha1.GroupVersion.String(),
			Kind:       "MultiTenantConfig",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "tenant-c",
			UID:  types.UID("ae4fbf11-2e83-4e8b-a9ec-f9ad6ca17a4e"),
		},
		Spec: tenantv1alpha1.MultiTenantConfigSpec{
			RoleBindings: []tenantv1alpha1.RoleBindingSpec{
				{
					Name: "access",
					RoleRef: rbacv1.RoleRef{
						APIGroup: rbacv1.GroupName,
						Kind:     "ClusterRole",
						Name:     "admin",
					},
					Subjects: []rbacv1.Subject{
						{Kind: rbacv1.UserKind, Name: "bob", APIGroup: rbacv1.GroupName},
						{Kind: rbacv1.GroupKind, Name: "platform-admins", APIGroup: rbacv1.GroupName},
					},
				},
			},
		},
	}

	err := CreateOrUpdateRoleBindings(context.Background(), cl, mtc, []string{"team-c"})
	if err != nil {
		t.Fatalf("CreateOrUpdateRoleBindings returned error: %v", err)
	}

	updated := &rbacv1.RoleBinding{}
	if err := cl.Get(context.Background(), client.ObjectKey{Namespace: "team-c", Name: "access"}, updated); err != nil {
		t.Fatalf("failed to get updated rolebinding: %v", err)
	}

	if got := updated.Labels["custom"]; got != "keep-me" {
		t.Fatalf("custom label should be preserved, got %q", got)
	}
	if got := updated.Labels[managedByLabelKey]; got != managedByLabelValue {
		t.Fatalf("managed-by label mismatch: got %q, want %q", got, managedByLabelValue)
	}
	if got := updated.Labels[multiTenantConfigNameLabelKey]; got != "tenant-c" {
		t.Fatalf("multitenantconfig label mismatch: got %q, want %q", got, "tenant-c")
	}

	wantRoleRef := mtc.Spec.RoleBindings[0].RoleRef
	if !reflect.DeepEqual(updated.RoleRef, wantRoleRef) {
		t.Fatalf("roleref mismatch: got %#v, want %#v", updated.RoleRef, wantRoleRef)
	}
	wantSubjects := mtc.Spec.RoleBindings[0].Subjects
	if !reflect.DeepEqual(updated.Subjects, wantSubjects) {
		t.Fatalf("subjects mismatch: got %#v, want %#v", updated.Subjects, wantSubjects)
	}

	if len(updated.OwnerReferences) != 1 {
		t.Fatalf("expected one owner reference, got %d", len(updated.OwnerReferences))
	}
	ownerRef := updated.OwnerReferences[0]
	if ownerRef.Kind != "MultiTenantConfig" || ownerRef.Name != "tenant-c" {
		t.Fatalf("owner reference mismatch: got kind=%q name=%q", ownerRef.Kind, ownerRef.Name)
	}
}