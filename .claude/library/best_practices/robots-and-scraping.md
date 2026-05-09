# Robots, Crawling, and Scraping Protection

## Principle

Control what automated agents can access on your site. Use `robots.txt` to guide well-behaved crawlers, but never rely on it as a security mechanism. Protect sensitive content with authentication and rate limiting, not with polite requests to bots.

## Why

- **robots.txt is a convention, not a barrier**: It's a plaintext file that asks crawlers to stay away. Malicious scrapers ignore it entirely. Treating it as access control is a security mistake.
- **AI crawlers are the new SEO scrapers**: LLM training crawlers (GPTBot, CCBot, Google-Extended, Bytespider) aggressively scrape content. If you don't explicitly block them, your content becomes training data.
- **Scraping costs you money**: Bots consume bandwidth, server resources, and can skew analytics. Unchecked scraping of API endpoints can look like a DDoS attack.

## Core Rules

### 1. Always Have a robots.txt

Every public-facing site needs a `robots.txt` at the root. Even if you allow everything, make it explicit.

```
# /robots.txt

# Allow search engines
User-agent: Googlebot
Allow: /

User-agent: Bingbot
Allow: /

# Block AI training crawlers
User-agent: GPTBot
Disallow: /

User-agent: CCBot
Disallow: /

User-agent: Google-Extended
Disallow: /

User-agent: Bytespider
Disallow: /

User-agent: anthropic-ai
Disallow: /

User-agent: ClaudeBot
Disallow: /

# Default: allow
User-agent: *
Allow: /

# Sitemap
Sitemap: https://example.com/sitemap.xml
```

**Key decisions to document**:
- Which crawlers are allowed and why
- Which paths are disallowed (admin, API, user content)
- Whether AI training crawlers are blocked (increasingly the default)

### 2. Never Put Sensitive Paths in robots.txt

```
BAD:
  Disallow: /admin/
  Disallow: /internal-api/
  Disallow: /staging/

  → Tells attackers exactly where your sensitive paths are
```

Sensitive paths should be protected by authentication. Don't advertise them in robots.txt. If a path shouldn't be public, it shouldn't be reachable without credentials — robots.txt is irrelevant.

### 3. Block AI Training Crawlers Explicitly

AI training crawlers are a separate concern from search engine crawlers. You can allow Google to index your site for search while blocking it from using your content for AI training.

| Crawler | Operator | Purpose |
|---------|----------|---------|
| `Googlebot` | Google | Search indexing |
| `Google-Extended` | Google | AI training (Gemini) |
| `GPTBot` | OpenAI | AI training |
| `ChatGPT-User` | OpenAI | ChatGPT browsing |
| `CCBot` | Common Crawl | Open dataset (used by many AI companies) |
| `Bytespider` | ByteDance | AI training |
| `anthropic-ai` | Anthropic | AI training |
| `ClaudeBot` | Anthropic | AI training |
| `Applebot-Extended` | Apple | AI training |
| `meta-externalagent` | Meta | AI training |

Block AI crawlers unless you explicitly want your content in training data. This is a business decision, not a technical one — document it.

### 4. Use Meta Tags for Per-Page Control

`robots.txt` controls crawling (fetching). Meta tags control indexing (appearing in results).

```html
<!-- Don't index this page, don't follow its links -->
<meta name="robots" content="noindex, nofollow">

<!-- Index but don't cache -->
<meta name="robots" content="noarchive">

<!-- Block AI training specifically (Google) -->
<meta name="google" content="nositelinkssearchbox, notranslate">
<meta name="google-extended" content="noindex">
```

Use meta tags for:
- Staging/preview environments (noindex everything)
- User-generated content pages you don't want indexed
- Pages with temporary content (event pages, promotions)

### 5. Rate Limit All Public Endpoints

Well-behaved bots respect `Crawl-delay` in robots.txt. Everything else needs server-side enforcement.

```
# robots.txt — advisory only
User-agent: *
Crawl-delay: 10
```

**Server-side enforcement**:
- Rate limit by IP on all public endpoints
- Rate limit by user-agent pattern for known bot signatures
- Return 429 with `Retry-After` header when limits are exceeded
- Consider CAPTCHA or proof-of-work challenges for suspicious traffic patterns

### 6. Protect APIs from Scraping

APIs are the most valuable scraping target — structured data, easy to automate.

**Strategies**:
- **Authentication**: Require API keys for all non-public endpoints
- **Rate limiting**: Per-key and per-IP limits
- **Pagination limits**: Cap page size (max 100 items), enforce cursor-based pagination
- **Response limiting**: Don't return more data than the UI needs (no "dump everything" endpoints)
- **Monitoring**: Alert on unusual access patterns (single IP fetching all pages sequentially)

### 7. Add a Security.txt

While not directly about robots, `/.well-known/security.txt` tells security researchers how to report vulnerabilities they find while examining your site.

```
# /.well-known/security.txt
Contact: mailto:security@example.com
Preferred-Languages: en
Canonical: https://example.com/.well-known/security.txt
Expires: 2027-01-01T00:00:00.000Z
```

## Monitoring

Track bot traffic to understand what's hitting your site:

- Log `User-Agent` strings on all requests
- Monitor for unusual traffic spikes from single IPs
- Track requests to `robots.txt` (crawlers fetch this first)
- Alert on high-volume sequential access patterns (scraping signature)

## When to Bend the Rules

- **Public APIs meant for third-party consumption**: Rate limit but don't block. That's the product.
- **Open-source documentation sites**: You may want AI crawlers to index your docs so LLMs can answer questions about your project. Allow selectively.
- **Marketing sites**: SEO is the priority. Allow all search crawlers, block only AI training crawlers. Make sure the sitemap is accurate and up to date.
- **Internal applications**: No robots.txt needed — the app shouldn't be reachable from the public internet at all.
