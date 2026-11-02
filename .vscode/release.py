#!/usr/bin/env python3
"""Builds native arm64 binaries, strips hotkeys, packages Bear.alfredworkflow, and opens a GitHub draft release."""

import json
import os
from pathlib import Path
import plistlib
import shutil
import subprocess
import sys
import tempfile
import zipfile

WORKFLOW_FILENAME = "Bear.alfredworkflow"
CMD_PACKAGES = [
    "create",
    "csearch",
    "link",
    "search",
    "setcursor",
    "toc",
    "wiki-link",
]


def build_env() -> dict:
    env = os.environ.copy()
    path_parts = env.get("PATH", "").split(os.pathsep)
    for extra in ("/opt/homebrew/bin", "/usr/local/bin"):
        if extra not in path_parts and os.path.isdir(extra):
            path_parts.insert(0, extra)
    env["PATH"] = os.pathsep.join(p for p in path_parts if p)
    env["GOOS"] = "darwin"
    env["GOARCH"] = "arm64"
    env["CGO_ENABLED"] = "1"
    env.setdefault("CGO_CFLAGS_ALLOW", "-Xpreprocessor -mmacosx-version-min")
    return env


def read_plist_version(plist_path: Path) -> str:
    data = plistlib.loads(plist_path.read_bytes())
    version = str(data.get("version", "")).strip()
    if not version:
        sys.exit("Error: <key>version</key> is empty in info.plist")
    return version


def build_arm64_binaries(repo_dir: Path, env: dict) -> None:
    go_bin = shutil.which("go", path=env.get("PATH"))
    if not go_bin:
        sys.exit("Error: 'go' compiler not found on PATH")

    file_cmd = shutil.which("file") or "/usr/bin/file"

    for pkg in CMD_PACKAGES:
        pkg_dir = repo_dir / "cmd" / pkg
        out_bin = pkg_dir / pkg
        if out_bin.exists():
            out_bin.unlink()

        print(f"Building cmd/{pkg}/{pkg} (darwin/arm64)...")
        subprocess.run([go_bin, "build", "-o", pkg], cwd=pkg_dir, env=env, check=True)

        arch_res = subprocess.run(
            [file_cmd, "-b", str(out_bin)],
            stdout=subprocess.PIPE,
            text=True,
            check=True,
        )
        arch_desc = arch_res.stdout.strip()
        if "arm64" not in arch_desc or "x86_64" in arch_desc:
            sys.exit(f"Error: expected pure arm64 binary for {out_bin}, got: {arch_desc}")


def write_sanitized_plist(src_plist: Path, dst_plist: Path) -> None:
    data = plistlib.loads(src_plist.read_bytes())
    cleared_hotkeys = 0
    for obj in data.get("objects", []):
        if obj.get("type") == "alfred.workflow.trigger.hotkey":
            cfg = obj.setdefault("config", {})
            if cfg.get("hotkey", 0) != 0 or cfg.get("hotmod", 0) != 0 or cfg.get("hotstring", "") != "":
                cleared_hotkeys += 1
            cfg["hotkey"] = 0
            cfg["hotmod"] = 0
            cfg["hotstring"] = ""
    dst_plist.write_bytes(plistlib.dumps(data))
    print(f"Sanitized staged info.plist (cleared {cleared_hotkeys} hotkey(s))")


def package_workflow(repo_dir: Path, out_path: Path) -> None:
    with tempfile.TemporaryDirectory(prefix="alfred-bear-release-") as tmp:
        stage_dir = Path(tmp)

        write_sanitized_plist(repo_dir / "info.plist", stage_dir / "info.plist")

        license_path = repo_dir / "LICENSE"
        if license_path.exists():
            shutil.copyfile(license_path, stage_dir / "LICENSE")

        for png in sorted(repo_dir.glob("*.png")):
            if not png.name.startswith("._"):
                shutil.copyfile(png, stage_dir / png.name)

        for pkg in CMD_PACKAGES:
            src_bin = repo_dir / "cmd" / pkg / pkg
            dst_bin = stage_dir / "cmd" / pkg / pkg
            dst_bin.parent.mkdir(parents=True, exist_ok=True)
            shutil.copyfile(src_bin, dst_bin)
            os.chmod(dst_bin, 0o755)

        out_path.parent.mkdir(parents=True, exist_ok=True)
        if out_path.exists():
            out_path.unlink()

        top_entries = sorted(p.name for p in stage_dir.iterdir())
        zip_cmd = shutil.which("zip") or "/usr/bin/zip"
        subprocess.run(
            [zip_cmd, "-q", "-r", "-X", str(out_path), *top_entries],
            cwd=stage_dir,
            check=True,
        )

    with zipfile.ZipFile(out_path, "r") as zf:
        names = set(zf.namelist())
        if "info.plist" not in names:
            sys.exit("Error: info.plist missing from packaged Bear.alfredworkflow")
        for pkg in CMD_PACKAGES:
            rel = f"cmd/{pkg}/{pkg}"
            if rel not in names:
                sys.exit(f"Error: {rel} missing from packaged Bear.alfredworkflow")

    size_mb = out_path.stat().st_size / (1024 * 1024)
    print(f"Packaged {out_path} ({size_mb:.1f} MB)")


