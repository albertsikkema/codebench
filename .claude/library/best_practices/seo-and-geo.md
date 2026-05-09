# SEO and Geo-Targeting

## Principle

Build sites that search engines can crawl, understand, and present well in results. For multi-region or multi-language sites, tell search engines which content is for which audience. SEO is not a hack — it's making your site's structure and intent machine-readable.

## Why

- **Search is the front door**: For most sites, organic search is the largest traffic source. If search engines can't crawl, understand, or index your pages, you're invisible.
- **Technical SEO is a developer's job**: Content strategy is marketing's responsibility. But crawlability, page speed, structured data, and canonical URLs are engineering decisions that marketing can't fix.
- **Geo-targeting mistakes hurt rankings**: Serving Dutch content to users searching in English (or vice versa) frustrates users and confuses search engines. Explicit signals prevent this.

## Core Rules

### 1. Make Every Page Crawlable

Search engines need to fetch, render, and understand your pages.

**Requirements**:
- Every public page is reachable via links from other pages (no orphan pages)
- Critical content is in the initial HTML, not loaded via JavaScript after page load
- Pages return proper HTTP status codes (200 for content, 301 for permanent redirects, 404 for missing)
- No infinite URL spaces (filters, sorts, pagination that generate unlimited URLs)

**For SPAs (React, Vue, Next.js, Nuxt)**:
- Use server-side rendering (SSR) or static site generation (SSG) for public-facing pages
- Client-side-only rendering is invisible to most crawlers
- Pre-rendering services (Prerender.io) are a fallback, not a solution

### 2. One URL Per Piece of Content

Every piece of content should have exactly one canonical URL. Duplicate content dilutes ranking.

**Use canonical tags** to declare the authoritative URL:

```html
<link rel="canonical" href="https://example.com/products/widget">
```

**Common duplication sources**:
- `http://` vs `https://` → redirect HTTP to HTTPS
- `www.` vs non-www → pick one, redirect the other
- Trailing slashes → pick one convention, redirect the other
- Query parameters for tracking (`?utm_source=...`) → canonical points to the clean URL
- Pagination → each page is canonical to itself, not to page 1

### 3. Write Meaningful Title and Meta Tags

Every page needs a unique, descriptive `<title>` and meta description.

```html
<head>
  <title>Widget Pro - Lightweight Project Management | Example</title>
  <meta name="description" content="Widget Pro helps small teams track tasks, deadlines, and progress without the overhead of enterprise tools. Free for up to 5 users.">
</head>
```

**Rules**:
- Title: 50–60 characters, unique per page, includes primary keyword
- Description: 150–160 characters, describes the page content, includes a call to action
- Both must be set server-side (not injected via JavaScript)

### 4. Use Structured Data

Structured data (JSON-LD) helps search engines understand what your content is — not just what it says.

```html
<script type="application/ld+json">
{
  "@context": "https://schema.org",
  "@type": "Product",
  "name": "Widget Pro",
  "description": "Lightweight project management for small teams",
  "offers": {
    "@type": "Offer",
    "price": "0",
    "priceCurrency": "EUR"
  }
}
</script>
```

**Common types**:
- `Organization` — company info, logo, social profiles
- `Product` — pricing, availability, reviews
- `Article` / `BlogPosting` — author, date, headline
- `FAQPage` — question-answer pairs (appear as rich results)
- `BreadcrumbList` — navigation path

Validate with Google's Rich Results Test before deploying.

### 5. Build a Sitemap

A sitemap tells search engines what pages exist and when they were last changed.

```xml
<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
  <url>
    <loc>https://example.com/</loc>
    <lastmod>2026-03-28</lastmod>
    <changefreq>weekly</changefreq>
    <priority>1.0</priority>
  </url>
  <url>
    <loc>https://example.com/products/widget</loc>
    <lastmod>2026-03-15</lastmod>
    <changefreq>monthly</changefreq>
    <priority>0.8</priority>
  </url>
</urlset>
```

**Rules**:
- Generate automatically from your content/routes — don't maintain manually
- Reference the sitemap in `robots.txt`: `Sitemap: https://example.com/sitemap.xml`
- Submit to Google Search Console and Bing Webmaster Tools
- Keep under 50,000 URLs per sitemap file (use sitemap index for larger sites)
- Only include canonical URLs (no duplicates, no non-indexable pages)

### 6. Implement Hreflang for Multi-Language/Region Sites

`hreflang` tells search engines which language/region variant of a page to show to which users.

