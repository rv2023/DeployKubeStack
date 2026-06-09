# Implementation Plan — DeployKubeStack Operator

> **Audience:** This plan is written for developers who **know Kubernetes well**
> but are **new to writing operators in Go** (controller-runtime / kubebuilder).
> We assume you're fluent with Deployments, Services, CRDs-as-a-user, RBAC, and
> `kubectl`. What's new is the *Go controller machinery*: the reconcile loop,
> controller-runtime's client, owner references in code, deepcopy/codegen, and
> the kubebuilder project layout. The concepts sections focus there.

---

## 1. Guiding Principles

1. **Vertical slices, one resource at a time.** We do **not** build all nine
   child resources at once ("big-bang"). We get a single resource —
   **Deployment** — working end-to-end (types → reconciler → RBAC → test →
   running on a real cluster), then add the next resource on top of a known-good
   base. This is the opposite of "vibe coding": every phase has a definition of
   done and a test before we move on.

2. **Walking skeleton first.** The first milestone is the smallest thing that
   actually runs: a CR that creates one Deployment and nothing else. Once that
   loop is proven, every new resource is a repeat of the same pattern.

3. **Each resource is an independent slice.** Per the project convention, one
   file per child resource, no cross-resource coupling, idempotent reconcile.
   This means each phase looks almost identical — you learn the pattern once and
   repeat it.

4. **Test before expand.** No new resource is started until the current one is
   unit-tested and manually verified on a cluster.

---

## 2. Concepts You Need First

> You already know what CRDs, CRs, owner references, status subresources, and RBAC
> *are* as a Kubernetes user. This section is about how they show up **in Go code**
> with controller-runtime — the machinery that's actually new.

### 2.1 The controller-runtime machinery (what's new vs. using Kubernetes)

| Term | The Go/controller-runtime angle | Where it lives |
|------|----------------------------------|----------------|
| **Manager** | The process-level object that wires up the client, cache, scheme, and starts all controllers. Set up once in `cmd/main.go`. | `cmd/main.go` |
| **Reconciler** | A struct with a `Reconcile(ctx, req) (ctrl.Result, error)` method. controller-runtime calls it on every relevant change; `req` is just a namespaced name — **you re-fetch the object yourself**, you are not handed the event. | `internal/controller/deploykubestack_controller.go` |
| **`SetupWithManager`** | Where you declare what to watch: `For(&DeployKubeStack{})` plus `Owns(&appsv1.Deployment{})` so changes to a child you own also trigger reconcile. | reconciler file |
| **`client.Client`** | The Go interface for the API server: `Get/List/Create/Update/Patch/Delete`. It's typed and cached. This replaces `kubectl` in code. | passed into every reconciler |
| **Scheme** | A registry mapping Go structs ↔ GroupVersionKind. The client and `SetControllerReference` need it to know how to (de)serialize and link types. | registered in `main.go` |
| **`ctrl.Result`** | Return value controlling requeue: `{}` = done, `{RequeueAfter: d}` = run again later. Returning a non-nil `error` auto-requeues with backoff. | every `Reconcile` |
| **`controllerutil.CreateOrUpdate`** | The idempotent get-or-create-then-mutate helper. Your mutate closure sets the spec **and** calls `SetControllerReference`. This is the required pattern (avoids drift). | every child reconciler |
| **`controllerutil.SetControllerReference`** | The *code* call that stamps the owner ref for GC. Easy to forget — without it, children are orphaned on CR delete. | inside every mutate closure |
| **deepcopy / `zz_generated.deepcopy.go`** | Kubernetes objects must implement `DeepCopyObject()`. `make generate` writes these for you — but only if you re-run it after editing `_types.go`. | generated |
| **kubebuilder markers** | Magic `//+kubebuilder:...` comments that generate the CRD schema (`//+kubebuilder:validation:...`, `//+kubebuilder:default=...`) and RBAC (`//+kubebuilder:rbac:...`). Code comments that drive codegen — unusual if you're new to Go tooling. | `_types.go`, reconciler |
| **envtest** | Spins up a real `kube-apiserver` + `etcd` binary (no nodes/kubelet) so tests exercise the actual API machinery, fast. | `test/` |

**The reconcile loop in one sentence:** *controller-runtime calls our `Reconcile`
with just a name; we re-fetch the CR, compute the child objects it should own, and
`CreateOrUpdate` them to match — then record the result in status.*

### 2.2 Go concepts you'll use constantly

