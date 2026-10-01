// kbd-rgb sets a VIA/Vial keyboard's RGB lighting to the current Omarchy theme accent.
package main

import (
	"bufio"
	"bytes"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	vid       = "00004653"
	pid       = "00000004"
	rawUsage  = "0660ff" // usage page 0xFF60 (VIA raw HID)
	reportLen = 32
)

func findDevice() (string, error) {
	paths, _ := filepath.Glob("/sys/class/hidraw/hidraw*")
	for _, p := range paths {
		uevent, err := os.ReadFile(filepath.Join(p, "device/uevent"))
		if err != nil || !bytes.Contains(uevent, []byte("HID_ID=0003:"+vid+":"+pid)) {
			continue
		}
		desc, err := os.ReadFile(filepath.Join(p, "device/report_descriptor"))
		if err != nil || !bytes.HasPrefix(desc, []byte{0x06, 0x60, 0xff}) {
			continue
		}
		return "/dev/" + filepath.Base(p), nil
	}
	return "", fmt.Errorf("Corne raw HID interface not found")
}

func send(f *os.File, cmd ...byte) ([]byte, error) {
	buf := make([]byte, reportLen+1) // leading 0x00 report ID
	copy(buf[1:], cmd)
	if _, err := f.Write(buf); err != nil {
		return nil, err
	}
	resp := make([]byte, reportLen)
	done := make(chan error, 1)
	go func() { _, err := f.Read(resp); done <- err }()
	select {
	case err := <-done:
		return resp, err
	case <-time.After(time.Second):
		return nil, fmt.Errorf("timeout waiting for reply to %x", cmd)
	}
}

func themeColor(key string) (string, error) {
	home, _ := os.UserHomeDir()
	f, err := os.Open(filepath.Join(home, ".local/state/omarchy/current/theme/colors.toml"))
	if err != nil {
		return "", err
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	for s.Scan() {
		k, v, ok := strings.Cut(s.Text(), "=")
		if ok && strings.TrimSpace(k) == key {
			return strings.Trim(strings.TrimSpace(v), `"`), nil
		}
	}
	return "", fmt.Errorf("%s not found in colors.toml", key)
}

// hexToHSV returns QMK-style 0-255 hue, saturation and value.
func hexToHSV(hex string) (byte, byte, byte, error) {
	n, err := strconv.ParseUint(strings.TrimPrefix(hex, "#"), 16, 32)
	if err != nil {
		return 0, 0, 0, err
	}
	r, g, b := float64(n>>16&0xff)/255, float64(n>>8&0xff)/255, float64(n&0xff)/255
	mx, mn := math.Max(r, math.Max(g, b)), math.Min(r, math.Min(g, b))
	d := mx - mn
	var h float64
	switch {
	case d == 0:
	case mx == r:
		h = math.Mod((g-b)/d, 6)
	case mx == g:
		h = (b-r)/d + 2
	default:
		h = (r-g)/d + 4
	}
	h *= 60
	if h < 0 {
		h += 360
	}
	s := 0.0
	if mx > 0 {
		s = d / mx
	}
	return byte(math.Round(h / 360 * 255)), byte(math.Round(s * 255)), byte(math.Round(mx * 255)), nil
}

func main() {
	dev, err := findDevice()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	f, err := os.OpenFile(dev, os.O_RDWR, 0)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer f.Close()

	if len(os.Args) > 1 && os.Args[1] == "probe" {
		for _, c := range [][]byte{{0x01}, {0xFE, 0x00}, {0x08, 0x40}, {0x08, 0x03, 0x01}, {0x08, 0x03, 0x02}, {0x08, 0x03, 0x04}} {
			r, err := send(f, c...)
			fmt.Printf("%x -> %x %v\n", c, r, err)
		}
		return
	}

	key := "accent"
	if len(os.Args) > 1 {
		key = os.Args[1]
	}
	hex := key
	if !strings.HasPrefix(key, "#") {
		if hex, err = themeColor(key); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
	h, s, _, err := hexToHSV(hex)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// VialRGB info: [.., version lo, version hi, max brightness]
	info, err := send(f, 0x08, 0x40)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	v := info[4]

	// LEDs wash out pastel colors to near-white, so push chromatic colors toward
	// full saturation while leaving greys/whites alone.
	if s > 20 {
		s = byte(math.Min(255, math.Max(float64(s)*2.5, 200)))
	}

	// VialRGB set mode: SOLID_COLOR (2), speed, h, s, v; then persist to EEPROM.
	if _, err := send(f, 0x07, 0x41, 0x02, 0x00, 0x80, h, s, v); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if _, err := send(f, 0x09); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("%s -> %s (h=%d s=%d v=%d)\n", key, hex, h, s, v)
}