```html
<head>
  <!-- Dutch version for Netherlands -->
  <link rel="alternate" hreflang="nl-NL" href="https://example.nl/producten/widget">
  <!-- Dutch version for Belgium -->
  <link rel="alternate" hreflang="nl-BE" href="https://example.be/producten/widget">
  <!-- English version (default) -->
  <link rel="alternate" hreflang="en" href="https://example.com/products/widget">
  <!-- Fallback for unmatched languages -->
  <link rel="alternate" hreflang="x-default" href="https://example.com/products/widget">
</head>
```

**Rules**:
- Every page in every language must link to all its variants (including itself)
- Always include `x-default` as the fallback
- `hreflang` is bidirectional — if page A points to page B, page B must point back to page A
- Use language-only (`hreflang="nl"`) when content is the same for all regions speaking that language. Use language-region (`hreflang="nl-NL"`) when content differs per region (pricing, legal terms, local references)

**URL strategies for multi-language**:

| Strategy | Example | Pros | Cons |
|----------|---------|------|------|
| ccTLD | `example.nl`, `example.be` | Strongest geo signal | Separate domains to manage |
| Subdomain | `nl.example.com` | Easy to set up | Weaker geo signal |
| Subdirectory | `example.com/nl/` | Single domain, easy to manage | Weakest geo signal |

Subdirectory is the most common and practical for most projects. ccTLDs are strongest for geo-targeting but add operational overhead.

### 7. Optimize Page Speed

Page speed is a ranking factor. Core Web Vitals (LCP, CLS, INP) directly affect search rankings.

This overlaps with the performance requirements in the baseline specifications. The key SEO-specific points:

- **Largest Contentful Paint (LCP) < 2.5s**: The main content must load fast. Optimize images, use CDN, minimize render-blocking resources.
- **Cumulative Layout Shift (CLS) < 0.1**: Set explicit `width` and `height` on images/videos. Use `font-display: swap`. Reserve space for dynamic content.
- **Interaction to Next Paint (INP) < 200ms**: Keep JavaScript execution fast. Break up long tasks. Use web workers for heavy computation.

### 8. Handle Redirects and Removed Content Properly

| Situation | Action | HTTP status |
|-----------|--------|-------------|
| Page moved permanently | Redirect old URL to new URL | 301 |
| Page moved temporarily | Redirect old URL to new URL | 302 |
| Page removed, no replacement | Show a useful 404 page | 404 |
| Page removed, similar content exists | Redirect to the closest alternative | 301 |

**Never**:
- Redirect all 404s to the homepage (confuses crawlers, bad UX)
- Chain redirects (A → B → C → D — each hop loses ranking value)
- Use JavaScript redirects for permanent moves (crawlers may not follow)

### 9. Geo-Targeting Without hreflang

For single-language sites that serve different regions:

- **Google Search Console**: Set geographic target per property (under Settings → International Targeting)
- **Server-side geo-detection**: Use the visitor's IP to show region-specific content (pricing, phone numbers, legal text) while keeping the same URL structure
- **Structured data**: Use `areaServed` in your Organization schema to indicate service regions

**Don't**: Automatically redirect users based on IP without providing an override. Users behind VPNs, travelers, and expats hate this. Show a banner suggesting the regional version instead.

### 10. Make 404 Pages Useful

A custom 404 page should:
- Clearly state the page wasn't found
- Provide navigation (search bar, links to popular pages, sitemap link)
- Return HTTP 404 status code (not 200 with "page not found" text)
- Match the site's design (not a generic server error page)

```html
<!-- Return 404 status, not 200 -->
<h1>Page not found</h1>
<p>The page you're looking for doesn't exist or has been moved.</p>
<ul>
  <li><a href="/">Go to homepage</a></li>
  <li><a href="/products">Browse products</a></li>
  <li><a href="/contact">Contact us</a></li>
</ul>
```

## Monitoring

- **Google Search Console**: Crawl errors, index coverage, Core Web Vitals, search performance
- **Bing Webmaster Tools**: Same, for Bing
- **Lighthouse CI**: Automated SEO and performance audits in the pipeline
- **Log analysis**: Monitor crawl frequency, crawl budget usage, and error rates from bot traffic

## When to Bend the Rules

- **Internal applications**: No SEO needed. Skip sitemaps, structured data, and hreflang entirely.
- **Authenticated content**: Search engines can't log in. If all content is behind auth, SEO applies only to the landing/marketing pages.
- **Single-language, single-region sites**: Skip hreflang. Set geographic target in Search Console if relevant.
- **API-only services**: No HTML to optimize. Focus robots.txt on blocking documentation you don't want indexed.
