# Visual contract

- audience: engineers judging whether the suite is a case catalog or a conformance path
- claim: 一条路径，不是目录。同一用例经公开 API 分别驱动 Python 与 Rust。共用一份期望，再强制比对。驱动只翻译，不实现路由或检查点。
- reader question: 这套件的形状是什么？驱动会不会自己实现路由与检查点？
- language: zh-CN
- grammar: one directional process spine with fork/join
- protagonist: Y-fork path — case splits to two public APIs, joins at shared expected, then parity
- canvas: 1280 × 720 (doc-wide / slide-16x9)
- medium: hand-placed editorial SVG inlined into single-page index.html (1200px-centred wrapper); page.svg stays the rebuildable source; rsvg-convert proof
- exceptions: missing participant is failure, not skip; fail branch at parity
- code detail: function/type/file names stay in provenance; case IDs, just verbs, Python/Rust may appear
- not this page: protocol handshake, JSON-on-stdout, normalize wipe/drop/keep, exit-code table (see driver-protocol-flow)
