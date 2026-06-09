# CLAUDE.md — DeployKubeStack Operator

## Project Overview

DeployKubeStack is a Kubernetes operator that provisions a full application stack from a single custom resource. One `kind: DeployKubeStack` manifest creates and manages: Deployment, Service, HPA, Ingress, ConfigMap, Secrets, NetworkPolicy, PodDisruptionBudget, and ServiceMonitor.

- **API Group:** `deploykubestack.io`
- **API Version:** `v1alpha1`
- **Kind:** `DeployKubeStack`
- **Repo language:** Go
- **Framework:** Kubebuilder (controller-runtime)

## Directory Structure

```
deploykubestack/
├── CLAUDE.md
├── Makefile
├── Dockerfile
├── go.mod / go.sum
├── cmd/
│   └── main.go                    # Entrypoint, manager setup
├── api/
│   └── v1alpha1/
│       ├── deploykubestack_types.go   # CRD Go types (spec, status)
│       ├── groupversion_info.go   # GV registration
│       └── zz_generated.deepcopy.go
├── internal/
│   └── controller/
│       ├── deploykubestack_controller.go   # Main reconciler
│       ├── deployment.go               # Deployment reconciliation
│       ├── service.go                  # Service reconciliation
│       ├── hpa.go                      # HPA reconciliation
│       ├── ingress.go                  # Ingress reconciliation
│       ├── configmap.go                # ConfigMap reconciliation
│       ├── secret.go                   # Secret reconciliation
│       ├── pdb.go                      # PDB reconciliation
│       ├── network_policy.go           # NetworkPolicy reconciliation
│       └── service_monitor.go          # ServiceMonitor reconciliation
├── config/
│   ├── crd/                       # Generated CRD manifests
│   ├── manager/                   # Operator deployment manifests
│   ├── rbac/                      # RBAC roles and bindings
│   ├── samples/                   # Example DeployKubeStack CRs
│   └── default/                   # Kustomize base
├── hack/                          # Helper scripts
└── test/
    ├── e2e/                       # End-to-end tests
    └── integration/               # Integration tests
```

## Development Approach

Development is **incremental and resource-by-resource**, not big-bang. We get one
child resource working end-to-end (types → reconciler → RBAC → test → verified on
a cluster) before starting the next. **Deployment is the first target** (the
"walking skeleton"); every later resource repeats the same per-resource recipe.

See **[implementation.md](implementation.md)** for the full plan — it includes a
concepts primer focused on the controller-runtime/Go machinery (for developers
fluent in Kubernetes but new to writing operators in Go), the phase order, the
per-resource recipe, and per-phase "definition of done" checklists.

**Rules of thumb:**
- Do not start a new resource until the current one is unit-tested and manually verified.
- Keep each `build<Resource>` a pure function (CR in → object out) so it's easy to test.
- Two manifests exist: `manifest-minimal.yml` (mandatory fields only) and
  `manifest-full.yml` (every field, for reference/extension).

## Key Commands

```bash
# Generate CRD manifests and deepcopy methods
make generate
make manifests

# Run locally against current kubeconfig cluster
make run

# Run tests
make test                   # Unit tests
make test-e2e               # E2e tests (requires a cluster)

# Build and push operator image
make docker-build IMG=registry.example.com/deploykubestack:latest
make docker-push  IMG=registry.example.com/deploykubestack:latest

# Install CRDs into cluster
make install

# Deploy operator into cluster
make deploy IMG=registry.example.com/deploykubestack:latest

# Uninstall
make undeploy
```

## Architecture & Reconciliation

The operator follows a single-controller pattern. `DeployKubeStackReconciler` is the only controller; it owns all child resources.

### Reconciliation flow

1. Fetch the `DeployKubeStack` CR
2. If `.spec.<resource>.enabled` is true (or the resource has no toggle), build the desired child resource
3. For each child resource: get current state → compare → create/update/delete
4. Set `ownerReferences` on every child so garbage collection is automatic on CR deletion
5. Update `status` subresource with per-resource conditions

