# multi-tenant-operator

[![Docker Repository on Quay](https://quay.io/repository/redhat-consulting-services/multi-tenant-operator/status "Docker Repository on Quay")](https://quay.io/repository/redhat-consulting-services/multi-tenant-operator)

multi-tenant-operator is a Kubernetes operator that automates the provisioning and lifecycle management of tenant namespaces on OpenShift clusters. It enforces consistent resource quotas, limit ranges, RBAC role bindings, network policies, and ArgoCD project configurations across all namespaces belonging to a tenant, using a single `MultiTenantConfig` custom resource.

## Description

multi-tenant-operator is built with [operator-sdk](https://sdk.operatorframework.io/) and manages three custom resource kinds:

- **`MultiTenantConfig`** (`tenant.openshift.io/v1alpha1`) — the central configuration object for a tenant. It declares which namespaces belong to the tenant and references a `NamespaceResourceQuota` and/or `NamespaceLimitRange` to be applied to them. It also controls optional per-namespace features such as OVN audit logging, user workload monitoring, OpenShift TLS certificate ConfigMap injection, Argo CD project creation, and RBAC role bindings. Configuration flags can be set globally for all namespaces and overridden on a per-namespace basis.

- **`NamespaceResourceQuota`** (`tenantconfig.openshift.io/v1alpha1`, short name `nrq`) — a cluster-scoped template that wraps a standard Kubernetes `ResourceQuota` spec. The operator applies it as a `ResourceQuota` to every namespace managed by a referencing `MultiTenantConfig`.

- **`NamespaceLimitRange`** (`tenantconfig.openshift.io/v1alpha1`, short name `nlr`) — a cluster-scoped template that wraps a standard Kubernetes `LimitRange` spec. The operator applies it as a `LimitRange` to every namespace managed by a referencing `MultiTenantConfig`.

When a `MultiTenantConfig` is created or updated, the operator reconciles the desired state by creating or updating namespaces, resource quotas, limit ranges, role bindings, config maps (e.g. the OpenShift CA bundle), and Argo CD `AppProject` resources as needed.

## Getting Started

### Installation

The operator is distributed through the [rh-consulting-catalog](https://quay.io/repository/redhat-consulting-services/rh-consulting-catalog) operator catalog. To install it on a cluster with the [Operator Lifecycle Manager (OLM)](https://olm.operatorframework.io/) installed, follow the steps below.

**1. Create a `CatalogSource` pointing to the RH Consulting catalog:**

```yaml
apiVersion: operators.coreos.com/v1alpha1
kind: CatalogSource
metadata:
  name: rh-consulting-catalog
  namespace: openshift-marketplace
spec:
  sourceType: grpc
  image: quay.io/redhat-consulting-services/rh-consulting-catalog:latest
  displayName: RH Consulting Catalog
  publisher: Red Hat Consulting
```

```sh
kubectl apply -f catalogsource.yaml
```

**2. Create the target namespace for the operator:**

```sh
kubectl create namespace multi-tenant-operator
```

**3. Create a `Subscription` in the `multi-tenant-operator` namespace:**

```yaml
apiVersion: operators.coreos.com/v1alpha1
kind: Subscription
metadata:
  name: multi-tenant-operator
  namespace: multi-tenant-operator
spec:
  channel: alpha
  name: multi-tenant-operator
  source: rh-consulting-catalog
  sourceNamespace: openshift-marketplace
```

```sh
kubectl apply -f subscription.yaml
```

Once the `Subscription` is created, OLM will automatically install the operator in the `multi-tenant-operator` namespace.

### Usage

nrq-small.yaml:

```yaml
apiVersion: tenantconfig.openshift.io/v1alpha1
kind: NamespaceResourceQuota
metadata:
  name: small
spec:
  hard:
    cpu: "2"
    memory: "8Gi"
    pods: "20"
    configmaps: "10"
    persistentvolumeclaims: "10"
    replicationcontrollers: "5"
    secrets: "20"
    services: "20"
    services.loadbalancers: "0"
```

nlr-small.yaml:

```yaml
apiVersion: tenantconfig.openshift.io/v1alpha1
kind: NamespaceLimitRange
metadata:
  name: small
spec:
  limits:
    - default:
        cpu: 500m
        memory: 512Mi
      defaultRequest:
        cpu: 10m
        memory: 32Mi
      max:
        cpu: "2"
        memory: "4Gi"
      min:
        cpu: 10m
        memory: 32Mi
      type: Container
```

mto.yaml:

```yaml
apiVersion: tenant.openshift.io/v1alpha1
kind: MultiTenantConfig
metadata:
  name: mto-light
  labels:
    access.network.openshift.io/default: ""
spec:
  resourceQuotaReference: small
  limitRangeReference: small
  argocd:
    instanceName: openshift-gitops
    instanceNamespace: openshift-gitops
    project:
      enabled: true
      name: mto-light
      destinations:
        - server: https://kubernetes.default.svc
          namespace: mto-light-front
        - server: https://kubernetes.default.svc
          namespace: mto-light-back
        - server: https://kubernetes.default.svc
          namespace: mto-light-center
        - server: https://kubernetes.default.svc
          namespace: mto-light-shared
      sourceRepos:
        - https://github.com/leonsteinhaeuser/openshift-cluster-config.git
      clusterResourceBlacklist:
        - group: '*'
          kind: '*'
      namespaceResourceBlacklist:
        - group: networking.k8s.io
          kind: NetworkPolicy
  configSpec:
    enableAuditLogging: true
    enableUserWorkloadMonitoring: true
    enableCertificateConfigMapCreation: true
    enableArgoCDControllerManagement: true
    enableNamePrefix: true
    enableNetworkPolicyTenantInternalAllow: false
  namespaces:
    - name: front
    - name: center
    - name: back
    - name: shared
      configSpec:
        enableNetworkPolicyTenantInternalAllow: true
```

Generated namespaces:

```yaml
mto-light-back
mto-light-center
mto-light-front
shared
```

Namespace labels:

```yaml
access.network.openshift.io/default: ""
tenant.openshift.io/multi-tenant-config: mto-light
tenant.openshift.io/name: mto-light
```

## Development

For development instructions, see [docs/dev.md](docs/dev.md).

## Contributing

We welcome contributions! Please open an issue to discuss the change you would like to make before submitting a pull request. Ensure all new code is covered by unit tests and that `make test` passes locally. Follow the existing code style and use `gofmt` / `golint` to format your changes. For larger features, consider opening a GitHub Discussion first to align on the design.

**NOTE:** Run `make help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.
