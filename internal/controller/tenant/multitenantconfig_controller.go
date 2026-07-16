/*
Copyright 2026 Red Hat, Inc.

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

package tenant

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	tenantv1alpha1 "github.com/redhat-consulting-services/multi-tenant-operator/api/tenant/v1alpha1"
	tenantconfigv1alpha1 "github.com/redhat-consulting-services/multi-tenant-operator/api/tenantconfig/v1alpha1"
	"github.com/redhat-consulting-services/multi-tenant-operator/internal/controller/tenant/namespaced"
)

// MultiTenantConfigReconciler reconciles a MultiTenantConfig object
type MultiTenantConfigReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=tenant.openshift.io,resources=multitenantconfigs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=tenant.openshift.io,resources=multitenantconfigs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=tenant.openshift.io,resources=multitenantconfigs/finalizers,verbs=update

// +kubebuilder:rbac:groups=tenantconfig.openshift.io,resources=namespacelimitranges,verbs=get;list;watch
// +kubebuilder:rbac:groups=tenantconfig.openshift.io,resources=namespaceresourcequotas,verbs=get;list;watch

// +kubebuilder:rbac:groups=core,resources=namespaces,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=limitranges,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=resourcequotas,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=rolebindings,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=networking.k8s.io,resources=networkpolicies,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=argoproj.io,resources=appprojects,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *MultiTenantConfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	mtc := &tenantv1alpha1.MultiTenantConfig{}
	err := r.Get(ctx, client.ObjectKey{Name: req.Name, Namespace: req.Namespace}, mtc)
	if err != nil {
		return ctrl.Result{}, err
	}

	// create or update namespaces based on the MultiTenantConfig spec
	namespaces, err := namespaced.CreateOrUpdateNamespaces(ctx, r.Client, mtc)
	if err != nil {
		log.Error(err, "Failed to create or update namespaces")
		return ctrl.Result{}, err
	}

	// get namespace limit range spec if reference is set in MultiTenantConfig
	nlr := &tenantconfigv1alpha1.NamespaceLimitRange{}
	if mtc.Spec.LimitRangeReference != "" {
		err = r.Get(ctx, client.ObjectKey{Name: mtc.Spec.LimitRangeReference}, nlr)
		if err != nil {
			log.Error(err, "Failed to get NamespaceLimitRange")
			return ctrl.Result{}, err
		}
	}

	// get namespace resource quota spec if reference is set in MultiTenantConfig
	nrr := &tenantconfigv1alpha1.NamespaceResourceQuota{}
	if mtc.Spec.ResourceQuotaReference != "" {
		err = r.Get(ctx, client.ObjectKey{Name: mtc.Spec.ResourceQuotaReference}, nrr)
		if err != nil {
			log.Error(err, "Failed to get NamespaceResourceQuota")
			return ctrl.Result{}, err
		}
	}

	for _, ns := range namespaces {
		mcs := ns.GetMergedConfigSpec(mtc.Spec.ConfigSpec)
		ns.ConfigSpec = &mcs

		clog := log.WithValues("mtcName", mtc.GetName(), "namespace", ns.Name)
		clog.Info("Ensuring all components exist in namespace")

		err = namespaced.CreateOrUpdateConfigMap(ctx, r.Client, mtc, ns.Name)
		if err != nil {
			clog.Error(err, "Failed to create or update ConfigMap in namespace", "namespace", ns.Name)
			return ctrl.Result{}, err
		}

		// create or update limit range in namespace
		err = namespaced.CreateOrUpdateLimitRange(ctx, r.Client, mtc, nlr, ns.Name)
		if err != nil {
			log.Error(err, "Failed to create or update LimitRanges in tenant namespaces")
			return ctrl.Result{}, err
		}

		// create or update resource quota in namespace
		err = namespaced.CreateOrUpdateResourceQuota(ctx, r.Client, mtc, nrr, ns.Name)
		if err != nil {
			log.Error(err, "Failed to create or update ResourceQuotas in tenant namespaces")
			return ctrl.Result{}, err
		}

		// create or update RoleBinding in tenant namespaces based on the MultiTenantConfig spec
		err = namespaced.CreateOrUpdateRoleBinding(ctx, r.Client, mtc, ns.Name)
		if err != nil {
			log.Error(err, "Failed to create or update RoleBindings in tenant namespaces")
			return ctrl.Result{}, err
		}

		// create or update NetworkPolicies in tenant namespaces based on the MultiTenantConfig spec
		err = namespaced.CreateOrUpdateNetworkPolicyTenantInternalAllow(ctx, r.Client, mtc, ns)
		if err != nil {
			log.Error(err, "Failed to create or update tenant-internal NetworkPolicies in tenant namespaces")
			return ctrl.Result{}, err
		}

		err = namespaced.CreateOrUpdateNetworkPolicyAllDeny(ctx, r.Client, mtc, ns)
		if err != nil {
			log.Error(err, "Failed to create or update all-deny NetworkPolicies in tenant namespaces")
			return ctrl.Result{}, err
		}

		err = namespaced.CreateOrUpdateNetworkPolicyNamespaceLocalAllow(ctx, r.Client, mtc, ns)
		if err != nil {
			log.Error(err, "Failed to create or update namespace-local NetworkPolicies in tenant namespaces")
			return ctrl.Result{}, err
		}
	}

	// create or update Argo CD AppProject in the Argo CD instance namespace based on the MultiTenantConfig spec
	err = namespaced.CreateOrUpdateArgoCDProject(ctx, r.Client, mtc, mtc.Spec.GetNamespaceNames())
	if err != nil {
		log.Error(err, "Failed to create or update Argo CD AppProject")
		return ctrl.Result{}, err
	}

	mtc.Status.LimitRangeReference = mtc.Spec.LimitRangeReference
	mtc.Status.QuotaReference = mtc.Spec.ResourceQuotaReference
	mtc.Status.ManagedNamespaceCount = len(namespaces)
	err = r.Client.Status().Update(ctx, mtc)
	if err != nil {
		log.Error(err, "Failed to update MultiTenantConfig status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MultiTenantConfigReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&tenantv1alpha1.MultiTenantConfig{}).
		Named("tenant-multitenantconfig").
		Complete(r)
}
