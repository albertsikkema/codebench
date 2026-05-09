# Baseline Requirements — Evidence & Sources

This document provides authoritative evidence backing each requirement in `baseline-requirements.md`. Sources include international standards, RFCs, peer-reviewed research, industry frameworks, and established best-practice guides.

---

## Security

### REQ-SEC-001 — HTTPS everywhere

- **OWASP ASVS v4.0, §V9 — Communication Security**: Requires TLS for all connections, especially those containing sensitive data or credentials.
  https://owasp.org/www-project-application-security-verification-standard/
- **NIST SP 800-52 Rev. 2 — Guidelines for TLS Implementations**: Federal systems must use TLS 1.2+ and shall not use TLS 1.0/1.1 or SSL.
  https://csrc.nist.gov/publications/detail/sp/800-52/rev-2/final
- **Google Search ranking signal**: Since 2014, Google uses HTTPS as a ranking signal; Chrome marks HTTP sites "Not Secure."
  https://developers.google.com/search/blog/2014/08/https-as-ranking-signal
- See also: `.claude/library/compliance_rules/cryptography.md` (ISO A.8.24, ASVS V9.1)

### REQ-SEC-002 — Encryption & hashing

- **OWASP Password Storage Cheat Sheet**: Recommends bcrypt (cost >= 10), argon2id, or scrypt. Never store plaintext or reversible encryption.
  https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html
- **NIST SP 800-132 — Recommendation for Password-Based Key Derivation**: Defines requirements for secure password hashing functions.
  https://csrc.nist.gov/publications/detail/sp/800-132/final
- **NIST SP 800-175B — Guideline for Using Cryptographic Standards**: AES-256 recommended for data at rest.
  https://csrc.nist.gov/publications/detail/sp/800-175b/rev-1/final
- See also: `.claude/library/compliance_rules/cryptography.md` (ISO A.8.24, NIS2 Art. 21(2)(h))

### REQ-SEC-003 — No secrets in source code

- **CWE-798 — Use of Hard-coded Credentials**: Classified as a critical weakness. Credentials in source are discoverable via decompilation, SCM history, or repository exposure.
  https://cwe.mitre.org/data/definitions/798.html
- **OWASP Top 10 Proactive Controls 2024, C10 — Stop Server Side Request Forgery / Secrets Management**: Secrets must be externalized.
  https://top10proactive.owasp.org/
- **The Twelve-Factor App, §III — Config**: *"The twelve-factor app stores config in environment variables... unlike config files, there is little chance of them being checked into the code repo accidentally."*
  https://12factor.net/config
- See also: `.claude/library/compliance_rules/configuration-security.md` (ISO A.8.9)

### REQ-SEC-012 — No secrets in Docker images

- **Docker Best Practices — Build secrets**: *"Don't leak build secrets... use Docker BuildKit `--mount=type=secret`."*
  https://docs.docker.com/build/building/best-practices/
- **CIS Docker Benchmark v1.6.0, §4.1**: Do not store secrets in Dockerfiles.
  https://www.cisecurity.org/benchmark/docker

### REQ-SEC-004 — Input validation

- **OWASP ASVS v4.0, §V5 — Validation, Sanitization, Encoding**: *"The most common web application security weakness is the failure to properly validate input."*
  https://github.com/OWASP/ASVS/blob/master/4.0/en/0x13-V5-Validation-Sanitization-Encoding.md
- **OWASP Top 10 Proactive Controls 2024, C3 — Validate all Input & Handle Exceptions**
  https://top10proactive.owasp.org/the-top-10/c3-validate-input-and-handle-exceptions/
- **CWE-20 — Improper Input Validation**: Top-ranked weakness in multiple CWE Top 25 lists.
  https://cwe.mitre.org/data/definitions/20.html
- See also: `.claude/library/compliance_rules/secure-coding.md` (ISO A.8.28, ASVS V5)

### REQ-SEC-005 — Parameterized queries

- **OWASP SQL Injection Prevention Cheat Sheet**: *"Use of Prepared Statements (with Parameterized Queries) is how all developers should first be taught how to write database queries."*
  https://cheatsheetseries.owasp.org/cheatsheets/SQL_Injection_Prevention_Cheat_Sheet.html
- **CWE-89 — Improper Neutralization of Special Elements in SQL Command**: Consistently in the CWE Top 25 Most Dangerous Software Weaknesses.
  https://cwe.mitre.org/data/definitions/89.html
- See also: `.claude/library/compliance_rules/secure-coding.md` (ASVS V5.3.4)

### REQ-SEC-006 — XSS prevention

- **OWASP XSS Prevention Cheat Sheet**: Context-aware output encoding is the primary defense against XSS.
  https://cheatsheetseries.owasp.org/cheatsheets/Cross_Site_Scripting_Prevention_Cheat_Sheet.html
- **CWE-79 — Improper Neutralization of Input During Web Page Generation**: #2 on CWE Top 25 (2023).
  https://cwe.mitre.org/data/definitions/79.html
- See also: `.claude/library/compliance_rules/secure-coding.md` (ASVS V5.3.3)

### REQ-SEC-007 — Restrictive CORS

- **OWASP CORS Misconfiguration**: Wildcard `Access-Control-Allow-Origin: *` with credentials is explicitly forbidden by the Fetch specification and creates credential-theft attack vectors.
  https://owasp.org/www-project-web-security-testing-guide/latest/4-Web_Application_Security_Testing/11-Client-side_Testing/07-Testing_Cross_Origin_Resource_Sharing
