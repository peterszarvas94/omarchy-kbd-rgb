#!/usr/bin/env bash
set -euo pipefail

repo_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
cd "$repo_dir"
go build -o kbd-rgb .
install -Dm755 kbd-rgb "$HOME/.local/bin/kbd-rgb"
install -Dm644 systemd/kbd-rgb.service "$HOME/.config/systemd/user/kbd-rgb.service"
install -Dm644 systemd/kbd-rgb.path "$HOME/.config/systemd/user/kbd-rgb.path"
systemctl --user daemon-reload
systemctl --user enable --now kbd-rgb.path
systemctl --user restart kbd-rgb.path
systemctl --user start kbd-rgb.service