| Concept | One-line explanation | Where it shows up |
|---------|----------------------|-------------------|
| **Struct** | A typed bag of fields. | `DeployKubeStackSpec` is a struct describing the CR. |
| **Struct tags** | The backtick strings like `` `json:"replicas"` ``. They tell the (de)serializer how to map YAML/JSON keys to fields. | Every field in `_types.go`. |
| **Pointer (`*int32`, `*bool`)** | A value that can be `nil` (absent). Used for **optional** fields so we can tell "user set 0" apart from "user set nothing". | `Replicas *int32` — nil means "use default". |
| **Method / receiver** | A function attached to a type: `func (r *DeployKubeStackReconciler) Reconcile(...)`. The `(r *...)` part is the receiver. | The reconciler and all `Reconcile<Resource>` helpers. |
| **Interface** | A contract of methods. `client.Client` is an interface for talking to the cluster (Get/Create/Update/Delete). | Passed into every reconciler; in tests we swap in a fake. |
| **`error` return** | Go has no exceptions; functions return an `error` value you must check (`if err != nil`). | Errors bubble up from child reconcilers to the main loop. |
| **`context.Context`** | Carries cancellation/deadlines and the logger through the call chain. | First argument of nearly every function: `ctx`. |
| **Package** | A folder of Go files sharing a namespace (`api/v1alpha1`, `internal/controller`). | Our code is split into these. |
| **`controllerutil.CreateOrUpdate`** | Helper that does get-or-create-then-mutate in one idempotent call. | The required pattern for every child resource (avoids drift). |

> **Tip:** Don't try to learn all of Go first. Learn these eight concepts, and
> learn the rest by reading the generated scaffolding kubebuilder produces.

---

## 3. Phase Breakdown

Each phase below is a complete vertical slice. The **per-resource recipe**
(Section 4) is identical for every resource — only the desired-object builder
changes.

### Phase 0 — Scaffolding (one-time)

**Goal:** A buildable, runnable, empty operator.

Steps:
1. `kubebuilder init --domain deploykubestack.io --repo github.com/<you>/deploykubestack`
2. `kubebuilder create api --group deploykubestack --version v1alpha1 --kind DeployKubeStack`
   (answer **yes** to both "create resource" and "create controller").
3. Confirm `make build`, `make manifests`, `make generate` all succeed.
4. `make install` then `make run` — the empty controller should start and watch.

**Concepts introduced:** project layout, the Makefile targets, the manager
(`cmd/main.go`), what `make generate`/`make manifests` actually regenerate.

**Definition of done:** `make run` starts cleanly; `kubectl get deploykubestack`
returns "No resources found" (CRD installed).

---

### Phase 1 — Deployment only (the walking skeleton) ⭐ **primary first target**

**Goal:** Applying a minimal CR creates a working Deployment; deleting the CR
deletes the Deployment.

1. **Types** — In `deploykubestack_types.go`, define only what Deployment needs:
   ```go
   type DeployKubeStackSpec struct {
       Deployment DeploymentSpec `json:"deployment"`
   }
   type DeploymentSpec struct {
       Replicas  *int32        `json:"replicas,omitempty"` // optional, default 1
       Container ContainerSpec `json:"container"`          // required
   }
   type ContainerSpec struct {
       Image string         `json:"image"`              // required
       Tag   string         `json:"tag"`                // required
       Ports []ContainerPort `json:"ports,omitempty"`
   }
   ```
   Run `make generate` (deepcopy) and `make manifests` (CRD) — **always after
   editing types.**
2. **Reconciler** — Create `internal/controller/deployment.go` exposing
   `ReconcileDeployment(ctx, ds) error`. Inside: build the desired
   `appsv1.Deployment`, call `controllerutil.CreateOrUpdate`, and set the owner
   reference. Call it from the main `Reconcile`.
3. **RBAC** — Add markers to the reconciler so the operator is allowed to manage
   Deployments:
   ```go
   //+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
   ```
   Re-run `make manifests`.
4. **Ownership / GC** — Verify `SetControllerReference` is called (Section 2.1).
5. **Test** — Unit test with `fake.NewClientBuilder`: given a CR, assert a
   Deployment with the right image/replicas is created. (Section 4.)
6. **Manual verify** — `make run`, apply `manifest-minimal.yml`, then
   `kubectl get deploy` shows the pods; `kubectl delete deploykubestack
   my-application` removes them.

**Definition of done:**
- [ ] `make generate && make manifests && make build` clean
- [ ] Unit test for the Deployment builder passes
- [ ] Apply CR → Deployment appears with correct image/tag/replicas
- [ ] Re-apply same CR → no changes (idempotent)
- [ ] Change `replicas` in CR → Deployment scales (update path works)
- [ ] Delete CR → Deployment garbage-collected (owner ref works)

> **This phase teaches the entire pattern.** Phases 2+ are repetitions.

---

### Phase 2 — Status reporting

**Goal:** Before adding more resources, make the operator *report* what it did.

- Add the `Status` struct from CLAUDE.md (`Phase`, `Conditions`, `Resources`).
- After Deployment reconcile, set `status.resources.deployment = "Created"` and
  `status.phase`.
- **Only write status when it changed** (avoids API spam — a CLAUDE.md pitfall).

**Concepts:** status subresource, `metav1.Condition`, `r.Status().Update()`.
**Definition of done:** `kubectl get deploykubestack -o yaml` shows a populated
`status` block.

