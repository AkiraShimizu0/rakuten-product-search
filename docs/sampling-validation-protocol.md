# Sampling validation protocol v1 (frozen 2026-10-07 JST)

Categories fixed: 204546 dehumidifiers, 303156 electric shredders, 568219 electric pressure cookers. Preserve standard-order and review-order experiments and all semantic configurations.

Candidate pool: current Rakuten item search, available listings, genre fixed, minPrice=3000/maxPrice=50000 and hasReviewFlag=1. Fetch six pages each under standard, +itemPrice, -itemPrice (30/page); no review-count ordering in the new pool. Pace 1.2 seconds, timeout 20s, retries 2. This is a bounded search pool, not a probability sample of all listings. Review presence is an explicit acquisition restriction; the unchanged eligibility filter still requires >=5 reviews and >=80 description runes. Stop with INCONCLUSIVE if fewer than 90 eligible unique products; do not enlarge/change the protocol after seeing data.

Deduplicate source/source_id, first occurrence retained. Sort eligible pool by price ascending then source/source_id; split rank positions into three nearly equal bands using floor(i*3/n), so price ties may cross bands. Select 30 per band by SHA256(seed|source|source_id), seed 20261007. Persist pool, acquisition counts, band boundaries and selected IDs.

Jev sample: 15 per price band (45/category); reuse identical saved v1 requests/results when available. Claude: at most 10 gate-pass products/category, allocated 3 low/3 mid/4 high and chosen by fixed hash, not by semantic score; a shortage leaves an unfilled slot. Reuse saved matching Claude input/model/rubric results. No reranker or diversification changes. These sample counts fit the additional $2 budget with a conservative preflight; stop paid work if estimate exceeds $2. Record actual requests/usage/failures/cost. Missing data is unknown.

Compare unchanged review-order baseline with the new protocol: pool/sample eligibility, Jev pass rate and raw mean/median, Claude mean and buyer/comparison/wrong-choice axes, family count/largest share/HHI. Review-order semantic sample was selected differently; report this confound rather than attributing every difference to acquisition alone.

Pre-registered sensitivity: INCONCLUSIVE if pool/sample quotas fail, any Claude group has <8 valid scored items, or incomplete Jev/billing data. Otherwise SAMPLE_SENSITIVE if absolute Claude mean difference >5 points (0-100), gate-rate difference >0.20, or category rank moves >=2 positions among these three. ROBUST otherwise. These are exploratory stability rules, not market/SEO GO rules. Do not change thresholds after results. Report direction and sample size, not significance claims from small nonrandom samples.

Standard-sort failures are decomposed with unchanged filter reasons plus clearly labelled title-based accessory proxy; distinguish overlapping descriptive counts from exclusive exclusion attribution. No claim that a failed acquisition sample proves the entire category weak.
