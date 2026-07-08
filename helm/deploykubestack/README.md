# DeployKubeStack Helm Chart

A Helm chart for deploying the DeployKubeStack Kubernetes operator, including its Custom Resource Definition (CRD).

## Prerequisites

- Kubernetes 1.20+
- Helm 3.0+

## Installation

### Add the Repository (when available)

```bash
# Add the Helm repository
helm repo add deploykubestack https://charts.deploykubestack.io
helm repo update
```

### Install from Local Chart

```bash
# Install with default values
helm install deploykubestack ./deploykubestack \
  --namespace deploykubestack-system \
  --create-namespace

# Install with custom values
helm install deploykubestack ./deploykubestack \
  --namespace deploykubestack-system \
  --create-namespace \
  --values custom-values.yaml

# Install from specific tag
helm install deploykubestack deploykubestack/deploykubestack \
  --version 1.0.0 \
  --namespace deploykubestack-system \
  --create-namespace
```

## Configuration

### Common Values

| Parameter | Description | Default |
|-----------|-------------|---------|
| `replicaCount` | Number of operator replicas | `1` |
| `image.registry` | Container registry | `ghcr.io` |
| `image.repository` | Image repository | `rv2023/DeployKubeStack` |
| `image.tag` | Image tag (auto-detects Chart.appVersion if empty) | `""` |
| `image.pullPolicy` | Image pull policy | `IfNotPresent` |
| `namespace` | Namespace to deploy operator | `deploykubestack-system` |
| `resources.requests.cpu` | CPU request | `100m` |
| `resources.requests.memory` | Memory request | `64Mi` |
| `resources.limits.cpu` | CPU limit | `500m` |
| `resources.limits.memory` | Memory limit | `128Mi` |

### CRD Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `crds.install` | Install the CRD | `true` |
| `crds.keep` | Keep CRD on chart deletion | `true` |

### Logging Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `logging.level` | Log level (debug, info, warn, error) | `info` |
| `logging.development` | Enable development logging | `false` |

### Feature Flags

| Parameter | Description | Default |
|-----------|-------------|---------|
| `metricsServer.enabled` | Enable metrics server | `true` |
| `metricsServer.port` | Metrics server port | `8080` |
| `healthProbe.enabled` | Enable health probes | `true` |
| `webhook.enabled` | Enable webhooks | `false` |

## Examples

### Minimal Installation

```bash
helm install deploykubestack ./deploykubestack \
  --namespace deploykubestack-system \
  --create-namespace
```

### Production Installation with Custom Resources

```yaml
# values-prod.yaml
replicaCount: 2
logging:
  level: warn
resources:
  requests:
    cpu: 200m
    memory: 128Mi
  limits:
    cpu: 1000m
    memory: 512Mi
nodeSelector:
  kubernetes.io/os: linux
```

```bash
helm install deploykubestack ./deploykubestack \
  --namespace deploykubestack-system \
  --create-namespace \
  --values values-prod.yaml
```

### Private Registry

```bash
helm install deploykubestack ./deploykubestack \
  --namespace deploykubestack-system \
  --create-namespace \
  --set image.registry=private.registry.com \
  --set imagePullSecrets[0].name=private-registry-secret
```

## Upgrading

```bash
# Update Helm repository
helm repo update

# Upgrade installation
helm upgrade deploykubestack deploykubestack/deploykubestack \
  --namespace deploykubestack-system

# Upgrade with custom values
helm upgrade deploykubestack deploykubestack/deploykubestack \
  --namespace deploykubestack-system \
  --values custom-values.yaml
```

## Uninstalling

```bash
# Uninstall (CRD is kept by default, set crds.keep=false to remove)
helm uninstall deploykubestack \
  --namespace deploykubestack-system

# Uninstall and remove CRD
helm uninstall deploykubestack \
  --namespace deploykubestack-system \
  --set crds.keep=false
```

## Verifying Installation

```bash
# Check deployment status
kubectl get deployment -n deploykubestack-system
kubectl logs deployment/deploykubestack-controller-manager -n deploykubestack-system

# Verify CRD is installed
kubectl get crd | grep deploykubestack

# Check RBAC
kubectl get clusterrole | grep deploykubestack
kubectl get clusterrolebinding | grep deploykubestack

# Check metrics service
kubectl get service -n deploykubestack-system
```

## Creating Applications

Once the operator is deployed, you can create Applications:

```yaml
apiVersion: apps.deploykubestack.com/v1alpha1
kind: Application
metadata:
  name: my-app
  namespace: default
spec:
  image: nginx:latest
  port: 80
  replicas: 2
  resources:
    requests:
      cpu: 200m
      memory: 256Mi
    limits:
      cpu: 500m
      memory: 512Mi
```

```bash
# Apply the Application
kubectl apply -f application.yaml

# Check Application status
kubectl get applications
kubectl describe application my-app
```

## Chart Structure

```
deploykubestack/
├── Chart.yaml                    # Chart metadata
├── values.yaml                   # Default values
├── README.md                     # This file
└── templates/
    ├── _helpers.tpl              # Template helpers
    ├── crd-application.yaml       # Application CRD
    ├── serviceaccount.yaml        # Service account
    ├── clusterrole.yaml           # Cluster role (RBAC)
    ├── clusterrolebinding.yaml    # Cluster role binding (RBAC)
    ├── deployment.yaml            # Controller deployment
    └── service-metrics.yaml       # Metrics service
```

## RBAC

The chart creates:
- **ServiceAccount**: `deploykubestack-controller-manager`
- **ClusterRole**: `deploykubestack-controller-manager-role` with permissions for:
  - Applications CRD (apps.deploykubestack.com)
  - Deployments (apps)
  - Services (core)
  - Leader election (if enabled)
  - Metrics (if enabled)
- **ClusterRoleBinding**: Binds role to service account

## Metrics

If metrics are enabled (default), the operator exposes Prometheus metrics on port 8080:

```bash
# Port-forward to access metrics
kubectl port-forward -n deploykubestack-system \
  svc/deploykubestack-controller-manager-metrics-service 8080:8080

# Scrape metrics
curl http://localhost:8080/metrics
```

## Troubleshooting

### Operator pod not running

```bash
# Check pod status
kubectl get pods -n deploykubestack-system

# View logs
kubectl logs deployment/deploykubestack-controller-manager \
  -n deploykubestack-system

# Describe pod for events
kubectl describe pod <pod-name> -n deploykubestack-system
```

### CRD not available

```bash
# Verify CRD is installed
kubectl get crd applications.apps.deploykubestack.com

# If missing, check helm values
helm get values deploykubestack -n deploykubestack-system | grep crds
```

### RBAC errors

```bash
# Check RBAC bindings
kubectl get clusterrole deploykubestack-controller-manager-role
kubectl get clusterrolebinding deploykubestack-controller-manager-rolebinding
```

## Contributing

For issues or improvements to this chart, please contribute to the main repository:
https://github.com/rv2023/DeployKubeStack

## License

Apache License 2.0

---

**For more information:**
- [DeployKubeStack GitHub](https://github.com/rv2023/DeployKubeStack)
- [Helm Chart Best Practices](https://helm.sh/docs/chart_best_practices/)
- [Kubernetes Documentation](https://kubernetes.io/docs/)
