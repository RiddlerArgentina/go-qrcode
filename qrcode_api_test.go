// go-qrcode
// Copyright 2014 Tom Harwood

package qrcode

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewEmptyContent(t *testing.T) {
	_, err := New("", Medium)
	if err == nil {
		t.Fatal("expected error encoding empty content")
	}
}

func TestNewInvalidRecoveryLevel(t *testing.T) {
	_, err := New("hello", RecoveryLevel(99))
	if err == nil {
		t.Fatal("expected error for invalid recovery level")
	}
}

func TestNewUTF8Content(t *testing.T) {
	q, err := New("こんにちは", Medium)
	if err != nil {
		t.Fatal(err)
	}
	if q.Content != "こんにちは" {
		t.Errorf("Content = %q", q.Content)
	}
	if len(q.Bitmap()) == 0 {
		t.Fatal("expected non-empty bitmap")
	}
}

func TestNewWithForcedVersion(t *testing.T) {
	q, err := NewWithForcedVersion("hello", 5, Medium)
	if err != nil {
		t.Fatal(err)
	}
	if q.VersionNumber != 5 {
		t.Errorf("VersionNumber = %d, want 5", q.VersionNumber)
	}

	if _, err := NewWithForcedVersion("hello", 0, Medium); err == nil {
		t.Error("expected error for version 0")
	}
	if _, err := NewWithForcedVersion("hello", 41, Medium); err == nil {
		t.Error("expected error for version 41")
	}

	tooLong := strings.Repeat("A", 100)
	if _, err := NewWithForcedVersion(tooLong, 1, Highest); err == nil {
		t.Error("expected error when content exceeds forced version capacity")
	}
}

func TestNewWithForcedVersionAllVersions(t *testing.T) {
	for version := 1; version <= 40; version++ {
		q, err := NewWithForcedVersion("x", version, Low)
		if err != nil {
			t.Errorf("version %d: %s", version, err)
			continue
		}
		if q.VersionNumber != version {
			t.Errorf("version %d: VersionNumber = %d", version, q.VersionNumber)
		}
		bitmap := q.Bitmap()
		if len(bitmap) == 0 || len(bitmap) != len(bitmap[0]) {
			t.Errorf("version %d: expected square non-empty bitmap, got %dx%d",
				version, len(bitmap), len(bitmap[0]))
		}
	}
}

func TestEncodeSpecialStrings(t *testing.T) {
	// Mixed numeric/alphanumeric strings from the add_special_string_tests branch.
	// These used to exercise segment-optimisation bugs.
	testStrings := []string{
		"8888888888",
		"88888888885555",
		"8888888888aaaa",
		"8888888888aaaa8888",
		"8888888888aaaa8888a8a8a8a",
		"8888888888aaaa8888a8a8a8o",
		"8aaaa8o",
		"8aaaa8oooooo8o8o8o8o",
		"16aaaa",
		"2aaaa",
		"3aaaa",
		"a3aaaa",
		"##3aaaa",
		"https://example.org",
		"HTTPS://EXAMPLE.ORG/123",
	}

	for _, content := range testStrings {
		for _, level := range []RecoveryLevel{Low, Medium, High, Highest} {
			q, err := New(content, level)
			if err != nil {
				t.Errorf("New(%q, %v): %s", content, level, err)
				continue
			}
			if q.Content != content {
				t.Errorf("Content = %q, want %q", q.Content, content)
			}
			if len(q.Bitmap()) == 0 {
				t.Errorf("empty bitmap for %q level=%v", content, level)
			}
		}
	}
}

func TestEncodePNGMagic(t *testing.T) {
	png, err := Encode("https://example.org", Medium, 256)
	if err != nil {
		t.Fatal(err)
	}
	if len(png) < 8 {
		t.Fatalf("PNG too short: %d bytes", len(png))
	}
	magic := []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a}
	if !bytes.Equal(png[:8], magic) {
		t.Errorf("missing PNG magic, got %x", png[:8])
	}
}

