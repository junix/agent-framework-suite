#!/usr/bin/env python3
"""Generate the agent-framework-suite Class A architecture remake.

CJK lives in unicode escapes so this file stays Write-tool safe.
Rebuild from the repository root:

    python3 docs/infographics/build.py
    rsvg-convert --width=2560 docs/infographics/page.svg -o docs/infographics/proof.png
"""

from __future__ import annotations

import hashlib
import json
import math
import xml.sax.saxutils
from pathlib import Path

from PIL import ImageFont

ROOT = Path(__file__).resolve().parent
FONT_REG = Path("/Users/junix/Library/Fonts/SourceHanSerifSC-Regular.otf")
FONT_MED = Path("/Users/junix/Library/Fonts/SourceHanSerifSC-Medium.otf")
FONT_BLD = Path("/Users/junix/Library/Fonts/SourceHanSerifSC-Bold.otf")
FONT_FAMILY = "Source Han Serif SC, Songti SC, serif"

PAPER = "#F7F4EE"
INK = "#17212B"
SLATE = "#5D6873"
FOG = "#D9E1E3"
OCEAN = "#356A79"
TEAL = "#2A9D8F"
MINT = "#DDF2EC"
CORAL = "#E76F51"
GOLD = "#E9C46A"

W, H = 1280, 720
SAFE = 48

# Reader-visible Chinese as escapes. Product names and just verbs stay Latin.
T = {
    "kicker": "agent-framework-suite \u00b7 Class A",
    "title": "\u4e00\u6761\u8def\u5f84\uff0c\u4e0d\u662f\u76ee\u5f55",
    "lede_1": (
        "\u540c\u4e00\u7528\u4f8b\u7ecf\u516c\u5f00 API \u5206\u522b\u9a71\u52a8 "
        "Python \u4e0e Rust\u3002"
    ),
    "lede_2": (
        "\u5171\u7528\u4e00\u4efd\u671f\u671b\uff0c\u518d\u5f3a\u5236\u6bd4\u5bf9\u3002"
        "\u9a71\u52a8\u53ea\u7ffb\u8bd1\uff0c\u4e0d\u5b9e\u73b0\u8def\u7531\u6216\u68c0\u67e5\u70b9\u3002"
    ),
    "finger": "hermetic \u00b7 \u65e0\u6a21\u578b \u00b7 \u65e0\u7f51\u7edc",
    "ids": (
        "WF-001 \u00b7 WF-002 \u00b7 WF-003 \u00b7 WF-004 \u00b7 WF-005 \u00b7 "
        "WF-006 \u00b7 WF-007 \u00b7 WF-008 \u00b7 WF-009 \u00b7 WF-010"
    ),
    "n_case": "\u7528\u4f8b",
    "n_py": "Python",
    "n_rs": "Rust",
    "n_api": "\u516c\u5f00 API",
    "n_oracle": "\u5171\u7528\u671f\u671b",
    "n_parity": "\u5f3a\u5236\u6bd4\u5bf9",
    "n_pass": "\u901a\u8fc7",
    "n_fail": "\u5931\u8d25",
    "e_dispatch": "just check \u00b7 \u5206\u53d1\u540c\u4e00\u4efd",
    "e_drive": "\u9a71\u52a8\u516c\u5f00 API",
    "e_obs": "\u5199\u51fa\u89c2\u6d4b",
    "e_check": "\u9010\u7aef\u6838\u5bf9",
    "e_same": "\u4e00\u81f4",
    "e_split": "\u5206\u53c9",
    "bound": "\u865a\u7ebf\u4ee5\u4e0a\u53ea\u7ffb\u8bd1\uff1b\u8def\u7531\u4e0e\u68c0\u67e5\u70b9\u5728\u5e93\u5185",
    "both": "\u4e24\u7aef\u5747\u4e3a\u5f3a\u5236",
    "miss": "\u7f3a\u4e00\u7aef\u4e0d\u662f\u8df3\u8fc7",
    "colophon": (
        "\u4f9d\u636e README \u4e0e\u6e90\u7801\u3002"
        "\u4ee3\u7801\u5750\u6807\u89c1\u672c\u76ee\u5f55\u6838\u9a8c\u8bb0\u5f55\u3002"
        "\u534f\u8bae\u7ec6\u8282\u89c1\u65e2\u6709\u6d41\u7a0b\u56fe\u3002"
        "\u65e7\u5206\u5c42\u56fe\u53e6\u5b58\u3002"
    ),
    "title_attr": (
        "agent-framework-suite\uff1a\u4e00\u6761\u8def\u5f84\uff0c\u4e0d\u662f\u76ee\u5f55\u3002"
        "\u540c\u4e00\u7528\u4f8b\u7ecf\u516c\u5f00 API \u5206\u522b\u9a71\u52a8 "
        "Python \u4e0e Rust\uff0c\u5171\u7528\u671f\u671b\uff0c\u5f3a\u5236\u6bd4\u5bf9\u3002"
        "\u9a71\u52a8\u4e0d\u5b9e\u73b0\u8def\u7531\u6216\u68c0\u67e5\u70b9\u3002"
    ),
    "desc_attr": (
        "just check \u628a\u7528\u4f8b\u5206\u53d1\u5230 Python \u4e0e Rust\u3002"
        "\u4e24\u7aef\u53ea\u8c03\u516c\u5f00 API\uff0c\u5199\u51fa\u89c2\u6d4b\u3002"
        "\u5148\u5bf9\u5171\u7528\u671f\u671b\u9010\u7aef\u6838\u5bf9\uff0c\u518d\u5f3a\u5236\u6bd4\u5bf9\u3002"
        "\u4e00\u81f4\u5219\u901a\u8fc7\uff0c\u5206\u53c9\u6216\u7f3a\u4e00\u7aef\u5219\u5931\u8d25\u3002"
        "\u7528\u4f8b\u4e3a WF-001 \u81f3 WF-010\u3002"
        "\u5c01\u95ed\u6267\u884c\uff0c\u65e0\u6a21\u578b\u3001\u65e0\u7f51\u7edc\u3002"
    ),
}