- **MDN Web Docs — CORS**: Detailed specification of the same-origin policy and how CORS relaxes it.
  https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS
- See also: `.claude/library/compliance_rules/secure-coding.md` (ASVS V14.5.3)

### REQ-SEC-008 — Token expiry & rotation

- **OWASP Session Management Cheat Sheet**: Short-lived access tokens limit the window of compromise. Refresh token rotation detects stolen tokens.
  https://cheatsheetseries.owasp.org/cheatsheets/Session_Management_Cheat_Sheet.html
- **RFC 6749, §10.4 — Refresh Token**: Refresh tokens should be bound to the client and rotated on use.
  https://www.rfc-editor.org/rfc/rfc6749#section-10.4
- See also: `.claude/library/compliance_rules/session-cookie.md` (ASVS V3.3)

### REQ-SEC-009 — Rate limiting

- **OWASP Blocking Brute Force Attacks**: Rate limiting is a primary defense against credential stuffing and brute-force.
  https://owasp.org/www-community/controls/Blocking_Brute_Force_Attacks
- **RFC 6585, §4 — 429 Too Many Requests**: Defines the standard status code for rate-limited responses.
  https://www.rfc-editor.org/rfc/rfc6585#section-4
- **IETF Draft — RateLimit Header Fields for HTTP**: Standardizes `RateLimit-Limit`, `RateLimit-Remaining`, `RateLimit-Reset` headers.
  https://datatracker.ietf.org/doc/draft-ietf-httpapi-ratelimit-headers/
- See also: `.claude/library/compliance_rules/auth-boundaries.md` (ASVS V2.2.1)

### REQ-SEC-010 — Security headers

- **OWASP HTTP Headers Cheat Sheet**: Recommends `Strict-Transport-Security: max-age=63072000`, `X-Content-Type-Options: nosniff`, `X-Frame-Options: DENY`, and Content-Security-Policy.
  https://cheatsheetseries.owasp.org/cheatsheets/HTTP_Headers_Cheat_Sheet.html
- **Mozilla Observatory**: Grades websites on security header adoption.
  https://observatory.mozilla.org/
- See also: `.claude/library/compliance_rules/secure-coding.md` (ASVS V14.4)

### REQ-SEC-011 — Dependency vulnerability scanning

- **OWASP Top 10:2021, A06 — Vulnerable and Outdated Components**: *"You are likely vulnerable if you do not know the versions of all components you use... [or] do not scan for vulnerabilities regularly."*
  https://owasp.org/Top10/A06_2021-Vulnerable_and_Outdated_Components/
- **NIST Cybersecurity Framework, ID.RA**: Risk assessment includes identifying vulnerabilities in software dependencies.
  https://www.nist.gov/cyberframework
- See also: `.claude/library/compliance_rules/supply-chain.md` (NIS2 Art. 21(2)(d), ASVS V14.2.1)

### REQ-SEC-013 — SAST & secret detection

- **NIST SP 800-218 — Secure Software Development Framework (SSDF)**: PW.7 — *"Review and/or analyze human-readable code to identify vulnerabilities and verify compliance with security requirements."*
  https://csrc.nist.gov/publications/detail/sp/800-218/final
- **OWASP Source Code Analysis Tools**: Recommends integrating SAST into CI/CD pipelines.
  https://owasp.org/www-community/Source_Code_Analysis_Tools
- See also: `.claude/library/compliance_rules/secure-development.md` (ISO A.8.29, NIS2 Art. 21(2)(e))

---

## Data

### REQ-DATA-001 — Versioned migrations

- **Martin Fowler — Evolutionary Database Design**: *"All database changes are migration scripts... versioned and kept in source control."*
  https://martinfowler.com/articles/evodb.html
- **Thoughtworks Technology Radar**: Database migrations as code has been in the "Adopt" ring consistently.
  https://www.thoughtworks.com/radar

### REQ-DATA-002 — PII minimization

- **GDPR Article 5(1)(c) — Data Minimisation**: *"Personal data shall be adequate, relevant and limited to what is necessary in relation to the purposes for which they are processed."*
  https://gdpr-info.eu/art-5-gdpr/
- **NIST Privacy Framework, CT.DM-P**: Control for data minimization — only collect data necessary for the stated purpose.
  https://www.nist.gov/privacy-framework
- See also: `.claude/library/compliance_rules/gdpr-processing-principles.md` (GDPR Art. 5(1)(c))

### REQ-DATA-003 — Audit trails

- **OWASP ASVS v4.0, §V7 — Error Handling and Logging**: Security-relevant events must be logged with sufficient detail for investigation.
  https://github.com/OWASP/ASVS
- **SOC 2 — CC7.2**: Monitoring and logging of system activities including authentication and authorization events.
- **GDPR Article 30**: Requires records of processing activities.
  https://gdpr-info.eu/art-30-gdpr/
- See also: `.claude/library/compliance_rules/audit-trail.md` (ISO A.8.15, NIS2 Art. 21(2)(b), ASVS V7)

### REQ-DATA-004, REQ-DATA-005 — Automated backups

- **AWS Well-Architected Framework, Reliability Pillar — REL 9**: *"Back up data, applications, and configuration to meet your recovery objectives."*
  https://docs.aws.amazon.com/wellarchitected/latest/reliability-pillar/back-up-data-applications-and-configuration.html
- **AWS RDS PITR**: Continuous WAL archiving enables point-in-time recovery to within 5 minutes.
  https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/USER_PIT.html

---

## API & Interface

