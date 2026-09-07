# VERIFICATION

Thesis: 一条路径，不是目录

## Kept vs changed

- Kept, not edited: `docs/architecture-infographic.tex`, `docs/architecture-infographic.svg`, `docs/architecture.html`.
- Kept, not edited: `docs/infographics/driver-protocol-flow.tex` and its derived artifacts (existing explainer; protocol / normalize / exit-code thesis).
- New: Class A path spine (`page.svg` + `index.html` with the SVG inlined + `proof.png`). Different thesis from the explainer: path-not-catalog, thin-driver boundary.

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

## 2026-09-06 refine

本次精修（fleet-refine b01）。本目录已入库（ab73aa7）；本次精修的改动未提交，由主会话稍后统一提交。

### 修复

- img-external-panel（high）：index.html 由 `<object data="page.svg">` 外挂改为内联 SVG，内联字节与 page.svg 一致；page.svg 仍是由 build.py 重建的源。页内零外部引用。
- width-rule（med）：CSS-only，body{max-width:1200px;margin:0 auto}、svg{width:100%;height:auto}；page.svg 画布仍 1280×720，几何未动。
- 文档随改：README 原称「页面是外部 SVG；index.html 只是薄包装」【后证不实，已修正】——现为内联 SVG、1200px 居中；contract 的 medium 行同步改为内联表述。

### 缓办（按本波范围记录，不在本次实现）

- no-claims-binding、no-poison、no-vacuum。
- fingerprint-gaps：指纹扩为 page.svg + index.html 两产物；全树指纹未做。
- other（低）：contract.md / VERIFICATION.md 仍为英文。
- form-mismatch / no-sidenote-track：单图形态与旁注轨未改。

### 门禁复跑（2026-09-06，仓库根执行）

- svg-linter objective 门（--require-complete --fail-on error）：0 错误。
- hygiene 与 collision 两组复查：各 0 发现。
- 从 index.html 正则抽取内联 SVG，与 page.svg 字节一致。
- 双跑 build.py：6 个产物文件字节一致（确定性）。
- rsvg-convert --width=2560 重渲染与既有 proof.png 字节一致。
- 页面代码坐标扫描（build.py 内置）：0 命中。

## Artifact fingerprints (this build)

- page.svg SHA-256: `b2513d295a65ed1715b6ac39d53811a8398f1a704524d9d3f7508a14d2f48774`
- page.svg bytes: 11035
- index.html SHA-256: `a85fcf97be0a3391a0fce9924f5d151cf6e18b4462b494fb47e10ed5673e7c80`
- index.html bytes: 11348
- index.html inlines page.svg byte-identically; zero external references
- proof.png is produced by `rsvg-convert --width=2560` (2x of 1280 x 720)
- CJK font at render: Source Han Serif SC

## Page code-coordinate sweep

Zero matches for file:line, type names, or function names on page.svg / index.html.
