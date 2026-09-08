#!/usr/bin/env python3
"""Audit gates for the agent-framework-suite infographic tree.

create-explainer audit batteries implemented here:
  1  detail/code-coordinate sweep (banned corpus frozen in build.py)
  2  claims gate: two-way Cxx binding + case-id recompute from CASES.md
  3  svg-linter objective/hygiene/collision gates (real binary from PATH)
  5  reverse digit sweep with reviewed in-tree exemption list
  +  inline-SVG identity, whole-tree fingerprint currency, doc-hash derivation
  +  poison pills: every gate must prove it can bite (throwaway copies only)

Run records land in data/audit/ (fingerprint-exempt, fixpoint rule).

Usage (repository root or anywhere; paths are tree-absolute):
  python3 tools/audit_gates.py all [--record]
  python3 tools/audit_gates.py pills [--record]
"""

from __future__ import annotations

import argparse
import html as htmllib
import json
import re
import shutil
import subprocess
import sys
import tempfile
from pathlib import Path

TREE = Path(__file__).resolve().parent.parent
REPO = TREE.parent.parent
REGISTRY_PATH = TREE / "data" / "claims.json"
PROVENANCE_PATH = TREE / "data" / "provenance.json"
PAGE_SVG = TREE / "page.svg"
INDEX = TREE / "index.html"
FINGERPRINT_TOOL = TREE / "tools" / "fingerprint.py"
FINGERPRINTS = TREE / "fingerprints.json"
AUDIT_DIR = TREE / "data" / "audit"
DATE = "2026-09-07"

sys.path.insert(0, str(TREE))
sys.dont_write_bytecode = True  # keep __pycache__ scratch out of the fingerprinted tree
import build  # noqa: E402  single source for the banned-token corpus

SVG_RE = re.compile(r"<svg[\s\S]*?</svg>")
STYLE_RE = re.compile(r"<style[\s\S]*?</style>")
SCRIPT_RE = re.compile(r"<script[\s\S]*?</script>")
TAG_RE = re.compile(r"<[^>]+>")
SVG_TEXT_RE = re.compile(r"<text[^>]*>([\s\S]*?)</text>")
SVG_TITLE_RE = re.compile(r"<title[^>]*>([\s\S]*?)</title>")
SVG_DESC_RE = re.compile(r"<desc[^>]*>([\s\S]*?)</desc>")
TOKEN_RE = re.compile(r"C\d+|WF-\d+|\d+")
CID_RE = re.compile(r"C\d+")
CHIP_PREFIX = "声明"  # sheng ming (claim) chip marker on the page

# Hashes quoted in README/VERIFICATION that are NOT tree artifacts. Every
# other 64-hex token in those docs must equal a fingerprints.json entry.
EXTERNAL_HASHES = {
    "fdfd45cabf9908c69853058ca51fba80878a99d21941eebf956258a904e55bb9":
        "svg-linter 0.1.0 rule catalog (external toolchain binary, not a tree artifact)",
}


def load_registry(path: Path | None = None) -> dict:
    return json.loads((path or REGISTRY_PATH).read_text(encoding="utf-8"))


def svg_text_projection(source: str) -> str:
    chunks = []
    for rx in (SVG_TEXT_RE, SVG_TITLE_RE, SVG_DESC_RE):
        chunks.extend(htmllib.unescape(m.group(1)) for m in rx.finditer(source))
    return "\n".join(chunks)


def html_text_projection(source: str) -> str:
    rest = SVG_RE.sub(" ", source)
    rest = STYLE_RE.sub(" ", rest)
    rest = SCRIPT_RE.sub(" ", rest)
    return htmllib.unescape(TAG_RE.sub("\n", rest))


def page_projection(html_source: str) -> str:
    return svg_text_projection(html_source) + "\n" + html_text_projection(html_source)


def casesmd_ids() -> set[str] | None:
    casesmd = REPO / "CASES.md"
    if not casesmd.is_file():
        return None
    return set(re.findall(r"^\| (WF-\d+) \|", casesmd.read_text(encoding="utf-8"), re.M))


# ---------------------------------------------------------------- gates


def gate_inline() -> dict:
    raw = INDEX.read_text(encoding="utf-8")
    m = SVG_RE.search(raw)
    if not m:
        return {"name": "inline", "ok": False, "error": "no inline <svg> found in index.html"}
    page = PAGE_SVG.read_text(encoding="utf-8")
    ok = m.group(0) + "\n" == page
    return {
        "name": "inline",
        "ok": ok,
        "inline_bytes": len(m.group(0).encode("utf-8")),
        "page_svg_bytes": len(page.encode("utf-8")),
        "note": "index.html ships page.svg inline byte-identically; zero external references",
    }


