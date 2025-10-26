#!/usr/bin/env python3
"""Download keymap files for supported editors into chore directories."""
from __future__ import annotations

import argparse
import pathlib
import sys
import urllib.error
import urllib.parse
import urllib.request

PRESETS = {
    "intellij": {
        "base_url": (
            "https://raw.githubusercontent.com/JetBrains/intellij-community/refs/heads/master/"
            "platform/platform-resources/src/keymaps"
        ),
        "files": [
            "$default.xml",
            "Default for GNOME.xml",
            # "Default for KDE.xml",
            "Default for XWin.xml",
            # "Emacs.xml",
            # "Mac OS X 10.5+.xml",
            "Mac OS X.xml",
            # "Sublime Text (Mac OS X).xml",
            # "Sublime Text.xml",
            # "macOS System Shortcuts.xml",
        ],
        "target_dir": pathlib.Path("chore") / "intellij",
    },
    "zed": {
        "base_url": (
            "https://raw.githubusercontent.com/zed-industries/zed/refs/heads/main/assets/keymaps"
        ),
        "files": [
            "default-linux.json",
            "default-macos.json",
            "default-windows.json",
            # "initial.json",
            # "linux/atom.json",
            # "linux/cursor.json",
            # "linux/emacs.json",
            # "linux/jetbrains.json",
            # "linux/sublime_text.json",
            # "macos/atom.json",
            # "macos/cursor.json",
            # "macos/emacs.json",
            # "macos/jetbrains.json",
            # "macos/sublime_text.json",
            # "macos/textmate.json",
            # "storybook.json",
            # "vim.json",
        ],
        "target_dir": pathlib.Path("chore") / "zed",
    },
}


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Download keymap files for supported editors into chore directories"
    )
    parser.add_argument(
        "--preset",
        choices=sorted(PRESETS.keys()),
        default="intellij",
        help="Preset to download (controls base URL, default files, and target directory)",
    )
    parser.add_argument(
        "files",
        metavar="FILE",
        nargs="*",
        help="Specific keymap filenames to download (defaults to preset files)",
    )
    parser.add_argument(
        "--base-url",
        default=None,
        help="Base URL to fetch keymap files from (defaults to preset base URL)",
    )
    parser.add_argument(
        "--target-dir",
        default=None,
        help="Directory to write the downloaded files to (defaults to preset directory)",
    )
    return parser.parse_args()


def download(url: str) -> bytes:
    with urllib.request.urlopen(url) as response:
        return response.read()


def main() -> int:
    args = parse_args()
    preset = PRESETS[args.preset]
    base_url = (args.base_url or preset["base_url"]).rstrip("/")
    files = args.files or preset["files"]

    target_dir_str = args.target_dir or str(preset["target_dir"])
    target_dir = pathlib.Path(target_dir_str)
    target_dir.mkdir(parents=True, exist_ok=True)

    for name in files:
        encoded = urllib.parse.quote(name)
        url = f"{base_url}/{encoded}"
        print(f"Downloading {name} -> {url}")
        try:
            data = download(url)
        except urllib.error.URLError as exc:
            print(f"Failed to download {name}: {exc}", file=sys.stderr)
            return 1
        destination = target_dir / name
        destination.parent.mkdir(parents=True, exist_ok=True)
        destination.write_bytes(data)
    return 0


if __name__ == "__main__":
    sys.exit(main())
