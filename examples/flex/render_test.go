package main

import (
	"testing"

	"github.com/Fiend3d/catatui"
)

// TestRender draws every flex mode at sizes from nothing to bigger than a
// screen, at a few spacings and scroll offsets. Rendering outside the area
// given panics in catatui, so this is what keeps the example honest.
func TestRender(t *testing.T) {
	sizes := [][2]uint16{{0, 0}, {1, 1}, {3, 2}, {10, 4}, {40, 12}, {80, 24}, {200, 60}}
	for selected := range tabs {
		for _, spacing := range []uint16{0, 1, 4} {
			for _, offset := range []uint16{0, 10, maxScrollOffset()} {
				for _, size := range sizes {
					a := &app{
						selectedTab:  tab(selected),
						spacing:      spacing,
						scrollOffset: offset,
					}
					terminal, err := catatui.NewTerminal(catatui.NewTestBackend(size[0], size[1]))
					if err != nil {
						t.Fatalf("tab %d, %dx%d: %v", selected, size[0], size[1], err)
					}
					err = terminal.Draw(func(f *catatui.Frame) {
						f.RenderWidget(a, f.Area())
					})
					if err != nil {
						t.Fatalf("tab %d, %dx%d: %v", selected, size[0], size[1], err)
					}
				}
			}
		}
	}
}

// TestScrollStaysInRange checks the scroll cannot run past the last example, or
// before the first.
func TestScrollStaysInRange(t *testing.T) {
	a := &app{}
	for range int(maxScrollOffset()) + 10 {
		a.down()
	}
	if got, want := a.scrollOffset, maxScrollOffset(); got != want {
		t.Errorf("scrolled to %d, want to stop at %d", got, want)
	}
	for range int(maxScrollOffset()) + 10 {
		a.up()
	}
	if a.scrollOffset != 0 {
		t.Errorf("scrolled back to %d, want 0", a.scrollOffset)
	}
}

// TestTabsStayInRange checks the tabs stop at both ends rather than wrapping,
// which is what ratatui's example does.
func TestTabsStayInRange(t *testing.T) {
	a := &app{}
	for range len(tabs) + 5 {
		a.nextTab()
	}
	if got, want := a.selectedTab, tab(len(tabs)-1); got != want {
		t.Errorf("selected tab %d, want %d", got, want)
	}
	for range len(tabs) + 5 {
		a.previousTab()
	}
	if a.selectedTab != 0 {
		t.Errorf("selected tab %d, want 0", a.selectedTab)
	}
}
