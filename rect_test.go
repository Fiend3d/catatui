// Tests ported from ratatui-core/src/layout/rect.rs @ ratatui-v0.30.2

package catatui

import (
	"reflect"
	"testing"
)

func TestRectNew(t *testing.T) {
	if got, want := NewRect(1, 2, 3, 4), (Rect{X: 1, Y: 2, Width: 3, Height: 4}); got != want {
		t.Errorf("NewRect(1,2,3,4) = %+v, want %+v", got, want)
	}
}

func TestRectArea(t *testing.T) {
	if got := NewRect(1, 2, 3, 4).Area(); got != 12 {
		t.Errorf("Area() = %d, want 12", got)
	}
}

func TestRectIsEmpty(t *testing.T) {
	if NewRect(1, 2, 3, 4).IsEmpty() {
		t.Error("3x4 rect should not be empty")
	}
	if !NewRect(1, 2, 0, 4).IsEmpty() {
		t.Error("zero-width rect should be empty")
	}
	if !NewRect(1, 2, 3, 0).IsEmpty() {
		t.Error("zero-height rect should be empty")
	}
}

func TestRectEdges(t *testing.T) {
	r := NewRect(1, 2, 3, 4)
	for _, c := range []struct {
		name string
		got  uint16
		want uint16
	}{
		{"Left", r.Left(), 1},
		{"Right", r.Right(), 4},
		{"Top", r.Top(), 2},
		{"Bottom", r.Bottom(), 6},
	} {
		if c.got != c.want {
			t.Errorf("%s() = %d, want %d", c.name, c.got, c.want)
		}
	}
}

func TestRectInner(t *testing.T) {
	got := NewRect(1, 2, 3, 4).Inner(NewMargin(1, 2))
	want := NewRect(2, 4, 1, 0)
	if got != want {
		t.Errorf("Inner() = %+v, want %+v", got, want)
	}
}

func TestRectOuter(t *testing.T) {
	cases := []struct {
		name string
		rect Rect
		m    Margin
		want Rect
	}{
		{
			"enough space to grow on all sides",
			NewRect(100, 200, 10, 20), NewMargin(20, 30), NewRect(80, 170, 50, 80),
		},
		{
			"left/top saturation truncates the size",
			NewRect(10, 20, 10, 20), NewMargin(20, 30), NewRect(0, 0, 40, 70),
		},
		{
			"right/bottom saturation truncates the size",
			NewRect(maxU16-20, maxU16-40, 10, 20), NewMargin(20, 30), NewRect(maxU16-40, maxU16-70, 40, 70),
		},
	}
	for _, c := range cases {
		if got := c.rect.Outer(c.m); got != c.want {
			t.Errorf("%s: Outer() = %+v, want %+v", c.name, got, c.want)
		}
	}
}

func TestRectOffset(t *testing.T) {
	cases := []struct {
		name string
		rect Rect
		off  Offset
		want Rect
	}{
		{"positive", NewRect(1, 2, 3, 4), Offset{5, 6}, NewRect(6, 8, 3, 4)},
		{"negative", NewRect(4, 3, 3, 4), Offset{-2, -1}, NewRect(2, 2, 3, 4)},
		{"negative saturates at zero", NewRect(1, 2, 3, 4), Offset{-5, -6}, NewRect(0, 0, 3, 4)},
		{
			"saturating at max keeps the size",
			NewRect(maxU16-500, maxU16-500, 100, 100), Offset{1000, 1000},
			NewRect(maxU16-100, maxU16-100, 100, 100),
		},
	}
	for _, c := range cases {
		if got := c.rect.Offset(c.off); got != c.want {
			t.Errorf("%s: Offset() = %+v, want %+v", c.name, got, c.want)
		}
	}
}

func TestRectUnion(t *testing.T) {
	got := NewRect(1, 2, 3, 4).Union(NewRect(2, 3, 4, 5))
	want := NewRect(1, 2, 5, 6)
	if got != want {
		t.Errorf("Union() = %+v, want %+v", got, want)
	}
}

