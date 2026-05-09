---
description: Index of applicable compliance standards and how they map to rule files
alwaysApply: false
---

# Applicable Compliance Standards

## ISO 27001:2022 (Annex A — Technological Controls)

Primary information security management standard. Code-relevant controls are in the A.8.x series:

| Control | Title | Rule File |
|---------|-------|-----------|
| A.8.3 | Information Access Restriction | `auth-boundaries.md` |
| A.8.4 | Access to Source Code | `secure-development.md` |
| A.8.5 | Secure Authentication | `auth-boundaries.md` |
| A.8.7 | Protection Against Malware | `secure-coding.md` |
| A.8.9 | Configuration Management | `configuration-security.md` |
| A.8.10 | Information Deletion | `data-lifecycle.md` |
| A.8.11 | Data Masking | `data-lifecycle.md` |
| A.8.12 | Data Leakage Prevention | `data-lifecycle.md` |
| A.8.15 | Logging | `audit-trail.md` |
| A.8.16 | Monitoring Activities | `audit-trail.md` |
| A.8.24 | Use of Cryptography | `cryptography.md` |
| A.8.25 | Secure Development Lifecycle | `secure-development.md` |
| A.8.26 | Application Security Requirements | `secure-development.md` |
| A.8.27 | Secure System Architecture and Engineering Principles | `secure-coding.md` |
| A.8.28 | Secure Coding | `secure-coding.md` |
| A.8.29 | Security Testing in Development and Acceptance | `secure-development.md` |
| A.8.31 | Separation of Development, Test and Production Environments | `configuration-security.md` |
| A.8.33 | Test Information | `configuration-security.md` |

## NIS2 Directive (2022/2555) — Article 21

EU cybersecurity risk-management measures. Implementing Regulation CIR 2024/2690 provides technical annex.

| Article | Topic | Rule File |
|---------|-------|-----------|
| 21(2)(a) | Risk analysis and information system security policies | `secure-coding.md` |
| 21(2)(b) | Incident handling | `audit-trail.md` |
| 21(2)(c) | Business continuity, backup management, disaster recovery, crisis management | `resilience.md` |
| 21(2)(d) | Supply chain security | `supply-chain.md` |
| 21(2)(e) | Security in network and information systems acquisition, development and maintenance, including vulnerability handling and disclosure | `secure-development.md` |
| 21(2)(f) | Policies and procedures to assess effectiveness of cybersecurity risk-management measures | `audit-trail.md` |
| 21(2)(g) | Basic cyber hygiene practices and cybersecurity training | `cyber-hygiene-training.md` |
| 21(2)(h) | Policies and procedures regarding use of cryptography and, where appropriate, encryption | `cryptography.md` |
| 21(2)(i) | Human resources security, access control policies and asset management | `auth-boundaries.md` |
| 21(2)(j) | Multi-factor authentication, continuous authentication, secured voice/video/text communications and secured emergency communication systems | `auth-boundaries.md` |
| Art. 23 | Reporting obligations | `audit-trail.md` |

## OWASP ASVS 4.0.3

Application Security Verification Standard. L1 and L2 requirements. ASVS 5.0.0 was released May 2025 — evaluate for future updates.

Note: some requirement IDs in ASVS 4.0.3 are marked DELETED (duplicates): V4.1.4 (duplicate of V4.1.3), V7.3.2 (duplicate of V7.3.1), V13.1.2 (duplicate of V4.3.1). These IDs exist in the numbering but have no active requirement.

| Chapter | Topic | Rule File |
|---------|-------|-----------|
| V2 | Authentication | `auth-boundaries.md` |
| V3 | Session Management | `session-cookie.md` |
| V4 | Access Control | `auth-boundaries.md` |
| V5 | Validation, Sanitization, Encoding | `secure-coding.md` |
| V7 | Error Handling and Logging | `audit-trail.md` |
| V8 | Data Protection | `data-lifecycle.md` |
| V9 | Communication Security | `cryptography.md` |
| V12 | Files and Resources | `secure-coding.md` |
| V13 | API and Web Service | `secure-coding.md` |
| V14 | Configuration | `configuration-security.md` |

## GDPR (Regulation 2016/679)

EU General Data Protection Regulation. Applies to all processing of personal data of EU residents.

| Article(s) | Topic | Rule File |
|------------|-------|-----------|
| Art. 5 | Processing principles (lawfulness, minimization, storage limitation, etc.) | `gdpr-processing-principles.md` |
| Art. 6 | Lawfulness of processing (legal bases) | `gdpr-processing-principles.md` |
| Art. 7 | Conditions for consent (+ EDPB Guidelines 05/2020) | `gdpr-processing-principles.md` |
| Art. 9 | Special categories (health, biometric, political, religious) | `gdpr-processing-principles.md` |
| Art. 12-14 | Transparency, information, privacy notices | `gdpr-data-subject-rights.md` |
| Art. 15 | Right of access | `gdpr-data-subject-rights.md` |
| Art. 16 | Right to rectification | `gdpr-data-subject-rights.md` |
| Art. 17 | Right to erasure ("right to be forgotten") | `gdpr-data-subject-rights.md` |
| Art. 20 | Right to data portability | `gdpr-data-subject-rights.md` |
| Art. 25 | Data protection by design and by default (+ EDPB Guidelines 4/2019) | `gdpr-processing-principles.md` |
| Art. 30 | Records of processing activities | `gdpr-processing-principles.md` |
| Art. 32 | Security of processing | `gdpr-processing-principles.md` |
| Art. 33 | Breach notification to supervisory authority (72h) | `gdpr-data-subject-rights.md` |
| Art. 34 | Breach communication to data subject | `gdpr-data-subject-rights.md` |
| Art. 35 | Data protection impact assessment (DPIA) | `gdpr-processing-principles.md` |
| Art. 44-49 | International transfers (data residency, SCCs, adequacy) | `gdpr-processing-principles.md` |

EDPB Guidelines referenced: 05/2020 (consent), 4/2019 (DPbDD), 9/2022 (breach notification).

## DAMA-DMBOK 2nd Edition — Data Quality Dimensions

Data Management Body of Knowledge (DAMA International). Chapter 13 defines 8 data quality dimensions used to ground the PR data integrity agent.

| Dimension | Definition | Used In |
|-----------|-----------|---------|
| Accuracy | Data correctly represents real-world entities | `pr-data-integrity` agent |
| Completeness | All required data is present | `pr-data-integrity` agent |
| Consistency | Same fact represented the same way everywhere | `pr-data-integrity` agent |
| Integrity | Relationships between data elements maintained correctly | `pr-data-integrity` agent |
| Reasonability | Values fall within logical, plausible ranges | `pr-data-integrity` agent |
| Timeliness | Data available when needed, reflects current state | `pr-data-integrity` agent |
| Uniqueness | Each entity represented only once | `pr-data-integrity` agent |
| Validity | Values conform to defined format, type, and domain | `pr-data-integrity` agent |

## Proportionality Note

NIS2 Article 21(1) requires measures to be "appropriate and proportionate" to the entity's size, exposure, and risk. Not every check applies equally to every project. Assess based on:
- Data sensitivity (PII, financial, health)
- Public exposure (internet-facing vs internal)
- User base size
- Regulatory sector (essential vs important entity under NIS2)

---

For authoritative sources, verification record, and direct links to official texts, see [`SOURCES.md`](SOURCES.md).
