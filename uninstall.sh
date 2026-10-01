#!/bin/sh
set -eu

systemctl --user disable --now kbd-rgb.path
systemctl --user stop kbd-rgb.service
rm -f "$HOME/.config/systemd/user/kbd-rgb.path" "$HOME/.config/systemd/user/kbd-rgb.service"
rm -f "$HOME/.local/bin/kbd-rgb"
systemctl --user daemon-reload
printf 'Uninstalled kbd-rgb. The keyboard retains its last saved color.\n'
