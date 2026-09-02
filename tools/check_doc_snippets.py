#!/usr/bin/env python3
"""Compile every Go snippet in the Markdown docs.

Documentation that does not compile is worse than none, and these pages are
mostly code. This extracts each ```go block from the README, the guides under
docs/concepts and the examples README, writes it into a throwaway module that
depends on this checkout, and builds it.

Blocks that are fragments rather than whole files - anything not starting with
`package` - are counted and skipped, so the summary says how much of the
documentation is actually verified.

    python tools/check_doc_snippets.py            # every documented page
    python tools/check_doc_snippets.py README.md  # just one
"""

import os
import re
import shutil
import subprocess
import sys
import tempfile

REPO = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))

DEFAULT_PAGES = [
    "README.md",
    "examples/README.md",
    "docs/concepts/events.md",
    "docs/concepts/layout.md",
    "docs/concepts/rendering.md",
    "docs/concepts/terminal.md",
    "docs/concepts/testing.md",
    "docs/concepts/text-and-style.md",
    "docs/concepts/widgets.md",
]

BLOCK = re.compile(r"^```go\n(.*?)^```", re.MULTILINE | re.DOTALL)

GO_MOD = """module docsnippets

go {go_version}

require github.com/Fiend3d/catatui v0.0.0

replace github.com/Fiend3d/catatui => {repo}
"""


def go_version():
    """The go directive from the repo's go.mod.

    The throwaway module has to ask for at least what catatui asks for, or the
    go command refuses to build against it.
    """
    with open(os.path.join(REPO, "go.mod"), encoding="utf-8") as f:
        for line in f:
            if line.startswith("go "):
                return line.split()[1]
    raise SystemExit("no go directive in go.mod")


def snippets(page):
    """Yield (line number, source, whole file?) for each Go block in page."""
    text = open(os.path.join(REPO, page), encoding="utf-8").read()
    for match in BLOCK.finditer(text):
        code = match.group(1)
        line = text[: match.start()].count("\n") + 1
        yield line, code, code.lstrip().startswith("package ")


# seed imports every package a snippet might use, so that one `go mod tidy`
# resolves the whole module before any snippet is built.
SEED = """package seed

import (
	_ "github.com/Fiend3d/catatui"
	_ "github.com/Fiend3d/catatui/palette/material"
	_ "github.com/Fiend3d/catatui/palette/tailwind"
	_ "github.com/Fiend3d/catatui/symbols"
	_ "github.com/Fiend3d/catatui/term"
	_ "github.com/Fiend3d/catatui/widgets"
)
"""


def resolve_dependencies(work):
    """Populate go.mod and go.sum for the throwaway module."""
    seed = os.path.join(work, "seed")
    os.mkdir(seed)
    open(os.path.join(seed, "seed.go"), "w", encoding="utf-8").write(SEED)
    proc = subprocess.run(
        ["go", "mod", "tidy"], cwd=work, capture_output=True, text=True
    )
    if proc.returncode != 0:
        raise SystemExit("go mod tidy failed:" + proc.stderr)
    shutil.rmtree(seed)


def main(argv):
    pages = argv[1:] or DEFAULT_PAGES
    work = tempfile.mkdtemp(prefix="catatui-docs-")
    failures = 0
    checked = 0
    fragments = 0
    try:
        open(os.path.join(work, "go.mod"), "w", encoding="utf-8").write(
            GO_MOD.format(repo=REPO.replace("\\", "/"), go_version=go_version())
        )
        shutil.copyfile(os.path.join(REPO, "go.sum"), os.path.join(work, "go.sum"))
        resolve_dependencies(work)

        for page in pages:
            for line, code, whole in snippets(page):
                if not whole:
                    fragments += 1
                    continue
                checked += 1
                # One directory per snippet: they are whole files, and several
                # of them declare package main.
                pkg = os.path.join(work, "s%d" % checked)
                os.mkdir(pkg)
                is_test = "func Test" in code or "func Example" in code
                name = "main_test.go" if is_test else "main.go"
                open(os.path.join(pkg, name), "w", encoding="utf-8").write(code)

                proc = subprocess.run(
                    ["go", "vet", "./s%d" % checked],
                    cwd=work,
                    capture_output=True,
                    text=True,
                )
                if proc.returncode != 0:
                    failures += 1
                    print("%s:%d: snippet does not build" % (page, line))
                    print(proc.stdout.strip())
                    print(proc.stderr.strip())
    finally:
        shutil.rmtree(work, ignore_errors=True)

    print(
        "checked %d snippets, %d failed, %d fragments skipped"
        % (checked, failures, fragments)
    )
    return 1 if failures else 0


if __name__ == "__main__":
    sys.exit(main(sys.argv))
