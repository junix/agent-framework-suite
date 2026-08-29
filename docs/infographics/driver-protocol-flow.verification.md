# driver-protocol-flow — 验证清单（Verification Manifest）

## 视觉契约

- **读者**：评估或为 agent-framework-suite 做贡献的工程师。
- **论点（一句话）**：同一个语言无关用例，经两个隔离子进程分别驱动 Python 与 Rust
  框架的公开 API，产出两份可比较的 workflow-core/v1 观测，再经 oracle 与 parity
  双保险归约为退出码。
- **读者问题**：一次 `just check` 如何把同一份用例变成对两个框架实现的行为判定？
- **与既有图的关系**：`docs/architecture-infographic.tex` 是六段式长图总览；本图为
  紧凑单论点流程图，主角是 driver 协议往返与「归一化边界」（抹平 / 去掉 / 保留）。
- **语言**：简体中文；代码标识符、文件路径、契约名保持原文拼写。
- **媒介**：单页紧凑 TikZ 图（约 17×17cm），复用 `docs/styles/showcase.sty` 样式令牌。
- **叙事顺序**：1 协议握手（`<driver> version`）→ 2 `run CASE_ID` 调用 ×2 →
  3 stdout 恰好一个 JSON 观测 → 归一化（抹平生成 ID/时间戳/原生枚举异常名，去掉
  participant，保留业务信号）→ 4 Oracle 验证 → 5 Parity 比对 → 终态
  （exit 0 / exit 1；exit 3 与 exit 2 以注释枚举）。

## 证据来源（全部经源码核实）

- `README.md`：CLI、退出码 0/1/2/3、hermetic 约束、`--timeout` 默认 30s。
- `CONTRACT.md`（workflow-core/v1）：归一化观测字段、parity 规则（仅去掉
  `participant` 后逐字段比对）、强制双参与者（缺失非 skip）。
- `DRIVER_PROTOCOL.md`：`version` / `run CASE_ID` 协议、stdout 恰好一个 JSON、
  driver 侧归一化范围。
- `internal/driver/client.go`：子进程环境（临时 HOME、TZ=UTC、C.UTF-8、NO_COLOR=1、
  stdout/stderr 各限 64KB）、`decodeOne` 只接受恰好一个 JSON 值。
- `internal/suite/runner.go`：oracle = `reflect.DeepEqual(期望, 归一化观测)`；
  parity = 两端归一化（去掉 participant）后 `reflect.DeepEqual`。
- `CASES.md` / `internal/suite/catalog.go`：10 个 WF-NNN 用例。

## 产物

| 文件 | 说明 | SHA-256 |
|---|---|---|
| `driver-protocol-flow.tex` | 规范可编辑源（canonical） | — |
| `driver-protocol-flow.pdf` | 矢量交付（1 页，字体嵌入） | `d33bda41…5a3881` |
| `driver-protocol-flow.png` | 300dpi 位图校对（2037×2061，不透明 Paper 底） | `4aaa4855…d87e` |
| `driver-protocol-flow.svg` | SVG 交付（pdftocairo 导出，path-font） | `68abd4cc…fff12` |
| `driver-protocol-flow.render.json` | 渲染器完整清单（三道 strict gate 全绿时产物） | — |

## 构建命令（便携）

```bash
cd docs
xelatex -interaction=nonstopmode -output-directory=infographics infographics/driver-protocol-flow.tex
pdftocairo -png -r 300 -singlefile infographics/driver-protocol-flow.pdf infographics/driver-protocol-flow
pdftocairo -svg infographics/driver-protocol-flow.pdf infographics/driver-protocol-flow.svg
```

注意：必须从 `docs/` 编译（`\usepackage{styles/showcase}` 相对它解析）。请勿改回
`\input{../styles/showcase}`——相对路径 `\input` 会触发 `LaTeX Warning: You have
requested package ''`，破坏 strict-warnings。

完整门禁复现（技能脚本，含 warnings/geometry/composition/svg 四道 strict gate）：

```bash
python3 ~/.agents/skills/create-infographic/scripts/render_tikz.py \
  docs/infographics/driver-protocol-flow.tex --workdir docs \
  --out-dir /tmp/tikz-figure --engine xelatex --dpi 300 --svg \
  --strict-warnings --strict-geometry --strict-composition --strict-svg \
  --min-labelled-connectors 0.6 --expect-text "一份用例"
```

## 门禁结果（2026-08-29，最终交付哈希同上）

- `lint_tikz_source.py`：clean。
- strict-warnings：0 条（无 Missing character / Overfull / 未定义引用）。
- strict-geometry：pass——0 文本行重叠、0 低于 6.5pt 文本、identity 匹配「一份用例」、
  页面 1 页、字体嵌入。
- strict-composition：0 violation；带标签连线比例 0.733（阈值 0.6）；
  raster.available=true（ink_density 0.07、quiet_cell_fraction 0.25、cell_density_cv 0.60）。
- strict-svg（svg-linter 0.1.0，catalog v3 `sha256:9079749f…`）：
  - error gate（`svg/duplicate-id`,`svg/dangling-reference`，--require-complete）：
    0 findings，两条规则均完整评估（交付文件复跑 exit 0）；
  - hygiene review（`svg/excessive-path-complexity`,`svg/unused-definition`,
    `svg/external-resource`,`svg/raster-upscale`,`svg/open-filled-path`）：0 findings；
  - collision review（`svg/line-overlap`,`svg/line-crossing`,`svg/line-text-overlap`,
    `svg/text-text-overlap`,`svg/canvas-edge-clearance`,`svg/ambiguous-junction`,
    `svg/connector-through-shape`,`svg/edge-congestion`,`svg/filter-clipping-risk`,
    `svg/low-contrast-text`）：0 findings。
- 人工审查：未标注 PNG 与编号 lint PNG 均在交付尺寸过目；标题/连线/端口/分支标签/
  图例/页脚无碰撞；CJK 字形完整。

## 迭代中已修复并复验的缺陷（均有裁剪图证据）

1. hero 归一化面板与 Oracle 盒重叠、与步骤 3 标题文字互叠 → 右列整体下移拉开。
2. `2 · run CASE_ID` 水平标签被斜向连线穿过（两次） → 改为 sloped 沿线标签，
   垂直偏移 1.5mm。
3. `4 · 逐端验证` 标签压在 Oracle 盒顶边 → 判定链节点下移，间距 ≥7.5mm。
4. `一致` 标签距 parity→passed 连线过近（svg/line-text-overlap）→ 偏移加大至 2.5mm。
5. 页脚文字贴近背景圆角弧 → 页脚上移左移、背景底边下移、圆角改 1.5mm。
6. `\input{../styles/showcase}` 触发 `requested package ''` 警告 → 改
   `\usepackage{styles/showcase}` 并以 `docs/` 为工作目录。

## 隔离与导出声明

- SVG 为 path-font 导出（pdftocairo）：`svg/out-of-bounds`、`svg/outlined-text` 等
  路径字体隔离规则未纳入门禁（按技能参考 svg-visual-qa.md 的 quarantine 清单）；
  越界/缺字信号由 TeX 侧 strict-warnings（Overfull / Missing character）覆盖。
- 全部修复均在规范源 `.tex` 完成并重新生成所有派生产物，未手工编辑 SVG。