def _font(path: Path, size: float) -> ImageFont.FreeTypeFont:
    return ImageFont.truetype(str(path), size)


def measure(text: str, size: float, weight: str = "reg") -> tuple[float, float]:
    path = {"reg": FONT_REG, "med": FONT_MED, "bld": FONT_BLD}[weight]
    font = _font(path, size)
    box = font.getbbox(text)
    return box[2] - box[0], box[3] - box[1]


def esc(text: str) -> str:
    return xml.sax.saxutils.escape(text)


def routed(points: list[tuple[float, float]], r: float = 8.0) -> str:
    if len(points) < 2:
        raise ValueError("need two points")
    cleaned = [points[0]]
    for p in points[1:]:
        if abs(p[0] - cleaned[-1][0]) > 0.01 or abs(p[1] - cleaned[-1][1]) > 0.01:
            cleaned.append(p)
    pts = cleaned
    if len(pts) == 2:
        return f"M {pts[0][0]:.1f},{pts[0][1]:.1f} L {pts[1][0]:.1f},{pts[1][1]:.1f}"
    parts = [f"M {pts[0][0]:.1f},{pts[0][1]:.1f}"]
    for i in range(1, len(pts) - 1):
        x0, y0 = pts[i - 1]
        x1, y1 = pts[i]
        x2, y2 = pts[i + 1]
        d1 = math.hypot(x1 - x0, y1 - y0)
        d2 = math.hypot(x2 - x1, y2 - y1)
        rr = min(r, d1 / 2.2, d2 / 2.2)
        ux, uy = (x1 - x0) / d1, (y1 - y0) / d1
        vx, vy = (x2 - x1) / d2, (y2 - y1) / d2
        sx, sy = x1 - ux * rr, y1 - uy * rr
        ex, ey = x1 + vx * rr, y1 + vy * rr
        parts.append(f"L {sx:.1f},{sy:.1f}")
        parts.append(f"Q {x1:.1f},{y1:.1f} {ex:.1f},{ey:.1f}")
    parts.append(f"L {pts[-1][0]:.1f},{pts[-1][1]:.1f}")
    return " ".join(parts)


def text_el(
    x: float,
    y: float,
    text: str,
    *,
    size: float,
    fill: str,
    weight: str = "reg",
    anchor: str = "start",
    baseline: str = "alphabetic",
    eid: str | None = None,
) -> str:
    fw = {"reg": 400, "med": 500, "bld": 700}[weight]
    attrs = [
        f'x="{x:.1f}"',
        f'y="{y:.1f}"',
        f'font-size="{size}"',
        f'font-weight="{fw}"',
        f'fill="{fill}"',
        f'font-family="{FONT_FAMILY}"',
        f'text-anchor="{anchor}"',
        f'dominant-baseline="{baseline}"',
    ]
    if eid:
        attrs.append(f'id="{eid}"')
    return f"<text {' '.join(attrs)}>{esc(text)}</text>"


def circle(cx: float, cy: float, r: float, fill: str, stroke: str, sw: float, eid: str) -> str:
    return (
        f'<circle id="{eid}" cx="{cx:.1f}" cy="{cy:.1f}" r="{r:.1f}" '
        f'fill="{fill}" stroke="{stroke}" stroke-width="{sw}"/>'
    )


def path_el(
    d: str,
    *,
    stroke: str,
    sw: float,
    eid: str,
    dashed: bool = False,
    marker: str | None = None,
) -> str:
    dash = ' stroke-dasharray="6 5"' if dashed else ""
    mk = f' marker-end="url(#{marker})"' if marker else ""
    return (
        f'<path id="{eid}" d="{d}" fill="none" stroke="{stroke}" '
        f'stroke-width="{sw}" stroke-linecap="round" stroke-linejoin="round"'
        f"{dash}{mk}/>"
    )


