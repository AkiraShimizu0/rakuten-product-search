# ChoiceLen typography and layout reset

Remove rhetorical opening copy, logo serif styling, opening split layout, tinted diagram cards, repeated brand messages and ornamental section rules. Keep a white background and Japanese system sans. Article body is 700px; H1 is 32px on desktop and 22px on mobile, left aligned with normal wrapping.

The original title, H1, meta description, canonical, frozen publication Markdown, robots, sitemap, all eight claim/source mappings, unknown Core Mini facts, formal affiliate destination and sponsored relation are unchanged. Diagram geometry and captions are retained and moved into the relevant article section.

Validation: go test ./..., go vet ./..., git diff --check; Desktop1440px and Mobile390px preview/production visual QA; no horizontal page overflow; SVG accessibility labels, keyboard focus and table horizontal scrolling; homepage/article/CSS/robots/sitemap HTTP200 with verified HTTPS; missing page HTTP404; three primary URLs HTTP200; existing Analytics beacon and sponsored href match baseline.

Production commit e29f0ad98b4d960f038f84ee9fe07a529cbb4455.
Draft PR #14, base PR #13; no merges.
Detailed audit, before/after screenshots and measurements remain in local outputs/typography-layout-reset, outside Git.
D+7 must not be interpreted as a causal design comparison.