def gate_codesweep(svg_source: str | None = None, html_source: str | None = None) -> dict:
    svg = svg_source if svg_source is not None else PAGE_SVG.read_text(encoding="utf-8")
    html = html_source if html_source is not None else INDEX.read_text(encoding="utf-8")
    hits = build.page_code_sweep(svg, html)
    return {"name": "codesweep", "ok": not hits, "hits": hits,
            "corpus": "banned token list frozen in build.py page_code_sweep"}


def gate_binding(html_source: str | None = None, registry: dict | None = None) -> dict:
    reg = registry or load_registry()
    html = html_source if html_source is not None else INDEX.read_text(encoding="utf-8")
    proj = page_projection(html)
    registry_ids = set(reg["claim_index"])
    page_ids = set(CID_RE.findall(proj))
    chip_ids = set(CID_RE.findall("\n".join(l for l in proj.splitlines() if CHIP_PREFIX in l)))
    anchors = {a["id"] for a in json.loads(PROVENANCE_PATH.read_text(encoding="utf-8"))["anchors"]}
    unknown = sorted(page_ids - registry_ids)
    missing = sorted(registry_ids - page_ids)
    chip_unknown = sorted(chip_ids - registry_ids)
    unbound_anchors = sorted(set(reg["claim_index"].values()) - anchors)
    live = casesmd_ids()
    expected = set(reg["expected_case_ids"])
    case_recompute = (
        {"source": "REPO/CASES.md", "ok": live == expected}
        if live is not None
        else {"source": "data/claims.json expected_case_ids (CASES.md absent)", "ok": True}
    )
    ok = not (unknown or missing or chip_unknown or unbound_anchors) and case_recompute["ok"]
    return {
        "name": "binding",
        "ok": ok,
        "registry_ids": sorted(registry_ids),
        "page_ids": sorted(page_ids),
        "unknown_on_page": unknown,
        "missing_from_page": missing,
        "chip_ids": sorted(chip_ids),
        "chip_unknown": chip_unknown,
        "claim_index_values_without_anchor": unbound_anchors,
        "case_id_recompute": {**case_recompute, "expected": sorted(expected),
                              "recomputed": sorted(live) if live is not None else None},
    }


def gate_digits(html_source: str | None = None, registry: dict | None = None) -> dict:
    reg = registry or load_registry()
    html = html_source if html_source is not None else INDEX.read_text(encoding="utf-8")
    tokens = TOKEN_RE.findall(page_projection(html))
    registry_ids = set(reg["claim_index"])
    expected_cases = set(reg["expected_case_ids"])
    exempt: dict[str, str] = {}
    for entry in reg["digit_exemptions"]:
        exempt.update({v: entry["reason"] for v in entry["values"]})
    bad_claims, bad_cases, unclaimed = [], [], []
    for tok in tokens:
        if tok.startswith("C"):
            if tok not in registry_ids:
                bad_claims.append(tok)
        elif tok.startswith("WF-"):
            if tok not in expected_cases:
                bad_cases.append(tok)
        elif tok in exempt:
            continue
        else:
            unclaimed.append(tok)
    ok = not (bad_claims or bad_cases or unclaimed)
    return {
        "name": "digits",
        "ok": ok,
        "tokens_total": len(tokens),
        "claim_ids": sum(1 for t in tokens if t.startswith("C")),
        "case_ids": sum(1 for t in tokens if t.startswith("WF-")),
        "exempt_bare_digits": sum(1 for t in tokens if not t.startswith(("C", "W")) and t in exempt),
        "unknown_claim_ids": sorted(set(bad_claims)),
        "unexpected_case_ids": sorted(set(bad_cases)),
        "unclaimed_digits": sorted(set(unclaimed)),
        "exemption_list": reg["digit_exemptions"],
    }


def _svg_linter(svg_path: Path, select: str, extra: list[str]) -> tuple[int, dict | None, str]:
    cmd = ["svg-linter", "--json", "check", str(svg_path), "--select", select, *extra]
    p = subprocess.run(cmd, capture_output=True, text=True)
    try:
        return p.returncode, json.loads(p.stdout) if p.stdout.strip() else None, p.stderr
    except json.JSONDecodeError:
        return 99, None, p.stdout + p.stderr


OBJ_SELECT = "svg/duplicate-id,svg/dangling-reference"
HYG_SELECT = ("svg/excessive-path-complexity,svg/unused-definition,svg/external-resource,"
              "svg/raster-upscale,svg/open-filled-path")
