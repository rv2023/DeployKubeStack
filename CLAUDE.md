# CLAUDE.md — DeployKubeStack Operator

## Project Overview

DeployKubeStack is a Kubernetes operator that provisions a full application stack from a single custom resource. One `kind: Application` manifest creates and manages: Deployment, Service, HPA, Ingress, ConfigMap, Secrets, NetworkPolicy, PodDisruptionBudget, and ServiceMonitor.

- **API Group:** `apps.deploykubestack.com`
- **API Version:** `v1alpha1`
- **Kind:** `Application`
- **Domain:** `deploykubestack.com`
- **Repo language:** Go 1.24.0+
- **Framework:** Kubebuilder v4.6.0 (controller-runtime v0.21.0)

## Current Status

**Phase 1: Deployment & Service Management** — ✅ **COMPLETE & TESTED**

- ✅ Application CRD with image, port, replicas, resources fields
- ✅ Deployment reconciliation with intelligent defaults
- ✅ Service reconciliation (ClusterIP)
- ✅ Automatic garbage collection via owner references
- ✅ Status tracking (phase, readiness flags)
- ✅ Comprehensive unit tests (82.2% coverage, 3/3 passing)
- ✅ Auto-generated RBAC and CRD schema
- ✅ Full documentation (see `docs/` directory)

**Usage:**
```yaml
apiVersion: apps.deploykubestack.com/v1alpha1
kind: Application
metadata:
  name: my-app
spec:
  image: nginx:latest
  port: 8080
  replicas: 2  # optional
  # resources: optional (defaults: 100m CPU, 100Mi memory)
```

This creates Deployment + Service with defaults applied if resources not specified.

## Directory Structure

```
deploykubestack/
├── CLAUDE.md                      # This file
├── implementation.md              # Detailed phase-by-phase implementation plan
├── README.md                      # User-facing documentation (has TODOs)
├── PROJECT                        # Kubebuilder project config (auto-generated)
├── Makefile                       # Build, test, deploy targets
├── Dockerfile                     # Operator image
├── .golangci.yml                  # Linter config
├── .devcontainer/                 # Dev container setup
├── go.mod / go.sum               # Go dependencies
├── cmd/
│   └── main.go                    # Entrypoint: manager setup, webhook config, metrics
├── api/
│   └── v1alpha1/
│       ├── application_types.go        # CRD Go types (ApplicationSpec/ApplicationStatus)
│       ├── groupversion_info.go        # GV registration (apps.deploykubestack.com/v1alpha1)
│       └── zz_generated.deepcopy.go    # Generated: DeepCopy/DeepCopyInto/DeepCopyObject
├── internal/
│   └── controller/
│       ├── application_controller.go        # Main reconciler (stub, no resources yet)
│       ├── application_controller_test.go   # Unit test scaffold
│       └── suite_test.go                    # Ginkgo test suite setup
├── config/
│   ├── crd/bases/                 # Generated CRD manifests (kustomize base)
│   ├── manager/                   # Operator Deployment & RBAC Service Account
│   ├── rbac/                      # RBAC roles, bindings, metrics auth
│   ├── samples/                   # Example Application CRs (kustomize samples)
│   ├── network-policy/            # Network policies for operator
│   └── default/                   # Kustomize base for all-in-one install
├── hack/                          # Boilerplate headers, helper scripts
├── .github/workflows/
│   ├── test.yml                   # Unit test CI
│   ├── lint.yml                   # Linter CI (.golangci.yml)
│   └── test-e2e.yml               # E2E test CI (requires cluster)
└── test/
    ├── e2e/
    │   ├── e2e_test.go            # E2E tests (Ginkgo, uses actual cluster)
    │   └── e2e_suite_test.go      # E2E test suite setup
    └── utils/
        └── utils.go               # Test helpers
```

## Development Approach

**Status: Phase 1 COMPLETE, Phase 2+ planned.**

The repo uses a **resource-by-resource, vertical-slice pattern**. Phase 1 (Deployment + Service) is complete and tested. Future resources follow the same proven recipe.

### Phase 1 (✅ COMPLETE): Deployment & Service

**What's implemented:**
1. ✅ `ApplicationSpec` with image, port, replicas, resources fields
2. ✅ `ApplicationStatus` with phase, readiness flags, error message
3. ✅ `ReconcileDeployment()` - Creates/updates Deployments with defaults
4. ✅ `ReconcileService()` - Creates/updates Services
5. ✅ Main `Reconcile()` - Orchestrates child reconcilers
6. ✅ Unit tests (3 test cases, 82.2% coverage)
7. ✅ RBAC auto-generated from markers
8. ✅ CRD validation schema

