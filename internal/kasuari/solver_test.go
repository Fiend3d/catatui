// Tests ported from kasuari's own suite @ v0.4.11, plus determinism tests for
// the ordered-iteration change this port makes.

package kasuari

import (
	"math"
	"testing"
)

func assertClose(t *testing.T, name string, got, want float64) {
	t.Helper()
	if math.Abs(got-want) > 1e-6 {
		t.Errorf("%s = %v, want %v", name, got, want)
	}
}

// TestQuadrilateral is kasuari's tests/quadrilateral.rs, up to the point where
// it starts using edit variables (which this port omits). Four corner points are
// pinned with increasing weights, midpoints are constrained to the averages of
// adjacent corners, and the corners must stay ordered and inside a 500x500 box.
//
// It is the strongest single check that the simplex core is correct: it mixes
// required equalities, weighted equalities, strong inequalities and bounds.
func TestQuadrilateral(t *testing.T) {
	type point struct{ x, y Variable }
	newPoint := func() point { return point{x: NewVariable(), y: NewVariable()} }

	points := [4]point{newPoint(), newPoint(), newPoint(), newPoint()}
	starts := [4][2]float64{{10, 10}, {10, 200}, {200, 200}, {200, 10}}
	midpoints := [4]point{newPoint(), newPoint(), newPoint(), newPoint()}

	s := NewSolver()

	// Pin each corner near its start, each one twice as firmly as the last.
	weight := 1.0
	for i := range points {
		must(t, s.AddConstraints(
			Relate(FromVariable(points[i].x), Equal, FromConstant(starts[i][0]), Weak*Strength(weight)),
			Relate(FromVariable(points[i].y), Equal, FromConstant(starts[i][1]), Weak*Strength(weight)),
		))
		weight *= 2
	}

	// Each midpoint sits halfway along its edge.
	for _, e := range [4][2]int{{0, 1}, {1, 2}, {2, 3}, {3, 0}} {
		start, end := e[0], e[1]
		must(t, s.AddConstraints(
			Relate(FromVariable(midpoints[start].x), Equal,
				FromVariable(points[start].x).AddVariable(points[end].x).Div(2), Required),
			Relate(FromVariable(midpoints[start].y), Equal,
				FromVariable(points[start].y).AddVariable(points[end].y).Div(2), Required),
		))
	}

	// Keep the corners in order, with at least 20 units between them.
	for _, c := range [8][2]int{{0, 2}, {0, 3}, {1, 2}, {1, 3}} {
		must(t, s.AddConstraint(Relate(
			FromVariable(points[c[0]].x).AddConstant(20), LessOrEqual,
			FromVariable(points[c[1]].x), Strong)))
	}
	for _, c := range [4][2]int{{0, 1}, {0, 2}, {3, 1}, {3, 2}} {
		must(t, s.AddConstraint(Relate(
			FromVariable(points[c[0]].y).AddConstant(20), LessOrEqual,
			FromVariable(points[c[1]].y), Strong)))
	}

	// And inside a 500x500 box.
	for _, p := range points {
		must(t, s.AddConstraints(
			Relate(FromVariable(p.x), GreaterOrEqual, FromConstant(0), Required),
			Relate(FromVariable(p.y), GreaterOrEqual, FromConstant(0), Required),
			Relate(FromVariable(p.x), LessOrEqual, FromConstant(500), Required),
			Relate(FromVariable(p.y), LessOrEqual, FromConstant(500), Required),
		))
	}

	want := [4][2]float64{{10, 105}, {105, 200}, {200, 105}, {105, 10}}
	for i, w := range want {
		assertClose(t, "midpoint x", s.GetValue(midpoints[i].x), w[0])
		assertClose(t, "midpoint y", s.GetValue(midpoints[i].y), w[1])
	}
}

// TestTwoBoxes is the worked example from kasuari's crate documentation: two
// elements laid out horizontally in a 300-unit window, the first left-aligned
// and the second right-aligned, with preferred widths of 50 and 100.
func TestTwoBoxes(t *testing.T) {
	windowWidth := NewVariable()
	box1Left, box1Right := NewVariable(), NewVariable()
	box2Left, box2Right := NewVariable(), NewVariable()

	s := NewSolver()
	must(t, s.AddConstraints(
		Relate(FromVariable(windowWidth), GreaterOrEqual, FromConstant(0), Required),
		Relate(FromVariable(box1Left), Equal, FromConstant(0), Required),
		Relate(FromVariable(box2Right), Equal, FromVariable(windowWidth), Required),
		Relate(FromVariable(box2Left), GreaterOrEqual, FromVariable(box1Right), Required),
		Relate(FromVariable(box1Left), LessOrEqual, FromVariable(box1Right), Required),
		Relate(FromVariable(box2Left), LessOrEqual, FromVariable(box2Right), Required),
		Relate(FromVariable(box1Right).SubVariable(box1Left), Equal, FromConstant(50), Weak),
		Relate(FromVariable(box2Right).SubVariable(box2Left), Equal, FromConstant(100), Weak),
		// Pin the window width, standing in for the edit variable the Rust
		// example uses.
		Relate(FromVariable(windowWidth), Equal, FromConstant(300), Strong),
	))

	assertClose(t, "windowWidth", s.GetValue(windowWidth), 300)
	assertClose(t, "box1Left", s.GetValue(box1Left), 0)
	assertClose(t, "box1Right", s.GetValue(box1Right), 50)
	assertClose(t, "box2Left", s.GetValue(box2Left), 200)
	assertClose(t, "box2Right", s.GetValue(box2Right), 300)
}