### REQ-API-001 — RFC 9457 Problem Details

- **RFC 9457 — Problem Details for HTTP APIs** (obsoletes RFC 7807): *"This document defines a 'problem detail' to carry machine-readable details of errors in HTTP response content to avoid the need to define new error response formats for HTTP APIs."*
  https://www.rfc-editor.org/rfc/rfc9457.html
- **Swagger/OpenAPI Blog**: *"RFC 9457 provides a standardized JSON format for consistent error handling across APIs."*
  https://swagger.io/blog/problem-details-rfc9457-doing-api-errors-well/

### REQ-API-002 — HTTP status codes

- **RFC 9110 — HTTP Semantics, §15**: Defines the canonical meaning of each status code class.
  https://www.rfc-editor.org/rfc/rfc9110#section-15
- **Richardson Maturity Model, Level 2**: Proper use of HTTP verbs and status codes is a fundamental REST maturity requirement.
  https://martinfowler.com/articles/richardsonMaturityModel.html

### REQ-API-003 — Pagination

- **OWASP REST Security Cheat Sheet**: Recommends pagination to prevent excessive data exposure and resource exhaustion.
  https://cheatsheetseries.owasp.org/cheatsheets/REST_Security_Cheat_Sheet.html
- **Google API Design Guide — Standard Methods**: List endpoints must support pagination.
  https://cloud.google.com/apis/design/standard_methods

### REQ-API-004 — Request payload validation

- **OWASP REST Security Cheat Sheet**: *"Use OpenAPI or JSON Schema to define and validate request structure."*
  https://cheatsheetseries.owasp.org/cheatsheets/REST_Security_Cheat_Sheet.html
- **OWASP API Security Top 10:2023, API3 — Broken Object Property Level Authorization**: Reject extra fields to prevent mass assignment.
  https://owasp.org/API-Security/editions/2023/en/0xa3-broken-object-property-level-authorization/

---

## Infrastructure

### REQ-INFRA-001, REQ-INFRA-005 — Health & readiness endpoints

- **Kubernetes Documentation — Configure Liveness, Readiness and Startup Probes**: Separate liveness (is the process alive?) and readiness (can it serve traffic?) probes.
  https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/
- **Microsoft Well-Architected Framework — Health Endpoint Monitoring pattern**: Health endpoints must not leak internal details.
  https://learn.microsoft.com/en-us/azure/architecture/patterns/health-endpoint-monitoring

### REQ-INFRA-002 — Config from environment

- **The Twelve-Factor App, §III — Config**: *"The twelve-factor app stores config in environment variables... Env vars are easy to change between deploys without changing any code."*
  https://12factor.net/config

### REQ-INFRA-003 — Graceful shutdown

- **Kubernetes — Pod Lifecycle, Termination**: Kubernetes sends SIGTERM and expects the process to finish in-flight work within the grace period.
  https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/#pod-termination
- **The Twelve-Factor App, §IX — Disposability**: *"Processes should strive to minimize startup time... [and] shut down gracefully when they receive a SIGTERM."*
  https://12factor.net/disposability

### REQ-INFRA-004 — Container-ready

- **The Twelve-Factor App, §VI — Processes**: *"Twelve-factor processes are stateless and share-nothing."*
  https://12factor.net/processes

### REQ-INFRA-006 — Minimal base images

- **Docker Best Practices**: *"A smaller base image not only offers portability and fast downloads, but also shrinks the size of your image and minimizes the number of vulnerabilities introduced through the dependencies."*
  https://docs.docker.com/build/building/best-practices/
- **Chainguard Images**: Distroless images reduce CVE surface to near-zero.
  https://www.chainguard.dev/chainguard-images

### REQ-INFRA-007 — Non-root containers

- **CIS Docker Benchmark v1.6.0, §4.1**: *"Do not run containers as root. Use a non-root user."*
  https://www.cisecurity.org/benchmark/docker
- **Docker Best Practices**: *"If a service can run without privileges, use USER to change to a non-root user."*
  https://docs.docker.com/build/building/best-practices/

### REQ-INFRA-008 — Image vulnerability scanning

- **NIST SP 800-190 — Application Container Security Guide, §3.1.3**: *"Image vulnerabilities... Organizations should use tools to detect and address vulnerabilities in images."*
  https://csrc.nist.gov/publications/detail/sp/800-190/final

### REQ-INFRA-009 — Multi-stage builds

- **Docker Best Practices — Multi-stage builds**: *"Multi-stage builds let you reduce the size of your final image, by creating a cleaner separation between the building of your image and the final output."*
  https://docs.docker.com/build/building/best-practices/

### REQ-INFRA-010 — Reproducible builds (pinned digests)

- **Docker Image Digests**: *"By pinning your images to a digest, you're guaranteed to always use the same image version, even if a publisher replaces the tag."*
  https://docs.docker.com/build/building/best-practices/
- **Chainguard Academy**: *"Digests address the integrity portion of the CIA Triad by providing a unique immutable identifier... Even a small change in the image content will result in a completely different digest."*
  https://edu.chainguard.dev/chainguard/chainguard-images/how-to-use/container-image-digests/

### REQ-INFRA-011 — OCI metadata labels

- **OCI Image Spec — Annotations**: Defines standard label keys for container image metadata.
  https://github.com/opencontainers/image-spec/blob/main/annotations.md

### REQ-INFRA-012 — Read-only root filesystem

- **CIS Docker Benchmark v1.6.0, §5.12**: *"Mount the container's root filesystem as read-only."*
  https://www.cisecurity.org/benchmark/docker
