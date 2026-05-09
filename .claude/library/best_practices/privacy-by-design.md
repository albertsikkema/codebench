# Privacy by Design

## Principle

Build privacy into the system architecture from the start, not as a compliance checkbox after launch. Collect only what you need, protect what you collect, give users control over their data, and make the system capable of forgetting. GDPR compliance is the floor, not the ceiling.

## Why

- **Retrofitting privacy is expensive**: Adding consent management, data deletion flows, and audit trails to an existing system is an order of magnitude harder than building them in. Data scattered across tables, caches, logs, backups, and third-party services creates a deletion nightmare.
- **Privacy breaches destroy user trust**: Trust is the hardest thing to rebuild. Users who learn their data was mishandled leave and don't come back. On top of that, GDPR fines can reach 4% of annual global turnover.
- **Users expect control**: "Delete my account," "export my data," and "stop tracking me" are features, not edge cases. Systems that can't do these things feel broken to users.

## Core Rules

### 1. Data Minimization

Collect only the data that is functionally necessary. For every field you store, answer: "What breaks if we don't have this?"

```
BAD:  Collect date of birth, gender, phone number "in case we need it later"
GOOD: Collect email (required for auth) and name (required for display). Nothing else.
```

**Rules**:
- Don't collect data "just in case" — collect it when you have a use case
- Don't store data you can derive (e.g. age from date of birth — compute at query time if needed)
- Don't retain data beyond its purpose (session data after logout, verification codes after use)
- Anonymize or pseudonymize where full identity isn't needed (analytics, A/B testing)

### 2. Consent Management

Before collecting personal data, obtain explicit, informed consent. Record what was consented to, when, by whom, and under which version of the policy.

**Requirements**:
- Present a clear consent UI *before* collecting data
- Support granular consent (analytics vs marketing vs functional, separately)
- Make consent revocable via user settings at any time
- Re-request consent when the privacy policy changes materially
- Never pre-check consent boxes

**What to record per consent**:

| Field | Example |
|-------|---------|
| User ID | `user_abc123` |
| What was consented to | `analytics`, `marketing_emails` |
| When | `2026-03-28T10:30:00Z` |
| Policy version | `privacy-policy-v2.3` |
| How (channel) | `web_signup_form` |
| IP address | `192.168.1.1` (for proof of consent) |

### 3. Right to Erasure

Users must be able to delete their account and personal data. The system must be able to fulfill this completely.

**Implementation**:
1. **Soft delete** with a grace period (7–30 days for account recovery)
2. **Hard delete** after the grace period — remove from all storage:
   - User profile and account data
   - User-generated content
   - Activity history and logs containing PII
   - Third-party systems notified of the data
   - Backups (within retention window)
3. **Log the erasure event** — but without the deleted PII. Log that "user X's data was deleted at time Y," not what was deleted.

**Fulfillment timeline**: Within 30 days (GDPR requirement).

