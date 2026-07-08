# Release Workflow Setup - Complete

## What's Been Set Up

### ✅ GitHub Actions Workflow Created

**File:** `.github/workflows/release.yml`

**Triggers:** Tag push (v* or release-*)  
**Actions:**
1. Checks out code
2. Sets up Docker Buildx for optimized builds
3. Authenticates with GitHub Container Registry
4. Builds Docker image (multi-stage, optimized)
5. Pushes image to GHCR
6. Creates GitHub Release with deployment instructions

### ✅ Dockerfile Ready

**File:** `Dockerfile`

**Features:**
- Multi-stage build (golang builder → distroless runtime)
- Minimal image size (distroless/static:nonroot)
- Non-root user (uid 65532) for security
- Cross-platform build support (linux/amd64, linux/arm64)
- Cached layers for faster rebuilds

### ✅ Documentation Complete

**File:** `docs/guides/RELEASE_PROCESS.md`

**Includes:**
- How to create releases
- Image naming conventions
- Usage examples (Docker, Kubernetes)
- Troubleshooting guide
- Release checklist
- CI/CD pipeline diagram

## Quick Start: Create Your First Release

### 1. Verify Everything Works

```bash
make test       # All tests passing
make lint       # No linting errors
make build      # Build succeeds
```

### 2. Create a Tag

```bash
git tag v1.0.0 -m "Release v1.0.0: Phase 1 Complete"
git push origin v1.0.0
```

### 3. GitHub Actions Automatically:
- ✅ Builds Docker image
- ✅ Pushes to ghcr.io/rv2023/DeployKubeStack:v1.0.0
- ✅ Creates GitHub Release
- ✅ Generates release notes

### 4. Use the Released Image

```bash
# Deploy to any Kubernetes cluster
make deploy IMG=ghcr.io/rv2023/DeployKubeStack:v1.0.0
```

## Workflow Features

### Automatic Tag Detection

The workflow triggers on:
- `v1.0.0` - Semantic versions
- `v1.0.0-rc1` - Pre-releases
- `release-2026-07-07` - Custom formats

Any other tag format: no action (workflow skips)

### Smart Image Tagging

The workflow automatically creates image tags based on git tags:
- `ghcr.io/rv2023/DeployKubeStack:v1.0.0` - From tag `v1.0.0`
- `ghcr.io/rv2023/DeployKubeStack:1.0` - Major.minor
- `ghcr.io/rv2023/DeployKubeStack:1` - Major version

### Cache Optimization

Uses GitHub Actions cache for BuildKit:
- Faster rebuilds
- Reduced build time
- Automatic cache cleanup

### Security

- Uses `GITHUB_TOKEN` (no manual secret needed)
- Non-root container user (uid 65532)
- Distroless base image (minimal attack surface)
- Private container registry (only accessible with auth)

## Image Accessibility

### Within GitHub Organization

```bash
# Pull with GitHub authentication
docker pull ghcr.io/rv2023/DeployKubeStack:v1.0.0
```

### Making Images Public (Optional)

To allow public pulls without authentication:

1. Go to GitHub repository settings
2. Navigate to: Settings → Packages → Container registry
3. Change visibility to "Public"
4. Anyone can pull without authentication

```bash
# Public pull (if configured)
docker pull ghcr.io/rv2023/DeployKubeStack:v1.0.0
```

## Workflow Permissions

**GitHub Actions Permissions (Principle of Least Privilege):**

The workflow explicitly declares only the permissions it needs:

| Permission | Purpose | Level |
|-----------|---------|-------|
| `contents: read` | Read repository code & tags | Workflow + Job |
| `packages: write` | Write Docker images to GHCR | Workflow + Job |
| `id-token: write` | OIDC token for signing (future) | Workflow + Job |

**Authentication Approach:**

- **Uses:** `${{ github.token }}` (automatic GitHub Actions token)
- **Not:** `${{ secrets.GITHUB_TOKEN }}` (implicit secret)
- **Why:** `github.token` is the modern, explicit approach recommended by GitHub
- **Scoped to:** Current workflow run only
- **Expires:** Automatically after workflow completes

**Security Best Practices Applied:**