- **NIST SP 800-190, §4.3.3**: Limit container filesystem writability to prevent runtime modification of application code.

### REQ-INFRA-013 — Drop all capabilities

- **CIS Docker Benchmark v1.6.0, §5.3/5.4**: *"Drop all capabilities and only add those specifically needed."*
  https://www.cisecurity.org/benchmark/docker
- **NIST SP 800-190, §4.3.2**: Run containers with least-privilege capabilities.

---

## Observability

### REQ-OBS-001 — Structured logging

- **Google SRE Book, Ch. 6 — Monitoring Distributed Systems**: Structured, machine-parseable logs are essential for automated alerting and analysis.
  https://sre.google/sre-book/monitoring-distributed-systems/
- **The Twelve-Factor App, §XI — Logs**: *"A twelve-factor app never concerns itself with routing or storage of its output stream"* — logs go to stdout as structured events.
  https://12factor.net/logs

### REQ-OBS-002, REQ-OBS-003 — Error context & correlation IDs

- **W3C Trace Context specification**: Standardizes trace propagation across distributed systems via `traceparent` and `tracestate` headers.
  https://www.w3.org/TR/trace-context/
- **Google SRE Book, Ch. 12 — Effective Troubleshooting**: Correlated request tracing is essential for diagnosing distributed failures.
  https://sre.google/sre-book/effective-troubleshooting/

---

## Code Quality

### REQ-QUAL-001 — Strict type checking

- **Gao et al. (2017) — "To Type or Not to Type"**: Study of JavaScript and TypeScript repos found that *"using static type analysis tools could have prevented 15% of the bugs"* in studied projects.
  https://dl.acm.org/doi/10.1145/3133872
- **Microsoft Research (2014) — "An Empirical Study on the Impact of Static Typing on Software Maintainability"**: Static types improve code comprehension and reduce defect rates.

### REQ-QUAL-002 — Linting & formatting in CI

- **Google Engineering Practices — Code Review Guidelines**: Consistent style enforced by automated tools reduces review friction and cognitive load.
  https://google.github.io/eng-practices/review/
- **Thoughtworks Technology Radar**: Automated formatting (Prettier, Black, gofmt) in the "Adopt" ring.

### REQ-QUAL-003 — Coverage targets (80%/70%)

- **Minimum Acceptable Code Coverage — industry consensus**: 80% is widely cited as the point of diminishing returns. Below 70%, critical paths are likely untested.
- **Martin Fowler — Test Coverage**: *"I would be suspicious of anything like 100%... the value of coverage is in finding which bits of your code aren't being tested."*
  https://martinfowler.com/bliki/TestCoverage.html

### REQ-QUAL-005, REQ-QUAL-006 — Test strategy & quality

- **Kent Beck — Test-Driven Development**: Tests should verify behavior, not implementation details.
- **Google Testing Blog — Testing on the Toilet**: Tests should be focused, readable, and catch real bugs.
  https://testing.googleblog.com/

### REQ-QUAL-004 — Pinned dependencies

- **OWASP Dependency-Check**: Unpinned dependencies allow supply-chain attacks via version substitution.
  https://owasp.org/www-project-dependency-check/
- **npm audit / Dependabot**: Lock files ensure deterministic installs; Dependabot automates update PRs with CVE alerts.

### REQ-QUAL-007 — Test framework from day one

- **Shift-Left Testing (IBM)**: *"Shift-left testing moves testing activities earlier in the development process... a bug caught during development can cost up to 30x less than the same bug in production."*
  https://www.ibm.com/think/topics/shift-left-testing
- **Microsoft — Shift testing left with unit tests**: Testing from project inception establishes conventions and catches regressions early.
  https://learn.microsoft.com/en-us/devops/develop/shift-left-make-testing-fast-reliable

---

## Versioning & Changelog

### REQ-VER-001 — Semantic Versioning

- **Semantic Versioning 2.0.0 (semver.org)**: The official specification authored by Tom Preston-Werner (GitHub co-founder).
  https://semver.org/

### REQ-VER-002 — Keep a Changelog

- **Keep a Changelog 1.1.0**: *"A changelog is a file which contains a curated, chronologically ordered list of notable changes for each version of a project."*
  https://keepachangelog.com/

### REQ-VER-003 — Conventional Commits

- **Conventional Commits 1.0.0**: *"A specification for adding human and machine readable meaning to commit messages."* Enables automated changelog generation and semantic versioning.
  https://www.conventionalcommits.org/

### REQ-VER-004 — Annotated git tags

- **Git Documentation — Tagging**: Annotated tags store tagger name, date, and message; recommended for releases over lightweight tags.
  https://git-scm.com/book/en/v2/Git-Basics-Tagging

---

## Deployment

### REQ-DEPLOY-001 — Zero-downtime deployments

- **Kubernetes — Rolling Update Strategy**: The default deployment strategy ensures old and new pods coexist.
  https://kubernetes.io/docs/concepts/workloads/controllers/deployment/#rolling-update-deployment
- **Martin Fowler — Blue Green Deployment**: Pattern for zero-downtime releases.
  https://martinfowler.com/bliki/BlueGreenDeployment.html

### REQ-DEPLOY-002 — Expand-contract migrations

- **Martin Fowler — Parallel Change (Expand and Contract)**: *"Make a change to an interface or schema in three phases: expand, migrate, contract."*
  https://martinfowler.com/bliki/ParallelChange.html
- **Pramod Sadalage & Martin Fowler — Refactoring Databases**: Established the expand-contract pattern for production database evolution.

