# Omarchy keyboard RGB

Match a Corne v4 keyboard's lighting to the active Omarchy theme accent, including Neon Glow's wallpaper-driven palette changes.

A small Go command reads `~/.local/state/omarchy/current/theme/colors.toml` and sets a solid color through raw HID using VialRGB. A systemd user path unit watches both `colors.toml` and `theme.name` and runs the command after changes.

## Requirements

- Linux with Omarchy and systemd user services.
- Go 1.27 or newer to build; no external Go dependencies.
- A Corne with USB vendor/product ID `4653:0004`, Vial firmware, and VialRGB support. These IDs are currently hardcoded in `main.go`; this is not a generic VIA keyboard tool.
- Read/write access to the keyboard's `/dev/hidraw*` interface through your existing device permissions.

The original setup was tested with a Corne v4 running VIA protocol 9 and VialRGB, on both keyboard halves. Its firmware reported a maximum brightness of 50. Other firmware and devices have not been verified.

## Install

```sh
git clone https://github.com/peterszarvas94/omarchy-kbd-rgb.git
cd omarchy-kbd-rgb
./install.sh
```

This builds and installs `~/.local/bin/kbd-rgb`, installs the two units under `~/.config/systemd/user/`, enables the watcher at login, and applies the current accent. Installing replaces existing files at those three locations. If the keyboard is disconnected or inaccessible, the final color application fails; the watcher remains installed.

If you previously installed a theme hook that runs `kbd-rgb`, remove that hook to avoid duplicate updates. The watcher handles both normal theme changes and Neon Glow palette changes.

## Manual use

```sh
kbd-rgb             # Current theme accent
kbd-rgb green       # Named color from colors.toml
kbd-rgb '#ff8800'    # Explicit six-digit RGB color
kbd-rgb probe       # Print firmware protocol responses
```

Colors with saturation above 20/255 are boosted to make pastel accents visible on LEDs. Greys and whites retain their saturation. Brightness always uses the maximum reported by the firmware, rather than the source color's brightness. Each color update is saved to the keyboard's EEPROM and survives unplugging; frequent palette changes also cause frequent EEPROM writes.

Two-color wallpapers use only the theme's accent. Gradients and secondary colors are not implemented.

## Troubleshooting

```sh
systemctl --user status kbd-rgb.path kbd-rgb.service
journalctl --user -u kbd-rgb.service
```

“Corne raw HID interface not found” means the expected USB device/interface was not found. “Permission denied” means your user cannot access its raw HID device; use your firmware's recommended device permissions. A timeout can indicate unsupported firmware or another application using the interface.

## Uninstall

```sh
systemctl --user disable --now kbd-rgb.path
systemctl --user stop kbd-rgb.service
rm ~/.config/systemd/user/kbd-rgb.path ~/.config/systemd/user/kbd-rgb.service
rm ~/.local/bin/kbd-rgb
systemctl --user daemon-reload
```

The last saved color remains on the keyboard.

## License

[MIT](LICENSE)
