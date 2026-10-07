# Family validation rulebook v1 (frozen 2026-10-07 JST)

Scope: dehumidifiers (204546), electric shredders (303156), electric pressure cookers (568219). No production resolver replacement. Labels are an independent AI/document review reference, not human ground truth.

Same family: identical basic model across shops, title punctuation/case/width/spacing variants, explicitly identified color suffixes, and bundles containing that same body model. Manufacturer must match. A bundle containing two different bodies is ambiguous.

Separate family: substantively different capacity or performance models, different model numbers/generations, upper/lower models, different manufacturers, body versus parts/consumables/accessories. A series name alone is insufficient to merge different basic models. Compatible-body numbers in a parts title do not identify the part itself.

Boundary cases: strip a terminal color suffix only when the title/document explicitly identifies the color. Do not strip arbitrary final letters or infer generations from resemblance. Accessories with no own part number, multiple candidate models, unidentified manufacturers/models, and contradictory titles/descriptions are AMBIGUOUS and excluded from primary metrics. Keep them in the gold file. No invented capacity/specification.

Sampling: reuse existing review-order 90 listings/category, deterministic hash seed 20261007, 40/category. Labels use title, description and explicit manufacturer/model text, never baseline family ID; retain provenance/reason/confidence. Pairwise metrics exclude ambiguous rows. Exact-family-match means the predicted partition containing each evaluated item equals its reference partition, independent of ID spelling. Zero positive-pair denominators are unavailable, not proof of perfect recall.

Baseline is the existing Cluster implementation. Candidate v2 may normalize manufacturer/model/color and separate body/parts, but remains experimental. Adoption proposal requires macro pairwise F1 improvement without increased false merges; otherwise HOLD. Do not edit this rulebook after viewing metrics.