### Owner references & GC

Every child resource MUST have `controllerutil.SetControllerReference` called. This ensures `kubectl delete deploykubestack my-app` cleans up everything.

### Status contract

```go
type DeployKubeStackStatus struct {
    Phase      string              `json:"phase"`                // Pending | Provisioning | Ready | Degraded
    Conditions []metav1.Condition  `json:"conditions,omitempty"`
    Resources  ResourceStatus      `json:"resources,omitempty"`
}

type ResourceStatus struct {
    Deployment string `json:"deployment,omitempty"` // Created | Updated | Failed
    Service    string `json:"service,omitempty"`
    HPA        string `json:"hpa,omitempty"`
    Ingress    string `json:"ingress,omitempty"`
    ConfigMap  string `json:"configMap,omitempty"`
}
```

## Coding Conventions

- **One file per child resource.** Each file in `internal/controller/` handles a single K8s resource type and exposes a `Reconcile<Resource>(ctx, ds) error` function.
- **No cross-resource coupling.** The deployment reconciler must not read ingress state or vice versa.
- **Idempotent reconciliation.** Every reconcile loop must be safe to re-run. Use `CreateOrUpdate` with a `MutateFn`.
- **Errors bubble up.** Return errors from child reconcilers; the main controller decides whether to requeue.
- **Use constants.** Label keys, annotation keys, finalizer names go in `api/v1alpha1/constants.go`.
- **Logging.** Use `log.FromContext(ctx)` with structured key-value pairs. No `fmt.Println`.
- **Tests.** Unit-test each child reconciler independently using `fake.NewClientBuilder`. E2e tests use envtest.

## CRD Spec Reference (abbreviated)

```yaml
spec:
  global:
    environment: production | staging | development
    imagePullSecrets: []{name}
    serviceAccountName: string

  deployment:
    replicas: int
    strategy: {type, rollingUpdate}
    container: {image, tag, ports[], resources, env[], volumeMounts[]}
    volumes: []
    healthChecks: {livenessProbe, readinessProbe}
    podDisruptionBudget: {enabled, minAvailable}

  service:
    enabled: bool
    type: ClusterIP | NodePort | LoadBalancer
    ports: []{name, port, targetPort, protocol}

  hpa:
    enabled: bool
    minReplicas: int
    maxReplicas: int
    metrics: []
    behavior: {scaleUp, scaleDown}

  ingress:
    enabled: bool
    className: string
    rules: []{host, paths[]}
    tls: []{secretName, hosts[]}

  configMap:
    enabled: bool
    name: string
    data: map[string]string

  secret:
    enabled: bool
    name: string
    type: Opaque | kubernetes.io/tls
    data: map[string]string

  serviceMonitor:
    enabled: bool
    interval: string
    path: string
    port: string

  networkPolicy:
    enabled: bool
    ingress: []
```

## Common Pitfalls

- **Forgetting `make generate` after changing `_types.go`.** DeepCopy methods won't update and the build will break.
- **Missing RBAC markers.** Every new resource type the controller touches needs `+kubebuilder:rbac` markers on the reconciler.
- **Hardcoded namespaces.** Always use `ds.Namespace`; never default to `"default"`.
- **Drift on updates.** Use `CreateOrUpdate` with a mutate function, not separate Get/Create/Update paths.
- **Large status updates.** Only update status when it actually changes to avoid unnecessary API calls and watch events.

## Environment & Tooling

- Go 1.22+
- Kubebuilder v4+
- controller-runtime v0.18+
- kubectl, kustomize
- A test cluster (kind, minikube, or remote) for `make run` and e2e

## Links

- [Kubebuilder book](https://book.kubebuilder.io)
- [controller-runtime docs](https://pkg.go.dev/sigs.k8s.io/controller-runtime)
- [Operator pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/)