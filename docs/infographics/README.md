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
| `index.html` | 单文件页：SVG 内联，零外部引用，1200px 居中 |
| `data/provenance.json` | 页上主张的源码锚点 |
| `data/claims.json` | 声明号登记：Cxx 与主张、用例号、数字豁免 |
| `VERIFICATION.md` | 门禁与目视 |
| `proof.png` | `rsvg-convert` 栅格校样 |
| `tools/audit_gates.py` | 审计门：声明绑定、数字反扫、代码坐标、svg-linter、毒丸 |
| `tools/fingerprint.py` | 全树指纹（豁免 fingerprints.json 与 data/audit/） |
| `tools/vacuum_rebuild.py` | 删除重建对账（真空电池） |
| `fingerprints.json` | 全树 SHA-256 清单 |
| `data/audit/` | 门禁运行记录（指纹豁免） |

改图时改 `build.py` 再写回。CJK 由 Python UTF-8 写入。

## 重建

```bash
python3 docs/infographics/build.py
rsvg-convert --width=2560 docs/infographics/page.svg -o docs/infographics/proof.png
```

`index.html` 内联 page.svg 字节，零外部引用；页面 1200px 居中。
`page.svg` 仍为可重建源。不启动 browser-harness。

## 审计

```bash
python3 docs/infographics/tools/audit_gates.py all --record
python3 docs/infographics/tools/audit_gates.py pills --record
python3 docs/infographics/tools/fingerprint.py
python3 docs/infographics/tools/fingerprint.py --check
python3 docs/infographics/tools/vacuum_rebuild.py
```

重建后必须重跑 `tools/fingerprint.py` 刷新清单；README/VERIFICATION 引用的树内哈希以 `fingerprints.json` 为准。