✅ Explicit permissions declaration (principle of least privilege)  
✅ Job-level permissions override (extra safety layer)  
✅ Auto-scoped token (doesn't require manual secret management)  
✅ OIDC token support ready (for future artifact signing)  
✅ Environment protection available (can add approval requirement)

No additional PAT (Personal Access Token) or secrets configuration needed.

## Next Steps

### 1. Create First Release (Optional - for testing)

```bash
# Create a test tag
git tag v0.1.0-test -m "Test release"
git push origin v0.1.0-test

# Monitor workflow: https://github.com/rv2023/DeployKubeStack/actions
# Should complete in 2-5 minutes
```

### 2. Create Production Release

```bash
git tag v1.0.0 -m "Release v1.0.0: Phase 1 Complete"
git push origin v1.0.0

# Verify at: https://github.com/rv2023/DeployKubeStack/releases
```

### 3. Deploy Released Image

```bash
# From release notes or manually
make deploy IMG=ghcr.io/rv2023/DeployKubeStack:v1.0.0

# Verify deployment
kubectl get deployment -n deploykubestack-system
```

## Monitoring & Troubleshooting

### View Workflow Runs

```bash
# List all workflow runs
gh run list --workflow=release.yml

# Watch a specific run
gh run watch <run-id>

# View logs
gh run view <run-id> --log
```

### Check Image in Registry

```bash
# List all images in GHCR
gh api orgs/rv2023/packages -t '{{ range .[] }}{{ .name }}: {{ .type }}{{ end }}'

# Or visit: https://github.com/rv2023/DeployKubeStack/pkgs/container/deploykubestack
```

### Troubleshoot Failed Build

1. **Go to Actions:** https://github.com/rv2023/DeployKubeStack/actions
2. **Click "Release - Build and Push Docker Image"**
3. **Select the failed run**
4. **Expand step that failed**
5. **Check the error message**

Common issues:
- Dockerfile not found → Verify Dockerfile exists in root
- Build failure → Check Dockerfile and `make build` locally
- GHCR auth failure → Verify secrets (should be auto-granted)

## Configuration Options

### To build for multiple architectures (optional)

Edit `.github/workflows/release.yml` line ~43:

```yaml
- name: Build and push Docker image
  uses: docker/build-push-action@v5
  with:
    context: .
    push: true
    platforms: linux/amd64,linux/arm64,linux/arm/v7  # Add this line
    tags: ${{ steps.meta.outputs.tags }}
    # ... rest of config
```

### To skip GitHub Release creation (optional)

Comment out or remove the "Create GitHub Release" step in the workflow.

### To push to Docker Hub instead of GHCR (optional)

Update login action:
```yaml
- uses: docker/login-action@v3
  with:
    username: ${{ secrets.DOCKER_USERNAME }}
    password: ${{ secrets.DOCKER_PASSWORD }}
```

## Files Created/Modified

| File | Status | Purpose |
|------|--------|---------|
| `.github/workflows/release.yml` | ✅ Created | Release workflow |
| `docs/guides/RELEASE_PROCESS.md` | ✅ Created | Release documentation |
| `docs/RELEASE_WORKFLOW_SETUP.md` | ✅ Created | This file |
| `docs/README.md` | ✅ Updated | Added release process link |
| `Dockerfile` | ✅ Existing | Already optimized |

## Success Criteria

✅ Workflow file created: `.github/workflows/release.yml`  
✅ Documentation complete: `docs/guides/RELEASE_PROCESS.md`  
✅ Dockerfile optimized for production  
✅ GitHub Packages (GHCR) integration configured  
✅ Automatic tagging logic implemented  
✅ GitHub Release generation enabled  
✅ Security best practices applied  

## Quick Reference

**Create a release:**
```bash
git tag v1.0.0 -m "Release description"
git push origin v1.0.0
```

**Pull the image:**
```bash
docker pull ghcr.io/rv2023/DeployKubeStack:v1.0.0
```

**Deploy to Kubernetes:**
```bash
make deploy IMG=ghcr.io/rv2023/DeployKubeStack:v1.0.0
```

**Check releases:**
```bash
# GitHub web UI
open https://github.com/rv2023/DeployKubeStack/releases

# Or via CLI
gh release list
```

---

**Status:** Release workflow ready for production  
**Next Action:** Create first tag when Phase 1 testing complete