def build_svg() -> str:
    case_c = (600.0, 232.0)
    py_c = (260.0, 356.0)
    rs_c = (980.0, 388.0)
    oracle_c = (620.0, 488.0)
    parity_c = (620.0, 572.0)
    pass_c = (480.0, 632.0)
    fail_c = (760.0, 632.0)
    r_case, r_arm, r_oracle, r_parity, r_end = 18.0, 16.0, 50.0, 14.0, 11.0

    case_s = case_c[1] + r_case
    case_x0 = case_c[0] - r_case
    py_port = case_x0 + (2 * r_case) * 1 / 3
    rs_port = case_x0 + (2 * r_case) * 2 / 3
    py_n = (py_c[0], py_c[1] - r_arm)
    rs_n = (rs_c[0], rs_c[1] - r_arm)
    py_e = (py_c[0] + r_arm, py_c[1])
    rs_w = (rs_c[0] - r_arm, rs_c[1])
    oracle_w = (oracle_c[0] - r_oracle, oracle_c[1])
    oracle_e = (oracle_c[0] + r_oracle, oracle_c[1])
    oracle_s = (oracle_c[0], oracle_c[1] + r_oracle)
    parity_n = (parity_c[0], parity_c[1] - r_parity)
    parity_s = parity_c[1] + r_parity
    parity_x0 = parity_c[0] - r_parity
    pass_port = parity_x0 + (2 * r_parity) * 1 / 3
    fail_port = parity_x0 + (2 * r_parity) * 2 / 3
    pass_n = (pass_c[0], pass_c[1] - r_end)
    fail_n = (fail_c[0], fail_c[1] - r_end)

    fork_py_y = 286.0
    fork_rs_y = 310.0
    py_mid_x = (py_e[0] + oracle_w[0]) / 2
    rs_mid_x = (rs_w[0] + oracle_e[0]) / 2
    term_y = 608.0
    bound_y = 416.0

    c_py = routed([(py_port, case_s), (py_port, fork_py_y), (py_n[0], fork_py_y), py_n])
    c_rs = routed([(rs_port, case_s), (rs_port, fork_rs_y), (rs_n[0], fork_rs_y), rs_n])
    c_py_obs = routed([py_e, (py_mid_x, py_c[1]), (py_mid_x, oracle_c[1]), oracle_w])
    c_rs_obs = routed([rs_w, (rs_mid_x, rs_c[1]), (rs_mid_x, oracle_c[1]), oracle_e])
    c_check = routed([oracle_s, parity_n])
    c_pass = routed(
        [(pass_port, parity_s), (pass_port, term_y), (pass_n[0], term_y), pass_n]
    )
    c_fail = routed(
        [(fail_port, parity_s), (fail_port, term_y), (fail_n[0], term_y), fail_n]
    )

    parts: list[str] = [
        (
            f'<svg xmlns="http://www.w3.org/2000/svg" width="{W}" height="{H}" '
            f'viewBox="0 0 {W} {H}" role="img" xml:lang="zh-CN" '
            f'aria-labelledby="fig-title fig-desc">'
        ),
        f'<title id="fig-title">{esc(T["title_attr"])}</title>',
        f'<desc id="fig-desc">{esc(T["desc_attr"])}</desc>',
        "<defs>",
        (
            f'<marker id="mk-flow" viewBox="0 0 10 8" refX="9" refY="4" '
            f'markerWidth="8" markerHeight="6.4" orient="auto">'
            f'<path d="M 0 0 L 10 4 L 0 8 z" fill="{OCEAN}"/></marker>'
        ),
        (
            f'<marker id="mk-warn" viewBox="0 0 10 8" refX="9" refY="4" '
            f'markerWidth="8" markerHeight="6.4" orient="auto">'
            f'<path d="M 0 0 L 10 4 L 0 8 z" fill="{CORAL}"/></marker>'
        ),
        "</defs>",
        f'<rect id="bg" x="0" y="0" width="{W}" height="{H}" fill="{PAPER}"/>',
    ]

    parts.append(text_el(SAFE, 52, T["kicker"], size=12, fill=OCEAN, weight="bld", eid="kicker"))
    parts.append(text_el(SAFE, 92, T["title"], size=32, fill=INK, weight="bld", eid="title"))
    parts.append(text_el(SAFE, 124, T["lede_1"], size=15, fill=INK, eid="lede-1"))
    parts.append(text_el(SAFE, 148, T["lede_2"], size=15, fill=INK, eid="lede-2"))
    parts.append(text_el(1232, 52, T["finger"], size=12, fill=SLATE, anchor="end", eid="finger"))
    parts.append(text_el(SAFE, 180, T["ids"], size=12, fill=SLATE, eid="ids"))

    # Tint under the shared oracle. No stroke: not a container. Kept
    # below the boundary so the rule does not transit the wash.
    parts.append(
        f'<ellipse id="zone-oracle" cx="{oracle_c[0]:.1f}" cy="{oracle_c[1]:.1f}" '
        f'rx="66" ry="54" fill="{MINT}"/>'
    )

    # Boundary gapped where observations cross into the shared expected.
    gap = 28.0
    parts.append(
        f'<line id="rule-a" x1="120" y1="{bound_y:.1f}" x2="{py_mid_x - gap:.1f}" '
        f'y2="{bound_y:.1f}" stroke="{OCEAN}" stroke-width="1.4" stroke-dasharray="7 5"/>'
    )
    parts.append(
        f'<line id="rule-b" x1="{py_mid_x + gap:.1f}" y1="{bound_y:.1f}" '
        f'x2="{rs_mid_x - gap:.1f}" y2="{bound_y:.1f}" stroke="{OCEAN}" '
        f'stroke-width="1.4" stroke-dasharray="7 5"/>'
    )
    parts.append(
        f'<line id="rule-c" x1="{rs_mid_x + gap:.1f}" y1="{bound_y:.1f}" x2="1160" '
        f'y2="{bound_y:.1f}" stroke="{OCEAN}" stroke-width="1.4" stroke-dasharray="7 5"/>'
    )
    parts.append(
        text_el(oracle_c[0], bound_y - 12, T["bound"], size=12, fill=SLATE, anchor="middle", eid="bound")
    )

    parts.append(path_el(c_py, stroke=OCEAN, sw=2.2, eid="e-py", marker="mk-flow"))
    parts.append(path_el(c_rs, stroke=OCEAN, sw=2.2, eid="e-rs", marker="mk-flow"))
    parts.append(path_el(c_py_obs, stroke=OCEAN, sw=2.4, eid="e-py-obs", marker="mk-flow"))
    parts.append(path_el(c_rs_obs, stroke=OCEAN, sw=2.4, eid="e-rs-obs", marker="mk-flow"))
    parts.append(path_el(c_check, stroke=OCEAN, sw=2.8, eid="e-check", marker="mk-flow"))
    parts.append(path_el(c_pass, stroke=OCEAN, sw=1.8, eid="e-pass", marker="mk-flow"))
    parts.append(
        path_el(c_fail, stroke=CORAL, sw=1.8, eid="e-fail", dashed=True, marker="mk-warn")
    )

    # Labels sit 8-12 px off the open segment.
    parts.append(
        text_el(case_c[0] - r_case - 10, case_c[1] + 4, T["n_case"], size=13, fill=INK, weight="bld", anchor="end", eid="lab-case")
    )
    parts.append(
        text_el(case_c[0] + r_case + 12, case_c[1] + 4, T["e_dispatch"], size=12, fill=INK, eid="lab-dispatch")
    )
    parts.append(
        text_el(py_c[0] - 18, fork_py_y + 22, T["e_drive"], size=12, fill=INK, anchor="end", eid="lab-drive-py")
    )
    parts.append(
        text_el(rs_c[0] + 18, fork_rs_y + 22, T["e_drive"], size=12, fill=INK, eid="lab-drive-rs")
    )
    parts.append(
        text_el(640, 328, T["both"], size=12, fill=SLATE, anchor="middle", eid="lab-both")
    )
    parts.append(
        text_el(py_mid_x + 12, 456, T["e_obs"], size=12, fill=INK, eid="lab-obs-py")
    )
    parts.append(
        text_el(rs_mid_x + 12, 456, T["e_obs"], size=12, fill=INK, eid="lab-obs-rs")
    )
    parts.append(
        text_el(oracle_c[0] + r_oracle + 12, (oracle_s[1] + parity_n[1]) / 2 + 4, T["e_check"], size=12, fill=INK, eid="lab-check")
    )
    parts.append(text_el(548, term_y - 10, T["e_same"], size=12, fill=INK, eid="lab-same"))
    parts.append(text_el(688, term_y - 10, T["e_split"], size=12, fill=INK, eid="lab-split"))
    parts.append(
        text_el(py_c[0] - r_arm - 12, py_c[1] - 4, T["n_py"], size=13, fill=INK, weight="bld", anchor="end", eid="lab-py")
    )
    parts.append(
        text_el(py_c[0] - r_arm - 12, py_c[1] + 14, T["n_api"], size=12, fill=SLATE, anchor="end", eid="lab-py-api")
    )
    parts.append(
        text_el(rs_c[0] + r_arm + 12, rs_c[1] - 4, T["n_rs"], size=13, fill=INK, weight="bld", eid="lab-rs")
    )
    parts.append(
        text_el(rs_c[0] + r_arm + 12, rs_c[1] + 14, T["n_api"], size=12, fill=SLATE, eid="lab-rs-api")
    )
    parts.append(
        text_el(
            parity_c[0] + r_parity + 12,
            parity_c[1] + 5,
            T["n_parity"],
            size=13,
            fill=INK,
            weight="bld",
            eid="lab-parity",
        )
    )
    parts.append(
        text_el(pass_c[0] - r_end - 10, pass_c[1] + 4, T["n_pass"], size=12, fill=INK, weight="bld", anchor="end", eid="lab-pass")
    )
    parts.append(
        text_el(fail_c[0] - r_end - 10, fail_c[1] + 4, T["n_fail"], size=12, fill=INK, weight="bld", anchor="end", eid="lab-fail")
    )
    parts.append(
        text_el(fail_c[0] + r_end + 12, fail_c[1] + 4, T["miss"], size=12, fill=SLATE, eid="lab-miss")
    )

    # Nodes last.
    parts.append(circle(*case_c, r_case, OCEAN, OCEAN, 0, "n-case"))
    parts.append(
        text_el(case_c[0], case_c[1] + 5, "01", size=12, fill=PAPER, weight="bld", anchor="middle", eid="num-01")
    )
    parts.append(circle(*py_c, r_arm, OCEAN, OCEAN, 0, "n-py"))
    parts.append(
        text_el(py_c[0], py_c[1] + 5, "02", size=12, fill=PAPER, weight="bld", anchor="middle", eid="num-02")
    )
    parts.append(circle(*rs_c, r_arm, OCEAN, OCEAN, 0, "n-rs"))
    parts.append(
        text_el(rs_c[0], rs_c[1] + 5, "03", size=12, fill=PAPER, weight="bld", anchor="middle", eid="num-03")
    )
    parts.append(circle(*oracle_c, r_oracle, OCEAN, OCEAN, 0, "n-oracle"))
    parts.append(
        text_el(oracle_c[0], oracle_c[1] - 8, "04", size=16, fill=PAPER, weight="bld", anchor="middle", eid="num-04")
    )
    parts.append(
        text_el(
            oracle_c[0],
            oracle_c[1] + 16,
            T["n_oracle"],
            size=15,
            fill=PAPER,
            weight="bld",
            anchor="middle",
            eid="n-oracle-t",
        )
    )
    parts.append(circle(*parity_c, r_parity, OCEAN, OCEAN, 0, "n-parity"))
    parts.append(
        text_el(parity_c[0], parity_c[1] + 5, "05", size=11, fill=PAPER, weight="bld", anchor="middle", eid="num-05")
    )
    parts.append(circle(*pass_c, r_end, PAPER, OCEAN, 1.8, "n-pass"))
    parts.append(circle(*fail_c, r_end, PAPER, CORAL, 1.8, "n-fail"))

    parts.append(text_el(SAFE, 688, T["colophon"], size=12, fill=SLATE, eid="colophon"))
    parts.append("</svg>")
    return "\n".join(parts) + "\n"