COL_SELECT = ("svg/line-overlap,svg/line-crossing,svg/line-text-overlap,svg/text-text-overlap,"
              "svg/canvas-edge-clearance,svg/ambiguous-junction,svg/connector-through-shape,"
              "svg/edge-congestion,svg/filter-clipping-risk,svg/low-contrast-text")


def gate_svg(svg_path: Path | None = None) -> dict:
    svg_path = svg_path or PAGE_SVG
    rc, rep, err = _svg_linter(svg_path, OBJ_SELECT, ["--require-complete", "--fail-on", "error"])
    if rep is None:
        return {"name": "svg", "ok": False, "error": f"svg-linter produced no JSON: {err}"}
    obj_eff = rep["selection"]["effective_rules"]
    obj_findings = rep["findings"]
    objective_ok = rc == 0 and not obj_findings and len(obj_eff) == 2
    groups = {"objective": {"rc": rc, "effective": len(obj_eff), "findings": len(obj_findings)}}
    for label, sel in (("hygiene", HYG_SELECT), ("collision", COL_SELECT)):
        rc2, rep2, err2 = _svg_linter(svg_path, sel, ["--fail-on", "never", "--max-findings", "80"])
        if rep2 is None:
            return {"name": "svg", "ok": False, "error": f"svg-linter {label} produced no JSON: {err2}"}
        groups[label] = {"rc": rc2, "effective": len(rep2["selection"]["effective_rules"]),
                         "findings": len(rep2["findings"])}
    ok = objective_ok and all(g["effective"] > 0 and g["findings"] == 0 for g in groups.values())
    return {"name": "svg", "ok": ok, "groups": groups,
            "note": "real svg-linter binary from PATH; empty effective_rules would read as clean (stale catalog trap)"}


def gate_fingerprint() -> dict:
    if not FINGERPRINTS.is_file():
        return {"name": "fingerprint", "ok": False,
                "error": "fingerprints.json missing; run tools/fingerprint.py"}
    p = subprocess.run([sys.executable, str(FINGERPRINT_TOOL), "--check"],
                       capture_output=True, text=True)
    return {"name": "fingerprint", "ok": p.returncode == 0,
            "stdout": p.stdout.strip(), "stderr": p.stderr.strip()}


def gate_dochash() -> dict:
    if not FINGERPRINTS.is_file():
        return {"name": "dochash", "ok": False,
                "error": "fingerprints.json missing; run tools/fingerprint.py"}
    manifest = json.loads(FINGERPRINTS.read_text(encoding="utf-8"))
    known = {v["sha256"] for v in manifest["files"].values()}
    bad = []
    quoted = 0
    for doc in ("README.md", "VERIFICATION.md"):
        for h in re.findall(r"\b[0-9a-f]{64}\b", (TREE / doc).read_text(encoding="utf-8")):
            quoted += 1
            if h in known or h in EXTERNAL_HASHES:
                continue
            bad.append({"doc": doc, "hash": h})
    return {"name": "dochash", "ok": not bad, "quoted_hashes": quoted, "violations": bad,
            "external_allowlist": EXTERNAL_HASHES,
            "note": "every tree-artifact hash quoted in README/VERIFICATION derives from fingerprints.json"}


ALL_GATES = [gate_inline, gate_codesweep, gate_binding, gate_digits, gate_svg,
             gate_fingerprint, gate_dochash]


# ---------------------------------------------------------------- pills

# One positive control per ban class of the detail gate (battery 1): the
# banned-token corpus itself lives in build.py; controls map to the banned
# entry that must fire.
PILL_CONTROLS = [
    ("type-name", "ValidateObservation", "ValidateObservation"),
    ("function-name", "run_from_checkpoint", "run_from_checkpoint"),
    ("go-file-name", "runner.go", "runner.go"),
    ("rust-file-name", "main.rs", "main.rs"),
    ("file-line-go", "selector.go:42", ".go:"),
    ("file-line-py", "driver.py:76", ".py:"),
]


