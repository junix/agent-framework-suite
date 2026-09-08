#!/usr/bin/env python3
"""Vacuum rebuild battery (create-explainer audit battery 5).

Flat /tmp copy of the tree -> snapshot A -> delete every rebuildable
artifact (page, wrapper, bitmaps, derived registries and docs; the frozen
evidence layer and the legacy driver-protocol-flow.* artifacts are NOT
deleted) -> full-chain rebuild from tree tools -> three-way byte compare
A / rebuilt / archived with real file counts. A 0-file PASS is a false
PASS, so compare_files must be > 0. The rebuilt copy also verifies its
own fingerprints.json (detached mode, battery 6).

If the renderer binary is not already on PATH the bitmap step is SKIPPED
with a recorded reason (no builds are attempted from here).
"""

from __future__ import annotations

import json
import shutil
import subprocess
import sys
import tempfile
from pathlib import Path

TREE = Path(__file__).resolve().parent.parent
AUDIT_DIR = TREE / "data" / "audit"
DATE = "2026-09-07"

REBUILDABLE = [
    "page.svg",
    "index.html",
    "README.md",
    "contract.md",
    "VERIFICATION.md",
    "data/provenance.json",
    "data/claims.json",
    "proof.png",
]


def main() -> int:
    renderer = shutil.which("rsvg-convert")
    if renderer is None:
        skip_bitmap = True
        targets = [r for r in REBUILDABLE if r != "proof.png"]
        reason = "rsvg-convert not on PATH; bitmap step skipped rather than built"
    else:
        skip_bitmap = False
        targets = list(REBUILDABLE)

    tmp = Path(tempfile.mkdtemp(prefix="afs-vacuum-"))
    try:
        work = tmp / "infographics"
        shutil.copytree(TREE, work)

        snapshot_a = {rel: (work / rel).read_bytes() for rel in targets if (work / rel).is_file()}
        for rel in snapshot_a:
            (work / rel).unlink()

        build_cmd = [sys.executable, str(work / "build.py")]
        build_run = subprocess.run(build_cmd, capture_output=True, text=True)
        bitmap_run = None
        if not skip_bitmap:
            bitmap_cmd = ["rsvg-convert", "--width=2560", str(work / "page.svg"),
                          "-o", str(work / "proof.png")]
            bitmap_run = subprocess.run(bitmap_cmd, capture_output=True, text=True)

        compare = {}
        for rel in targets:
            rebuilt = (work / rel).read_bytes() if (work / rel).is_file() else None
            archived = (TREE / rel).read_bytes() if (TREE / rel).is_file() else None
            compare[rel] = {
                "a_eq_rebuilt": snapshot_a.get(rel) == rebuilt,
                "rebuilt_eq_archived": rebuilt == archived,
            }
        compare_files = sum(
            1 for rel in targets
            if (work / rel).is_file() and all(compare[rel].values())
        )

        detached = subprocess.run(
            [sys.executable, str(work / "tools" / "fingerprint.py"), "--check"],
            capture_output=True, text=True,
        )

        chain_ok = build_run.returncode == 0 and (
            bitmap_run is None or bitmap_run.returncode == 0
        )
        ok = (
            chain_ok
            and len(snapshot_a) == len(targets)
            and compare_files == len(targets)
            and len(targets) > 0
            and detached.returncode == 0
        )
        record = {
            "date": DATE,
            "tmp_dir": str(tmp),
            "tmp_removed_after_run": True,
            "deleted_then_regrown": sorted(snapshot_a),
            "deleted_count": len(snapshot_a),
            "rebuild_commands": {
                "build.py": {"argv": build_cmd, "rc": build_run.returncode},
                "rsvg-convert": (
                    {"argv": bitmap_cmd, "rc": bitmap_run.returncode}
                    if bitmap_run is not None
                    else {"skipped": reason}
                ),
            },
            "compare": compare,
            "compare_files": compare_files,
            "false_pass_guard": "compare_files > 0 required (a 0-file PASS is a false PASS)",
            "detached_fingerprint_check": {
                "rc": detached.returncode, "stdout": detached.stdout.strip(),
            },
            "frozen_layer": "data sources live in build.py; legacy driver-protocol-flow.* not deleted",
            "ok": ok,
        }
    finally:
        shutil.rmtree(tmp, ignore_errors=True)

    AUDIT_DIR.mkdir(parents=True, exist_ok=True)
    out = AUDIT_DIR / f"vacuum-{DATE}.json"
    out.write_text(json.dumps(record, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    print(json.dumps(record, ensure_ascii=False, indent=2))
    print(f"record -> {out}")
    return 0 if ok else 1


if __name__ == "__main__":
    raise SystemExit(main())