**Default Resources (when not specified):**
- CPU: 100m request, 100m limit
- Memory: 100Mi request, 100Mi limit
- Replicas: 1

### Future Phases: Same Recipe

Each new phase (HPA, Ingress, ConfigMap, etc.) follows the proven pattern:
1. Add spec fields to `ApplicationSpec`
2. Create `internal/controller/<resource>_reconcile.go`
3. Add RBAC markers to main reconciler
4. Call from main `Reconcile()` function
5. Add unit tests
6. Update status fields

See **[docs/design/implementation.md](docs/design/implementation.md)** for:
- Full phase roadmap
- Concepts primer on controller-runtime/Go
- Per-resource recipe with checklists

**Key patterns used:**
- All reconciliation in `Reconcile()` — fetch CR, compute desired state, `CreateOrUpdate` to match
- One file per child resource: `<resource>_reconcile.go` with pure builder function
- `CreateOrUpdate` + `SetControllerReference` for idempotency and garbage collection
- Pure builder functions: `buildDeployment(app) → Deployment` (testable, no I/O)

## Key Commands

```bash
# Codegen (run after editing api/v1alpha1/application_types.go or controller)
make generate              # DeepCopy, DeepCopyInto, DeepCopyObject (zz_generated.deepcopy.go)
make manifests             # CRD manifests, RBAC from markers

# Verify (always run before committing)
make build                 # Build binary locally
make lint                  # Run golangci-lint (.golangci.yml)

# Testing
make test                  # Unit tests (Ginkgo, via envtest, no cluster required)
make test-e2e              # E2E tests (Ginkgo, requires active cluster in kubeconfig)

# Local development (requires active kubeconfig cluster, e.g., kind/minikube)
make install               # Apply CRDs to cluster
make run                   # Start operator locally (watches cluster, exits on C-c)
# In another terminal: kubectl apply -f config/samples/apps_v1alpha1_application.yaml

# Container image
make docker-build IMG=<registry>/deploykubestack:latest
make docker-push  IMG=<registry>/deploykubestack:latest

# Deploy to cluster
make deploy IMG=<registry>/deploykubestack:latest  # Install CRDs + deploy operator

# Cleanup
make undeploy              # Remove deployed operator
make uninstall             # Remove CRDs from cluster
```

## Architecture & Reconciliation

The operator follows a **single-controller pattern**: `ApplicationReconciler` is the only controller. It is responsible for all nine child resource types (Deployment, Service, HPA, Ingress, ConfigMap, Secret, NetworkPolicy, PodDisruptionBudget, ServiceMonitor).

### Reconciliation flow (high-level)

1. controller-runtime calls `Reconcile(ctx, req)` with just the CR name
2. Fetch the `Application` CR from the cluster
3. For each child resource type (when enabled):
   - Call the resource-specific reconciler: `ReconcileDeployment(ctx, app)`, `ReconcileService(ctx, app)`, etc.
   - Each returns an error; errors bubble up and trigger automatic requeue
4. Update `status.phase` and resource-specific status fields
5. Return success or error; requeue on error

### The mutate closure pattern (core pattern)

Every child resource uses `controllerutil.CreateOrUpdate` with a mutate closure:

```go
deployment := &appsv1.Deployment{}
op, err := controllerutil.CreateOrUpdate(ctx, r.Client, deployment, func() error {
    // Mutate closure: set desired state (spec, labels, etc.)
    deployment.Spec = /* builder output */
    deployment.Labels["app"] = app.Name
    
    // CRITICAL: set owner reference for garbage collection
    return controllerutil.SetControllerReference(app, deployment, r.Scheme)
})
// err == nil means created or updated; op shows which
if err != nil { return err }
```

### Owner references & garbage collection

Every child resource **must** call `SetControllerReference`. This sets the `ownerReferences` field, ensuring:
- When the `Application` CR is deleted, Kubernetes automatically deletes all children
- `kubectl delete application my-app` cleans up everything

### Status contract (current/evolving)

Will be defined as you build each resource. Start simple:

```go
type ApplicationStatus struct {
    Phase   string `json:"phase,omitempty"`   // Pending | Provisioning | Ready | Degraded
    DeploymentReady   bool   `json:"deploymentReady,omitempty"`
    ServiceReady      bool   `json:"serviceReady,omitempty"`
    // Add a field per resource as you build it
}
```

Add condition reporting later (via `meta.SetStatusCondition`) when ready.

## Coding Conventions

### File organization
- **One file per child resource:** `internal/controller/<resource>_reconcile.go` (e.g., `deployment_reconcile.go`, `service_reconcile.go`)
- Each file exports a single function: `Reconcile<Resource>(ctx, application *appsv1alpha1.Application, r *ApplicationReconciler) error`
- Put the desired-object builder in the same file: `build<Resource>(application *appsv1alpha1.Application) *appsv1.Deployment`
- **No cross-resource coupling.** Deployment reconciler does not peek at Ingress state. Each is independent.

