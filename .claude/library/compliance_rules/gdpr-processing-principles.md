---
description: GDPR processing principles — lawfulness, purpose limitation, data minimization, consent, special categories, data protection by design/default, international transfers
languages:
- python
- javascript
- typescript
- go
- rust
- java
alwaysApply: false
---

# GDPR Processing Principles

Covers: GDPR Articles 5, 6, 7, 9, 25, 30, 32, 35, 44-49 | EDPB Guidelines 05/2020 (consent), 4/2019 (DPbDD)

## Article 5 — Principles

### 5(1)(b) Purpose Limitation
- Data collected for purpose A is not reused for purpose B without compatibility check or fresh consent
- APIs do not expose data beyond the purpose for which it was collected

### 5(1)(c) Data Minimization
- Forms and APIs do not collect fields not strictly necessary for the stated purpose
- API responses do not include unnecessary personal data fields — no `SELECT *` on sensitive tables
- Log statements do not include personal data unless operationally necessary
- Error messages do not leak personal data

### 5(1)(d) Accuracy
- Users can update their own data through the UI
- Data from external sources has validation/verification steps

### 5(1)(e) Storage Limitation
- Retention periods defined per data category
- Automated deletion/anonymization jobs exist and are tested
- TTL set on caches and temporary stores containing personal data
- Database migrations do not extend retention beyond defined periods

### 5(1)(f) Integrity and Confidentiality
- Encryption at rest for personal data stores
- Encryption in transit (TLS) for all personal data transfers
- Access controls (RBAC/ABAC) on personal data endpoints
- Audit logging on personal data access
- No personal data in application logs without masking/redaction

### 5(2) Accountability
- Processing activities documented (code, architecture docs, or ROPA tooling)
- Data flow diagrams maintained alongside code
- Compliance decisions recorded (why a legal basis was chosen, why a field is collected)

## Article 6 — Lawfulness of Processing

- Every processing activity has a documented legal basis (consent, contract, legitimate interest, legal obligation, vital interests, public task)
- Consent-based processing (6(1)(a)) is gated on a recorded, valid consent signal — processing does not proceed without consent verification
- Contract-based processing (6(1)(b)) is tied to active contract/subscription status
- Legitimate interest processing (6(1)(f)) has documented balancing test and offers opt-out
- Legal basis is checked before processing begins, not after
- When legal basis is removed (e.g., consent withdrawn), processing stops immediately

## Article 7 — Consent (+ EDPB Guidelines 05/2020)

- **No pre-ticked boxes** — consent requires clear affirmative act (Recital 32)
- **Granular consent** — separate checkboxes/toggles for separate purposes, no bundled consent
- **Consent record** includes: timestamp, policy version shown, specific purposes, method of consent
- **Withdrawal as easy as giving consent** — same UI, same number of clicks, equally prominent
- Consent withdrawal triggers actual cessation of processing — data flows gated on consent status
- Consent is not a precondition for service access unless processing is strictly necessary (no cookie walls blocking essential functionality)
- Consent records stored immutably for audit
- Re-consent requested when purposes change
- **No dark patterns** — equal prominence for accept/reject, no manipulative language
- Scrolling or continued browsing does not constitute consent

## Article 9 — Special Categories

Processing PROHIBITED for: racial/ethnic origin, political opinions, religious/philosophical beliefs, trade union membership, genetic data, biometric data (for identification), health data, sex life/orientation.

- Database schemas reviewed for special category data fields — flagged with enhanced protections
- **Explicit consent** required (not just regular consent) — consent mechanism specifically names the special category data types
- Stricter access controls than regular personal data
- Separate or additional encryption (e.g., field-level encryption, separate keys)
- Comprehensive access logging
- API responses and data exports explicitly exclude special category data unless endpoint is designed for it
- No inferring special category data from regular data without same protections

## Article 25 — Data Protection by Design and by Default (+ EDPB Guidelines 4/2019)

### By Design
- Pseudonymization applied where feasible (separate identity from data attributes)
- Data minimization built into APIs and database schemas from the start
- New features processing personal data have documented privacy considerations
- Privacy is a design requirement, not a post-hoc addition

### By Default
- Default settings are most privacy-protective (opt-in, not opt-out for data sharing)
- Profile visibility defaults to private/restricted, not public
- Marketing communications off by default
- Analytics/tracking off by default or behind consent
- New features do not expand data collection without explicit user action
- Minimum amount of personal data processed for each function

## Article 30 — Records of Processing Activities

- Each data processing pipeline/endpoint has metadata: purpose, data categories, retention period, recipients
- New endpoints/services processing personal data include ROPA updates
- Schema migrations adding personal data fields trigger documentation updates
- Data flow documentation maintained alongside code

## Article 32 — Security of Processing

- Personal data at rest encrypted (database-level, disk-level, or field-level)
- Personal data in transit uses TLS 1.2+
- Pseudonymization used where full identification not required (analytics, testing, development)
- Test/staging environments do not contain real personal data — use anonymized/synthetic data
- Authentication and authorization on all personal data endpoints
- Backup and recovery procedures exist and are tested for personal data systems
- Security testing in CI/CD pipeline

## Article 35 — Data Protection Impact Assessment (DPIA)

- New features involving profiling, automated decision-making, special category data, or large-scale processing have documented DPIA
- DPIA referenced in code comments or architecture decision records
- Risk mitigation measures from DPIA are implemented in code
- ML models and algorithmic scoring systems have associated DPIAs
- Systematic monitoring features (location tracking, behavior analysis) have DPIAs

## Articles 44-49 — International Transfers

- Data storage location documented per service/database (which region, which country)
- Cloud infrastructure configuration enforces data residency (AWS region, Azure geography, GCP location)
- Third-party services documented with their data processing locations
- CDN and caching configurations do not replicate personal data to non-adequate countries without safeguards
- Database replication does not cross borders without documentation and legal basis
- Sub-processors and their locations documented
- Transfer mechanisms (SCCs, adequacy, BCRs) documented per data flow
- Backup storage locations comply with data residency requirements
- Log aggregation services configured for EU data residency when processing EU personal data

---

For authoritative sources, verification record, and direct links to official texts, see [`SOURCES.md`](SOURCES.md).
