# multi-tenant-operator

[![Docker Repository on Quay](https://quay.io/repository/redhat-consulting-services/multi-tenant-operator/status "Docker Repository on Quay")](https://quay.io/repository/redhat-consulting-services/multi-tenant-operator)

multi-tenant-operator is a Kubernetes operator that automates the provisioning and lifecycle management of tenant namespaces on OpenShift clusters. It enforces consistent resource quotas, limit ranges, RBAC role bindings, network policies, and Argo CD project configurations across all namespaces belonging to a tenant, using a single `MultiTenantConfig` custom resource.

## Description

multi-tenant-operator is built with [operator-sdk](https://sdk.operatorframework.io/) and manages three custom resource kinds:

- **`MultiTenantConfig`** (`tenant.openshift.io/v1alpha1`) — the central configuration object for a tenant. It declares which namespaces belong to the tenant and references a `NamespaceResourceQuota` and/or `NamespaceLimitRange` to be applied to them. It also controls optional per-namespace features such as OVN audit logging, user workload monitoring, OpenShift TLS certificate ConfigMap injection, Argo CD project creation, and RBAC role bindings. Configuration flags can be set globally for all namespaces and overridden on a per-namespace basis.

- **`NamespaceResourceQuota`** (`tenantconfig.openshift.io/v1alpha1`, short name `nrq`) — a cluster-scoped template that wraps a standard Kubernetes `ResourceQuota` spec. The operator applies it as a `ResourceQuota` to every namespace managed by a referencing `MultiTenantConfig`.

- **`NamespaceLimitRange`** (`tenantconfig.openshift.io/v1alpha1`, short name `nlr`) — a cluster-scoped template that wraps a standard Kubernetes `LimitRange` spec. The operator applies it as a `LimitRange` to every namespace managed by a referencing `MultiTenantConfig`.

When a `MultiTenantConfig` is created or updated, the operator reconciles the desired state by creating or updating namespaces, resource quotas, limit ranges, role bindings, config maps (e.g. the OpenShift CA bundle), and Argo CD `AppProject` resources as needed.

## Getting Started

### Prerequisites

- go version v1.24.0+
- docker version 17.03+.
- kubectl version v1.11.3+.
- Access to a Kubernetes v1.11.3+ cluster.

### To Deploy on the cluster

#### Build and Push the Image

> **NOTE**: Do not build and push the image manually. To release a new version, create a git tag and push it to GitHub. The CI pipeline will automatically build the new image and push it to [quay.io](https://quay.io/repository/redhat-consulting-services/multi-tenant-operator).

```sh
git tag v<version>
git push origin v<version>
```

#### Deploy the Operator via OLM (recommended)

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

#### Development environment

**Install the CRDs into the cluster:**

```sh
make install
```

**Deploy the Manager to the cluster with the image specified by `IMG`:**

```sh
make deploy IMG=<some-registry>/multi-tenant-operator:tag
```

> **NOTE**: If you encounter RBAC errors, you may need to grant yourself cluster-admin
privileges or be logged in as admin.

**Create instances of your solution**
You can apply the samples (examples) from the config/sample:

```sh
kubectl apply -k config/samples/
```

>**NOTE**: Ensure that the samples has default values to test it out.

### To Uninstall

**Delete the instances (CRs) from the cluster:**

```sh
kubectl delete -k config/samples/
```

**Delete the APIs(CRDs) from the cluster:**

```sh
make uninstall
```

**UnDeploy the controller from the cluster:**

```sh
make undeploy
```

## Project Distribution

Following the options to release and provide this solution to the users.

### By providing a bundle with all YAML files

1. Build the installer for the image built and published in the registry:

```sh
make build-installer IMG=<some-registry>/multi-tenant-operator:tag
```

**NOTE:** The makefile target mentioned above generates an 'install.yaml'
file in the dist directory. This file contains all the resources built
with Kustomize, which are necessary to install this project without its
dependencies.

2. Using the installer

Users can just run 'kubectl apply -f <URL for YAML BUNDLE>' to install
the project, i.e.:

```sh
kubectl apply -f https://raw.githubusercontent.com/<org>/multi-tenant-operator/<tag or branch>/dist/install.yaml
```

### By providing a Helm Chart

1. Build the chart using the optional helm plugin

```sh
operator-sdk edit --plugins=helm/v1-alpha
```

2. See that a chart was generated under 'dist/chart', and users
can obtain this solution from there.

**NOTE:** If you change the project, you need to update the Helm Chart
using the same command above to sync the latest changes. Furthermore,
if you create webhooks, you need to use the above command with
the '--force' flag and manually ensure that any custom configuration
previously added to 'dist/chart/values.yaml' or 'dist/chart/manager/manager.yaml'
is manually re-applied afterwards.

## Contributing

We welcome contributions! Please open an issue to discuss the change you would like to make before submitting a pull request. Ensure all new code is covered by unit tests and that `make test` passes locally. Follow the existing code style and use `gofmt` / `golint` to format your changes. For larger features, consider opening a GitHub Discussion first to align on the design.

**NOTE:** Run `make help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.