### REQ-DEPLOY-003 — Reversible, idempotent migrations

- **Martin Fowler — Evolutionary Database Design**: Every change must have both up and down scripts.
  https://martinfowler.com/articles/evodb.html

---

## Accessibility

### REQ-A11Y-001 through REQ-A11Y-010 — WCAG 2.2 Level AA

All accessibility requirements trace directly to the **W3C Web Content Accessibility Guidelines (WCAG) 2.2**, a W3C Recommendation:
https://www.w3.org/TR/WCAG22/

Specific success criteria:

| Requirement | WCAG Success Criterion |
|---|---|
| REQ-A11Y-001 | Conformance Level AA (all Level A + AA criteria) |
| REQ-A11Y-002 | SC 2.1.1 Keyboard (A), SC 2.4.7 Focus Visible (AA) |
| REQ-A11Y-003 | SC 1.1.1 Non-text Content (A) |
| REQ-A11Y-004 | SC 1.4.1 Use of Color (A), SC 1.4.3 Contrast Minimum (AA), SC 1.4.11 Non-text Contrast (AA) |
| REQ-A11Y-005 | Supporting automated verification of the above |
| REQ-A11Y-006 | SC 1.3.1 Info and Relationships (A), SC 3.3.1 Error Identification (A), SC 3.3.2 Labels or Instructions (A) |
| REQ-A11Y-007 | SC 2.3.1 Three Flashes or Below (A), SC 2.2.2 Pause Stop Hide (A) |
| REQ-A11Y-008 | SC 1.4.4 Resize Text (AA) |
| REQ-A11Y-009 | SC 4.1.3 Status Messages (AA) |
| REQ-A11Y-010 | SC 2.4.2 Page Titled (A), SC 3.2.3 Consistent Navigation (AA), SC 2.4.5 Multiple Ways (AA) |

**Contrast ratio derivation**: The 4.5:1 ratio for normal text compensates for vision loss equivalent to 20/40 acuity. It derives from ISO 9241-3 (3:1 baseline for normal vision) multiplied by the ~1.5 contrast sensitivity loss typical at age 80.
- **W3C Understanding SC 1.4.3**: https://www.w3.org/WAI/WCAG21/Understanding/contrast-minimum.html

**WHO disability statistics**: Over 1 billion people (16% of the global population) experience significant disability. Over 2.2 billion have a vision impairment.
- https://www.who.int/health-topics/disability

---

## Documentation

### REQ-DOC-001 — OpenAPI documentation

- **OpenAPI Initiative**: The industry-standard specification for describing RESTful APIs, supported by all major API tooling.
  https://www.openapis.org/
- **SmartBear State of API Report**: OpenAPI/Swagger is the most widely used API documentation format.

### REQ-DOC-002 — README standard

- **GitHub — About READMEs**: *"A README is often the first item a visitor will see when visiting your repository. It should tell people why this project is useful, what they can do with the project, and how they can use it."*
  https://docs.github.com/en/repositories/managing-your-repositorys-settings-and-features/customizing-your-repository/about-readmes

### REQ-DOC-004 — Configuration documentation

- **The Twelve-Factor App, §III — Config**: Environment configuration must be documented since env vars lack self-describing types or defaults.
  https://12factor.net/config

---

## Frontend & UI

### REQ-UI-001 — Mobile-first design

- **StatCounter GlobalStats (2024)**: Mobile accounts for ~60% of global web traffic.
  https://gs.statcounter.com/platform-comparison-chart
- **Luke Wroblewski — Mobile First (2011)**: The foundational text on mobile-first design strategy, advocating that starting with mobile constraints leads to better design decisions.
  https://www.lukew.com/ff/entry.asp?933
- **Google — Mobile-first indexing**: Google indexes the mobile version of sites first.
  https://developers.google.com/search/docs/crawling-indexing/mobile/mobile-sites-mobile-first-indexing

### REQ-UI-002 — Responsive design

- **Nielsen Norman Group — Responsive Design Breakpoints**: Content-driven breakpoints produce better results than device-specific ones. Accommodate 2-3 breakpoints.
  https://www.nngroup.com/articles/breakpoints-in-responsive-design/
- **Ethan Marcotte — Responsive Web Design (2010)**: The original article that coined the term and established the practice.
  https://alistapart.com/article/responsive-web-design/

### REQ-UI-003 — Semantic HTML

- **MDN Web Docs — HTML and Accessibility**: *"A great deal of web content can be made accessible just by making sure the correct HTML elements are used for the correct purpose at all times."* Native `<button>` provides keyboard accessibility, screen reader recognition, and focus management that `<div onClick>` does not.
  https://developer.mozilla.org/en-US/docs/Learn_web_development/Core/Accessibility/HTML
- **W3C — Using Semantic HTML**: Semantic elements create the document structure that assistive technologies navigate. Screen readers use headings and landmarks (`<nav>`, `<main>`) as navigation affordances.
  https://www.w3.org/TR/WCAG22/

### REQ-UI-004 — Image optimization

- **web.dev — Optimize images**: Lazy loading, responsive images via `srcset`, and modern formats significantly improve LCP.
  https://web.dev/articles/optimize-images
- **HTTP Archive (2024)**: Images account for ~50% of page weight on median sites. WebP is 25-35% smaller than JPEG at equivalent quality.
  https://httparchive.org/

### REQ-UI-005 — Font loading strategy