def battery_pills() -> dict:
    svg = PAGE_SVG.read_text(encoding="utf-8")
    html = INDEX.read_text(encoding="utf-8")
    pills: list[dict] = []
    tmp = Path(tempfile.mkdtemp(prefix="afs-pills-"))
    try:
        # P1 detail/code-coordinate sweep: inject one control per ban class.
        controls = " ".join(c for _, c, _ in PILL_CONTROLS)
        poisoned_svg = svg.replace(
            "</svg>", f'<text x="48" y="704" font-size="12">{controls}</text>\n</svg>'
        )
        hits = build.page_code_sweep(poisoned_svg, html)
        caught = [{"class": cls, "control": ctl, "flagged": banned in hits}
                  for cls, ctl, banned in PILL_CONTROLS]
        clean_hits = build.page_code_sweep(svg, html)
        pills.append({
            "gate": "codesweep", "mode": "inject-into-throwaway-copy",
            "controls": len(PILL_CONTROLS), "caught": sum(1 for c in caught if c["flagged"]),
            "detail": caught, "clean_corpus_hits": clean_hits,
            "ok": all(c["flagged"] for c in caught) and not clean_hits,
        })

        # P2 svg-linter objective gate: delete one referenced <defs> id.
        pill_svg = tmp / "pill-defs-removed.svg"
        pill_svg.write_text(re.sub(r'<marker id="mk-flow"[\s\S]*?</marker>', "", svg), encoding="utf-8")
        prc, prep, _ = _svg_linter(pill_svg, OBJ_SELECT, ["--require-complete", "--fail-on", "error"])
        pfindings = prep["findings"] if prep else None
        clean_svg = tmp / "clean-copy.svg"
        clean_svg.write_text(svg, encoding="utf-8")
        crc, crep, _ = _svg_linter(clean_svg, OBJ_SELECT, ["--require-complete", "--fail-on", "error"])
        pills.append({
            "gate": "svg-objective", "mode": "delete-referenced-defs-id",
            "removed_def": "mk-flow", "poison_rc": prc,
            "poison_findings": len(pfindings) if pfindings is not None else None,
            "poison_rules": sorted({f["rule"] for f in pfindings}) if pfindings else [],
            "clean_rc": crc, "clean_findings": len(crep["findings"]) if crep else None,
            "ok": prc != 0 and bool(pfindings) and crc == 0,
        })

        # P3a claims binding: unknown Cxx injected onto a throwaway page copy.
        poison_unknown = html.replace("</body>", "<p>" + CHIP_PREFIX + " C99</p>\n</body>")
        r_unknown = gate_binding(html_source=poison_unknown)
        # P3b claims binding: registry id removed from the page copy.
        poison_missing = CID_RE.sub(lambda m: "" if m.group(0) == "C05" else m.group(0), html)
        r_missing = gate_binding(html_source=poison_missing)
        r_clean = gate_binding(html_source=html)
        pills.append({
            "gate": "binding", "mode": "two-directional",
            "unknown_id_pill": {"injected": "C99", "ok": (not r_unknown["ok"])
                                and "C99" in r_unknown["unknown_on_page"]},
            "missing_id_pill": {"removed": "C05", "ok": (not r_missing["ok"])
                                and "C05" in r_missing["missing_from_page"]},
            "clean_ok": r_clean["ok"],
            "ok": (not r_unknown["ok"] and "C99" in r_unknown["unknown_on_page"]
                   and not r_missing["ok"] and "C05" in r_missing["missing_from_page"]
                   and r_clean["ok"]),
        })

        # P4 digit sweep: stray unclaimed number on a throwaway page copy.
        poison_digits = html.replace("</body>", "<p>42</p>\n</body>")
        r_digits = gate_digits(html_source=poison_digits)
        r_digits_clean = gate_digits(html_source=html)
        pills.append({
            "gate": "digits", "mode": "inject-stray-number",
            "injected": "42", "poison_unclaimed": r_digits["unclaimed_digits"],
            "clean_unclaimed": r_digits_clean["unclaimed_digits"],
            "ok": (not r_digits["ok"]) and "42" in r_digits["unclaimed_digits"]
            and r_digits_clean["ok"],
        })
    finally:
        shutil.rmtree(tmp, ignore_errors=True)
    return {"date": DATE, "corpus": "throwaway /tmp copies; frozen layer untouched",
            "pills": pills, "all_ok": all(p["ok"] for p in pills)}


# ---------------------------------------------------------------- cli


def record(name: str, payload: dict) -> Path:
    AUDIT_DIR.mkdir(parents=True, exist_ok=True)
    out = AUDIT_DIR / f"{name}-{DATE}.json"
    out.write_text(json.dumps(payload, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    return out


def main() -> int:
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument("command", choices=["all", "pills"])
    ap.add_argument("--record", action="store_true",
                    help="write the run record into data/audit/ (fingerprint-exempt)")
    args = ap.parse_args()

    if args.command == "pills":
        res = battery_pills()
        print(json.dumps(res, ensure_ascii=False, indent=2))
        if args.record:
            print(f"record -> {record('pills', res)}")
        return 0 if res["all_ok"] else 1

    results = [g() for g in ALL_GATES]
    ok = all(r["ok"] for r in results)
    payload = {"date": DATE, "gates": results, "all_ok": ok}
    print(json.dumps(payload, ensure_ascii=False, indent=2))
    if args.record:
        print(f"record -> {record('gates', payload)}")
    return 0 if ok else 1


if __name__ == "__main__":
    raise SystemExit(main())
