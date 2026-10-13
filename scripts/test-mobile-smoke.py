"""Exercise real native builds against synthetic compatible and broken modules."""

from pathlib import Path
import shutil
import subprocess
import sys
import tempfile


def main():
    if len(sys.argv) != 2 or sys.argv[1] not in ("android", "apple"):
        raise SystemExit("Usage: python3 scripts/test-mobile-smoke.py android|apple")
    target = sys.argv[1]
    source = Path(__file__).resolve().parent.parent
    with tempfile.TemporaryDirectory(prefix="barometer-smoke-test-") as directory:
        root = Path(directory)
        (root / "go.mod").write_text(
            "module github.com/tv-rain/barometer-go\n\ngo 1.26.0\n"
        )
        shutil.copy(source / "mobile_smoke.go", root)
        (root / "scripts").mkdir()
        shutil.copy(source / "scripts/mobile-smoke.sh", root / "scripts")
        (root / "tools/mobile").mkdir(parents=True)
        for name in ("go.mod", "go.sum", "dependencies.go"):
            shutil.copy(source / "tools/mobile" / name, root / "tools/mobile")

        def check(label, expected=None):
            result = subprocess.run(
                ["bash", "scripts/mobile-smoke.sh", target],
                cwd=root, text=True, stdout=subprocess.PIPE, stderr=subprocess.STDOUT,
                timeout=600,
            )
            if expected is None:
                valid = result.returncode == 0 and "Mobile smoke passed:" in result.stdout
            else:
                valid = result.returncode != 0 and expected in result.stdout
            if not valid:
                raise SystemExit(f"FAIL: {label}\n{result.stdout}")
            print(f"PASS: {target}: {label}", flush=True)

        check("compatible exported boundary builds")
        broken = root / "broken.go"
        broken.write_text(
            "package barometer\n\n"
            "func Unsupported() map[string]string { return nil }\n"
        )
        check("skipped unsupported export fails", "unsupported exports were skipped")
        (root / "internal/desktoponly").mkdir(parents=True)
        (root / "internal/desktoponly/only.go").write_text(
            "//go:build !android && !ios\n\npackage desktoponly\n"
        )
        broken.write_text(
            "package barometer\n\n"
            'import _ "github.com/tv-rain/barometer-go/internal/desktoponly"\n'
        )
        check("incompatible dependency fails", "build constraints exclude all Go files")


if __name__ == "__main__":
    main()
