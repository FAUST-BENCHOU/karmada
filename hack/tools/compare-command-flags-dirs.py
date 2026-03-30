#!/usr/bin/env python3
# Copyright 2024 The Karmada Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

"""Compare command flag docs from two directories (e.g. extract vs binary).

- Strips extract-only "Deprecated flags:" blocks from the reference (first) dir.
- Normalizes whitespace (collapse runs, lowercase) so formatting differences are ignored.
- For karmadactl.txt, compares only sections with the same === karmadactl ... === header
  present in both files (intersection).
"""

from __future__ import annotations

import argparse
import re
import sys
from pathlib import Path


def strip_deprecated_sections(s: str) -> str:
    while True:
        m = re.search(r"\n\nDeprecated flags:\n", s)
        if not m:
            break
        start = m.start()
        rest = s[m.end() :]
        end_rel = re.search(r"\n\n(?===|\Z)", rest)
        end = m.end() + (end_rel.start() if end_rel else len(rest))
        s = s[:start] + s[end:]
    return s


def normalize(s: str) -> str:
    return " ".join(s.split()).lower()


def parse_karmadactl_sections(text: str) -> dict[str, str]:
    parts = re.split(r"^=== karmadactl (.+) ===\s*\n", text, flags=re.M)
    out: dict[str, str] = {"__root__": parts[0].rstrip()}
    for i in range(1, len(parts), 2):
        if i + 1 < len(parts):
            out[parts[i].strip()] = parts[i + 1].rstrip()
    return out


def long_flags(s: str) -> set[str]:
    return set(re.findall(r"(?<![\w-])--[\w-]+", s))


def compare_file(ref_path: Path, other_path: str, name: str) -> list[str]:
    errors: list[str] = []
    with open(ref_path, encoding="utf-8") as f:
        ref_raw = f.read()
    with open(other_path, encoding="utf-8") as f:
        other_raw = f.read()

    if name == "karmadactl.txt":
        ref_secs = parse_karmadactl_sections(strip_deprecated_sections(ref_raw))
        other_secs = parse_karmadactl_sections(other_raw)
        keys = sorted(set(ref_secs) & set(other_secs))
        for k in keys:
            a = normalize(ref_secs[k])
            b = normalize(other_secs[k])
            if a == b:
                continue
            fa, fb = long_flags(ref_secs[k]), long_flags(other_secs[k])
            if fa != fb:
                errors.append(f"  {name} section [{k}]: mismatch (long flags differ)")
                ob = sorted(fa - fb)[:12]
                oc = sorted(fb - fa)[:12]
                if ob:
                    errors.append(f"    only in ref: {ob}")
                if oc:
                    errors.append(f"    only in other: {oc}")
        return errors

    ref_stripped = strip_deprecated_sections(ref_raw)
    ref = normalize(ref_stripped)
    other = normalize(other_raw)
    if ref == other:
        return errors
    if long_flags(ref_stripped) != long_flags(other_raw):
        errors.append(f"  {name}: long flags differ after strip deprecated")
        fa, fb = long_flags(ref_stripped), long_flags(other_raw)
        ob = sorted(fa - fb)[:12]
        oc = sorted(fb - fa)[:12]
        if ob:
            errors.append(f"    only in ref: {ob}")
        if oc:
            errors.append(f"    only in other: {oc}")
    return errors


def main() -> int:
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument(
        "ref_dir",
        nargs="?",
        default="docs/command-flags",
        help="Reference dir (extract-flags output; deprecated blocks stripped)",
    )
    p.add_argument(
        "other_dir",
        nargs="?",
        default="docs/binary-command-flags",
        help="Other dir (e.g. binary --help output)",
    )
    args = p.parse_args()
    ref_root = Path(args.ref_dir)
    other_root = Path(args.other_dir)
    if not ref_root.is_dir() or not other_root.is_dir():
        print("Both arguments must be existing directories", file=sys.stderr)
        return 2

    ref_files = {f.name for f in ref_root.glob("*.txt")}
    other_files = {f.name for f in other_root.glob("*.txt")}
    common = sorted(ref_files & other_files)
    if not common:
        print("No common .txt files", file=sys.stderr)
        return 2

    all_errors: list[str] = []
    for name in common:
        errs = compare_file(ref_root / name, str(other_root / name), name)
        all_errors.extend(errs)

    if all_errors:
        print("Compare (ignore format + strip ref Deprecated flags blocks):\n")
        print("\n".join(all_errors))
        return 1
    print("All common files match after normalization (and karmadactl section intersection).")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