### Reconciliation patterns
- **Idempotent reconciliation.** Use `CreateOrUpdate` + `SetControllerReference` so re-running is safe.
- **Builder as pure function.** `buildDeployment(app) → Deployment` — takes Application, returns desired Deployment, no I/O.
- **Errors bubble up.** Every reconciler returns an error. `Reconcile()` calls each and returns the first error to trigger requeue.
- **No manual retries.** Return the error; controller-runtime requeues with exponential backoff.

### Constants and helpers
- Label keys, annotation keys, finalizer names: put in a `constants.go` file (TBD)
- Shared helpers (e.g., `getOwnerName()`, `makeSelector()`) in `internal/controller/helpers.go` (TBD)

### Logging
- Use `logr.FromContext(ctx)` (injected by `logf` in Ginkgo tests)
- Structured key-value: `log.Info("created deployment", "name", dep.Name, "namespace", dep.Namespace)`
- No `fmt.Println`, no global log variables

### Testing
- **Unit tests:** `internal/controller/application_controller_test.go`. Test reconcilers and builders independently with `fake.NewClientBuilder()` (no cluster needed).
- **E2E tests:** `test/e2e/e2e_test.go`. Run against an actual cluster (via envtest or real cluster). Create a CR, verify child resources appear.
- Builder tests are pure logic: no client, no clusters — just `buildDeployment(app) == expected`.

### Kubebuilder markers
- **RBAC markers** on the reconciler: `//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=create;get;update;patch`
- **Validation markers** on types: `//+kubebuilder:validation:Required`, `//+kubebuilder:validation:MinItems=1`, etc.
- Always run `make manifests` after changing markers — they drive CRD and RBAC generation.

## Application CRD Spec (to be built)

**Current state:** `ApplicationSpec` is empty (scaffold). Fields are added during development, one resource at a time.

**Target spec (architecture overview):**

```yaml
apiVersion: apps.deploykubestack.com/v1alpha1
kind: Application
metadata:
  name: my-app
  namespace: default
spec:
  global:
    environment: production | staging | development
    imagePullSecrets: [{name: string}]
    serviceAccountName: string

  deployment:
    replicas: int (default: 1)
    strategy: {type: RollingUpdate|Recreate, rollingUpdate: {...}}
    container:
      image: string
      tag: string
      ports: [{name, port, protocol}]
      resources: {requests: {...}, limits: {...}}
      env: [{name, value|valueFrom}]
      volumeMounts: [{name, mountPath, readOnly}]
    volumes: [...]
    healthChecks: {livenessProbe: {...}, readinessProbe: {...}}
    podDisruptionBudget: {enabled: bool, minAvailable: int}

  service:
    enabled: bool (default: true)
    type: ClusterIP | NodePort | LoadBalancer
    ports: [{name, port, targetPort, protocol}]

  hpa:
    enabled: bool (default: false)
    minReplicas: int
    maxReplicas: int
    metrics: [{type, resource|pods|...}]
    behavior: {scaleUp: {...}, scaleDown: {...}}

  ingress:
    enabled: bool (default: false)
    className: string
    rules: [{host, paths: [{path, pathType, backend}]}]
    tls: [{secretName, hosts}]

  configMap:
    enabled: bool (default: false)
    name: string
    data: {key: value}

  secret:
    enabled: bool (default: false)
    name: string
    type: Opaque | kubernetes.io/tls
    data: {key: base64-value}

  serviceMonitor:
    enabled: bool (default: false)
    interval: string (e.g., "30s")
    path: string (e.g., "/metrics")
    port: string

  networkPolicy:
    enabled: bool (default: false)
    ingress: [{from: [{podSelector}], ports: [{protocol, port}]}]

status:
  phase: Pending | Provisioning | Ready | Degraded
  deploymentReady: bool
  serviceReady: bool
  # Add per resource as implemented
```

**You define the spec fields as you implement each resource.** Start with just Deployment (small spec), then expand.

## Common Pitfalls

- **Forgetting `make generate` after editing `application_types.go`.** DeepCopy methods (`zz_generated.deepcopy.go`) won't regenerate, and the build will fail with "DeepCopyObject not found". Always run `make generate && make manifests` after any type change.

- **Missing RBAC markers.** Every Kubernetes resource the controller creates/updates/patches needs a `+kubebuilder:rbac:...` marker on the reconciler. Without them, the operator has no permission to create Deployments, Services, etc. Add markers first, then `make manifests`.

