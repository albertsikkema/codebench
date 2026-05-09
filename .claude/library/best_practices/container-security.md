# Container Security

## Principle

A container image is a deployable artifact with an attack surface. Minimize that surface: use minimal base images, run as non-root, drop capabilities, make the filesystem read-only, never bake in secrets, and scan for vulnerabilities. Every layer of the image is a decision about what an attacker can use if they get in.

## Why

- **Containers are not sandboxes**: A container shares the host kernel. A privileged container or one with excessive capabilities is one exploit away from host access.
- **Images are immutable records**: Every `RUN`, `COPY`, and `ENV` instruction creates a layer that is permanently stored. A secret added in layer 3 and "deleted" in layer 5 still exists in layer 3 and can be extracted.
- **Supply chain starts at the base image**: Your application inherits every vulnerability in its base image. A full Ubuntu image has hundreds of packages you don't use — each one is a potential attack vector you don't need.

## Core Rules

### 1. Use Minimal Base Images

| Image type | Use case | Size | Attack surface |
|-----------|----------|------|----------------|
| `distroless` (Google/Chainguard) | Compiled languages (Go, Rust) | ~2-20MB | Minimal — no shell, no package manager |
| `*-slim` (debian-slim, python-slim) | Interpreted languages | ~50-150MB | Small — base OS packages only |
| `alpine` | When musl libc is acceptable | ~5-50MB | Small — but musl can break native extensions |
| Full OS (ubuntu, debian) | Never in production | ~200-500MB | Large — hundreds of unnecessary packages |

**Rule**: Production images must use distroless or slim variants. Full OS images are never acceptable in production.

### 2. Multi-Stage Builds

Separate build dependencies from the runtime image. The final image should contain only what's needed to run.

```dockerfile
# Stage 1: Build
FROM golang:1.22 AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /app ./cmd/server

# Stage 2: Runtime — no compiler, no source code, no build tools
FROM gcr.io/distroless/static-debian12
COPY --from=builder /app /app
ENTRYPOINT ["/app"]
```

**What must NOT be in the final image**: compilers, package managers, source code, test files, development dependencies, IDE configs, `.git` directory.

Use `.dockerignore` to exclude: `.git`, `tests/`, `docs/`, `node_modules/`, `__pycache__/`, `*.md`, IDE files.

### 3. Run as Non-Root

Create a dedicated application user. Never run as root.

```dockerfile
# Use UID 65532 (matches distroless "nonroot" convention)
RUN addgroup --gid 65532 appgroup && \
    adduser --uid 65532 --ingroup appgroup --disabled-password --no-create-home appuser

# Application code owned by root (prevents self-modification)
COPY --from=builder --chown=root:root /app /app

# Writable directories owned by app user
RUN mkdir -p /data /tmp && chown 65532:65532 /data /tmp

USER 65532
```

**Why UID 65532?** It avoids collision with system users (0-999) and human users (1000-65000). It matches the distroless `nonroot` user convention.

### 4. Read-Only Filesystem

Run containers with `--read-only` to prevent runtime modification of application code.

```yaml
# Docker Compose
services:
  app:
    read_only: true
    tmpfs:
      - /tmp:noexec,nosuid
    volumes:
      - app-data:/data  # only explicit writable paths
```

Writable paths must be explicitly defined as tmpfs or volume mounts. This limits what an attacker can write to if they achieve code execution.

### 5. Drop All Capabilities

Linux capabilities grant specific privileges. Drop all of them and add back only what's needed.

```yaml
# Docker Compose
services:
  app:
    cap_drop:
      - ALL
    # cap_add:
    #   - NET_BIND_SERVICE  # only if binding to port < 1024
    security_opt:
      - no-new-privileges:true
```

**Never use**:
- `--privileged` (grants full host access)
- Docker socket mounts (`/var/run/docker.sock`) — this is equivalent to root on the host
- `SYS_ADMIN` capability (too broad)

**Prefer binding to high ports** (8080, 3000) over adding `NET_BIND_SERVICE` to bind to 80/443. Use a reverse proxy for port mapping.

### 6. Never Bake Secrets into Images

Secrets in image layers are permanent and extractable with `docker history` or `docker save`.

| Phase | Method | Safe? |
|-------|--------|-------|
| Build-time | `ARG SECRET` / `ENV SECRET` | No — visible in `docker history` |
| Build-time | `COPY .env /app/` | No — stored in layer |
| Build-time | BuildKit `--mount=type=secret` | Yes — tmpfs, not in layer |
| Runtime | `-e SECRET=value` / `--env-file` | Yes — not in image |
| Runtime | Mounted secrets (`/run/secrets/`) | Yes — not in image |

```dockerfile
# GOOD: BuildKit secret mount (build-time)
RUN --mount=type=secret,id=npm_token \
    NPM_TOKEN=$(cat /run/secrets/npm_token) npm ci

# BAD: ARG is visible in docker history
ARG NPM_TOKEN
RUN npm ci
```

**Verify with**: `docker history <image>` — no layer should contain secret values.

### 7. Pin Base Images by Digest

Tags are mutable — `python:3.12-slim` today may be different from `python:3.12-slim` next month. Pin by digest for reproducible builds.

```dockerfile
FROM python:3.12-slim@sha256:abc123...
```

Also pin OS packages to exact versions in `RUN` commands, and ensure application dependencies are pinned via lock files.

### 8. Scan for Vulnerabilities

Run a vulnerability scanner (Trivy, Grype, or equivalent) in CI on every build.

```bash
# Scan image for vulnerabilities
trivy image --severity CRITICAL,HIGH --exit-code 1 myapp:latest

# Scan Dockerfile for misconfigurations
trivy config --exit-code 1 .
```

- **Fail on**: CRITICAL severity
- **Flag on**: HIGH severity (with `--ignore-unfixed` to skip unpatched CVEs)
- **Maintain**: `.trivyignore` for accepted risks with justification comments
- **Schedule**: Nightly scans of deployed images for newly discovered CVEs

### 9. Add OCI Labels

Include metadata so images are traceable back to source code:

```dockerfile
LABEL org.opencontainers.image.title="myapp" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.created="${BUILD_DATE}" \
      org.opencontainers.image.revision="${VCS_REF}" \
      org.opencontainers.image.source="https://github.com/org/repo" \
      org.opencontainers.image.licenses="MIT"
```

Pass `BUILD_DATE`, `VCS_REF`, and `VERSION` as build-args.

## Complete Example

```dockerfile
# --- Build stage ---
FROM golang:1.22-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app ./cmd/server

# --- Runtime stage ---
FROM gcr.io/distroless/static-debian12

ARG BUILD_DATE
ARG VCS_REF
ARG VERSION

LABEL org.opencontainers.image.title="myapp" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.created="${BUILD_DATE}" \
      org.opencontainers.image.revision="${VCS_REF}" \
      org.opencontainers.image.source="https://github.com/org/repo"

COPY --from=builder /app /app

USER 65532:65532
ENTRYPOINT ["/app"]
```

## When to Bend the Rules

- **Development images**: Use full images with shell access for debugging. Never deploy them.
- **Alpine with musl issues**: Switch to `debian:*-slim` if native extensions (Python C extensions, Node.js native modules) break on musl libc.
- **CI/CD runner images**: Build tool images (compilers, linters) don't need the same hardening as production runtime images.
- **Debugging production issues**: Temporarily run with a shell (ephemeral debug container in Kubernetes) rather than weakening the production image.
