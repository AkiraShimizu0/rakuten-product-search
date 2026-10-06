# Day 6 Canary production publication

Decision: **GO_WITH_AFFILIATE_PENDING**. One article published; stop at D+0.

- Home: https://choicelen.page/
- Article / canonical: https://choicelen.page/articles/compact-air-purifier-placement/
- Published: 2026-10-06 14:41:24 JST (Wrangler custom-domain publish completion).
- Cloudflare account-token authentication succeeded inside GitHub Actions. Project: choicelen-canary. No token values, account IDs or Apple credentials saved in repository/site.
- Registered domain: CloudFlare, Inc.; RDAP status add period / client transfer prohibited. Cloudflare nameservers and authorized Workers custom-domain binding confirmed. HTTPS active with certificate validation.
- HTTP 200: home, article, CSS, robots, sitemap. Missing page: HTTP 404. Correct canonicals, no noindex in HTML or response headers. Robots allows crawling; sitemap contains exactly home and article, lastmod 2026-10-06. This is crawlability, not proof of Google indexing.
- Browser mobile viewport 390x844: body font 16px, no document overflow; tables have their own horizontal scroll wrappers. Top/article navigation works. Advertising disclosure visible above body. Mobile screenshot and DOM evidence saved.
- Article source Markdown unchanged (SHA-256 in manifest); CP04 and CP07 removed assertions remain removed; HU02 resolved calculation remains unchanged. Existing critical=0, major=0; minor 2 removed + 1 resolved. No fresh independent audit claimed.
- Required primary sources: 3/3 GET HTTP200, checked 2026-10-06T05:27:05Z. Removed source P15 also HTTP200, without restoring removed claims.
- Current rendered price snapshots: Core Mini 7,980 JPY, Blue Mini Max 14,300 JPY, FU-TC01 19,800 JPY. Prices are snapshots, no coupon assumptions. FU-TC01 listing explicitly sold out. Other two: stock unknown, ordering UI is not a guarantee. Product snapshot records public shop/model observations only.
- Affiliate links: **0**, AFFILIATE_LINK_NOT_CONFIGURED. No genuine ID in existing environment; no fabricated links. Disclosure preserved. Conversion, current commission and revenue unknown.
- SEARCH_CONSOLE_NOT_CONFIGURED / ANALYTICS_NOT_CONFIGURED: no available configuration used. No account-wide absence asserted; no tracking adopted. Cloudflare standard-statistics availability not checked independently.
- All measurements remain unknown: indexed, impressions, clicks, CTR, position, affiliate clicks, orders, commission. D+0/3/7/14/28/56 dates anchored to actual publication; no automation scheduled.
- Git branch day6-1-production-publication; Draft PR #9 based on day6-canary-publication. PR #8 and earlier stack remain unmerged. Deployment does not require merging the stack.
- Tests: go test ./..., go vet ./..., git diff --check passed; official pinned Wrangler 4.147.0 dry-run and production deployment succeeded; public verification passed.

## Reproduction and limitations

See README for build/verify commands. Workflow runs deployment only with workflow_dispatch or an explicit [deploy-choicelen] marker; ordinary pushes do not redeploy. Until workflow exists on the default branch, marked branch pushes provide the explicit deployment trigger. Publication artifacts use create-only writes; retain earlier blocked Day6 artifacts separately.

Cloudflare Workers Static Assets and custom domain provisioning follow official documentation:
https://developers.cloudflare.com/workers/static-assets/get-started/
https://developers.cloudflare.com/workers/configuration/routing/custom-domains/
https://developers.cloudflare.com/workers/ci-cd/external-cicd/github-actions/

Remaining: genuine affiliate setup; Search Console/measurement availability; sold-out representative listing; no proof of traffic, indexing, revenue or SEO performance. No second article, model changes, SNS, ads or SEO operations performed.
