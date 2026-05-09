---
description: GDPR data subject rights — access, rectification, erasure, portability, transparency, breach notification
languages:
- python
- javascript
- typescript
- go
- rust
- java
alwaysApply: false
---

# GDPR Data Subject Rights & Breach Notification

Covers: GDPR Articles 12-17, 20, 33, 34 | EDPB Guidelines 9/2022 (breach notification)

## Articles 12-14 — Transparency

- Privacy notice displayed at every data collection point (registration, checkout, contact forms, newsletter)
- Privacy notice versioned — version shown recorded alongside consent
- Data collection forms link to or display the privacy notice before submission
- Privacy dashboard or settings page exists where users can view and manage their data
- Automated decision-making (profiling, scoring) disclosed with "meaningful information about the logic involved" (Art. 13(2)(f))
- Third-party data sources documented and disclosed to data subjects (Art. 14)

## Article 15 — Right of Access

- Data export endpoint or admin function retrieves all personal data for a given data subject
- Export includes data from ALL systems/databases (not just the primary user table)
- Export format is machine-readable (JSON, CSV, XML)
- Identity verification performed before serving access requests (Recital 64)
- Response includes metadata: purposes, categories, recipients, retention periods
- System can identify all personal data associated with a user across all storage locations (databases, caches, third-party services)
- Response within one month (extendable by two months for complex requests, with notification to the subject)

## Article 16 — Right to Rectification

- Users can edit their own profile/personal data through the UI
- Admin/support function exists to rectify data on behalf of users
- Rectification propagates to all systems holding copies (caches, replicas, third-party integrations)
- Rectification events logged for audit
- Data validation rules do not prevent legitimate corrections (e.g., allowing name changes)

## Article 17 — Right to Erasure ("Right to Be Forgotten")

- User deletion/erasure function is comprehensive (not just soft-delete of main record)
- Erasure covers ALL storage:
  - Primary database
  - Caches (Redis, Memcached)
  - Search indexes (Elasticsearch)
  - Analytics stores
  - Log files (where feasible; document approach for logs)
  - Backups (documented process for backup erasure or exclusion on restore)
  - Third-party services (via API calls or documented manual process)
  - File storage (uploads, avatars)
- Soft-delete must eventually lead to hard-delete within defined timeframe
- Cascading deletion handles foreign key relationships properly
- Erasure of shared data triggers notification to third parties who received it (Art. 17(2))
- Exceptions implemented: data needed for legal claims, legal obligations, or public health is retained with documentation
- Anonymization as alternative to deletion is truly irreversible
- Erasure requests logged (without storing the deleted data itself)
- Automated tests verify erasure is complete across all storage systems

## Article 20 — Right to Data Portability

- Data export endpoint produces structured, machine-readable output (JSON, CSV, XML)
- Export includes all data the subject provided (not derived/inferred data)
- Export format uses common standards where available (vCard for contacts, iCal for calendar, etc.)
- Export downloadable by the authenticated user
- Where technically feasible, API-to-API transfer mechanism exists
- Export does not include other users' personal data
- Portability applies only to consent-based and contract-based processing

## Article 33 — Breach Notification to Supervisory Authority (+ EDPB Guidelines 9/2022)

72-hour notification requirement after becoming aware of breach.

- Security incident detection mechanisms exist (monitoring, anomaly detection, log analysis)
- Breach detection triggers automated alerts to security/DPO team
- Breach logging captures: timestamp of detection, nature of breach, affected data categories, number of records affected, remedial actions
- Processor-to-controller notification channels implemented (if application acts as processor)
- All breaches documented, even those not requiring notification
- The 72-hour clock supported by automated timestamps in incident management
- "Awareness" starts when controller has "reasonable degree of certainty" a breach occurred

## Article 34 — Breach Communication to Data Subject

Required when breach poses "high risk to rights and freedoms."

- Mechanism exists to send breach notifications to affected users (email, in-app, SMS)
- Notification templates use clear, plain language (not legalese)
- System can identify which users are affected by a specific breach (mapping breached data to accounts)
- Encryption status trackable per record/field (Art. 34(3)(a) exemption: if data was encrypted, notification may not be required)
- Public communication channel for large-scale breaches (status page, website banner)

## Cross-Cutting Code Review Checklist

Apply to every PR touching personal data:

| Category | Check |
|----------|-------|
| **Collection** | Only necessary fields? Purpose documented? Legal basis identified? Privacy notice displayed? |
| **Consent** | Affirmative action? Granular? No pre-ticked? Withdrawal equal to consent? Recorded with metadata? |
| **Storage** | Encrypted at rest? Retention defined? Auto-deletion? No PII in logs? |
| **Access Control** | Auth required? Authz enforced? Least privilege? Audit logged? |
| **Transit** | TLS 1.2+? No plain HTTP for PII? |
| **Subject Rights** | Export exists? Deletion comprehensive? Rectification propagates? Portability format machine-readable? |
| **Special Categories** | Identified? Enhanced protections? Explicit consent? Separate encryption? |
| **International Transfers** | Data residency documented? Regions configured? Transfer mechanism in place? |
| **Breach Readiness** | Monitoring in place? Alerting automated? Affected user identification possible? |
| **By Design/Default** | Defaults most private? Pseudonymization used? Privacy docs for new features? |

---

For authoritative sources, verification record, and direct links to official texts, see [`SOURCES.md`](SOURCES.md).