func TestWriteAndWriteFile(t *testing.T) {
	q, err := New("hello", Medium)
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if err := q.Write(64, &buf); err != nil {
		t.Fatal(err)
	}
	pngBytes, err := q.PNG(64)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(buf.Bytes(), pngBytes) {
		t.Error("Write() output does not match PNG()")
	}

	path := filepath.Join(t.TempDir(), "qr.png")
	if err := q.WriteFile(64, path); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, pngBytes) {
		t.Error("WriteFile() contents do not match PNG()")
	}

	if err := WriteFile("hello", Medium, 64, filepath.Join(t.TempDir(), "pkg.png")); err != nil {
		t.Fatal(err)
	}
}

func TestWriteColorFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "color.png")
	fg := color.RGBA{R: 0x33, G: 0x33, B: 0x66, A: 0xff}
	bg := color.RGBA{R: 0xef, G: 0xef, B: 0xef, A: 0xff}

	if err := WriteColorFile("hello", Medium, 64, bg, fg, path); err != nil {
		t.Fatal(err)
	}

	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	img, err := png.Decode(f)
	if err != nil {
		t.Fatal(err)
	}
	pimg, ok := img.(*image.Paletted)
	if !ok {
		t.Fatalf("expected paletted image, got %T", img)
	}
	if !colorsEqual(pimg.Palette[0], bg) {
		t.Errorf("background = %v, want %v", pimg.Palette[0], bg)
	}
	if !colorsEqual(pimg.Palette[1], fg) {
		t.Errorf("foreground = %v, want %v", pimg.Palette[1], fg)
	}
}

func TestWriteColorFileTooLong(t *testing.T) {
	// Regression for skip2/go-qrcode#67: error must be returned, not a nil deref.
	tooLong := strings.Repeat("#", 3000)
	path := filepath.Join(t.TempDir(), "too-long.png")
	err := WriteColorFile(tooLong, Low, 64, color.White, color.Black, path)
	if err == nil {
		t.Fatal("expected error for content that exceeds QR capacity")
	}
}

func TestBitmapFinderPatternsAndQuietZone(t *testing.T) {
	q, err := New("A", Medium)
	if err != nil {
		t.Fatal(err)
	}

	bitmap := q.Bitmap()
	// Version 1 is 21 modules; default quiet zone is 4 on each side.
	want := 21 + 8
	if len(bitmap) != want || len(bitmap[0]) != want {
		t.Fatalf("bitmap size %dx%d, want %dx%d", len(bitmap), len(bitmap[0]), want, want)
	}

	// Quiet zone is unset.
	for i := 0; i < 4; i++ {
		for j := 0; j < want; j++ {
			if bitmap[i][j] || bitmap[want-1-i][j] || bitmap[j][i] || bitmap[j][want-1-i] {
				t.Fatalf("quiet zone module set at border offset %d", i)
			}
		}
	}

	assertFinderPattern(t, bitmap, 4, 4)
	assertFinderPattern(t, bitmap, want-4-7, 4)
	assertFinderPattern(t, bitmap, 4, want-4-7)
}

func TestDisableBorder(t *testing.T) {
	q, err := New("A", Medium)
	if err != nil {
		t.Fatal(err)
	}
	q.DisableBorder = true

	bitmap := q.Bitmap()
	if len(bitmap) != 21 || len(bitmap[0]) != 21 {
		t.Fatalf("borderless bitmap size %dx%d, want 21x21", len(bitmap), len(bitmap[0]))
	}
	assertFinderPattern(t, bitmap, 0, 0)
}

func TestImageSizes(t *testing.T) {
	q, err := New("A", Medium)
	if err != nil {
		t.Fatal(err)
	}

	img := q.Image(256)
	if img.Bounds().Dx() != 256 || img.Bounds().Dy() != 256 {
		t.Errorf("Image(256) size = %v, want 256x256", img.Bounds())
	}

	// Too-small size is silently raised to the symbol size (29 with quiet zone).
	small := q.Image(8)
	if small.Bounds().Dx() != 29 {
		t.Errorf("Image(8) width = %d, want 29", small.Bounds().Dx())
	}

	// Negative size: each module is |size| pixels.
	variable := q.Image(-4)
	if variable.Bounds().Dx() != 29*4 {
		t.Errorf("Image(-4) width = %d, want %d", variable.Bounds().Dx(), 29*4)
	}
}