- **web.dev — Best practices for fonts**: *"`font-display: swap` tells the browser that text using this font should be displayed immediately using a system font, and the custom font should be swapped in when ready."*
  https://web.dev/articles/font-best-practices
- **CSS Fonts Module Level 4 — font-display**: W3C specification for controlling font loading behavior.
  https://drafts.csswg.org/css-fonts/#font-display-desc

### REQ-UI-006 — Touch targets (44x44px)

- **WCAG 2.2 SC 2.5.8 Target Size (Minimum)**: Interactive targets must be at least 24x24 CSS pixels (Level AA), with 44x44px recommended. Apple HIG and Material Design both specify 44pt/48dp minimum.
  https://www.w3.org/TR/WCAG22/#target-size-minimum
- **Apple Human Interface Guidelines**: 44x44pt minimum touch target.
  https://developer.apple.com/design/human-interface-guidelines/accessibility
- **Material Design — Touch target size**: 48x48dp minimum recommended.
  https://m3.material.io/foundations/interaction/states

### REQ-UI-007 — Loading, empty, and error states

- **Nielsen Norman Group — Skeleton Screens 101**: *"Skeleton screens reduce the perception of a long loading time by providing clues for how the page will ultimately look."* Perceived wait is 11-15% shorter with feedback. Skeleton screens outperform spinners for full-page loads.
  https://www.nngroup.com/articles/skeleton-screens/
- **Viget Research (2017) — "The effect of skeleton screens"**: Users perceived skeleton-loaded pages as loading faster than spinner-loaded pages at identical speeds.
  https://www.researchgate.net/publication/326858669

### REQ-UI-008 — Dark/light theme

- **W3C Media Queries Level 5 — `prefers-color-scheme`**: Standard media feature for detecting user OS-level preference.
  https://drafts.csswg.org/mediaqueries-5/#prefers-color-scheme
- **Android/iOS System Defaults**: Both platforms ship with system-wide dark mode, and user adoption is substantial — surveys consistently show 70-80% of users enable dark mode.
- **WCAG**: Both themes must meet AA contrast ratios, requiring explicit design for both schemes.

---

## Performance & Caching

### REQ-PERF-001 — Response-time targets

- **Google SRE Book, Ch. 4 — Service Level Objectives**: SLOs should define latency targets (e.g., p99 < Xms). Measure and alert on breaches.
  https://sre.google/sre-book/service-level-objectives/
- **Jakob Nielsen — Response Times: The 3 Important Limits**: 0.1s (instant), 1s (flow maintained), 10s (attention lost). Published 1993, repeatedly validated.
  https://www.nngroup.com/articles/response-times-3-important-limits/

### REQ-PERF-002 — Core Web Vitals

- **Google web.dev — Web Vitals**: Defines LCP < 2.5s, CLS < 0.1, INP < 200ms at the 75th percentile. *"Optimizing for quality of user experience is key to the long-term success of any site on the web."*
  https://web.dev/articles/vitals
- **Google Search — Page Experience**: Core Web Vitals are a ranking signal in Google Search.
  https://developers.google.com/search/docs/appearance/core-web-vitals

### REQ-PERF-003 — HTTP cache headers

- **RFC 9111 — HTTP Caching**: Defines `Cache-Control`, `ETag`, conditional requests, and content-addressed caching.
  https://www.rfc-editor.org/rfc/rfc9111
- **web.dev — HTTP caching**: Best practices for cache headers on static vs dynamic content.
  https://web.dev/articles/http-cache

### REQ-PERF-004 — N+1 query prevention

- **Martin Fowler — Object-Relational Mapping**: The N+1 problem is a well-documented antipattern where retrieving N records causes N+1 database queries instead of 1-2.
- **Django Documentation — select_related / prefetch_related**: Framework-level evidence that N+1 prevention is a standard concern.
  https://docs.djangoproject.com/en/stable/ref/models/querysets/#select-related

### REQ-PERF-005 — Caching strategy

- **AWS Well-Architected Framework, Performance Pillar**: *"Use caching to reduce redundant data retrieval, computation, and processing."*
  https://docs.aws.amazon.com/wellarchitected/latest/performance-efficiency-pillar/caching.html

---

## Monitoring & Alerting

### REQ-MON-001 — RED metrics

- **Tom Wilkie — The RED Method (2015)**: Monitor **R**ate, **E**rrors, and **D**uration for every service. *"The RED Method is a good proxy to how happy your customers will be."*
  https://grafana.com/blog/the-red-method-how-to-instrument-your-services/
- **Google SRE Book, Ch. 6 — The Four Golden Signals**: Latency, traffic, errors, and saturation.
  https://sre.google/sre-book/monitoring-distributed-systems/

### REQ-MON-002 — Dashboards as code

- **Grafana — Grafana as Code**: Store dashboard definitions as code (JSON, Terraform) for version control and reproducibility.
  https://grafana.com/blog/2022/12/06/a-complete-guide-to-managing-grafana-as-code-tools-tips-and-tricks/

### REQ-MON-003 — Alerting with severity

- **Google SRE Book, Ch. 11 — Being On-Call**: Alerts must be actionable. Alert fatigue is an anti-pattern.
  https://sre.google/sre-book/being-on-call/
- **PagerDuty Incident Response Guide**: Severity-based alerting ensures critical issues get immediate attention.
  https://response.pagerduty.com/

---

## Resilience

### REQ-RES-001 — Explicit timeouts

- **Google SRE Book, Ch. 22 — Addressing Cascading Failures**: *"Setting either no deadline or an extremely high deadline may cause short-term problems that have long since passed to continue to consume server resources."*
  https://sre.google/sre-book/addressing-cascading-failures/
