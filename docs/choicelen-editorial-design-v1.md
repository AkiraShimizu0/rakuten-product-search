# ChoiceLen Editorial Design v1

Scope: visual design and information hierarchy of the existing home, one article and 404. No second article, new category research, ranking/rubric changes or SEO experiment. This branch starts from the actual production branch `day6-1-production-publication` (PR #9), independently of experiment PRs #10/#11. No PR is merged.

## Before visual audit

Actual production was inspected in the browser at 1440×1000 and 390×844 (iPhone-size viewport). Screenshots are retained privately in `outputs/editorial-design-v1/screenshots/`, outside Git.

| Area | Observed issue | Design response |
|---|---|---|
| Brand / first impression | Home is a white text document on a gray field; no shared navigation/footer | Text wordmark, warm paper tone, restrained green ink, consistent shell |
| Typography / measure | Mostly one sans-serif scale; long paragraphs and metadata share weight | System Japanese serif headings, system sans body, article measure 760px |
| Hierarchy / whitespace | Metadata/disclosure/prose compete at the top; no reading route | Article header, exact existing takeaway as dek, contents navigation and ruled sections |
| Tables | Header cells are `td`; mobile scrolling has no explanation | Caption, thead/th, row headers, named focusable horizontal region |
| Checklist | Seven actions form one paragraph | Same sentences, separated into a numbered static list |
| Sources | Internal P/CP identifiers resemble an audit report | Reader reference numbers, organization/type/checked date, external-link label |
| Disclosure / affiliate | Plain paragraphs, no distinction from ordinary prose | Visible top disclosure and restrained labelled sales-conditions area at the end |
| Footer / navigation | Missing publication-wide structure | Home/guides/policy links and concise editorial footer |
| AI/report impression | Raw CP IDs, `family` heading, uniform text blocks, preparation-style language | IDs move to attributes/JSON; heading uses モデル; original caveats remain intact |

No fake editorial staff, credentials, hands-on experience, popularity or independent measurements were introduced. The design communicates editorial organization, not a claim that the article was human-authored or tested in person.

## Design system and content preservation

Warm background `#f6f3ec`, ink `#252c27`, muted `#61685d`, accent `#315945`, thin border `#d8d9cf`; reusable widths, spacing, surface and radius tokens. No gradient, hero artwork, card grid, framework, webfont or new executable JavaScript. Existing Cloudflare Analytics injection is left to Cloudflare configuration, unchanged.

`internal/editorial` wraps frozen content after the original Day 6 renderer/affiliate setup. `cmd/build-choiceLen` still checks the original Markdown SHA256 and refuses existing build output directories. HTML/CSS site files reproduce exactly from the build. Original Markdown, robots, sitemap, URL, canonical, title/meta-description and affiliate destination remain unchanged. Dates remain the original publication 2026-10-06 and source check 2026-10-05; visual deployment does not masquerade as new evidence review.

Presentation changes: common header/footer; existing first conclusion sentence moved to article dek; section anchors; `代表3family` heading displayed as `代表3モデル`; existing checklist sentences become list items; source-description `claim ID` becomes `参照番号`; the sales section label becomes `販売条件を確認する`. Original factual prose/caveats are otherwise preserved. No paragraphs containing uncertainty or limitations are deleted. Removed CP04/CP07 assertions remain removed.

## Claim traceability

Inline CP markers become superscript source references. Every reference retains `data-claim-id`; source rows retain `data-source-id`. An inert `application/json` script `claim-source-map` stores claim IDs, source IDs, Verified/Derived classification and removed claims. Source numbers: 1=P14 Sharp manual; 2=P12 JEMA announcement; 3=P16 Blueair manual. CP03-D6 retains the Day 6 Sharp source substitution; CP09 remains a dimension-only Derived claim; the CP08-based placement rule remains explicitly editorial inference. No source was added to justify a new claim.

The raw CP identifiers remain accessible to internal audits through HTML/metadata, not visible body text. Source-link destinations and the complete issued affiliate query remain identical, with one `rel="sponsored"` link after the purchase checklist. No click was generated to test affiliate conversion.

## QA and validation

Browser preview: home/article at Desktop and iPhone-size viewport; checklist, sources, sales area and focusable table inspected. At Mobile, page has no horizontal overflow; the two tables have dedicated horizontal scroll. ArrowRight moved a focused table 40px. Main links have visible focus treatment, navigation/source/affiliate targets have usable spacing, tables have semantic headers, and a skip link reaches main. This is scoped visual/semantic QA, not a certification of full WCAG conformance or a real iOS device test.

Tests cover original prose sentence retention, exact external destinations, metadata and canonical retention, single sponsored link, source mapping, removed assertions, semantic table structure, static generation and broken source-anchor rejection. The existing source Markdown hash and exact committed-site reproduction tests continue to pass. `go test ./...`, `go vet ./...`, `git diff --check` passed before deployment. Source URLs returned HTTP 200 on 2026-10-07; no new source facts were extracted.

Deployment uses the existing pinned Wrangler/Workers Static Assets configuration and credentials. The existing workflow accepts this branch only on manual dispatch or an explicit `[deploy-choicelen]` marker. Ordinary subsequent log/doc commits do not deploy. Production timestamps, code SHA, actual HTTP/TLS checks, analytics/mobile evidence and measurement caveat are recorded after deployment in a separate log. D+7 traffic differences must not be attributed causally to this design change.

Unresolved: actual usability research, assistive-technology user testing, iOS hardware behavior and business effects are unmeasured. The original Core Mini evidence limitations remain. No claims of increased trust, CTR, ranking or revenue are made from visual QA alone.
