/*
Copyright 2026.

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
// +kubebuilder:rbac:groups=tenantconfig.openshift.io,resources=namespaceresourcequotalists,verbs=get;list;watch

// +kubebuilder:rbac:groups=core,resources=namespaces,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=limitranges,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=resourcequotas,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=rolebindings,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=argoproj.io,resources=appprojects,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the MultiTenantConfig object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.21.0/pkg/reconcile
func (r *MultiTenantConfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	mtc := &tenantv1alpha1.MultiTenantConfig{}
	err := r.Client.Get(ctx, client.ObjectKey{Name: req.Name, Namespace: req.Namespace}, mtc)
	if err != nil {
		return ctrl.Result{}, err
	}

	// create or update namespaces based on the MultiTenantConfig spec
	namespaces, err := namespaced.CreateOrUpdateNamespaces(ctx, r.Client, mtc)
	if err != nil {
		log.Error(err, "Failed to create or update namespaces")
		return ctrl.Result{}, err
	}

	// create or update ConfigMaps in tenant namespaces based on the MultiTenantConfig spec
	err = namespaced.CreateOrUpdateConfigMaps(ctx, r.Client, mtc, namespaces)
	if err != nil {
		log.Error(err, "Failed to create or update ConfigMaps in tenant namespaces")
		return ctrl.Result{}, err
	}

	if mtc.Spec.LimitRangeReference != "" {
		nlr := &tenantconfigv1alpha1.NamespaceLimitRange{}
		err = r.Client.Get(ctx, client.ObjectKey{Name: mtc.Spec.LimitRangeReference}, nlr)
		if err != nil {
			log.Error(err, "Failed to get NamespaceLimitRange")
			return ctrl.Result{}, err
		}

		// create or update LimitRanges in tenant namespaces based on the MultiTenantConfig spec
		err = namespaced.CreateOrUpdateLimitRanges(ctx, r.Client, mtc, nlr, namespaces)
		if err != nil {
			log.Error(err, "Failed to create or update LimitRanges in tenant namespaces")
			return ctrl.Result{}, err
		}
	}

	if mtc.Spec.QuotaReference != "" {
		nrr := &tenantconfigv1alpha1.NamespaceResourceQuota{}
		err = r.Client.Get(ctx, client.ObjectKey{Name: mtc.Spec.QuotaReference}, nrr)
		if err != nil {
			log.Error(err, "Failed to get NamespaceResourceQuota")
			return ctrl.Result{}, err
		}

		// create or update ResourceQuotas in tenant namespaces based on the MultiTenantConfig spec
		err = namespaced.CreateOrUpdateResourceQuotas(ctx, r.Client, mtc, nrr, namespaces)
		if err != nil {
			log.Error(err, "Failed to create or update ResourceQuotas in tenant namespaces")
			return ctrl.Result{}, err
		}
	}

	// create or update RoleBindings in tenant namespaces based on the MultiTenantConfig spec
	err = namespaced.CreateOrUpdateRoleBindings(ctx, r.Client, mtc, namespaces)
	if err != nil {
		log.Error(err, "Failed to create or update RoleBindings in tenant namespaces")
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
