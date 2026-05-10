# Release Workflow Templates

Language-specific GitHub Actions workflow templates for `.github/workflows/release.yml`.

## Go

```yaml
name: Release

on:
  push:
    tags:
      - "v*.*.*"

permissions:
  contents: write

jobs:
  release:
    name: Build and Release
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v5
      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod

      - name: Validate tag format
        run: |
          if ! echo "${GITHUB_REF_NAME}" | grep -qE '^v[0-9]+\.[0-9]+\.[0-9]+$'; then
            echo "::error::Invalid tag format: ${GITHUB_REF_NAME}"
            exit 1
          fi

      - name: Build binaries
        run: |
          tag="${GITHUB_REF_NAME}"
          name="PROJECT_NAME"  # Replace with actual binary name
          for pair in linux/amd64 linux/arm64 darwin/amd64 darwin/arm64; do
            os="${pair%/*}"
            arch="${pair#*/}"
            output="${name}-${os}-${arch}"
            GOOS="${os}" GOARCH="${arch}" CGO_ENABLED=0 \
              go build -ldflags "-s -w -X main.version=${tag}" -o "${output}"
          done

      - name: Generate checksums
        run: sha256sum PROJECT_NAME-* > checksums.txt

      - name: Create release
        uses: softprops/action-gh-release@v2
        with:
          generate_release_notes: true
          files: |
            PROJECT_NAME-*
            checksums.txt
```

## Node.js

```yaml
name: Release

on:
  push:
    tags:
      - "v*.*.*"

permissions:
  contents: write

jobs:
  release:
    name: Release
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v5
      - uses: actions/setup-node@v4
        with:
          node-version-file: .node-version
          registry-url: https://registry.npmjs.org

      - name: Validate tag format
        run: |
          if ! echo "${GITHUB_REF_NAME}" | grep -qE '^v[0-9]+\.[0-9]+\.[0-9]+$'; then
            echo "::error::Invalid tag format: ${GITHUB_REF_NAME}"
            exit 1
          fi

      - run: npm ci
      - run: npm run build --if-present
      - run: npm test

      - name: Create release
        uses: softprops/action-gh-release@v2
        with:
          generate_release_notes: true

      # Uncomment to publish to npm:
      # - run: npm publish
      #   env:
      #     NODE_AUTH_TOKEN: ${{ secrets.NPM_TOKEN }}
```

## Rust

```yaml
name: Release

on:
  push:
    tags:
      - "v*.*.*"

permissions:
  contents: write

jobs:
  release:
    name: Build and Release
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v5
      - uses: dtolnay/rust-toolchain@stable

      - name: Validate tag format
        run: |
          if ! echo "${GITHUB_REF_NAME}" | grep -qE '^v[0-9]+\.[0-9]+\.[0-9]+$'; then
            echo "::error::Invalid tag format: ${GITHUB_REF_NAME}"
            exit 1
          fi

      - name: Build release binaries
        run: cargo build --release

      - name: Create release
        uses: softprops/action-gh-release@v2
        with:
          generate_release_notes: true
          files: target/release/PROJECT_NAME
```

## Python

```yaml
name: Release

on:
  push:
    tags:
      - "v*.*.*"

permissions:
  contents: write

jobs:
  release:
    name: Release
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v5
      - uses: actions/setup-python@v5
        with:
          python-version-file: pyproject.toml

      - name: Validate tag format
        run: |
          if ! echo "${GITHUB_REF_NAME}" | grep -qE '^v[0-9]+\.[0-9]+\.[0-9]+$'; then
            echo "::error::Invalid tag format: ${GITHUB_REF_NAME}"
            exit 1
          fi

      - run: pip install build
      - run: python -m build

      - name: Create release
        uses: softprops/action-gh-release@v2
        with:
          generate_release_notes: true
          files: dist/*

      # Uncomment to publish to PyPI:
      # - run: pip install twine
      # - run: twine upload dist/*
      #   env:
      #     TWINE_USERNAME: __token__
      #     TWINE_PASSWORD: ${{ secrets.PYPI_TOKEN }}
```

## Docker-only

```yaml
name: Release

on:
  push:
    tags:
      - "v*.*.*"

permissions:
  contents: write
  packages: write

jobs:
  release:
    name: Build and Push
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v5

      - name: Validate tag format
        run: |
          if ! echo "${GITHUB_REF_NAME}" | grep -qE '^v[0-9]+\.[0-9]+\.[0-9]+$'; then
            echo "::error::Invalid tag format: ${GITHUB_REF_NAME}"
            exit 1
          fi

      - uses: docker/login-action@v3
        with:
          registry: ghcr.io
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}

      - uses: docker/build-push-action@v6
        with:
          push: true
          tags: |
            ghcr.io/${{ github.repository }}:${{ github.ref_name }}
            ghcr.io/${{ github.repository }}:latest

      - name: Create release
        uses: softprops/action-gh-release@v2
        with:
          generate_release_notes: true
```