- **Forgetting `SetControllerReference`.** Without it in the mutate closure, children aren't owned by the Application CR, so they don't get garbage-collected when the CR is deleted. Always call it inside the `CreateOrUpdate` mutate closure.

- **Hardcoded namespaces.** Always use `application.Namespace` when creating child objects — never default to `"default"`. Child objects must be in the same namespace as the parent CR.

- **Drift on updates.** Never do separate `Get` → `Create` → `Update` paths. Always use `CreateOrUpdate` with a deterministic mutate function — this ensures idempotence and avoids drift if the user manually edits a child object.

- **Mutations outside the mutate closure.** The mutate function is the *only* place to set desired state. If you mutate the object *before* calling `CreateOrUpdate`, your changes might be lost if the object already exists.

- **No logging in builders.** Builder functions are pure — they don't have access to a logger or context. Log in the reconciler after the builder returns, not inside it.

- **Testing against the real cluster in unit tests.** Use `fake.NewClientBuilder()` with fake objects in unit tests — this is fast and deterministic. Save real-cluster testing for E2E tests.

- **Large status updates.** Only update `status` when it actually changes. Unnecessary status updates trigger reconciliation loops in other controllers watching this CR.

## Environment & Tooling

**Required:**
- Go 1.24.0+ (see `go.mod`)
- Kubebuilder v4.6.0+ (used to scaffold and generate code)
- controller-runtime v0.21.0+ (see `go.mod`)
- Docker or Podman (for `make docker-build`)
- kubectl (to interact with cluster)
- kustomize (for manifest generation; comes with kubebuilder)

**For development:**
- A test cluster reachable via current kubeconfig (kind, minikube, remote) for `make run` and E2E tests
- `golangci-lint` (for `make lint`; runs in CI)
- Ginkgo (test framework; already in `go.mod`)

**To set up a local test cluster:**
```bash
# Using kind (recommended for testing operators)
kind create cluster --name deploykubestack

# Or minikube
minikube start

# Verify kubeconfig
kubectl cluster-info
```

## Getting Started

**For a 5-minute quickstart:** See [docs/quick-start/QUICK_START.md](docs/quick-start/QUICK_START.md)

### Build & Test

```bash
# Install dependencies
go mod download

# Build, lint, test
make build       # Compile operator
make lint        # Run linter
make test        # Run unit tests (should see 3 PASSED)
```

### Run Locally

```bash
# Set up test cluster (one-time)
kind create cluster --name deploykubestack
kubectl cluster-info

# Install CRD and run operator
make install
make run    # Leave running, opens new terminal for next steps

# Create an Application in another terminal
kubectl apply -f config/samples/apps_v1alpha1_application.yaml

# Watch it work
kubectl get applications
kubectl describe application application-sample
kubectl get deployment,service -l app=application-sample
```

### Documentation by Use Case

| Goal | Read |
|------|------|
| Get running fast | [Quick Start](docs/quick-start/QUICK_START.md) |
| Understand architecture | [Implementation Summary](docs/guides/IMPLEMENTATION_SUMMARY.md) |
| Deploy to cluster | [Deployment Guide](docs/guides/DEPLOYMENT_GUIDE.md) |
| See test results | [Completion Report](docs/reports/COMPLETION_REPORT.md) |
| Check Phase 1 features | [Phase 1 Complete](docs/guides/PHASE1_COMPLETE.md) |
| Plan Phase 2+ | [Implementation Plan](docs/design/implementation.md) |

**All documentation:** See [docs/README.md](docs/README.md)

## Project Documentation

**Start here:** [docs/README.md](docs/README.md) — Index of all documentation

### Key Guides
- [Quick Start](docs/quick-start/QUICK_START.md) — Get running in 5 minutes
- [Phase 1 Complete](docs/guides/PHASE1_COMPLETE.md) — What's implemented and how to use it
- [Implementation Summary](docs/guides/IMPLEMENTATION_SUMMARY.md) — Deep technical details and architecture
- [Deployment Guide](docs/guides/DEPLOYMENT_GUIDE.md) — Production deployment, troubleshooting, monitoring
- [Completion Report](docs/reports/COMPLETION_REPORT.md) — Phase 1 metrics, tests, quality assurance
- [Implementation Plan](docs/design/implementation.md) — Original phase-by-phase plan and roadmap

## External References

- [Kubebuilder book](https://book.kubebuilder.io) — definitive reference for markers, patterns, and troubleshooting
- [controller-runtime docs](https://pkg.go.dev/sigs.k8s.io/controller-runtime) — Go API reference
- [Operator pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/) — conceptual overview
- [Ginkgo testing framework](https://onsi.github.io/ginkgo/) — used for unit and E2E tests in this repo
- [envtest reference](https://book.kubebuilder.io/reference/envtest.html) — testing against a local kube-apiserver