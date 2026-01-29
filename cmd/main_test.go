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

// setupTestIconDir creates tmpDir with assets/icon and Resources/assets/icon containing iconName.png, sets AMM_ICON_DIR and Chdir. Caller must defer os.Chdir(origWd) and os.Unsetenv("AMM_ICON_DIR"). Returns tmpDir.
func setupTestIconDir(t *testing.T, iconName string) string {
	t.Helper()
	tmpDir := t.TempDir()
	assetsDir := filepath.Join(tmpDir, "assets", "icon")
	resDir := filepath.Join(tmpDir, "Resources", "assets", "icon")
	iconPathAssets := filepath.Join(assetsDir, iconName+".png")
	iconPathRes := filepath.Join(resDir, iconName+".png")
	if err := writeTestPNG(iconPathAssets); err != nil {
		t.Fatalf("writeTestPNG assets error: %v", err)
	}
	if err := writeTestPNG(iconPathRes); err != nil {
		t.Fatalf("writeTestPNG resources error: %v", err)
	}
	origWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Chdir: %v", err)
	}
	t.Cleanup(func() { os.Chdir(origWd) })
	os.Setenv("AMM_ICON_DIR", tmpDir)
	t.Cleanup(func() { os.Unsetenv("AMM_ICON_DIR") })
	return tmpDir
}

func TestGetIconOriginalAndInvalidName(t *testing.T) {
	tmpDir := t.TempDir()
	assetsDir := filepath.Join(tmpDir, "assets", "icon")
	resDir := filepath.Join(tmpDir, "Resources", "assets", "icon")
	mousePathAssets := filepath.Join(assetsDir, "mouse.png")
	mousePathRes := filepath.Join(resDir, "mouse.png")
	if err := writeTestPNG(mousePathAssets); err != nil {
		t.Fatalf("writeTestPNG assets error: %v", err)
	}
	if err := writeTestPNG(mousePathRes); err != nil {
		t.Fatalf("writeTestPNG resources error: %v", err)
	}
	origWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Chdir: %v", err)
	}
	defer os.Chdir(origWd)
	os.Setenv("AMM_ICON_DIR", tmpDir)
	defer os.Unsetenv("AMM_ICON_DIR")

	// call getTrayIcon for valid name, inactive (alpha reduced to alphaInactive)
	b := getTrayIcon("mouse", false, "")
	if len(b) == 0 {
		t.Fatalf("getTrayIcon returned empty bytes")
	}
	img, err := png.Decode(bytes.NewReader(b))
	if err != nil {
		t.Fatalf("decode returned bytes error: %v", err)
	}
	if bnd := img.Bounds(); bnd.Dx() != 2 || bnd.Dy() != 2 {
		t.Fatalf("expected 2x2 image (test PNG), got %dx%d", bnd.Dx(), bnd.Dy())
	}
	r, g, bl, a := getRGBA8(img, 0, 0)
	if g != 0 || bl != 0 {
		t.Fatalf("expected (?,0,0,?) at (0,0), got (%d,%d,%d,%d)", r, g, bl, a)
	}
	if r == 0 {
		t.Fatalf("expected non-zero R at (0,0), got (%d,%d,%d,%d)", r, g, bl, a)
	}
	expectedAlpha := uint8(255 * alphaInactive) // 153
	if a != expectedAlpha {
		t.Fatalf("expected alpha %d (inactive) at (0,0), got %d", expectedAlpha, a)
	}
	// transparent pixel should remain transparent
	_, _, _, a2 := getRGBA8(img, 0, 1)
	if a2 != 0 {
		t.Fatalf("expected alpha 0 at (0,1), got %d", a2)
	}

	// calling with invalid name must default to mouse -> validate pixel
	b2 := getTrayIcon("invalid-name", false, "")
	img2, err := png.Decode(bytes.NewReader(b2))
	if err != nil {
		t.Fatalf("decode invalid-name returned bytes error: %v", err)
	}
	r3, _, _, a3 := getRGBA8(img2, 0, 0)
	expectedAlpha3 := uint8(255 * alphaInactive)
	if a3 != expectedAlpha3 {
		t.Fatalf("invalid-name did not default to mouse image (inactive alpha); got a=%d want %d", a3, expectedAlpha3)
	}
	if r3 == 0 {
		t.Fatalf("invalid-name did not default to mouse image; got r=%d", r3)
	}
}

func TestGetIconActiveRecolors(t *testing.T) {
	tmpDir := t.TempDir()
	assetsDir := filepath.Join(tmpDir, "assets", "icon")
	resDir := filepath.Join(tmpDir, "Resources", "assets", "icon")
	mousePathAssets := filepath.Join(assetsDir, "mouse.png")
	mousePathRes := filepath.Join(resDir, "mouse.png")
	if err := writeTestPNG(mousePathAssets); err != nil {
		t.Fatalf("writeTestPNG assets error: %v", err)
	}
	if err := writeTestPNG(mousePathRes); err != nil {
		t.Fatalf("writeTestPNG resources error: %v", err)
	}
	origWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Chdir: %v", err)
	}
	defer os.Chdir(origWd)
	os.Setenv("AMM_ICON_DIR", tmpDir)
	defer os.Unsetenv("AMM_ICON_DIR")

	// call getTrayIcon with active=true and color "blue" -> recolor to blue
	b := getTrayIcon("mouse", true, "blue")
	if len(b) == 0 {
		t.Fatalf("getTrayIcon active+blue returned empty bytes")
	}
	img, err := png.Decode(bytes.NewReader(b))
	if err != nil {
		t.Fatalf("decode active returned bytes error: %v", err)
	}
	// non-transparent pixels should be recolored to blue (30,144,255,255)
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

