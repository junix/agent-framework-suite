#!/usr/bin/env python3
"""Whole-tree SHA-256 fingerprints (create-explainer audit battery 6).

Covers data, tools, render and docs — every file in the tree — with exactly
two exemptions, each justified in VERIFICATION.md (2026-09-07 refine):

  fingerprints.json  self-hash fixpoint: the manifest cannot hash itself
  data/audit/        run records: re-running a gate must not change the
                     tree fingerprint (battery 7 fixpoint rule)

Idempotent by construction: re-running rewrites identical bytes.
"""

from __future__ import annotations

import argparse
import hashlib
import json
from pathlib import Path

TREE = Path(__file__).resolve().parent.parent
OUT = TREE / "fingerprints.json"
EXEMPT_FILES = {"fingerprints.json": "self-hash fixpoint (manifest cannot hash itself)"}
EXEMPT_DIRS = {"data/audit": "run records; gate re-runs must not change the tree fingerprint"}


def iter_files():
    for p in sorted(TREE.rglob("*")):
        if not p.is_file():
            continue
        rel = p.relative_to(TREE).as_posix()
        if rel in EXEMPT_FILES:
            continue
        if any(rel.startswith(d + "/") for d in EXEMPT_DIRS):
            continue
        yield rel, p


def compute() -> dict:
    files = {}
    for rel, p in iter_files():
        data = p.read_bytes()
        files[rel] = {"sha256": hashlib.sha256(data).hexdigest(), "bytes": len(data)}
    canon = "".join(f"{rel} {files[rel]['sha256']} {files[rel]['bytes']}\n" for rel in sorted(files))
    return {
        "algorithm": "sha256",
        "generated_by": "tools/fingerprint.py",
        "file_count": len(files),
        "exemptions": {"files": EXEMPT_FILES, "dirs": EXEMPT_DIRS},
        "files": files,
        "tree_sha256": hashlib.sha256(canon.encode("utf-8")).hexdigest(),
    }


def main() -> int:
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument("--check", action="store_true",
                    help="recompute and compare against the stored manifest; exit 1 on drift")
    args = ap.parse_args()
    cur = compute()
    if args.check:
        if not OUT.is_file():
            print("FAIL fingerprints.json missing (run without --check to write it)")
            return 1
        stored = json.loads(OUT.read_text(encoding="utf-8"))
        if stored == cur:
            print(f"OK fingerprints current: {cur['file_count']} files, "
                  f"tree_sha256={cur['tree_sha256']}")
            return 0
        problems = []
        for rel in sorted(set(stored["files"]) | set(cur["files"])):
            a, b = stored["files"].get(rel), cur["files"].get(rel)
            if a is None:
                problems.append(f"+ {rel} (untracked in manifest)")
            elif b is None:
                problems.append(f"- {rel} (listed but gone)")
            elif a != b:
                problems.append(f"~ {rel}: {a['sha256'][:12]} -> {b['sha256'][:12]}")
        print("FAIL fingerprints drifted (rebuild or edit landed without refresh):")
        for line in problems:
            print("  " + line)
        return 1
    OUT.write_text(json.dumps(cur, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    print(f"wrote {OUT}: {cur['file_count']} files, tree_sha256={cur['tree_sha256']}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