func TestRectIntersection(t *testing.T) {
	got := NewRect(1, 2, 3, 4).Intersection(NewRect(2, 3, 4, 5))
	if want := NewRect(2, 3, 2, 3); got != want {
		t.Errorf("Intersection() = %+v, want %+v", got, want)
	}
	// Non-overlapping rects must not underflow into a huge size.
	got = NewRect(1, 1, 2, 2).Intersection(NewRect(4, 4, 2, 2))
	if want := NewRect(4, 4, 0, 0); got != want {
		t.Errorf("Intersection() underflow = %+v, want %+v", got, want)
	}
}

func TestRectIntersects(t *testing.T) {
	if !NewRect(1, 2, 3, 4).Intersects(NewRect(2, 3, 4, 5)) {
		t.Error("overlapping rects should intersect")
	}
	if NewRect(1, 2, 3, 4).Intersects(NewRect(5, 6, 7, 8)) {
		t.Error("disjoint rects should not intersect")
	}
}

// TestRectMutualIntersect checks that Intersects is symmetric, including for
// rects that merely touch at a corner or edge (which do not intersect).
func TestRectMutualIntersect(t *testing.T) {
	cases := []struct {
		name string
		a, b Rect
	}{
		{"corner", NewRect(0, 0, 10, 10), NewRect(10, 10, 20, 20)},
		{"edge", NewRect(0, 0, 10, 10), NewRect(10, 0, 20, 10)},
		{"no intersect", NewRect(0, 0, 10, 10), NewRect(11, 11, 20, 20)},
		{"contains", NewRect(0, 0, 20, 20), NewRect(5, 5, 10, 10)},
	}
	for _, c := range cases {
		if c.a.Intersects(c.b) != c.b.Intersects(c.a) {
			t.Errorf("%s: Intersects is not symmetric for %+v and %+v", c.name, c.a, c.b)
		}
	}
}

func TestRectContains(t *testing.T) {
	r := NewRect(1, 2, 3, 4)
	cases := []struct {
		name string
		p    Position
		want bool
	}{
		{"inside top left", Position{1, 2}, true},
		{"inside top right", Position{3, 2}, true},
		{"inside bottom left", Position{1, 5}, true},
		{"inside bottom right", Position{3, 5}, true},
		{"outside left", Position{0, 2}, false},
		{"outside right", Position{4, 2}, false},
		{"outside top", Position{1, 1}, false},
		{"outside bottom", Position{1, 6}, false},
		{"outside top left", Position{0, 1}, false},
		{"outside bottom right", Position{4, 6}, false},
	}
	for _, c := range cases {
		if got := r.Contains(c.p); got != c.want {
			t.Errorf("%s: Contains(%+v) = %v, want %v", c.name, c.p, got, c.want)
		}
	}
}

// TestRectSizeTruncation is the guarantee that a Rect never extends past
// uint16 max: NewRect shrinks the size rather than letting Right() wrap.
func TestRectSizeTruncation(t *testing.T) {
	got := NewRect(maxU16-100, maxU16-1000, 200, 2000)
	want := Rect{X: maxU16 - 100, Y: maxU16 - 1000, Width: 100, Height: 1000}
	if got != want {
		t.Errorf("NewRect truncation = %+v, want %+v", got, want)
	}
	// A size that exactly fits must be preserved untouched.
	got = NewRect(maxU16-100, maxU16-1000, 100, 1000)
	if got != want {
		t.Errorf("NewRect preservation = %+v, want %+v", got, want)
	}
}