func TestGetTrayIconActiveNoColor(t *testing.T) {
	setupTestIconDir(t, "mouse")

	b := getTrayIcon("mouse", true, "")
	if len(b) == 0 {
		t.Fatalf("getTrayIcon returned empty bytes")
	}
	img, err := png.Decode(bytes.NewReader(b))
	if err != nil {
		t.Fatalf("decode returned bytes error: %v", err)
	}
	if bnd := img.Bounds(); bnd.Dx() != 2 || bnd.Dy() != 2 {
		t.Fatalf("expected 2x2 image (test PNG), got %dx%d", bnd.Dx(), bnd.Dy())
	}
	r, g, bl, a := getRGBA8(img, 0, 0)
	if r != 255 || g != 0 || bl != 0 || a != 255 {
		t.Fatalf("expected (255,0,0,255) at (0,0), got (%d,%d,%d,%d)", r, g, bl, a)
	}
	_, _, _, a2 := getRGBA8(img, 0, 1)
	if a2 != 0 {
		t.Fatalf("expected alpha 0 at (0,1), got %d", a2)
	}

	// result must match loadIconFile (original bytes)
	orig := getMenuIcon("mouse")
	if !bytes.Equal(b, orig) {
		t.Fatalf("getTrayIcon(active, no color) should return same bytes as loadIconFile")
	}
}

func TestGetTrayIconColorsRedWhite(t *testing.T) {
	setupTestIconDir(t, "mouse")

	t.Run("red", func(t *testing.T) {
		b := getTrayIcon("mouse", true, "red")
		if len(b) == 0 {
			t.Fatalf("getTrayIcon red returned empty bytes")
		}
		img, err := png.Decode(bytes.NewReader(b))
		if err != nil {
			t.Fatalf("decode: %v", err)
		}
		r, g, bl, a := getRGBA8(img, 0, 0)
		if r != 255 || g != 0 || bl != 0 || a != 255 {
			t.Fatalf("expected colorRed (255,0,0,255) at (0,0), got (%d,%d,%d,%d)", r, g, bl, a)
		}
		_, _, _, a2 := getRGBA8(img, 0, 1)
		if a2 != 0 {
			t.Fatalf("expected alpha 0 at (0,1), got %d", a2)
		}
	})

	t.Run("white", func(t *testing.T) {
		b := getTrayIcon("mouse", true, "white")
		if len(b) == 0 {
			t.Fatalf("getTrayIcon white returned empty bytes")
		}
		img, err := png.Decode(bytes.NewReader(b))
		if err != nil {
			t.Fatalf("decode: %v", err)
		}
		r, g, bl, a := getRGBA8(img, 0, 0)
		if r != 255 || g != 255 || bl != 255 || a != 255 {
			t.Fatalf("expected colorWhite (255,255,255,255) at (0,0), got (%d,%d,%d,%d)", r, g, bl, a)
		}
		_, _, _, a2 := getRGBA8(img, 0, 1)
		if a2 != 0 {
			t.Fatalf("expected alpha 0 at (0,1), got %d", a2)
		}
	})
}

func TestGetMenuIcon(t *testing.T) {
	setupTestIconDir(t, "mouse")

	menuIcon := getMenuIcon("mouse")
	if len(menuIcon) == 0 {
		t.Fatalf("getMenuIcon returned empty bytes")
	}
	img, err := png.Decode(bytes.NewReader(menuIcon))
	if err != nil {
		t.Fatalf("decode getMenuIcon bytes: %v", err)
	}
	if bnd := img.Bounds(); bnd.Dx() != 2 || bnd.Dy() != 2 {
		t.Fatalf("expected 2x2 image, got %dx%d", bnd.Dx(), bnd.Dy())
	}
	// getMenuIcon should return same as loadIconFile -> same as getTrayIcon(active, no color)
	orig := getTrayIcon("mouse", true, "")
	if !bytes.Equal(menuIcon, orig) {
		t.Fatalf("getMenuIcon should return same bytes as getTrayIcon(active, no color)")
	}
}

func TestLoadIconFilePanicWhenNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	origWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Chdir: %v", err)
	}
	t.Cleanup(func() { os.Chdir(origWd) })
	os.Setenv("AMM_ICON_DIR", tmpDir)
	t.Cleanup(func() { os.Unsetenv("AMM_ICON_DIR") })
	// tmpDir has no assets/icon/mouse.png, so loadIconFile (via getTrayIcon) must panic
	var panicked interface{}
	func() {
		defer func() { panicked = recover() }()
		getTrayIcon("mouse", true, "")
	}()
	if panicked == nil {
		t.Fatalf("expected panic when icon file not found")
	}
	errMsg, ok := panicked.(string)
	if !ok {
		t.Fatalf("expected string panic message, got %T: %v", panicked, panicked)
	}
	if errMsg != "Failed to load icon: mouse.png" {
		t.Fatalf("expected panic \"Failed to load icon: mouse.png\", got %q", errMsg)
	}
}