def write_utf8(path: Path, text: str) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(text, encoding="utf-8")


def build_index_html(svg: str) -> str:
    # The page ships the SVG inline (byte-identical to page.svg, which stays
    # the rebuildable source) so index.html has zero external references and
    # renders 1200 CSS px wide centred (width rule, CSS-only).
    title = T["title"]
    return (
        "<!DOCTYPE html>\n"
        '<html lang="zh-CN">\n'
        "<head>\n"
        '<meta charset="utf-8"/>\n'
        f"<title>{esc(title)} \u00b7 agent-framework-suite</title>\n"
        "<style>html{margin:0;background:#F7F4EE} "
        "body{margin:0 auto;max-width:1200px;background:#F7F4EE} "
        "svg{display:block;width:100%;height:auto}</style>\n"
        "</head>\n"
        "<body>\n"
        + svg
        + "</body>\n"
        "</html>\n"
    )


def build_readme() -> str:
    return "\n".join(
        [
            "# agent-framework-suite \u4fe1\u606f\u56fe\u76ee\u5f55",
            "",
            "\u5df2\u6709 `driver-protocol-flow`\uff1a\u534f\u8bae\u63e1\u624b\u3001"
            "\u5f52\u4e00\u5316\u4e0e\u9000\u51fa\u7801\u3002"
            "\u672c\u9875\u8bba\u70b9\u4e0d\u540c\uff1a\u5957\u4ef6\u662f\u4e00\u6761\u5c01\u95ed\u53cc\u9a71\u8def\u5f84\uff0c"
            "\u4e0d\u662f\u7528\u4f8b\u76ee\u5f55\uff1b\u9a71\u52a8\u4e0d\u5b9e\u73b0\u8def\u7531\u6216\u68c0\u67e5\u70b9\u3002",
            "",
            "## \u8bba\u70b9",
            "",
            T["title"] + "\u2014" + T["lede_1"] + T["lede_2"],
            "",
            "## \u7d22\u5f15",
            "",
            "| \u6587\u4ef6 | \u8bba\u70b9 | \u5904\u7f6e |",
            "|---|---|---|",
            (
                "| `docs/architecture-infographic.tex` "
                "\u4e0e `docs/architecture.html` | \u516d\u6bb5\u5f0f\u5206\u5c42 + \u7528\u4f8b\u76ee\u5f55\uff08\u65e7\uff09 | "
                "\u539f\u6837\u4fdd\u7559\uff0c\u672c\u6ce2\u4e0d\u7f16\u8f91 |"
            ),
            (
                "| `driver-protocol-flow.*` | \u4e00\u4efd\u7528\u4f8b\uff0c\u4e24\u4efd\u89c2\u6d4b\uff0c"
                "\u4e00\u6b21\u6bd4\u5bf9\uff08\u534f\u8bae / \u5f52\u4e00\u5316 / \u9000\u51fa\u7801\uff09 | "
                "\u65e2\u6709 explainer\uff0c\u539f\u6837\u4fdd\u7559 |"
            ),
            (
                "| `page.svg` | Class A\uff1a\u7528\u4f8b \u2192 \u53cc\u516c\u5f00 API \u2192 "
                "\u5171\u7528\u671f\u671b \u2192 \u5f3a\u5236\u6bd4\u5bf9 | \u672c\u9875 |"
            ),
            "",
            "## \u672c\u9875\u6750\u6599",
            "",
            "| \u6587\u4ef6 | \u804c\u8d23 |",
            "|---|---|",
            "| `contract.md` | \u8bfb\u8005\u3001\u8bba\u70b9\u3001\u8bed\u6cd5\u3001\u5c3a\u5bf8 |",
            "| `page.svg` | \u5c01\u95ed\u53cc\u9a71\u8def\u5f84 |",
            "| `index.html` | \u5355\u6587\u4ef6\u9875\uff1aSVG \u5185\u8054\uff0c\u96f6\u5916\u90e8\u5f15\u7528\uff0c1200px \u5c45\u4e2d |",
            "| `data/provenance.json` | \u9875\u4e0a\u4e3b\u5f20\u7684\u6e90\u7801\u951a\u70b9 |",
            "| `VERIFICATION.md` | \u95e8\u7981\u4e0e\u76ee\u89c6 |",
            "| `proof.png` | `rsvg-convert` \u6805\u683c\u6821\u6837 |",
            "",
            "\u6539\u56fe\u65f6\u6539 `build.py` \u518d\u5199\u56de\u3002CJK \u7531 Python UTF-8 \u5199\u5165\u3002",
            "",
            "## \u91cd\u5efa",
            "",
            "```bash",
            "python3 docs/infographics/build.py",
            "rsvg-convert --width=2560 docs/infographics/page.svg -o docs/infographics/proof.png",
            "```",
            "",
            "`index.html` \u5185\u8054 page.svg \u5b57\u8282\uff0c\u96f6\u5916\u90e8\u5f15\u7528\uff1b\u9875\u9762 1200px \u5c45\u4e2d\u3002",
            "`page.svg` \u4ecd\u4e3a\u53ef\u91cd\u5efa\u6e90\u3002\u4e0d\u542f\u52a8 browser-harness\u3002",
            "",
        ]
    )