// TestStrengthOrdering checks that a stronger constraint wins when two conflict.
func TestStrengthOrdering(t *testing.T) {
	v := NewVariable()
	s := NewSolver()
	must(t, s.AddConstraints(
		Relate(FromVariable(v), Equal, FromConstant(10), Weak),
		Relate(FromVariable(v), Equal, FromConstant(20), Strong),
	))
	assertClose(t, "v", s.GetValue(v), 20)
}

// TestRequiredBeatsEverything checks that a required constraint is never given up.
func TestRequiredBeatsEverything(t *testing.T) {
	v := NewVariable()
	s := NewSolver()
	must(t, s.AddConstraints(
		Relate(FromVariable(v), Equal, FromConstant(10), Strong),
		Relate(FromVariable(v), Equal, FromConstant(42), Required),
	))
	assertClose(t, "v", s.GetValue(v), 42)
}

func TestUnsatisfiableConstraint(t *testing.T) {
	v := NewVariable()
	s := NewSolver()
	must(t, s.AddConstraint(Relate(FromVariable(v), Equal, FromConstant(10), Required)))
	err := s.AddConstraint(Relate(FromVariable(v), Equal, FromConstant(20), Required))
	if err == nil {
		t.Fatal("two conflicting required constraints should be unsatisfiable")
	}
}

func TestDuplicateConstraint(t *testing.T) {
	v := NewVariable()
	c := Relate(FromVariable(v), Equal, FromConstant(10), Required)
	s := NewSolver()
	must(t, s.AddConstraint(c))
	if err := s.AddConstraint(c); err != ErrDuplicateConstraint {
		t.Errorf("adding the same constraint twice = %v, want ErrDuplicateConstraint", err)
	}
}

// TestDeterminism is the reason this port sorts its map iterations. The system
// below is deliberately degenerate: three variables that must sum to a fixed
// total, each equally weighted toward the same value, so the solver has a free
// choice of pivot. Go randomizes map iteration per run, so without the ordered
// scans this could resolve differently on each run, and a terminal layout would
// visibly shift between launches of the same program.
func TestDeterminism(t *testing.T) {
	solve := func() [3]float64 {
		a, b, c := NewVariable(), NewVariable(), NewVariable()
		s := NewSolver()
		must(t, s.AddConstraints(
			Relate(FromVariable(a).AddVariable(b).AddVariable(c), Equal, FromConstant(100), Required),
			Relate(FromVariable(a), Equal, FromVariable(b), Medium),
			Relate(FromVariable(b), Equal, FromVariable(c), Medium),
			Relate(FromVariable(a), GreaterOrEqual, FromConstant(0), Required),
			Relate(FromVariable(b), GreaterOrEqual, FromConstant(0), Required),
			Relate(FromVariable(c), GreaterOrEqual, FromConstant(0), Required),
		))
		return [3]float64{s.GetValue(a), s.GetValue(b), s.GetValue(c)}
	}

	first := solve()
	for i := range 200 {
		if got := solve(); got != first {
			t.Fatalf("run %d gave %v, but the first run gave %v; the solver is not deterministic", i, got, first)
		}
	}
	// It should also actually be the right answer.
	total := first[0] + first[1] + first[2]
	assertClose(t, "total", total, 100)
	for i, v := range first {
		assertClose(t, "share", v, 100.0/3.0)
		if v < 0 {
			t.Errorf("share %d is negative: %v", i, v)
		}
	}
}

func TestGetValueOfUnknownVariable(t *testing.T) {
	s := NewSolver()
	if got := s.GetValue(NewVariable()); got != 0 {
		t.Errorf("an unconstrained variable should read as 0, got %v", got)
	}
}

func TestStrengthHelpers(t *testing.T) {
	if got := NewStrength(-5); got != 0 {
		t.Errorf("NewStrength should clamp below at 0, got %v", got)
	}
	if got := NewStrength(1e12); got != Required {
		t.Errorf("NewStrength should clamp above at Required, got %v", got)
	}
	// CreateStrength blends the three components, which is how ratatui builds
	// its layout strength ladder.
	if got, want := CreateStrength(1, 0, 0, 1), Strong; got != want {
		t.Errorf("CreateStrength(1,0,0,1) = %v, want %v", got, want)
	}
	if got, want := CreateStrength(0, 1, 0, 1), Medium; got != want {
		t.Errorf("CreateStrength(0,1,0,1) = %v, want %v", got, want)
	}
}

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected solver error: %v", err)
	}
}
