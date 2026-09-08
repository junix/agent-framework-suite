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

## 2026-09-07 refine（审计硬化）

本次审计硬化（fleet-refine w2）：补齐缓办的审计机械五类。冻结证据层 data/provenance.json 字节未动（重建前后 sha256 一致）；legacy driver-protocol-flow.* 仍原样保留。改动未提交，由主会话统一提交。

### 新增

- no-claims-binding：页上加「声明 Cxx」随面板脚注（7 枚：C01/C10/C07–C09/C02·C06/C03/C04/C05）与页尾「声明对照表」（index.html 图下 HTML 表，C01–C10）；声明号机器登记在 data/claims.json（Cxx ↔ provenance 锚点 slug + 主张转写 + 用例号 + 数字豁免）。tools/audit_gates.py gate binding 两向校验：页上 Cxx 集合 == 登记集合（unknown/missing 均空）；claim_index 的 slug 全部存在于 provenance anchors；WF 用例号从仓库 CASES.md 重算与登记一致。
- no-reverse-sweep（数字反扫）：gate digits 抽取读者可见投影（SVG <text>/<title>/<desc> + 去 <style> 后的 HTML 文本）的全部数字串并逐个归类。本轮 38 个 token = 声明号 19（chips 9 + 表 10）+ 用例号 14（WF-001…WF-010，C08 声明值）+ 版式序号 5；未申领数字 = 0。豁免表（data/claims.json digit_exemptions，逐条理由）：01–05 = 路径节点序号（01 用例、02/03 双驱、04 共用期望、05 强制比对），版式序号非数据主张。无年份豁免（页上无年份）。
- no-poison（毒丸，全部注入 /tmp 一次性副本，冻结层不动）：detail 门 6 类对照（类型名/函数名/Go 文件名/Rust 文件名/.go: 与 .py: 坐标各一）6/6 命中、干净语料 0；svg-linter 客观门删去被引用的 <defs> 标记 mk-flow → rc=1、6 条 svg/dangling-reference，复原本 rc=0、0 条；claims 门双向（注入 C99 = 未知号、移除 C05 = 登记缺失）2/2 失败、复原通过；数字门注入 42 → 未申领 [42]、复原 0。运行记录 data/audit/pills-2026-09-07.json。
- fingerprint-gaps：tools/fingerprint.py 产出全树 SHA-256 清单 fingerprints.json（数据、工具、渲染、文档全入册；文件数见清单 file_count）。恰两项豁免并在此披露理由：fingerprints.json 自身（自指不动点）、data/audit/（运行记录，门禁重跑不得改指纹）。README/VERIFICATION 引用的 64 位树内哈希均派生自该清单（gate dochash 校验；唯一白名单外部哈希 = svg-linter 0.1.0 规则目录，外部工具链非树内产物）。工具幂等：重写字节不变。
- no-vacuum：tools/vacuum_rebuild.py——/tmp 平拷贝 → 快照 A → 删除 8 个可重建产物（page.svg、index.html、README.md、contract.md、VERIFICATION.md、data/provenance.json、data/claims.json、proof.png）→ build.py + rsvg-convert --width=2560 全链重建 → 三方字节比对 A/重建/归档：compare_files = 8 > 0，逐文件 a_eq_rebuilt 与 rebuilt_eq_archived 全真（0 文件 PASS 视为假 PASS，已防）；/tmp 副本内 fingerprint --check 通过（脱离验证）。运行记录 data/audit/vacuum-2026-09-07.json。

### 门禁复跑（2026-09-07，树内工具）

- tools/audit_gates.py all：7 门全绿（inline / codesweep / binding / digits / svg / fingerprint / dochash），记录 data/audit/gates-2026-09-07.json。
- svg-linter 三组（客观 --require-complete --fail-on error / hygiene / collision）：0 错误、0 发现；三组 effective_rules 分别为 2/5/10（非空，防陈旧目录假绿）。
- rsvg-convert 双渲染字节一致；build.py 双跑文本产物字节一致（确定性保持）。
- 重建后必须重跑 tools/fingerprint.py 刷新清单（README 审计节已写明）。

## Artifact fingerprints (this build)

- page.svg SHA-256: `5f2ab6f821e1cc2e142d9894eaf4255a5c586270595dbb71e9d27573f6dd385e`
- page.svg bytes: 12509
- index.html SHA-256: `23c371640ffa05797f210ebb84f3a72b1831abb2f116529c979aeb945486b808`
- index.html bytes: 14991
- index.html inlines page.svg byte-identically; zero external references
- proof.png is produced by `rsvg-convert --width=2560` (2x of 1280 x 720)
- CJK font at render: Source Han Serif SC

## Page code-coordinate sweep

Zero matches for file:line, type names, or function names on page.svg / index.html.
