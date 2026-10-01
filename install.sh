#!/bin/sh
set -eu

for command in go curl tar install systemctl; do
  command -v "$command" >/dev/null 2>&1 || {
    printf 'Missing dependency: %s\n' "$command" >&2
    exit 1
  }
done

if [ "$(uname -s)" != Linux ]; then
  printf 'This installer requires Linux.\n' >&2
  exit 1
fi

work_dir=$(mktemp -d)
trap 'rm -rf "$work_dir"' 0
trap 'exit 1' HUP INT TERM
curl -fsSL https://github.com/peterszarvas94/omarchy-kbd-rgb/archive/refs/heads/master.tar.gz -o "$work_dir/source.tar.gz"
tar -xzf "$work_dir/source.tar.gz" -C "$work_dir"
cd "$work_dir/omarchy-kbd-rgb-master"
go build -o "$work_dir/kbd-rgb" .
install -Dm755 "$work_dir/kbd-rgb" "$HOME/.local/bin/kbd-rgb"
install -Dm644 systemd/kbd-rgb.service "$HOME/.config/systemd/user/kbd-rgb.service"
install -Dm644 systemd/kbd-rgb.path "$HOME/.config/systemd/user/kbd-rgb.path"
systemctl --user daemon-reload
systemctl --user enable --now kbd-rgb.path
systemctl --user restart kbd-rgb.path
if ! systemctl --user start kbd-rgb.service; then
  printf 'Installed, but applying the current color failed. Check the keyboard connection and journalctl --user -u kbd-rgb.service.\n' >&2
  exit 1
fi
printf 'Installed kbd-rgb and enabled automatic theme updates.\n'