def build_contract() -> str:
    return "\n".join(
        [
            "# Visual contract",
            "",
            "- audience: engineers judging whether the suite is a case catalog or a conformance path",
            f"- claim: {T['title']}\u3002{T['lede_1']}{T['lede_2']}",
            "- reader question: \u8fd9\u5957\u4ef6\u7684\u5f62\u72b6\u662f\u4ec0\u4e48\uff1f\u9a71\u52a8\u4f1a\u4e0d\u4f1a\u81ea\u5df1\u5b9e\u73b0\u8def\u7531\u4e0e\u68c0\u67e5\u70b9\uff1f",
            "- language: zh-CN",
            "- grammar: one directional process spine with fork/join",
            "- protagonist: Y-fork path \u2014 case splits to two public APIs, joins at shared expected, then parity",
            "- canvas: 1280 \u00d7 720 (doc-wide / slide-16x9)",
            "- medium: hand-placed editorial SVG inlined into single-page index.html (1200px-centred wrapper); page.svg stays the rebuildable source; rsvg-convert proof",
            "- exceptions: missing participant is failure, not skip; fail branch at parity",
            "- code detail: function/type/file names stay in provenance; case IDs, just verbs, Python/Rust may appear",
            "- not this page: protocol handshake, JSON-on-stdout, normalize wipe/drop/keep, exit-code table (see driver-protocol-flow)",
            "",
        ]
    )


