---
description: Authoritative sources used to create and verify the compliance rules in this directory
alwaysApply: false
---

# Sources

This document lists the authoritative sources that the compliance rules in this directory are based on, and the references used to verify their accuracy.

## Primary Standards

### ISO/IEC 27001:2022

The international standard for information security management systems. The Annex A technological controls (A.8.x series) are the basis for security requirements throughout these rule files.

- **Official standard**: ISO/IEC 27001:2022 — Information security, cybersecurity and privacy protection — Information security management systems — Requirements
- **Companion standard**: ISO/IEC 27002:2022 — Information security, cybersecurity and privacy protection — Information security controls (provides implementation guidance for each Annex A control)
- **Publisher**: International Organization for Standardization (ISO)
- **Purchase**: https://www.iso.org/standard/27001

Verification references:
- ICT Institute — ISO 27002:2022 Explained: Technological Controls: https://ictinstitute.nl/iso270022022-explained-technological-controls/
- ISMS.online — ISO 27001:2022 Annex A Explained: https://www.isms.online/iso-27001/annex-a-2022/
- HighTable — ISO 27001 Annex A Controls: Complete 2022 Reference List: https://hightable.io/iso-27001-annex-a-controls-reference-guide/
- DataGuard — Technological Controls ISO 27001: https://www.dataguard.com/iso-27001/annex-a/8-technological-controls/
- Secureframe — ISO 27001 Controls Explained: https://secureframe.com/hub/iso-27001/controls

### NIS2 Directive (EU) 2022/2555

The EU directive on measures for a high common level of cybersecurity across the Union. Article 21 defines cybersecurity risk-management measures; Article 23 defines incident reporting obligations.

- **Official text**: Directive (EU) 2022/2555 of the European Parliament and of the Council of 14 December 2022
- **EUR-Lex**: https://eur-lex.europa.eu/legal-content/EN/TXT/HTML/?uri=CELEX:32022L2555

Implementing regulation:
- **CIR 2024/2690**: Commission Implementing Regulation (EU) 2024/2690 of 17 October 2024 — laying down rules for the application of Directive (EU) 2022/2555 as regards technical and methodological requirements of cybersecurity risk-management measures and further specification of the cases in which an incident is considered to be significant
- **EUR-Lex**: https://eur-lex.europa.eu/legal-content/EN/TXT/HTML/?uri=OJ:L_202402690

Verification references:
- NIS-2-Directive.com — Article 21 reference: https://www.nis-2-directive.com/NIS_2_Directive_Article_21.html
- NIS-2-Directive.com — Article 23 reference: https://www.nis-2-directive.com/NIS_2_Directive_Article_23.html

### GDPR (EU) 2016/679

The EU General Data Protection Regulation. Applies to all processing of personal data of EU residents. Articles 5-7, 9, 12-17, 20, 25, 30, 32-35, and 44-49 are the basis for the privacy and data protection rules in this directory.

- **Official text**: Regulation (EU) 2016/679 of the European Parliament and of the Council of 27 April 2016
- **EUR-Lex**: https://eur-lex.europa.eu/legal-content/EN/TXT/HTML/?uri=CELEX:32016R0679

Individual article sources (gdpr-info.eu provides the full text with linked recitals):
- Art. 5 (Principles): https://gdpr-info.eu/art-5-gdpr/
- Art. 6 (Lawfulness): https://gdpr-info.eu/art-6-gdpr/
- Art. 7 (Consent conditions): https://gdpr-info.eu/art-7-gdpr/
- Art. 9 (Special categories): https://gdpr-info.eu/art-9-gdpr/
- Art. 12 (Transparency): https://gdpr-info.eu/art-12-gdpr/
- Art. 13 (Information at collection): https://gdpr-info.eu/art-13-gdpr/
- Art. 14 (Information not from subject): https://gdpr-info.eu/art-14-gdpr/
- Art. 15 (Right of access): https://gdpr-info.eu/art-15-gdpr/
- Art. 16 (Right to rectification): https://gdpr-info.eu/art-16-gdpr/
- Art. 17 (Right to erasure): https://gdpr-info.eu/art-17-gdpr/
- Art. 20 (Data portability): https://gdpr-info.eu/art-20-gdpr/
- Art. 25 (By design/default): https://gdpr-info.eu/art-25-gdpr/
- Art. 30 (Records of processing): https://gdpr-info.eu/art-30-gdpr/
- Art. 32 (Security of processing): https://gdpr-info.eu/art-32-gdpr/
- Art. 33 (Breach notification to authority): https://gdpr-info.eu/art-33-gdpr/
- Art. 34 (Breach communication to subject): https://gdpr-info.eu/art-34-gdpr/
- Art. 35 (DPIA): https://gdpr-info.eu/art-35-gdpr/
- Art. 44 (Transfer principles): https://gdpr-info.eu/art-44-gdpr/

EDPB (European Data Protection Board) guidelines:
- Guidelines 05/2020 on consent under Regulation 2016/679: https://www.edpb.europa.eu/our-work-tools/our-documents/guidelines/guidelines-052020-consent-under-regulation-2016679_en
- Guidelines 4/2019 on Article 25 — Data Protection by Design and by Default: https://www.edpb.europa.eu/our-work-tools/our-documents/guidelines/guidelines-42019-article-25-data-protection-design-and_en
- Guidelines 9/2022 on personal data breach notification: https://www.edpb.europa.eu/our-work-tools/documents/public-consultations/2022/guidelines-92022-personal-data-breach_en

