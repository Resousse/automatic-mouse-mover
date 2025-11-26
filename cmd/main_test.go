package main

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

// helper: create a 2x2 PNG with known pixels and write to path
func writeTestPNG(path string) error {
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	// (0,0) opaque red
	img.Set(0, 0, color.RGBA{255, 0, 0, 255})
	// (1,0) opaque green
	img.Set(1, 0, color.RGBA{0, 255, 0, 255})
	// (0,1) transparent
	img.Set(0, 1, color.RGBA{0, 0, 0, 0})
	// (1,1) opaque white
	img.Set(1, 1, color.RGBA{255, 255, 255, 255})

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	fh, err := os.Create(path)
	if err != nil {
		return err
	}
	defer fh.Close()
	return png.Encode(fh, img)
}

func getRGBA8(img image.Image, x, y int) (r, g, b, a uint8) {
	rr, gg, bb, aa := img.At(x, y).RGBA()
	return uint8(rr >> 8), uint8(gg >> 8), uint8(bb >> 8), uint8(aa >> 8)
}

func TestGetIconOriginalAndInvalidName(t *testing.T) {
	ex, err := os.Executable()
	if err != nil {
		t.Fatalf("os.Executable error: %v", err)
	}
	dir := filepath.Dir(ex)

	assetsDir := filepath.Join(dir, "..", "assets", "icon")
	resDir := filepath.Join(dir, "..", "Resources", "assets", "icon")

	mousePathAssets := filepath.Join(assetsDir, "mouse.png")
	mousePathRes := filepath.Join(resDir, "mouse.png")

	// create both locations to be robust to either branch in getIcon
	if err := writeTestPNG(mousePathAssets); err != nil {
		t.Fatalf("writeTestPNG assets error: %v", err)
	}
	if err := writeTestPNG(mousePathRes); err != nil {
		t.Fatalf("writeTestPNG resources error: %v", err)
	}

	// call getIcon for valid name, inactive
	b := getIcon("mouse", false, "")
	if len(b) == 0 {
		t.Fatalf("getIcon returned empty bytes")
	}
	img, err := png.Decode(bytes.NewReader(b))
	if err != nil {
		t.Fatalf("decode returned bytes error: %v", err)
	}
	r, g, bl, a := getRGBA8(img, 0, 0)
	if r != 255 || g != 0 || bl != 0 || a != 255 {
		t.Fatalf("expected (255,0,0,255) at (0,0), got (%d,%d,%d,%d)", r, g, bl, a)
	}
	// transparent pixel should remain transparent
	_, _, _, a2 := getRGBA8(img, 0, 1)
	if a2 != 0 {
		t.Fatalf("expected alpha 0 at (0,1), got %d", a2)
	}

	// calling with invalid name must default to mouse -> validate pixel
	b2 := getIcon("invalid-name", false, "")
	img2, err := png.Decode(bytes.NewReader(b2))
	if err != nil {
		t.Fatalf("decode invalid-name returned bytes error: %v", err)
	}
	r3, _, _, a3 := getRGBA8(img2, 0, 0)
	if r3 != 255 || a3 != 255 {
		t.Fatalf("invalid-name did not default to mouse image; got (%d,%d)", r3, a3)
	}
}

func TestGetIconActiveRecolors(t *testing.T) {
	ex, err := os.Executable()
	if err != nil {
		t.Fatalf("os.Executable error: %v", err)
	}
	dir := filepath.Dir(ex)

	assetsDir := filepath.Join(dir, "..", "assets", "icon")
	resDir := filepath.Join(dir, "..", "Resources", "assets", "icon")

	mousePathAssets := filepath.Join(assetsDir, "mouse.png")
	mousePathRes := filepath.Join(resDir, "mouse.png")

	// write test PNGs in both locations
	if err := writeTestPNG(mousePathAssets); err != nil {
		t.Fatalf("writeTestPNG assets error: %v", err)
	}
	if err := writeTestPNG(mousePathRes); err != nil {
		t.Fatalf("writeTestPNG resources error: %v", err)
	}

	// call getIcon with active=true
	b := getIcon("mouse", true, "")
	if len(b) == 0 {
		t.Fatalf("getIcon active returned empty bytes")
	}
	img, err := png.Decode(bytes.NewReader(b))
	if err != nil {
		t.Fatalf("decode active returned bytes error: %v", err)
	}
	// non-transparent pixels should be recolored to (30,144,255,255)
	r, g, bl, a := getRGBA8(img, 0, 0)
	if r != 30 || g != 144 || bl != 255 || a != 255 {
		t.Fatalf("expected recolored (30,144,255,255) at (0,0), got (%d,%d,%d,%d)", r, g, bl, a)
	}
	// transparent pixel should remain transparent
	_, _, _, a2 := getRGBA8(img, 0, 1)
	if a2 != 0 {
		t.Fatalf("expected alpha 0 at (0,1) after active recolor, got %d", a2)
	}
}