def provenance() -> dict:
    return {
        "thesis": T["title"],
        "page_policy": (
            "Case IDs, just verbs, and Python/Rust may appear; "
            "function/type/file names stay here"
        ),
        "anchors": [
            {
                "id": "hermetic-e2e",
                "claim": "Suite drives both installed libraries through public APIs, checks a shared oracle, then requires Python/Rust parity",
                "file": "README.md",
                "lines": "3-9",
            },
            {
                "id": "driver-no-impl",
                "claim": "Drivers must not implement routing, checkpointing, request tracking, or iteration behavior",
                "file": "CONTRACT.md",
                "lines": "24-28",
                "protocol": "DRIVER_PROTOCOL.md:16-17",
            },
            {
                "id": "shared-expected",
                "claim": "Each case carries one Expected observation; both participants are checked against that same record",
                "file": "internal/suite/catalog.go",
                "lines": "20-72",
                "oracle": "internal/suite/runner.go:76-81,105-117",
            },
            {
                "id": "parity-after-oracle",
                "claim": "After both observations exist, parity compares them with only participant removed",
                "file": "internal/suite/runner.go",
                "lines": "91-122",
                "contract": "CONTRACT.md:80-88",
            },
            {
                "id": "both-mandatory",
                "claim": "Both participants are mandatory; a missing driver is infrastructure failure, not a skip",
                "file": "README.md",
                "lines": "35-36,45",
                "runner": "internal/suite/runner.go:85-89",
                "tests": "internal/suite/suite_test.go:478-491",
            },
            {
                "id": "public-api-drive",
                "claim": "Python and Rust drivers construct graphs and call library public APIs; they do not reimplement fan-out, switch, or checkpoint restore",
                "python": "drivers/python/driver.py:76,151-154,253,280",
                "rust": "drivers/rust/src/main.rs:102-106,199-200,312,359",
            },
            {
                "id": "just-check",
                "claim": "just check probes both drivers then runs every hermetic case",
                "file": "justfile",
                "lines": "52-55",
                "readme": "README.md:16-21",
            },
            {
                "id": "case-ids",
                "claim": "Required cases are WF-001 through WF-010",
                "file": "CASES.md",
                "lines": "5-14",
                "catalog": "internal/suite/catalog.go:61-71",
            },
            {
                "id": "cli-verbs",
                "claim": "Product verbs are doctor, list, run, version; just exposes doctor/list/run/check",
                "file": "cmd/agent-framework-suite/main.go",
                "lines": "40-53",
                "just": "justfile:40-55",
            },
            {
                "id": "hermetic-no-net",
                "claim": "No case uses a model, network, ambient credential, or external service",
                "file": "README.md",
                "lines": "22",
            },
        ],
        "not_on_page": [
            "ValidateObservation",
            "Equivalent",
            "Observation",
            "CaseResult",
            "WorkflowBuilder",
            "reflect.DeepEqual",
            "decodeOne",
            "runner.go",
            "catalog.go",
            "client.go",
            "main.go",
            "driver.py",
            "main.rs",
            "selector.go",
            "types.go",
            "run_from_checkpoint",
            "add_fan_out_edges",
            "with_checkpointing",
        ],
    }


