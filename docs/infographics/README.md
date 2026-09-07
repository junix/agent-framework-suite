# agent-framework-suite 信息图目录

已有 `driver-protocol-flow`：协议握手、归一化与退出码。本页论点不同：套件是一条封闭双驱路径，不是用例目录；驱动不实现路由或检查点。

## 论点

一条路径，不是目录—同一用例经公开 API 分别驱动 Python 与 Rust。共用一份期望，再强制比对。驱动只翻译，不实现路由或检查点。

## 索引

| 文件 | 论点 | 处置 |
|---|---|---|
| `docs/architecture-infographic.tex` 与 `docs/architecture.html` | 六段式分层 + 用例目录（旧） | 原样保留，本波不编辑 |
| `driver-protocol-flow.*` | 一份用例，两份观测，一次比对（协议 / 归一化 / 退出码） | 既有 explainer，原样保留 |
| `page.svg` | Class A：用例 → 双公开 API → 共用期望 → 强制比对 | 本页 |

## 本页材料

| 文件 | 职责 |
|---|---|
| `contract.md` | 读者、论点、语法、尺寸 |
| `page.svg` | 封闭双驱路径 |
| `index.html` | 薄页，外挂 SVG，零 CDN |
| `data/provenance.json` | 页上主张的源码锚点 |
| `VERIFICATION.md` | 门禁与目视 |
| `proof.png` | `rsvg-convert` 栅格校样 |

改图时改 `build.py` 再写回。CJK 由 Python UTF-8 写入。

## 重建

```bash
python3 docs/infographics/build.py
rsvg-convert --width=2560 docs/infographics/page.svg -o docs/infographics/proof.png
```

页面是外部 SVG；`index.html` 只是薄包装。
不启动 browser-harness。
