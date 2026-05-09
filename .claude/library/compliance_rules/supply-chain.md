---
description: Supply chain security requirements — dependency management, pinning, vulnerability scanning, license compliance, SRI
languages:
- python
- javascript
- typescript
- go
- rust
- java
alwaysApply: false
---

# Supply Chain Security

Covers: NIS2 Art. 21(2)(d) | ISO 27001 A.8.28 (dependency management) | OWASP ASVS V14.2

## Dependency Pinning (NIS2 Art. 21(2)(d))

- All dependencies pinned to exact versions
- Lockfiles present and committed: `package-lock.json`, `go.sum`, `Cargo.lock`, `poetry.lock`, `yarn.lock`, `pnpm-lock.yaml`
- No floating version ranges in production dependencies (no `^`, `~`, `*`, `latest`)

## Vulnerability Scanning (NIS2 Art. 21(2)(d), ASVS V14.2.1)

- SCA (Software Composition Analysis) integrated in CI pipeline
- No dependencies with known critical/high CVEs merged without documented justification
- Security patches applied within defined SLAs: critical = days, high = weeks
- Tools: `npm audit`, `pip-audit`, `govulncheck`, Snyk, Dependabot, or equivalent

## Dependency Provenance (NIS2 Art. 21(2)(d), ASVS V14.2.4)

- Dependencies sourced from official registries only
- No references to private/unknown registries without documentation and review
- Container base images: official/minimal, pinned by digest (not just tag), no `latest` in production
- Vendored/copied third-party code reviewed for security and kept up to date

## Integrity Verification (NIS2 Art. 21(2)(d), ASVS V14.2.3)

- Checksums/hashes verified for dependencies (handled by lockfiles)
- Subresource Integrity (SRI) for CDN-loaded scripts: `<script>` and `<link>` tags for external resources have `integrity` attribute (ASVS V14.2.3)
- No external script loads without hash verification (ASVS V12.3.6)

## Minimal Dependency Surface (NIS2 Art. 21(2)(d))

- New dependencies justified — no unnecessary transitive dependency trees
- Prefer standard library where feasible
- All unneeded features, samples, documentation removed from production builds (ASVS V14.2.2)

## License Compliance (NIS2 Art. 21(2)(d))

- License of newly added packages checked before merge
- No copyleft licenses (GPL, AGPL, EUPL) in proprietary projects without legal review
- Bundled code includes license attribution
- Dependency version upgrades checked for license term changes

### License Classification

| Category | Licenses | Action |
|----------|----------|--------|
| Permissive (OK) | MIT, Apache-2.0, BSD-2-Clause, BSD-3-Clause, ISC, Unlicense, CC0 | Allowed |
| Weak copyleft (Review) | LGPL-2.1, LGPL-3.0, MPL-2.0, EPL-2.0 | Allowed if dynamically linked; review if bundled |
| Strong copyleft (Block) | GPL-2.0, GPL-3.0, AGPL-3.0, SSPL, EUPL-1.2 | Blocked in proprietary projects unless entire project is under same license |
| Unknown | No license, custom license | Blocked until reviewed |

## Software Bill of Materials (ASVS V14.2.5)

- SBOM generation in build pipeline (CycloneDX, SPDX format)
- SBOM maintained and available for incident response and compliance audits

## Third-Party Encapsulation (ASVS V14.2.6)

- Third-party libraries wrapped in facade/adapter — not directly exposed throughout codebase
- Enables replacement without widespread code changes

---

For authoritative sources, verification record, and direct links to official texts, see [`SOURCES.md`](SOURCES.md).