func TestImageCustomColors(t *testing.T) {
	q, err := New("A", Medium)
	if err != nil {
		t.Fatal(err)
	}
	q.ForegroundColor = color.RGBA{R: 0xff, A: 0xff}
	q.BackgroundColor = color.RGBA{B: 0xff, A: 0xff}

	img, ok := q.Image(64).(*image.Paletted)
	if !ok {
		t.Fatal("expected paletted image")
	}
	if !colorsEqual(img.Palette[0], q.BackgroundColor) {
		t.Errorf("palette[0] = %v, want %v", img.Palette[0], q.BackgroundColor)
	}
	if !colorsEqual(img.Palette[1], q.ForegroundColor) {
		t.Errorf("palette[1] = %v, want %v", img.Palette[1], q.ForegroundColor)
	}
}

func TestEncodeIdempotent(t *testing.T) {
	q, err := New("hello", Medium)
	if err != nil {
		t.Fatal(err)
	}

	first := q.Bitmap()
	second := q.Bitmap()
	if !bitmapsEqual(first, second) {
		t.Error("Bitmap() is not stable across repeated calls")
	}

	// Image() also calls encode(); it must not panic or change the symbol.
	_ = q.Image(64)
	third := q.Bitmap()
	if !bitmapsEqual(first, third) {
		t.Error("Image() changed the encoded bitmap")
	}
}

func TestToStringAndToSmallString(t *testing.T) {
	q, err := New("hello", Medium)
	if err != nil {
		t.Fatal(err)
	}

	s := q.ToString(false)
	if !strings.Contains(s, "█") {
		t.Error("ToString() missing block characters")
	}
	if !strings.Contains(s, "\n") {
		t.Error("ToString() missing newlines")
	}

	inv := q.ToString(true)
	if s == inv {
		t.Error("inverse ToString() matched non-inverse output")
	}

	small := q.ToSmallString(false)
	if small == "" {
		t.Error("ToSmallString() empty")
	}
	if len(strings.Split(strings.TrimRight(small, "\n"), "\n")) >= len(strings.Split(strings.TrimRight(s, "\n"), "\n")) {
		t.Error("ToSmallString() should use fewer rows than ToString()")
	}
}

func TestISOAnnexIExampleVersion(t *testing.T) {
	q, err := New("01234567", Medium)
	if err != nil {
		t.Fatal(err)
	}
	if q.VersionNumber != 1 {
		t.Errorf("VersionNumber = %d, want 1", q.VersionNumber)
	}
	_ = q.Bitmap()
	if q.mask != 2 {
		t.Errorf("mask = %d, want 2", q.mask)
	}
}

func assertFinderPattern(t *testing.T, bitmap [][]bool, x, y int) {
	t.Helper()
	pattern := [][]bool{
		{true, true, true, true, true, true, true},
		{true, false, false, false, false, false, true},
		{true, false, true, true, true, false, true},
		{true, false, true, true, true, false, true},
		{true, false, true, true, true, false, true},
		{true, false, false, false, false, false, true},
		{true, true, true, true, true, true, true},
	}
	for dy, row := range pattern {
		for dx, want := range row {
			if bitmap[y+dy][x+dx] != want {
				t.Fatalf("finder at (%d,%d) module (%d,%d) = %v, want %v",
					x, y, dx, dy, bitmap[y+dy][x+dx], want)
			}
		}
	}
}

func colorsEqual(a, b color.Color) bool {
	ar, ag, ab, aa := a.RGBA()
	br, bg, bb, ba := b.RGBA()
	return ar == br && ag == bg && ab == bb && aa == ba
}

func bitmapsEqual(a, b [][]bool) bool {
	if len(a) != len(b) {
		return false
	}
	for y := range a {
		if len(a[y]) != len(b[y]) {
			return false
		}
		for x := range a[y] {
			if a[y][x] != b[y][x] {
				return false
			}
		}
	}
	return true
}