- **Michael Nygard — Release It! (2007, 2nd ed. 2018)**: Timeouts are the first stability pattern. Every outbound call must have one.

### REQ-RES-002 — Exponential backoff with jitter

- **AWS Architecture Blog — Exponential Backoff and Jitter**: Demonstrates that "Full Jitter" results in better performance, with approximately constant call rate beyond the initial spike. Most AWS SDKs implement this natively.
  https://aws.amazon.com/blogs/architecture/exponential-backoff-and-jitter/
- **Google SRE Book, Ch. 22**: *"Always use randomized exponential backoff when scheduling retries... If retries aren't randomly distributed, a small perturbation can cause retry ripples."*
  https://sre.google/sre-book/addressing-cascading-failures/

### REQ-RES-003 — Circuit breakers

- **Martin Fowler — Circuit Breaker (2014)**: *"Michael Nygard popularized the Circuit Breaker pattern to prevent this kind of catastrophic cascade."* Three states: Closed → Open → Half-Open. *"Circuit breakers are a valuable place for monitoring."*
  https://martinfowler.com/bliki/CircuitBreaker.html
- **Michael Nygard — Release It!, Ch. 5**: Original exposition of the circuit breaker as a stability pattern.

### REQ-RES-004 — Graceful degradation

- **Google SRE Book, Ch. 22**: *"Graceful degradation takes the concept of load shedding one step further by reducing the amount of work that needs to be performed."* Example: serving cached stale data or using a faster but less accurate algorithm.
  https://sre.google/sre-book/addressing-cascading-failures/
- **AWS Builders' Library — Avoiding fallback in distributed systems**
  https://aws.amazon.com/builders-library/avoiding-fallback-in-distributed-systems/

### REQ-RES-005 — Backpressure & bounded operations

- **Google SRE Book, Ch. 22**: *"Load shedding drops some proportion of load... per-task throttling based on CPU, memory, or queue length."*
  https://sre.google/sre-book/addressing-cascading-failures/
- **Reactive Manifesto — Back-Pressure**: Systems must signal upstream when they are under pressure rather than failing silently.
  https://www.reactivemanifesto.org/glossary#Back-Pressure

---

## Internationalization

### REQ-I18N-001 — Externalized strings

- **W3C Internationalization — String Externalization**: All user-visible text must be in resource files for translation.
  https://www.w3.org/International/
- **GitLab Development Guidelines — Externalization**: *"Non-hardcoded strings can be easily localized... If you see any un-bracketed source text during pseudo-localization testing, those strings were hardcoded."*
  https://docs.gitlab.com/development/i18n/externalization/

### REQ-I18N-002 — Locale-aware formatting

- **ECMA-402 — ECMAScript Internationalization API**: The `Intl` API provides locale-sensitive formatting for dates, numbers, currency, and plurals. *"By adopting Intl, you reduce dependencies, shrink bundle sizes, and improve performance."*
  https://tc39.es/ecma402/
- **W3C Guide to the ECMAScript Internationalization API**
  https://w3c.github.io/i18n-drafts/articles/intl/index.en.html

### REQ-I18N-003 — Text expansion accommodation

- **IBM Globalization Guidelines**: Translated text typically expands 30-50% from English. Layouts must not break.
  https://www.ibm.com/docs/en/i/7.5?topic=design-text-expansion
- **W3C Internationalization — String Length**: Short English strings can expand 200-300% in some languages.
  https://www.w3.org/International/articles/article-text-size

### REQ-I18N-004 — RTL support

- **W3C Internationalization — Inline markup and bidirectional text**: Use CSS logical properties and the `dir` attribute.
  https://www.w3.org/International/articles/inline-bidi-markup/
- **MDN — CSS Logical Properties**: `margin-inline-start` instead of `margin-left` for proper bidi support.
  https://developer.mozilla.org/en-US/docs/Web/CSS/CSS_logical_properties_and_values

### REQ-I18N-005 — ICU MessageFormat

- **Unicode ICU — Message Format**: The standard for handling plurals, gender, and select across languages. Naive concatenation breaks in languages with different word order.
  https://unicode-org.github.io/icu/userguide/format_parse/messages/
- **W3C Internationalization — Compound Messages**: *"Do not compose sentences from multiple separately translated strings."*
  https://www.w3.org/International/

---

## Privacy & Compliance

### REQ-PRIV-001 — Consent management

- **GDPR Article 7 — Conditions for Consent**: Consent must be freely given, specific, informed, and unambiguous. Must be as easy to withdraw as to give.
  https://gdpr-info.eu/art-7-gdpr/
- **GDPR Recital 32**: *"Consent should be given by a clear affirmative act."*

### REQ-PRIV-002 — Right to erasure

- **GDPR Article 17 — Right to Erasure**: Data subjects can request deletion when data is no longer necessary, consent is withdrawn, or processing was unlawful. Controllers must comply "without undue delay."
  https://gdpr-info.eu/art-17-gdpr/

### REQ-PRIV-003 — Data portability

- **GDPR Article 20 — Right to Data Portability**: *"The data subject shall have the right to receive the personal data... in a structured, commonly used and machine-readable format."*
  https://gdpr-info.eu/art-20-gdpr/

### REQ-PRIV-004 — Cookie compliance

- **ePrivacy Directive, Article 5(3)**: Storing cookies requires informed consent, except for strictly necessary cookies.
  https://eur-lex.europa.eu/legal-content/EN/TXT/?uri=CELEX:32002L0058
