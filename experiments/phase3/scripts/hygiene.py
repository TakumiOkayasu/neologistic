#!/usr/bin/env python3
from __future__ import annotations

import json
import re
import sys
from pathlib import Path

PHASE3_ROOT = Path(__file__).resolve().parents[1]
REPO_ROOT = PHASE3_ROOT.parents[1]
ROOT_DOCS = (
    "AGENTS.md",
    "README.md",
    "PROJECT_ORIGIN.md",
    "DECISION_LINEAGE.md",
    "EXPERIMENT_PLAN.md",
)
PERSONAL_PATHS = (
    re.compile("/" + "Users" + r"/[^/\s]+/"),
    re.compile("/" + "home" + r"/[^/\s]+/"),
    re.compile(r"[A-Za-z]:\\" + "Users" + r"\\[^\\\s]+\\"),
)
OBSOLETE_FIXTURE_PREFIX = "p3-" + "01-"
BOX_DRAWING = re.compile(r"[\u2500-\u257f]")
ASCII_ART = (
    re.compile(r"(?m)^\s*\+[-=]{3,}\+"),
    re.compile(r"(?m)^\s*(?:\|--|\+--|\\--)")
)
TEXT_SUFFIXES = {".md", ".json", ".jsonl", ".go", ".sh", ".py", ".yml", ".yaml"}
MODEL_COMMANDS = tuple("".join(parts) for parts in (("co", "dex"), ("cla", "ude"), ("open", "ai"), ("curl",), ("wget",)))
FORBIDDEN_GO_IMPORTS = ('"' + "net/" + "http" + '"', '"' + "os/" + "exec" + '"')


def iter_files() -> list[Path]:
    files: set[Path] = set()
    for rel in ROOT_DOCS:
        path = REPO_ROOT / rel
        if path.is_file():
            files.add(path)
    for path in PHASE3_ROOT.rglob("*"):
        if path.is_file() and path.suffix.lower() in TEXT_SUFFIXES:
            files.add(path)
    return sorted(files)


def relative(path: Path) -> str:
    return path.relative_to(REPO_ROOT).as_posix()


def validate_json(path: Path, text: str, errors: list[str]) -> None:
    try:
        if path.suffix == ".json":
            json.loads(text, object_pairs_hook=_reject_duplicates)
            return
        if path.suffix == ".jsonl":
            for line_number, line in enumerate(text.splitlines(), 1):
                if line.strip():
                    json.loads(line, object_pairs_hook=_reject_duplicates)
    except (ValueError, json.JSONDecodeError) as exc:
        errors.append(f"{relative(path)}: invalid JSON: {exc}")


def _reject_duplicates(pairs: list[tuple[str, object]]) -> dict[str, object]:
    value: dict[str, object] = {}
    for key, item in pairs:
        if key in value:
            raise ValueError(f"duplicate JSON key {key!r}")
        value[key] = item
    return value


def main() -> int:
    errors: list[str] = []
    for path in iter_files():
        text = path.read_text(encoding="utf-8")
        rel = relative(path)
        for line_number, line in enumerate(text.splitlines(), 1):
            if line != line.rstrip():
                errors.append(f"{rel}:{line_number}: trailing whitespace")
        for pattern in PERSONAL_PATHS:
            match = pattern.search(text)
            if match:
                errors.append(f"{rel}: personal absolute path {match.group(0)!r}")
        if BOX_DRAWING.search(text):
            errors.append(f"{rel}: box-drawing character found; use Mermaid for diagrams")
        if any(pattern.search(text) for pattern in ASCII_ART):
            errors.append(f"{rel}: ASCII-art diagram pattern found; use Mermaid")
        if OBSOLETE_FIXTURE_PREFIX in text:
            errors.append(f"{rel}: obsolete semantic fixture identifier found")
        if path.suffix in {".sh", ".py"}:
            for command in MODEL_COMMANDS:
                command_pattern = re.compile(r"(?m)(?:^|[;&|]\s*)" + re.escape(command) + r"(?:\s|$)")
                if command_pattern.search(text):
                    errors.append(f"{rel}: executable script invokes forbidden external/model command {command!r}")
        if path.suffix == ".go":
            for forbidden_import in FORBIDDEN_GO_IMPORTS:
                if forbidden_import in text:
                    errors.append(f"{rel}: deterministic harness imports forbidden runtime package {forbidden_import}")
        if path.suffix in {".json", ".jsonl"}:
            validate_json(path, text, errors)

    if errors:
        print("hygiene failed:", file=sys.stderr)
        for error in errors:
            print(f"- {error}", file=sys.stderr)
        return 1
    print(f"hygiene: PASS files={len(iter_files())}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
