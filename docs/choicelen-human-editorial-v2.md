# ChoiceLen Human Editorial v2

## Scope and base

Based on `choicelen-editorial-design-v1` / draft PR #12, independently of research PRs #10–11. One existing article only. No merge, new article, category research, Search Console change or Analytics setting change. The title/H1 and meta description stay exact; domain, slug, canonical, robots, sitemap and the formally issued sponsored link remain unchanged. The frozen Day 6 Markdown is retained as the factual baseline; `internal/editorial/human-article.html` is the new editorial presentation, not a new evidence source.

## Visual audit and reference observations

1440 × 1000 and 390 × 844 browser views of the live v1 home/article were saved before editing. The long chain of eyebrow, title, dek, dates, disclosure, lead, TOC and conclusion delayed the actual question. Uniform paragraph/section treatment, two dense tables, repeated audit language and a large mission section made the site look like a reusable validation template. A one-article site had more navigation and institutional explanation than it needed.

Reference pages were visually opened on 2026-10-07 (Asia/Tokyo):

- [価格.com product guide](https://kakaku.com/kaden/air-purifier/guide_2160/): topic images and specific purchase questions make the subject visible; its extensive navigation is unsuitable for our one-article scale.
- [無印良品 editorial column](https://www.muji.net/lab/living/100120.html): concrete actions and short subject-specific subheads move the story forward; no wording or layout was copied.
- [価格.comマガジン](https://kakakumag.com/): topic-specific visual entry points rather than abstract brand promises. We did not borrow its carousel, scale, ads, ratings or product claims.

These are design references only, never added to the article's evidence set. No third-party text, photographs or illustrations are reproduced.

## Editorial decisions

- Home: wordmark, one short line, one article with its own concept diagram, a one-sentence introduction and advertising note. Single article-list header link; compact footer.
- Opening: unchanged H1 and date, visible ad disclosure, then “置ける。でも、そのまま使える？” with a diagram. No TOC/dek/conclusion stack.
- Headings describe this specific task: 棚に入るだけでは足りない / 寸法は、まず外枠で比べる / 弱運転で使うなら、最大値だけ見ない / 置き場所で迷ったら、ここに戻る / 買う前に、置く場所と型番をひと続きで見る.
- Two tables become one decision table; model facts become individual short notes; airflow uses a simple inline numerical comparison. Seven checklist sentences are consolidated into five steps without dropping their decisions.
- Repeated no-firsthand/performance/health, generalization and price/stock caveats are grouped in “この記事の調べ方”. Model-specific unknowns and numerical qualifications remain adjacent to the affected information. Figure-specific limitations remain in captions.
- Factual content has been edited for clarity, not expanded with new specifications, installation distances, measured results or author identity. “Human Editorial” describes the intended reading experience, not proof of human authorship.

## Figures and audit trail

The first inline SVG is explicitly a generic conceptual diagram. Its lines do not encode actual clearances, airflow positions, maintenance directions or permission to put any model on a shelf. The second figure uses same-scale rectangular dimension envelopes for FU-TC01 (190 × 330 mm) and Mini Max (173 × 290 mm), not inferred device shapes. Core Mini is not drawn to an invented size. Differences 17 mm / 40 mm are calculated by Go, retaining CP09/CALC02/CALC03's limitation to dimensions. Color does not carry the meaning alone; each SVG has a title and description.

Retained mappings remain CP01→P14, CP02→P14, CP03-D6→P14, CP05→P12, CP06→P14, CP08→P16, CP09→P14/P16, CP10→P16. HTML `data-claim-id`, `data-classification`, source row IDs and the inert JSON `claim-source-map` separate internal audit information from reader references. CP04 and CP07 stay removed; Core Mini clearances, washing, intervals and SKU/manual revision remain unverified. The CP08-derived placement advice remains explicitly editorial inference. Conditional filter intervals, CADR units, JEM standard incompatibility and weak/strong airflow differences remain qualified.

Claim audit: critical 0, major 0, remaining minor 0 in the edited copy. This is an evidence-retention/source-mapping audit against the previously verified baseline and Day 6 claim-resolution record, not a new independent physical test or a fresh verification of every PDF statement. Unknowns are not upgraded. Audit records and before/preview/production screenshots live outside Git under `outputs/human-editorial-v2/`.

## Verification and limits

Meaningful regressions cover frozen baseline hash, reproducible build, retained mappings and unknowns, removed assertions, exact H1/title/meta/canonical, source and affiliate destinations, sponsored disclosure, diagram scale, accessible descriptions and missing audited input rejection. Full Go tests, vet and diff whitespace checks are required before deploy. Browser QA checks desktop/mobile diagram legibility, table-only scrolling, sources, discreet sponsored link, internal anchors and focus. No framework, external font, raster/stock/AI image, new executable JavaScript or new tracking is introduced.

This scoped visual/semantic QA is not a human usability study, real iOS-device test, full WCAG certification, SEO experiment or promise of improved traffic/latency. Core Mini's primary-document gap remains. Source HTTP availability is recorded separately from substantive fact verification. D+7 must not be interpreted as a causal v1/v2 comparison: v2 changes layout, visuals and editorial wording while keeping core claims, URLs and search intent.

Production completion timestamp, deployed commit and workflow are appended to `docs/measurement-log.jsonl` after successful deployment.