- **GDPR Recital 30**: Cookies that identify users qualify as personal data and are subject to GDPR.
  https://gdpr-info.eu/recitals/no-30/

### REQ-PRIV-005 — Privacy policy route

- **GDPR Articles 13 & 14 — Information to be provided**: Data subjects must be informed of processing details at the point of collection.
  https://gdpr-info.eu/art-13-gdpr/

### REQ-PRIV-006 — Third-party data sharing controls

- **GDPR Article 13(1)(e)**: Recipients or categories of recipients of personal data must be disclosed.
  https://gdpr-info.eu/art-13-gdpr/
- **ePrivacy Directive**: Non-essential third-party tracking requires explicit consent.

---

## Authorization

### REQ-AUTHZ-001 — RBAC/ABAC model

- **NIST SP 800-162 — Guide to ABAC Definition and Considerations**: Defines attribute-based access control as a flexible authorization model.
  https://csrc.nist.gov/publications/detail/sp/800-162/final
- **OWASP Top 10 Proactive Controls 2024, C1 — Implement Access Control**: The #1 proactive control. *"Broken Access Control tops the list."*
  https://top10proactive.owasp.org/the-top-10/c1-accesscontrol/

### REQ-AUTHZ-002 — Least privilege

- **NIST SP 800-53, AC-6 — Least Privilege**: *"Employ the principle of least privilege, allowing only authorized accesses for users."*
  https://csrc.nist.gov/publications/detail/sp/800-53/rev-5/final
- **OWASP ASVS v4.0, §V4.1**: Default-deny; access requires explicit grant.

### REQ-AUTHZ-003 — Server-side enforcement

- **OWASP Top 10:2021, A01 — Broken Access Control**: *"Access control is only effective in trusted server-side code... where the attacker cannot modify the access control check."*
  https://owasp.org/Top10/A01_2021-Broken_Access_Control/

### REQ-AUTHZ-004 — Admin audit logging

- **SOC 2, CC6.1**: Administrative actions must be logged and reviewable.
- **OWASP ASVS v4.0, §V7**: Privileged actions must produce audit records.

### REQ-AUTHZ-005 — Object-level authorization (IDOR prevention)

- **OWASP API Security Top 10:2023, API1 — Broken Object Level Authorization**: *"Attackers can exploit API endpoints that are vulnerable to broken object-level authorization by manipulating the ID of an object."* Mapped to CWE-285 (Improper Authorization) and CWE-639 (Authorization Bypass Through User-Controlled Key).
  https://owasp.org/API-Security/editions/2023/en/0xa1-broken-object-level-authorization/

---

## Source Index

| Source | Type | URL |
|---|---|---|
| OWASP Top 10 (2021) | Industry Standard | https://owasp.org/Top10/ |
| OWASP ASVS v4.0 | Verification Standard | https://owasp.org/www-project-application-security-verification-standard/ |
| OWASP API Security Top 10 (2023) | Industry Standard | https://owasp.org/API-Security/ |
| OWASP Proactive Controls (2024) | Best Practices | https://top10proactive.owasp.org/ |
| OWASP Cheat Sheet Series | Best Practices | https://cheatsheetseries.owasp.org/ |
| NIST SP 800-53 Rev. 5 | Government Standard | https://csrc.nist.gov/publications/detail/sp/800-53/rev-5/final |
| NIST SP 800-190 | Government Standard | https://csrc.nist.gov/publications/detail/sp/800-190/final |
| NIST SP 800-218 (SSDF) | Government Standard | https://csrc.nist.gov/publications/detail/sp/800-218/final |
| CIS Docker Benchmark v1.6.0 | Industry Benchmark | https://www.cisecurity.org/benchmark/docker |
| GDPR | Regulation | https://gdpr-info.eu/ |
| WCAG 2.2 | W3C Recommendation | https://www.w3.org/TR/WCAG22/ |
| The Twelve-Factor App | Methodology | https://12factor.net/ |
| Google SRE Book | Industry Practice | https://sre.google/sre-book/table-of-contents/ |
| AWS Well-Architected Framework | Cloud Best Practice | https://aws.amazon.com/architecture/well-architected/ |
| Martin Fowler (various articles) | Thought Leadership | https://martinfowler.com/ |
| Michael Nygard — Release It! | Book | ISBN 978-1680502398 |
| RFC 9457 (Problem Details) | IETF Standard | https://www.rfc-editor.org/rfc/rfc9457.html |
| RFC 9110 (HTTP Semantics) | IETF Standard | https://www.rfc-editor.org/rfc/rfc9110 |
| RFC 9111 (HTTP Caching) | IETF Standard | https://www.rfc-editor.org/rfc/rfc9111 |
| Semantic Versioning 2.0.0 | Specification | https://semver.org/ |
| Conventional Commits 1.0.0 | Specification | https://www.conventionalcommits.org/ |
| Keep a Changelog 1.1.0 | Specification | https://keepachangelog.com/ |
| OCI Image Spec | Specification | https://github.com/opencontainers/image-spec |
| Docker Best Practices | Vendor Docs | https://docs.docker.com/build/building/best-practices/ |
| Nielsen Norman Group | UX Research | https://www.nngroup.com/ |
| web.dev (Google) | Web Best Practices | https://web.dev/ |
| MDN Web Docs | Reference | https://developer.mozilla.org/ |
| Unicode ICU | Standard Library | https://unicode-org.github.io/icu/ |
| W3C Internationalization | Standards | https://www.w3.org/International/ |