def check_git_state(repo_dir: Path, env: dict) -> None:
    git_bin = shutil.which("git", path=env.get("PATH"))
    if not (repo_dir / ".git").exists() or not git_bin:
        return
    status = subprocess.run(
        [git_bin, "status", "--porcelain"],
        cwd=repo_dir,
        capture_output=True,
        text=True,
        check=False,
    )
    if status.returncode == 0 and status.stdout.strip():
        print("Note: You have uncommitted changes. Commit and push to GitHub before publishing the draft release.")
        return

    branch_status = subprocess.run(
        [git_bin, "status", "-sb"],
        cwd=repo_dir,
        capture_output=True,
        text=True,
        check=False,
    )
    if branch_status.returncode == 0 and "[ahead " in branch_status.stdout:
        print("Note: Local branch is ahead of origin. Run 'git push' before publishing the draft release.")


def upload_github_draft(repo_dir: Path, out_path: Path, tag: str, env: dict) -> None:
    gh_bin = shutil.which("gh", path=env.get("PATH"))
    if not gh_bin:
        print("Note: 'gh' CLI not found (install via 'brew install gh && gh auth login' for auto-upload).")
        subprocess.run(["open", "-R", str(out_path)], check=False)
        return

    auth = subprocess.run([gh_bin, "auth", "status"], cwd=repo_dir, env=env, capture_output=True, text=True, check=False)
    if auth.returncode != 0:
        print("Note: 'gh' CLI is not logged in (run 'gh auth login' for auto-upload). Revealing file in Finder.")
        subprocess.run(["open", "-R", str(out_path)], check=False)
        return

    check_git_state(repo_dir, env)

    api_res = subprocess.run(
        [gh_bin, "api", "repos/:owner/:repo/releases"],
        cwd=repo_dir,
        env=env,
        stdout=subprocess.PIPE,
        text=True,
        check=True,
    )
    releases = json.loads(api_res.stdout)
    existing = next((r for r in releases if r.get("tag_name") == tag), None)

    if existing:
        if not existing.get("draft", False):
            sys.exit(
                f"Error: Release {tag} is already published on GitHub. "
                "Bump <key>version</key> in info.plist before creating a new release."
            )
        print(f"Updating {WORKFLOW_FILENAME} on existing draft release {tag} (preserving draft notes)...")
        subprocess.run(
            [gh_bin, "release", "upload", tag, str(out_path), "--clobber"],
            cwd=repo_dir,
            env=env,
            check=True,
        )
        html_url = existing["html_url"]
    else:
        print(f"Creating GitHub draft release {tag} and uploading {WORKFLOW_FILENAME}...")
        create_res = subprocess.run(
            [gh_bin, "release", "create", tag, str(out_path), "--draft", "--title", tag, "--notes", ""],
            cwd=repo_dir,
            env=env,
            stdout=subprocess.PIPE,
            text=True,
            check=True,
        )
        html_url = create_res.stdout.strip().splitlines()[-1]

    edit_url = html_url.replace("/releases/tag/", "/releases/edit/")
    print(f"Opening draft release editor: {edit_url}")
    subprocess.run(["open", edit_url], check=False)


def main() -> None:
    repo_dir = Path(__file__).resolve().parent.parent
    env = build_env()
    version = read_plist_version(repo_dir / "info.plist")
    tag = f"v{version}"
    out_path = Path.home() / "Downloads" / WORKFLOW_FILENAME

    build_arm64_binaries(repo_dir, env)
    package_workflow(repo_dir, out_path)
    upload_github_draft(repo_dir, out_path, tag, env)


if __name__ == "__main__":
    main()
