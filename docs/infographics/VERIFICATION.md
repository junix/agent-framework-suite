# VERIFICATION

Thesis: 一条路径，不是目录

## Kept vs changed

- Kept, not edited: `docs/architecture-infographic.tex`, `docs/architecture-infographic.svg`, `docs/architecture.html`.
- Kept, not edited: `docs/infographics/driver-protocol-flow.tex` and its derived artifacts (existing explainer; protocol / normalize / exit-code thesis).
- New: Class A path spine (`page.svg` + thin `index.html` + `proof.png`). Different thesis from the explainer: path-not-catalog, thin-driver boundary.

## Mechanisms checked in source

- README: hermetic dual drive of public APIs, shared oracle, required parity; drivers do not implement routing/checkpoint.
- CONTRACT participant boundary: drivers translate and normalize; they must not implement routing, checkpointing, request tracking, or iteration.
- Catalog stores one Expected per case; runner checks each observation against that same Expected, then compares the two after dropping participant.
- Missing participant count != 2 fails; missing driver is infrastructure, not skip.
- Python/Rust drivers call library graph/build/run/resume APIs for fan-out, switch, and checkpoint cases.
- `just check` builds, doctors both drivers, and runs every case.
- Case IDs WF-001..WF-010 verified in CASES.md and the catalog.

## Build commands

```bash
python3 docs/infographics/build.py
rsvg-convert --width=2560 docs/infographics/page.svg -o docs/infographics/proof.png
```

svg-linter (repository root, after the two commands):

```bash
svg-linter --json check docs/infographics/page.svg \
  --select svg/duplicate-id,svg/dangling-reference \
  --require-complete --fail-on error
svg-linter --json check docs/infographics/page.svg \
  --select svg/excessive-path-complexity,svg/unused-definition,svg/external-resource,svg/raster-upscale,svg/open-filled-path \
  --fail-on never --max-findings 80
svg-linter --json check docs/infographics/page.svg \
  --select svg/line-overlap,svg/line-crossing,svg/line-text-overlap,svg/text-text-overlap,svg/canvas-edge-clearance,svg/ambiguous-junction,svg/connector-through-shape,svg/edge-congestion,svg/filter-clipping-risk,svg/low-contrast-text \
  --fail-on never --max-findings 80
```

Toolchain inspected with `python3 ~/.cursor/skills/create-infographic/scripts/inspect_toolchain.py --json --project-root .` (read-only; mutated=false).
Route: hand-placed editorial SVG (web-svg-backend). No Vega/D3/ELK: topology is editorial and under twelve nodes.

Objective error gate: 0 errors; `svg/duplicate-id` and `svg/dangling-reference` both evaluated (`--require-complete`).
Hygiene review: 0 findings on the 4 evaluated rules; `svg/raster-upscale` not applicable (no rasters).
Collision review: 0 findings on the 9 evaluated geometry/contrast rules; `svg/filter-clipping-risk` not applicable.
svg-linter 0.1.0, catalog 7 (`sha256:fdfd45cabf9908c69853058ca51fba80878a99d21941eebf956258a904e55bb9`).
Unannotated `proof.png` inspected at 2560×1440; numbered collision lint PNG had no candidates after the layout pass.
Saturated roles on the page: Ocean (flow) and Coral (fail). Teal/Gold unused. Fail is the same Coral hue with a dashed stroke.

## Visual review

- Thumbnail: Y-fork into a large shared-expected disc remains the silhouette.
- Title occupies the upper left; upper right stays quiet except the hermetic fingerprint.
- Case IDs are one slate line, not a ten-box catalog.
- Boundary rule is gapped where observations cross; label sits above the stroke.
- Both branch outcomes are labelled (`一致` / `分叉`); missing participant is annotated at fail.
- CJK glyphs present in the SVG source and the rsvg-convert proof.

## Unverified

- Live `just check` against installed Python/Rust frameworks (this task did not run the suite).
- Whether a future case beyond WF-010 would still fit the one-line ID strip.
- browser-harness / Playwright (not used; rsvg-convert only).
- CJK glyph coverage under a machine without Source Han Serif SC.

## Artifact fingerprints (this build)

- page.svg SHA-256: `b2513d295a65ed1715b6ac39d53811a8398f1a704524d9d3f7508a14d2f48774`
- page.svg bytes: 11035
- proof.png is produced by `rsvg-convert --width=2560` (2x of 1280 x 720)
- CJK font at render: Source Han Serif SC

## Page code-coordinate sweep

Zero matches for file:line, type names, or function names on page.svg / index.html.