Verification references:
- ICO Guide to Data Protection Principles: https://ico.org.uk/for-organisations/uk-gdpr-guidance-and-resources/data-protection-principles/a-guide-to-the-data-protection-principles/
- ICO Guide to Data Security: https://ico.org.uk/for-organisations/uk-gdpr-guidance-and-resources/security/a-guide-to-data-security/
- ICO Data Protection by Design and Default: https://ico.org.uk/for-organisations/uk-gdpr-guidance-and-resources/accountability-and-governance/guide-to-accountability-and-governance/data-protection-by-design-and-by-default/

### OWASP ASVS 4.0.3

The OWASP Application Security Verification Standard provides a framework of security requirements for web applications. These rules target version 4.0.3 (October 2021). Version 5.0.0 was released May 2025.

- **Official specification**: OWASP Application Security Verification Standard 4.0.3
- **GitHub repository**: https://github.com/OWASP/ASVS
- **Version 4.0.3 tag**: https://github.com/OWASP/ASVS/tree/v4.0.3

Individual chapter sources (v4.0.3):
- V2 Authentication: https://github.com/OWASP/ASVS/blob/v4.0.3/4.0/en/0x11-V2-Authentication.md
- V3 Session Management: https://github.com/OWASP/ASVS/blob/v4.0.3/4.0/en/0x12-V3-Session-management.md
- V4 Access Control: https://github.com/OWASP/ASVS/blob/v4.0.3/4.0/en/0x12-V4-Access-Control.md
- V5 Validation, Sanitization, Encoding: https://github.com/OWASP/ASVS/blob/v4.0.3/4.0/en/0x13-V5-Validation-Sanitization-Encoding.md
- V7 Error Handling and Logging: https://github.com/OWASP/ASVS/blob/v4.0.3/4.0/en/0x15-V7-Error-Logging.md
- V8 Data Protection: https://github.com/OWASP/ASVS/blob/v4.0.3/4.0/en/0x16-V8-Data-Protection.md
- V9 Communications: https://github.com/OWASP/ASVS/blob/v4.0.3/4.0/en/0x17-V9-Communications.md
- V12 Files and Resources: https://github.com/OWASP/ASVS/blob/v4.0.3/4.0/en/0x20-V12-Files-Resources.md
- V13 API and Web Service: https://github.com/OWASP/ASVS/blob/v4.0.3/4.0/en/0x21-V13-API.md
- V14 Configuration: https://github.com/OWASP/ASVS/blob/v4.0.3/4.0/en/0x22-V14-Config.md

### DAMA-DMBOK 2nd Edition

The Data Management Body of Knowledge, published by DAMA International. Chapter 13 (Data Quality Management) defines the 8 data quality dimensions used in the PR data integrity agent. The DAMA Wheel defines 11 knowledge areas for data management.

- **Official publication**: DAMA-DMBOK: Data Management Body of Knowledge, 2nd Edition (2017), ISBN 978-1634622349
- **Publisher**: DAMA International / Technics Publications
- **DAMA International**: https://dama.org/
- **Learning resources**: https://dama.org/learning-resources/dama-data-management-body-of-knowledge-dmbok/

Verification references:
- Atlan — DAMA DMBOK Framework: An Ultimate Guide: https://atlan.com/dama-dmbok-framework/
- Data Crossroads — DAMA-DMBOK in a Nutshell: https://datacrossroads.nl/2018/06/25/dama-dmbok-in-a-nutshell/
- Dataversity — The Many Dimensions of Data Quality: https://www.dataversity.net/the-many-dimensions-of-data-quality/
- OvalEdge — What Is DAMA-DMBOK: https://www.ovaledge.com/blog/dama-dmbok-data-governance-framework
- DAMA-NL — Dimensions of Data Quality Research Paper (2020): https://dama-nl.org/wp-content/uploads/2020/09/DDQ-Dimensions-of-Data-Quality-Research-Paper-version-1.2-d.d.-3-Sept-2020.pdf

## Supplementary References

- OWASP Top 10 (2021): https://owasp.org/Top10/
- OWASP ASVS 5.0.0 announcement: https://owasp.org/blog/2025/04/09/asvs-rc1-review
- What's New in ASVS 5.0: https://softwaremill.com/whats-new-in-asvs-5-0/

## Verification Record

All control numbers, requirement IDs, and article references in these rule files were verified against the sources listed above on 2026-03-21. Corrections applied:

- ISO A.8.25, A.8.27, A.8.29, A.8.31: titles updated to match official full names
- NIS2 Art. 21(2) subsections (a), (c), (e), (f), (g), (h), (i), (j): descriptions expanded to match official EUR-Lex text
- NIS2 Art. 21(2)(g): new rule file `cyber-hygiene-training.md` created (was missing entirely)
- NIS2 Art. 23 final report deadline: clarified as 1 month after incident notification, not discovery
- OWASP ASVS: noted deleted requirement IDs (V4.1.4, V7.3.2, V13.1.2) and ASVS 5.0.0 availability
- DAMA-DMBOK: 8 data quality dimensions integrated into PR data integrity agent (`pr-data-integrity.md`)
- GDPR: Articles 5-7, 9, 12-17, 20, 25, 30, 32-35, 44-49 verified against gdpr-info.eu full text and EUR-Lex
- GDPR: EDPB Guidelines 05/2020, 4/2019, 9/2022 key requirements verified against EDPB published texts
- GDPR: Recitals 32 (consent), 39 (principles), 64 (identity verification), 78 (technical measures) cross-referenced
