package main

import (
	"strings"
	"testing"

	"github.com/Fiend3d/catatui"
)

// TestRender draws the banner at sizes from nothing to bigger than a screen,
// including the 68x16 it is meant for. Rendering outside the area given panics
// in catatui, so this is what keeps the example honest when the library
// changes.
func TestRender(t *testing.T) {
	sizes := [][2]uint16{{0, 0}, {1, 1}, {3, 2}, {10, 4}, {40, 12}, {68, 16}, {80, 24}, {200, 60}}
	for _, size := range sizes {
		terminal, err := catatui.NewTerminal(catatui.NewTestBackend(size[0], size[1]))
		if err != nil {
			t.Fatalf("%dx%d: %v", size[0], size[1], err)
		}
		err = terminal.Draw(func(f *catatui.Frame) { render(f, "0.1.0", "Ratatouille") })
		if err != nil {
			t.Fatalf("%dx%d: %v", size[0], size[1], err)
		}
	}
}

// TestBannerContents checks the banner drawn at its own size holds what it is
// for: the version, and every package named on the menu.
func TestBannerContents(t *testing.T) {
	backend := catatui.NewTestBackend(68, 16)
	terminal, err := catatui.NewTerminal(backend)
	if err != nil {
		t.Fatal(err)
	}
	if err := terminal.Draw(func(f *catatui.Frame) { render(f, "1.2.3", "Bryndza") }); err != nil {
		t.Fatal(err)
	}

	screen := backend.Buffer().String()
	for _, want := range append(append([]string{
		`v1.2.3 "Bryndza"`, "Main Courses", "Pairings",
	}, mainDishes...), pairings...) {
		if !strings.Contains(screen, want) {
			t.Errorf("the banner does not show %q:\n%s", want, screen)
		}
	}

	// The two menu blocks overlap by a row, so the border they share is drawn
	// once, as a tee rather than as two corners.
	if !strings.Contains(screen, "├") {
		t.Errorf("the menu blocks did not merge their borders:\n%s", screen)
	}
}

// TestGradientCoversEveryLetterAndRow checks every letter of the wordmark and
// every row of the glow mixes to a real colour, since the tables are indexed
// directly.
func TestGradientCoversEveryLetterAndRow(t *testing.T) {
	for letter := range 7 {
		for row := range len(ambientGradient) {
			color := gradientColor(rainbow(letter), row)
			if _, _, _, ok := color.RGB(); !ok {
				t.Errorf("letter %d row %d is not an RGB colour", letter, row)
			}
		}
	}
	// A row off the end of the gradient falls back rather than panicking, which
	// is what a shorter area would ask for.
	if got := gradientColor(red, len(ambientGradient)); got != bgColor {
		t.Errorf("a row past the gradient gave %v, want the background", got)
	}
}