func TestRectResize(t *testing.T) {
	if got, want := NewRect(10, 20, 5, 5).Resize(Size{30, 40}), NewRect(10, 20, 30, 40); got != want {
		t.Errorf("Resize() = %+v, want %+v", got, want)
	}
	got := NewRect(maxU16-2, maxU16-3, 1, 1).Resize(Size{10, 10})
	if want := NewRect(maxU16-2, maxU16-3, 2, 3); got != want {
		t.Errorf("Resize() clamped = %+v, want %+v", got, want)
	}
}

func TestRectClamp(t *testing.T) {
	area := NewRect(10, 10, 100, 100)
	cases := []struct {
		name string
		rect Rect
		want Rect
	}{
		{"inside", NewRect(20, 20, 10, 10), NewRect(20, 20, 10, 10)},
		{"up left", NewRect(5, 5, 10, 10), NewRect(10, 10, 10, 10)},
		{"up", NewRect(20, 5, 10, 10), NewRect(20, 10, 10, 10)},
		{"up right", NewRect(105, 5, 10, 10), NewRect(100, 10, 10, 10)},
		{"left", NewRect(5, 20, 10, 10), NewRect(10, 20, 10, 10)},
		{"right", NewRect(105, 20, 10, 10), NewRect(100, 20, 10, 10)},
		{"down left", NewRect(5, 105, 10, 10), NewRect(10, 100, 10, 10)},
		{"down", NewRect(20, 105, 10, 10), NewRect(20, 100, 10, 10)},
		{"down right", NewRect(105, 105, 10, 10), NewRect(100, 100, 10, 10)},
		{"too wide", NewRect(5, 20, 200, 10), NewRect(10, 20, 100, 10)},
		{"too tall", NewRect(20, 5, 10, 200), NewRect(20, 10, 10, 100)},
		{"too large", NewRect(0, 0, 200, 200), NewRect(10, 10, 100, 100)},
	}
	for _, c := range cases {
		if got := c.rect.Clamp(area); got != c.want {
			t.Errorf("%s: Clamp() = %+v, want %+v", c.name, got, c.want)
		}
	}
}

func TestRectRows(t *testing.T) {
	got := NewRect(0, 0, 3, 2).Rows()
	want := []Rect{NewRect(0, 0, 3, 1), NewRect(0, 1, 3, 1)}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Rows() = %+v, want %+v", got, want)
	}
}

func TestRectColumns(t *testing.T) {
	got := NewRect(0, 0, 3, 2).Columns()
	want := []Rect{NewRect(0, 0, 1, 2), NewRect(1, 0, 1, 2), NewRect(2, 0, 1, 2)}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Columns() = %+v, want %+v", got, want)
	}
}

func TestRectPositions(t *testing.T) {
	got := NewRect(1, 1, 2, 2).Positions()
	want := []Position{{1, 1}, {2, 1}, {1, 2}, {2, 2}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Positions() = %+v, want %+v", got, want)
	}
}

func TestRectAsPositionAsSize(t *testing.T) {
	r := NewRect(1, 2, 3, 4)
	if got, want := r.AsPosition(), (Position{1, 2}); got != want {
		t.Errorf("AsPosition() = %+v, want %+v", got, want)
	}
	if got, want := r.AsSize(), (Size{3, 4}); got != want {
		t.Errorf("AsSize() = %+v, want %+v", got, want)
	}
}

func TestSaturatingHelpers(t *testing.T) {
	if got := satAdd(maxU16, 1); got != maxU16 {
		t.Errorf("satAdd overflow = %d, want %d", got, maxU16)
	}
	if got := satSub(0, 1); got != 0 {
		t.Errorf("satSub underflow = %d, want 0", got)
	}
	if got := satMul(maxU16, 2); got != maxU16 {
		t.Errorf("satMul overflow = %d, want %d", got, maxU16)
	}
	if got := satAdd(2, 3); got != 5 {
		t.Errorf("satAdd(2,3) = %d, want 5", got)
	}
	if got := satSub(5, 3); got != 2 {
		t.Errorf("satSub(5,3) = %d, want 2", got)
	}
}