**What to keep after deletion**:
- Anonymized aggregated data (can't be traced back to the user)
- Audit trail of the deletion itself
- Legal-hold data (if applicable, documented separately)

### 4. Data Portability

Users must be able to export their personal data in a portable, machine-readable format.

**Export must include**:
- Profile data (name, email, settings)
- User-generated content (posts, comments, files)
- Activity history (orders, interactions)

**Export must exclude**:
- Derived data (recommendations, scores)
- Internal identifiers (database IDs, foreign keys)
- Other users' data

**Format**: JSON (preferred) or CSV. Provide via user settings or API endpoint.

### 5. Cookie Compliance

Categorize cookies and obtain consent before setting non-essential ones.

| Category | Consent required? | Examples |
|----------|------------------|---------|
| **Strictly necessary** | No | Session cookies, CSRF tokens, auth cookies, user-requested preferences (language, theme, locale) |
| **Analytics** | Yes | Google Analytics |
| **Marketing** | Yes | Ad tracking, retargeting pixels |

**Note on client-side storage**: The ePrivacy Directive covers all client-side storage (cookies, localStorage, sessionStorage, IndexedDB), not just cookies. User-requested preferences like theme and language are defensibly "strictly necessary" — they directly serve the user's explicit choice, involve no tracking or profiling, and no data leaves the browser. Document this classification decision rather than assuming it silently.

**Implementation**:
- Show a cookie banner with Accept / Reject / Customize options
- Don't set non-essential cookies before consent
- Persist the user's choice for returning visits
- Provide a way to change preferences later (settings or re-showing the banner)

### 6. Third-Party Data Sharing

Inventory every third-party service that receives user data.

| Service | Data shared | Purpose | DPA in place? |
|---------|------------|---------|--------------|
| Stripe | Email, payment info | Payment processing | Yes |
| Sentry | IP, user agent, error context | Error tracking | Yes |
| Analytics | Page views, anonymized user ID | Usage analytics | Yes |

**Requirements**:
- Disclose each third party in the privacy policy
- Ensure Data Processing Agreements (DPAs) are in place
- Provide user controls to opt out of non-essential sharing
- Respect opt-out in code — don't just hide the UI

### 7. Privacy Policy

Provide a dedicated route/page for the privacy policy, linked from:
- Site footer (every page)
- Registration/signup forms
- Consent dialogs

**Implementation**: Render the policy from a structured source (CMS, markdown file, config) so it can be updated without a code deployment.

## Data Lifecycle

Design for the full lifecycle of personal data:

```
Collection → Processing → Storage → Access → Portability → Deletion
    ↑            ↑           ↑         ↑          ↑            ↑
 Consent    Minimize     Encrypt    Authorize   Export      Erasure
                        + Audit    + Log                   + Backups
```

Every stage has a privacy concern. Skipping any stage creates a gap.

## Implementation Notes

### Database Design for Erasure

Design your schema so that deleting a user cascades cleanly:

```sql
-- User data in one place, relationships via foreign keys with CASCADE
CREATE TABLE users (
    id UUID PRIMARY KEY,
    email VARCHAR(320) NOT NULL UNIQUE,
    name VARCHAR(255),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE user_content (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Audit trail: reference user by ID, but don't store PII
CREATE TABLE audit_log (
    id UUID PRIMARY KEY,
    actor_id UUID,  -- no FK constraint — user may be deleted
    action VARCHAR(100) NOT NULL,
    timestamp TIMESTAMP NOT NULL DEFAULT NOW()
);
```

**Key pattern**: `ON DELETE CASCADE` on user-owned data. No FK constraint on audit logs (the user may be deleted, but the audit entry must survive).

### Consent Storage

```sql
CREATE TABLE consent_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    consent_type VARCHAR(50) NOT NULL,  -- 'analytics', 'marketing', etc.
    granted BOOLEAN NOT NULL,
    policy_version VARCHAR(20) NOT NULL,
    ip_address INET,
    granted_at TIMESTAMP NOT NULL DEFAULT NOW(),
    revoked_at TIMESTAMP  -- NULL if still active
);
```

### Checking Consent in Code

```python
# Before setting analytics cookies or sending data to analytics
async def track_event(user_id: UUID, event: str):
    if not await consent_service.has_consent(user_id, "analytics"):
        return  # silently skip — user hasn't consented

    await analytics.track(event)
```

```typescript
// Before sharing data with third party (TypeScript)
async function trackEvent(userId: string, event: string): Promise<void> {
  const hasConsent = await consentService.hasConsent(userId, "analytics");
  if (!hasConsent) return; // silently skip — user hasn't consented

  await analytics.track(event);
}
```

```go
// Before sharing data with third party
func (s *Service) SendToAnalytics(ctx context.Context, userID uuid.UUID, event Event) error {
    hasConsent, err := s.consent.HasConsent(ctx, userID, "analytics")
    if err != nil || !hasConsent {
        return nil // skip silently
    }
    return s.analytics.Track(event)
}
```

## When to Bend the Rules

- **Internal tools with no external users**: Data minimization still applies, but consent flows and cookie banners don't. Still support data deletion for departing employees.
- **B2B SaaS**: Consent may be handled at the organization level (via DPA) rather than per-user cookie banners. Data portability and erasure still apply per-user.
- **Anonymous/unauthenticated services**: If you truly collect no personal data (no IP logging, no cookies, no analytics), most rules don't apply. Document this decision explicitly.
- **Legal holds**: Erasure requests may be overridden by legal requirements (tax records, regulatory compliance). Document which data categories have legal holds and for how long.

## References

- **ePrivacy Directive** (Directive 2009/136/EC, consolidated): https://eur-lex.europa.eu/legal-content/EN/TXT/?uri=CELEX%3A02009L0136-20201221 — Article 5(3) covers consent for client-side storage (cookies, localStorage, etc.)
- **GDPR** (Regulation 2016/679): https://eur-lex.europa.eu/legal-content/EN/TXT/?uri=CELEX%3A32016R0679 — the overarching data protection regulation