def build_verification() -> str:
    return "\n".join(
        [
            "# VERIFICATION",
            "",
            f"Thesis: {T['title']}",
            "",
            "## Kept vs changed",
            "",
            "- Kept, not edited: `docs/architecture-infographic.tex`, `docs/architecture-infographic.svg`, `docs/architecture.html`.",
            "- Kept, not edited: `docs/infographics/driver-protocol-flow.tex` and its derived artifacts (existing explainer; protocol / normalize / exit-code thesis).",
            "- New: Class A path spine (`page.svg` + `index.html` with the SVG inlined + `proof.png`). Different thesis from the explainer: path-not-catalog, thin-driver boundary.",
            "",
            "## Mechanisms checked in source",
            "",
            "- README: hermetic dual drive of public APIs, shared oracle, required parity; drivers do not implement routing/checkpoint.",
            "- CONTRACT participant boundary: drivers translate and normalize; they must not implement routing, checkpointing, request tracking, or iteration.",
            "- Catalog stores one Expected per case; runner checks each observation against that same Expected, then compares the two after dropping participant.",
            "- Missing participant count != 2 fails; missing driver is infrastructure, not skip.",
            "- Python/Rust drivers call library graph/build/run/resume APIs for fan-out, switch, and checkpoint cases.",
            "- `just check` builds, doctors both drivers, and runs every case.",
            "- Case IDs WF-001..WF-010 verified in CASES.md and the catalog.",
            "",
            "## Build commands",
            "",
            "```bash",
            "python3 docs/infographics/build.py",
            "rsvg-convert --width=2560 docs/infographics/page.svg -o docs/infographics/proof.png",
            "```",
            "",
            "svg-linter (repository root, after the two commands):",
            "",
            "```bash",
            "svg-linter --json check docs/infographics/page.svg \\",
            "  --select svg/duplicate-id,svg/dangling-reference \\",
            "  --require-complete --fail-on error",
            "svg-linter --json check docs/infographics/page.svg \\",
            "  --select svg/excessive-path-complexity,svg/unused-definition,svg/external-resource,svg/raster-upscale,svg/open-filled-path \\",
            "  --fail-on never --max-findings 80",
            "svg-linter --json check docs/infographics/page.svg \\",
            "  --select svg/line-overlap,svg/line-crossing,svg/line-text-overlap,svg/text-text-overlap,svg/canvas-edge-clearance,svg/ambiguous-junction,svg/connector-through-shape,svg/edge-congestion,svg/filter-clipping-risk,svg/low-contrast-text \\",
            "  --fail-on never --max-findings 80",
            "```",
            "",
            "Toolchain inspected with `python3 ~/.cursor/skills/create-infographic/scripts/inspect_toolchain.py --json --project-root .` (read-only; mutated=false).",
            "Route: hand-placed editorial SVG (web-svg-backend). No Vega/D3/ELK: topology is editorial and under twelve nodes.",
            "",
            "Objective error gate: 0 errors; `svg/duplicate-id` and `svg/dangling-reference` both evaluated (`--require-complete`).",
            "Hygiene review: 0 findings on the 4 evaluated rules; `svg/raster-upscale` not applicable (no rasters).",
            "Collision review: 0 findings on the 9 evaluated geometry/contrast rules; `svg/filter-clipping-risk` not applicable.",
            "svg-linter 0.1.0, catalog 7 (`sha256:fdfd45cabf9908c69853058ca51fba80878a99d21941eebf956258a904e55bb9`).",
            "Unannotated `proof.png` inspected at 2560\u00d71440; numbered collision lint PNG had no candidates after the layout pass.",
            "Saturated roles on the page: Ocean (flow) and Coral (fail). Teal/Gold unused. Fail is the same Coral hue with a dashed stroke.",
            "",
            "## Visual review",
            "",
            "- Thumbnail: Y-fork into a large shared-expected disc remains the silhouette.",
            "- Title occupies the upper left; upper right stays quiet except the hermetic fingerprint.",
            "- Case IDs are one slate line, not a ten-box catalog.",
            "- Boundary rule is gapped where observations cross; label sits above the stroke.",
            "- Both branch outcomes are labelled (`\u4e00\u81f4` / `\u5206\u53c9`); missing participant is annotated at fail.",
            "- CJK glyphs present in the SVG source and the rsvg-convert proof.",
            "",
            "## Unverified",
            "",
            "- Live `just check` against installed Python/Rust frameworks (this task did not run the suite).",
            "- Whether a future case beyond WF-010 would still fit the one-line ID strip.",
            "- browser-harness / Playwright (not used; rsvg-convert only).",
            "- CJK glyph coverage under a machine without Source Han Serif SC.",
            "",
            "## 2026-09-06 refine",
            "",
            "\u672c\u6b21\u7cbe\u4fee\uff08fleet-refine b01\uff09\u3002\u672c\u76ee\u5f55\u5df2\u5165\u5e93\uff08ab73aa7\uff09\uff1b\u672c\u6b21\u7cbe\u4fee\u7684\u6539\u52a8\u672a\u63d0\u4ea4\uff0c\u7531\u4e3b\u4f1a\u8bdd\u7a0d\u540e\u7edf\u4e00\u63d0\u4ea4\u3002",
            "",
            "### \u4fee\u590d",
            "",
            "- img-external-panel\uff08high\uff09\uff1aindex.html \u7531 `<object data=\"page.svg\">` \u5916\u6302\u6539\u4e3a\u5185\u8054 SVG\uff0c\u5185\u8054\u5b57\u8282\u4e0e page.svg \u4e00\u81f4\uff1bpage.svg \u4ecd\u662f\u7531 build.py \u91cd\u5efa\u7684\u6e90\u3002\u9875\u5185\u96f6\u5916\u90e8\u5f15\u7528\u3002",
            "- width-rule\uff08med\uff09\uff1aCSS-only\uff0cbody{max-width:1200px;margin:0 auto}\u3001svg{width:100%;height:auto}\uff1bpage.svg \u753b\u5e03\u4ecd 1280\xd7720\uff0c\u51e0\u4f55\u672a\u52a8\u3002",
            "- \u6587\u6863\u968f\u6539\uff1aREADME \u539f\u79f0\u300c\u9875\u9762\u662f\u5916\u90e8 SVG\uff1bindex.html \u53ea\u662f\u8584\u5305\u88c5\u300d\u3010\u540e\u8bc1\u4e0d\u5b9e\uff0c\u5df2\u4fee\u6b63\u3011\u2014\u2014\u73b0\u4e3a\u5185\u8054 SVG\u30011200px \u5c45\u4e2d\uff1bcontract \u7684 medium \u884c\u540c\u6b65\u6539\u4e3a\u5185\u8054\u8868\u8ff0\u3002",
            "",
            "### \u7f13\u529e\uff08\u6309\u672c\u6ce2\u8303\u56f4\u8bb0\u5f55\uff0c\u4e0d\u5728\u672c\u6b21\u5b9e\u73b0\uff09",
            "",
            "- no-claims-binding\u3001no-poison\u3001no-vacuum\u3002",
            "- fingerprint-gaps\uff1a\u6307\u7eb9\u6269\u4e3a page.svg + index.html \u4e24\u4ea7\u7269\uff1b\u5168\u6811\u6307\u7eb9\u672a\u505a\u3002",
            "- other\uff08\u4f4e\uff09\uff1acontract.md / VERIFICATION.md \u4ecd\u4e3a\u82f1\u6587\u3002",
            "- form-mismatch / no-sidenote-track\uff1a\u5355\u56fe\u5f62\u6001\u4e0e\u65c1\u6ce8\u8f68\u672a\u6539\u3002",
            "",
            "### \u95e8\u7981\u590d\u8dd1\uff082026-09-06\uff0c\u4ed3\u5e93\u6839\u6267\u884c\uff09",
            "",
            "- svg-linter objective \u95e8\uff08--require-complete --fail-on error\uff09\uff1a0 \u9519\u8bef\u3002",
            "- hygiene \u4e0e collision \u4e24\u7ec4\u590d\u67e5\uff1a\u5404 0 \u53d1\u73b0\u3002",
            "- \u4ece index.html \u6b63\u5219\u62bd\u53d6\u5185\u8054 SVG\uff0c\u4e0e page.svg \u5b57\u8282\u4e00\u81f4\u3002",
            "- \u53cc\u8dd1 build.py\uff1a6 \u4e2a\u4ea7\u7269\u6587\u4ef6\u5b57\u8282\u4e00\u81f4\uff08\u786e\u5b9a\u6027\uff09\u3002",
            "- rsvg-convert --width=2560 \u91cd\u6e32\u67d3\u4e0e\u65e2\u6709 proof.png \u5b57\u8282\u4e00\u81f4\u3002",
            "- \u9875\u9762\u4ee3\u7801\u5750\u6807\u626b\u63cf\uff08build.py \u5185\u7f6e\uff09\uff1a0 \u547d\u4e2d\u3002",
            "",
        ]
    )


