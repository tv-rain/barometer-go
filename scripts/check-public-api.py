"""Compile the documented API signatures in an isolated native binding fixture."""

from pathlib import Path
import re
import shutil
import subprocess
import sys
import tempfile


def main():
    if len(sys.argv) != 2 or sys.argv[1] not in ("android", "apple"):
        raise SystemExit("Usage: python3 scripts/check-public-api.py android|apple")
    target = sys.argv[1]
    source = Path(__file__).resolve().parent.parent
    document = (source / "docs/public-api.md").read_text()
    section = document.split("<!-- public-api-signatures:start -->", 1)[1]
    section = section.split("<!-- public-api-signatures:end -->", 1)[0]
    signatures = re.search(r"```go\n(.*?)\n```", section, re.S).group(1)
    fixture = "\n".join(
        line + ' { panic("signature fixture only") }'
        if line.startswith("func ") else line
        for line in signatures.splitlines()
    ) + "\n"
    with tempfile.TemporaryDirectory(prefix="barometer-api-fixture-") as directory:
        root = Path(directory)
        shutil.copy(source / "go.mod", root)
        (root / "api.go").write_text(fixture)
        (root / "scripts").mkdir()
        shutil.copy(source / "scripts/mobile-build.sh", root / "scripts")
        (root / "tools/mobile").mkdir(parents=True)
        for name in ("go.mod", "go.sum", "dependencies.go"):
            shutil.copy(source / "tools/mobile" / name, root / "tools/mobile")
        subprocess.run(["gofmt", "-w", "api.go"], cwd=root, check=True)
        subprocess.run(
            ["bash", "scripts/mobile-build.sh", target], cwd=root, check=True,
            timeout=600,
        )
        print(f"PASS: documented API signatures compile for {target}; fixture only")


if __name__ == "__main__":
    main()
