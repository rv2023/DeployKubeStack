# Release Process - Building and Publishing Docker Images

## Overview

The DeployKubeStack operator uses GitHub Actions to automatically build and publish Docker images to GitHub Container Registry (GHCR) whenever a tag is created.

## Workflow: `.github/workflows/release.yml`

**Triggers:** Tag push (v* or release-*)  
**Actions:**
1. Checks out code
2. Sets up Docker Buildx for multi-platform builds
3. Authenticates with GitHub Container Registry
4. Builds Docker image
5. Pushes to GHCR
6. Creates GitHub Release with instructions

## How to Create a Release

### Step 1: Ensure Tests Pass

```bash
# Run all tests locally first
make test
make lint
make build
```

### Step 2: Create and Push a Tag

```bash
# Create a semantic version tag
git tag v1.0.0 -m "Release v1.0.0: Phase 1 Complete"

# Push the tag to GitHub
git push origin v1.0.0
```

**Tag format options:**
- `v1.0.0` - Semantic version (recommended)
- `v1.0.0-rc1` - Release candidate
- `release-2026-07-07` - Date-based release
- `v2.0.0-alpha` - Pre-release

### Step 3: GitHub Actions Workflow Runs Automatically

The workflow will:
1. ✅ Build the Docker image
2. ✅ Push to `ghcr.io/<org>/deploykubestack:<tag>`
3. ✅ Create a GitHub Release with instructions
4. ✅ Generate automatic release notes

### Step 4: Verify the Release

```bash
# View releases on GitHub
open https://github.com/rv2023/DeployKubeStack/releases

# Or pull the image locally
docker pull ghcr.io/rv2023/DeployKubeStack:v1.0.0
```

## Image Naming Convention

The workflow creates images with these tags:

| Tag Format | Example | Use Case |
|------------|---------|----------|
| `v1.0.0` | `ghcr.io/rv2023/DeployKubeStack:v1.0.0` | Release version |
| `latest` | Auto-tagged on main branch pushes | Development (if configured) |
| `sha-abc123` | Commit-based tag | Debugging specific commits |

**Current configuration:** Only creates versioned tags for tag pushes (v* and release-*)

## Using Released Images

### In Docker

```bash
# Pull the latest release
docker pull ghcr.io/rv2023/DeployKubeStack:v1.0.0

# Run locally (development)
docker run -it ghcr.io/rv2023/DeployKubeStack:v1.0.0
```

### In Kubernetes (Local Cluster)

```bash
# Deploy operator using released image
make deploy IMG=ghcr.io/rv2023/DeployKubeStack:v1.0.0

# Create test application
kubectl apply -f config/samples/apps_v1alpha1_application.yaml

# Verify deployment
kubectl get deployment -n deploykubestack-system
```

### In Your Helm Chart / Kustomize

```yaml
# kustomization.yaml
images:
  - name: deploykubestack
    newName: ghcr.io/rv2023/DeployKubeStack
    newTag: v1.0.0
```

```yaml
# values.yaml (Helm)
deploykubestack:
  image:
    repository: ghcr.io/rv2023/DeployKubeStack
    tag: v1.0.0
    pullPolicy: IfNotPresent
```

## Workflow Details

### Permissions

The workflow requires these GitHub permissions:
- `contents: read` - Read repository contents
- `packages: write` - Write to GitHub Packages (GHCR)

These are automatically granted to the `GITHUB_TOKEN` secret.

### Build Configuration

**Platform:** Linux (ubuntu-latest)  
**Buildx Cache:** GitHub Actions cache (enables faster rebuilds)  
**Docker Builder:** BuildKit (modern image building)

### Authentication

The workflow uses `${{ secrets.GITHUB_TOKEN }}` which is automatically available to all GitHub Actions workflows. No manual secret configuration needed.

## Troubleshooting

### Image not appearing in GHCR

**Check 1: Verify tag was pushed**
```bash
git tag -l
# Should show your tags
```

**Check 2: Check GitHub Actions workflow status**
- Go to: https://github.com/rv2023/DeployKubeStack/actions
- Click "Release - Build and Push Docker Image"
- Look for your tag in the run history
- Check logs for errors

**Check 3: Verify GHCR permissions**
- Go to: https://github.com/settings/packages
- Ensure personal access token has `write:packages` scope (if using PAT instead of GITHUB_TOKEN)

### Build fails with "No Dockerfile found"

**Solution:** Ensure `Dockerfile` exists in repository root
```bash
ls -la Dockerfile
# Should exist and be readable
```

### Image size is large

**Optimization tips:**
1. Use multi-stage builds in Dockerfile
2. Use smaller base images (alpine, distroless)
3. Remove unnecessary dependencies

Current Dockerfile should be optimized already.

## Advanced: Custom Release Configuration

### To also build for multiple architectures

Update the workflow to include:
```yaml
- name: Build and push Docker image
  uses: docker/build-push-action@v5
  with:
    context: .
    push: true
    platforms: linux/amd64,linux/arm64
    tags: ${{ steps.meta.outputs.tags }}
    labels: ${{ steps.meta.outputs.labels }}
```

### To create release only for main branch tags

Update trigger in workflow:
```yaml
on:
  push:
    tags:
      - 'v*'
    branches:
      - main
```

### To skip release notes generation

Remove or comment out the "Create GitHub Release" step.

## Release Checklist

Before creating a release tag:

- [ ] All tests passing: `make test`
- [ ] Linting passes: `make lint`
- [ ] Build succeeds: `make build`
- [ ] CLAUDE.md updated with new features
- [ ] docs/ updated with relevant changes
- [ ] Version bump in code (if applicable)
- [ ] Release notes prepared (auto-generated, but review them)
- [ ] Tag follows semantic versioning

After tag is pushed:

- [ ] GitHub Actions workflow completes successfully
- [ ] GitHub Release appears at: github.com/rv2023/DeployKubeStack/releases
- [ ] Docker image accessible: `docker pull ghcr.io/rv2023/DeployKubeStack:vX.Y.Z`
- [ ] Release notes are clear and accurate

## Example Release Commands

### Phase 1 Release (First Release)

```bash
git tag v1.0.0 -m "Release v1.0.0: Phase 1 Complete - Deployment and Service Management"
git push origin v1.0.0
```

### Bug Fix Release

```bash
git tag v1.0.1 -m "Release v1.0.1: Fix for issue #123"
git push origin v1.0.1
```

### Release Candidate

```bash
git tag v1.1.0-rc1 -m "Release Candidate: Phase 2 Preview"
git push origin v1.1.0-rc1
```

## CI/CD Pipeline

The complete release pipeline:

```
Developer pushes tag
    ↓
GitHub Actions triggers release workflow
    ↓
Docker image built (BuildKit)
    ↓
Image pushed to GHCR
    ↓
GitHub Release created with instructions
    ↓
Image available for deployment
```

## Links

- **Releases:** https://github.com/rv2023/DeployKubeStack/releases
- **Packages:** https://github.com/rv2023/DeployKubeStack/pkgs/container/deploykubestack
- **Actions:** https://github.com/rv2023/DeployKubeStack/actions
- **Workflow:** `.github/workflows/release.yml`

---

**Status:** Release workflow ready for first tag  
**Next Release:** After Phase 1 verification complete