def page_code_sweep(svg_text: str, html_text: str) -> list[str]:
    banned = [
        "ValidateObservation",
        "Equivalent",
        "Observation",
        "CaseResult",
        "WorkflowBuilder",
        "reflect.DeepEqual",
        "decodeOne",
        "runner.go",
        "catalog.go",
        "client.go",
        "main.go",
        "driver.py",
        "main.rs",
        "selector.go",
        "types.go",
        "run_from_checkpoint",
        "add_fan_out_edges",
        "with_checkpointing",
        ".go:",
        ".py:",
        ".rs:",
    ]
    blob = svg_text + "\n" + html_text
    return [name for name in banned if name in blob]


def assert_fits() -> None:
    title_w, _ = measure(T["title"], 32, "bld")
    if title_w > 720:
        raise SystemExit(f"title too wide: {title_w:.1f}")
    oracle_w, _ = measure(T["n_oracle"], 15, "bld")
    if oracle_w > 88:
        raise SystemExit(f"oracle label too wide for disc: {oracle_w:.1f}")
    ids_w, _ = measure(T["ids"], 12)
    if ids_w > 1180:
        raise SystemExit(f"case-id line too wide: {ids_w:.1f}")


def main() -> None:
    for path in (FONT_REG, FONT_MED, FONT_BLD):
        if not path.is_file():
            raise SystemExit(f"missing font: {path}")
    assert_fits()

    svg = build_svg()
    html = build_index_html(svg)
    hits = page_code_sweep(svg, html)
    if hits:
        raise SystemExit(f"code coordinates leaked onto the page: {hits}")

    write_utf8(ROOT / "page.svg", svg)
    write_utf8(ROOT / "index.html", html)
    write_utf8(ROOT / "README.md", build_readme())
    write_utf8(ROOT / "contract.md", build_contract())
    write_utf8(ROOT / "VERIFICATION.md", build_verification())
    write_utf8(
        ROOT / "data" / "provenance.json",
        json.dumps(provenance(), ensure_ascii=False, indent=2) + "\n",
    )
    svg_path = ROOT / "page.svg"
    digest = hashlib.sha256(svg_path.read_bytes()).hexdigest()
    html_path = ROOT / "index.html"
    hdigest = hashlib.sha256(html_path.read_bytes()).hexdigest()
    extra = (
        "\n## Artifact fingerprints (this build)\n\n"
        f"- page.svg SHA-256: `{digest}`\n"
        f"- page.svg bytes: {svg_path.stat().st_size}\n"
        f"- index.html SHA-256: `{hdigest}`\n"
        f"- index.html bytes: {html_path.stat().st_size}\n"
        "- index.html inlines page.svg byte-identically; zero external references\n"
        "- proof.png is produced by `rsvg-convert --width=2560` (2x of 1280 x 720)\n"
        "- CJK font at render: Source Han Serif SC\n"
        "\n## Page code-coordinate sweep\n\n"
        "Zero matches for file:line, type names, or function names on page.svg / index.html.\n"
    )
    ver = ROOT / "VERIFICATION.md"
    ver.write_text(ver.read_text(encoding="utf-8").rstrip() + "\n" + extra, encoding="utf-8")
    print(f"wrote {svg_path} ({len(svg)} bytes) sha256={digest}")


if __name__ == "__main__":
    main()