---

### Phase 3+ — Add resources one at a time

Each of these is a **repeat of the per-resource recipe** in Section 4. Suggested
order (simplest/most-depended-on first):

| Phase | Resource | Why this order |
|-------|----------|----------------|
| 3 | **Service** | Tiny, almost everyone needs it, exercises the `enabled` toggle. |
| 4 | **ConfigMap** | Introduces config data; Deployment can mount it. |
| 5 | **Secret** | Like ConfigMap but with the `enabled: false` default and types. |
| 6 | **Ingress** | First resource with nested rules/TLS; more complex builder. |
| 7 | **HPA** | Teaches metrics & behavior, and watching a scaling subresource. |
| 8 | **PodDisruptionBudget** | Small, tied to deployment availability. |
| 9 | **NetworkPolicy** | Security; mostly pass-through of user-provided rules. |
| 10 | **ServiceMonitor** | Requires the Prometheus CRD; handle "CRD not installed" gracefully. |

For each: add the spec sub-struct, the `Reconcile<Resource>` file, RBAC markers,
the `enabled` toggle (build when on / delete when off), a unit test, and a sample
in the manifest. **Do not start the next phase until the current one's checklist
is green.**

---

## 4. The Per-Resource Recipe (repeat for every resource)

This is the muscle-memory pattern. Once Phase 1 clicks, every resource is:

1. **Spec** — Add a sub-struct in `_types.go` with `json` tags. Use pointers for
   optional fields. Add an `Enabled bool` if the resource is toggleable.
2. `make generate` (deepcopy) + `make manifests` (CRD + RBAC).
3. **Builder + reconciler** — New file `internal/controller/<resource>.go`:
   ```go
   func ReconcileService(ctx context.Context, c client.Client, ds *v1alpha1.DeployKubeStack) error {
       log := log.FromContext(ctx)
       if !ds.Spec.Service.Enabled {
           // ensure it's deleted, then return (toggle-off path)
       }
       desired := buildService(ds)            // pure function, easy to unit-test
       _, err := controllerutil.CreateOrUpdate(ctx, c, desired, func() error {
           // mutate desired's spec here
           return controllerutil.SetControllerReference(ds, desired, scheme)
       })
       return err
   }
   ```
4. **RBAC markers** on the reconciler for the new group/resource; re-run
   `make manifests`.
5. **Wire it** into the main `Reconcile` (call `ReconcileService(...)`, bubble
   the error).
6. **Unit test** the pure `build<Resource>` function and the toggle behavior with
   `fake.NewClientBuilder`.
7. **Sample** — add the block to a sample manifest; manually verify on cluster.
8. **Status** — record `Created/Updated/Failed` for the resource.
9. Tick the resource's **definition of done** checklist.

> Keep `build<Resource>` a **pure function** (CR in → object out, no cluster
> calls). Pure functions are trivial to unit-test and contain no I/O.

---

## 5. Testing Strategy

- **Unit tests (per resource, fast):** `sigs.k8s.io/controller-runtime/pkg/client/fake`.
  Test the builder output and the create/update/delete-on-toggle behavior. No
  cluster needed. This is the primary safety net.
- **Integration tests (envtest):** Spin up a fake API server, run the real
  reconciler, assert children appear. Catches RBAC/scheme wiring bugs.
- **E2e (real/kind cluster):** `make test-e2e`. Apply a full CR, assert the whole
  stack provisions and GC works on delete. Run before releases, not on every
  change.

**Rule:** every phase adds at least one unit test before it's "done".

---

## 6. Milestones

| Milestone | Contents | Outcome |
|-----------|----------|---------|
| **M0** | Phase 0 | Operator scaffold runs against a cluster. |
| **M1** | Phases 1–2 | CR → Deployment + status. The pattern is proven. |
| **M2** | Phases 3–5 | Service, ConfigMap, Secret. Core app stack. |
| **M3** | Phases 6–7 | Ingress + HPA. Externally reachable + autoscaling. |
| **M4** | Phases 8–10 | PDB, NetworkPolicy, ServiceMonitor. Production hardening. |

---

## 7. Common Pitfalls (carried from CLAUDE.md)

- Forgetting `make generate` after editing `_types.go` → build breaks.
- Missing `+kubebuilder:rbac` markers → operator gets "forbidden" at runtime.
- Hardcoding `"default"` namespace → always use `ds.Namespace`.
- Separate Get/Create/Update instead of `CreateOrUpdate` → drift bugs.
- Writing status every reconcile → API spam; only write on change.
- Forgetting `SetControllerReference` → orphaned resources on CR delete.

---

## 8. What "Done" Looks Like for the Whole Project

A user applies one `DeployKubeStack` CR and gets a full, owned, garbage-collected
application stack; `kubectl get deploykubestack` reports accurate health; every
resource is unit-tested; and the operator survives re-applies (idempotent) and
deletes (GC). We reach it **one resource at a time**, never all at once.
