---
description: Cyber hygiene and security training requirements — secure practices, awareness, developer training
languages:
- python
- javascript
- typescript
- go
- rust
- java
- c
- cpp
alwaysApply: false
---

# Basic Cyber Hygiene & Cybersecurity Training

Covers: NIS2 Art. 21(2)(g)

## Secure Coding Awareness

- Developers trained on OWASP Top 10 and common vulnerability patterns relevant to their stack
- Security training refreshed periodically — not a one-time onboarding event
- New team members receive security orientation covering project-specific risks and controls

## Code Hygiene Practices

- No credentials, tokens, or secrets in source code, commit messages, or PR descriptions
- Dependencies kept up to date — automated update notifications enabled
- Unused code, dependencies, and services removed rather than left dormant
- Development environments kept patched and secured

## Security Awareness for All Contributors

- Phishing awareness — recognize social engineering targeting developer accounts (npm, PyPI, GitHub)
- Recognise signs of supply chain attacks: typosquatting, dependency confusion, compromised maintainer accounts
- Report suspicious activity through defined channels

## Proportionality Note

NIS2 Art. 21(2)(g) requires "basic cyber hygiene practices and cybersecurity training." The scope and depth of training should be proportionate to the entity's size, risk exposure, and the sensitivity of the systems being developed — per Art. 21(1).

---

For authoritative sources, verification record, and direct links to official texts, see [`SOURCES.md`](SOURCES.md).
