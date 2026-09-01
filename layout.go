// Port of ratatui-core/src/layout/layout.rs, constraint.rs, direction.rs and
// flex.rs @ ratatui-v0.30.2
//
// Layout divides an area into smaller ones according to a list of Constraints.
// It is solved with Cassowary rather than by simple arithmetic, because the
// constraints can conflict: a Length(20) beside a Min(30) in a 40-column
// terminal cannot both be honoured, and the solver decides which gives way based
// on a fixed ladder of strengths.

package catatui

import (
	"fmt"
	"math"
	"strings"

	"github.com/Fiend3d/catatui/internal/kasuari"
)

// floatPrecisionMultiplier scales cell counts before they reach the solver, so
// that a solution landing between two cells still rounds the way ratatui's does.
const floatPrecisionMultiplier = 100.0

// Direction is the axis a Layout splits along.
type Direction uint8

const (
	// Vertical splits into rows, top to bottom. It is the zero value, because
	// ratatui marks Vertical as Direction's default even though it lists
	// Horizontal first. Matching the default matters here; matching the
	// declaration order would not, since nothing depends on the ordinal.
	Vertical Direction = iota
	// Horizontal splits into columns, left to right.
	Horizontal
)

// String returns the direction's name.
func (d Direction) String() string {
	if d == Horizontal {
		return "Horizontal"
	}
	return "Vertical"
}

// --- Constraint -----------------------------------------------------------

type constraintKind uint8

const (
	constraintMin constraintKind = iota
	constraintMax
	constraintLength
	constraintPercentage
	constraintRatio
	constraintFill
)

// Constraint describes how much of an area one segment of a Layout should take.
//
// Constraints are requests, not guarantees. When they cannot all be met, the
// solver honours them in this order of priority: Min and Max first, then Length,
// then Percentage, then Ratio, then Fill.
//
// The zero value is Min(0), matching the first variant of ratatui's enum, but
// build them with the constructors rather than relying on that.
type Constraint struct {
	kind constraintKind
	a    uint32
	b    uint32
}

// Min requests at least n cells. It has the highest priority alongside Max, and
// grows to fill leftover space.
func Min(n uint16) Constraint { return Constraint{kind: constraintMin, a: uint32(n)} }

// Max requests at most n cells.
func Max(n uint16) Constraint { return Constraint{kind: constraintMax, a: uint32(n)} }

// Length requests exactly n cells.
func Length(n uint16) Constraint { return Constraint{kind: constraintLength, a: uint32(n)} }

// Percentage requests p percent of the available space, rounded to the nearest
// cell.
func Percentage(p uint16) Constraint { return Constraint{kind: constraintPercentage, a: uint32(p)} }

// Ratio requests num/den of the available space. A zero denominator is treated
// as one.
func Ratio(num, den uint32) Constraint {
	return Constraint{kind: constraintRatio, a: num, b: den}
}

// Fill requests whatever space is left, shared with the other Fill constraints
// in proportion to their scaling factors. It has the lowest priority.
func Fill(scale uint16) Constraint { return Constraint{kind: constraintFill, a: uint32(scale)} }

// String formats the constraint the way ratatui's Display impl does.
func (c Constraint) String() string {
	switch c.kind {
	case constraintMin:
		return fmt.Sprintf("Min(%d)", c.a)
	case constraintMax:
		return fmt.Sprintf("Max(%d)", c.a)
	case constraintLength:
		return fmt.Sprintf("Length(%d)", c.a)
	case constraintPercentage:
		return fmt.Sprintf("Percentage(%d)", c.a)
	case constraintRatio:
		return fmt.Sprintf("Ratio(%d, %d)", c.a, c.b)
	default:
		return fmt.Sprintf("Fill(%d)", c.a)
	}
}

func (c Constraint) isFill() bool { return c.kind == constraintFill }
func (c Constraint) isMin() bool  { return c.kind == constraintMin }

// --- Flex -----------------------------------------------------------------

// Flex decides what to do with space left over once every Constraint is
// satisfied.
type Flex uint8

const (
	// FlexStart packs segments at the start and leaves the excess at the end.
	// This is the zero value and ratatui's default.
	FlexStart Flex = iota
	// FlexLegacy puts the excess into the last constraint, which is how ratatui
	// and tui-rs behaved before Flex existed.
	FlexLegacy
	// FlexEnd packs segments at the end and leaves the excess at the start.
	FlexEnd
	// FlexCenter centers the segments, splitting the excess between both ends.
	FlexCenter
	// FlexSpaceBetween spreads the excess evenly between segments, with none at
	// the ends.
	FlexSpaceBetween
	// FlexSpaceEvenly spreads the excess evenly between segments and at both
	// ends.
	FlexSpaceEvenly
	// FlexSpaceAround spreads the excess between segments, giving the ends half
	// as much as the gaps between segments.
	FlexSpaceAround
)

// String returns the flex mode's name.
func (f Flex) String() string {
	switch f {
	case FlexLegacy:
		return "Legacy"
	case FlexEnd:
		return "End"
	case FlexCenter:
		return "Center"
	case FlexSpaceBetween:
		return "SpaceBetween"
	case FlexSpaceEvenly:
		return "SpaceEvenly"
	case FlexSpaceAround:
		return "SpaceAround"
	default:
		return "Start"
	}
}

func (f Flex) isLegacy() bool { return f == FlexLegacy }

// --- Spacing --------------------------------------------------------------

// Spacing is the gap between segments. A positive spacing separates them; an
// overlap makes them share cells, which is how collapsed borders are drawn.
type Spacing struct {
	value int16
}

// Space returns a spacing of n cells between segments.
func Space(n uint16) Spacing { return Spacing{value: int16(min(int(n), math.MaxInt16))} }

// Overlap returns a negative spacing, so that adjacent segments share n cells.
func Overlap(n uint16) Spacing { return Spacing{value: -int16(min(int(n), math.MaxInt16))} }

// String returns the spacing's description.
func (s Spacing) String() string {
	if s.value < 0 {
		return fmt.Sprintf("Overlap(%d)", -s.value)
	}
	return fmt.Sprintf("Space(%d)", s.value)
}

// --- Layout ---------------------------------------------------------------

// Layout splits an area into segments according to a list of Constraints.
//
// Build one with NewLayout, Horizontal or Vertical layouts via the direction
// helpers, then call Split:
//
//	rows := catatui.VerticalLayout(catatui.Length(3), catatui.Fill(1)).Split(area)
//
// A Layout is a value; every builder returns a modified copy, so a Layout can be
// stored and reused safely.
type Layout struct {
	direction   Direction
	constraints []Constraint
	margin      Margin
	flex        Flex
	spacing     Spacing
}

// NewLayout returns a vertical layout with no constraints, matching ratatui's
// Layout::default.
func NewLayout() Layout { return Layout{} }

// VerticalLayout returns a top-to-bottom layout with the given constraints.
func VerticalLayout(constraints ...Constraint) Layout {
	return Layout{direction: Vertical, constraints: constraints}
}

// HorizontalLayout returns a left-to-right layout with the given constraints.
func HorizontalLayout(constraints ...Constraint) Layout {
	return Layout{direction: Horizontal, constraints: constraints}
}

// Direction returns a copy of l splitting along the given axis.
func (l Layout) Direction(d Direction) Layout { l.direction = d; return l }

// Horizontal returns a copy of l splitting into columns.
func (l Layout) Horizontal() Layout { l.direction = Horizontal; return l }

// Vertical returns a copy of l splitting into rows.
func (l Layout) Vertical() Layout { l.direction = Vertical; return l }

// Constraints returns a copy of l with the given constraints.
func (l Layout) Constraints(constraints ...Constraint) Layout {
	l.constraints = constraints
	return l
}

// Margin returns a copy of l inset by n cells on every side.
func (l Layout) Margin(n uint16) Layout {
	l.margin = Margin{Horizontal: n, Vertical: n}
	return l
}

// HorizontalMargin returns a copy of l inset by n cells on the left and right.
func (l Layout) HorizontalMargin(n uint16) Layout { l.margin.Horizontal = n; return l }

// VerticalMargin returns a copy of l inset by n cells on the top and bottom.
func (l Layout) VerticalMargin(n uint16) Layout { l.margin.Vertical = n; return l }

// Flex returns a copy of l with the given handling of leftover space.
func (l Layout) Flex(f Flex) Layout { l.flex = f; return l }

// Spacing returns a copy of l with the given gap between segments.
func (l Layout) Spacing(s Spacing) Layout { l.spacing = s; return l }

// GetDirection returns the layout's axis.
func (l Layout) GetDirection() Direction { return l.direction }

// GetConstraints returns the layout's constraints.
func (l Layout) GetConstraints() []Constraint { return l.constraints }

// GetMargin returns the layout's margin.
func (l Layout) GetMargin() Margin { return l.margin }

// GetFlex returns how the layout handles leftover space.
func (l Layout) GetFlex() Flex { return l.flex }

// GetSpacing returns the gap between segments.
func (l Layout) GetSpacing() Spacing { return l.spacing }

// String describes the layout, for debugging.
func (l Layout) String() string {
	cs := make([]string, len(l.constraints))
	for i, c := range l.constraints {
		cs[i] = c.String()
	}
	return fmt.Sprintf("Layout{%s, [%s], margin=%+v, flex=%s, spacing=%s}",
		l.direction, strings.Join(cs, ", "), l.margin, l.flex, l.spacing)
}

// Split divides area into one Rect per constraint.
//
// The returned slice always has exactly len(constraints) entries. Segments may
// be empty when the area is too small to satisfy everything.
func (l Layout) Split(area Rect) []Rect {
	segments, _ := l.SplitWithSpacers(area)
	return segments
}

// SplitWithSpacers divides area into one Rect per constraint, and also returns
// the gaps between them.
//
// There is always one more spacer than there are segments: one before the first
// segment, one between each adjacent pair, and one after the last. Widgets that
// draw in the gaps, such as collapsed borders, use these.
func (l Layout) SplitWithSpacers(area Rect) (segments, spacers []Rect) {
	solver := kasuari.NewSolver()
	inner := area.Inner(l.margin)

	var areaStart, areaEnd float64
	if l.direction == Horizontal {
		areaStart = float64(inner.X) * floatPrecisionMultiplier
		areaEnd = float64(inner.Right()) * floatPrecisionMultiplier
	} else {
		areaStart = float64(inner.Y) * floatPrecisionMultiplier
		areaEnd = float64(inner.Bottom()) * floatPrecisionMultiplier
	}

	// Variables alternate between spacer and segment boundaries:
	//
	//   v0    v1                  v2   v3                  v4   v5
	//   ┌   ┐┌──────────────────┐┌   ┐┌──────────────────┐┌   ┐
	//   spacer      segment      spacer     segment      spacer
	//
	// so spacers are the pairs (v0,v1), (v2,v3), ... and segments are the pairs
	// starting one along: (v1,v2), (v3,v4), ...
	variableCount := len(l.constraints)*2 + 2
	variables := make([]kasuari.Variable, variableCount)
	for i := range variables {
		variables[i] = kasuari.NewVariable()
	}

	spacerElems := make([]element, 0, len(l.constraints)+1)
	for i := 0; i+1 < variableCount; i += 2 {
		spacerElems = append(spacerElems, element{start: variables[i], end: variables[i+1]})
	}
	segmentElems := make([]element, 0, len(l.constraints))
	for i := 1; i+1 < variableCount; i += 2 {
		segmentElems = append(segmentElems, element{start: variables[i], end: variables[i+1]})
	}

	spacing := l.spacing.value
	areaElem := element{start: variables[0], end: variables[variableCount-1]}

	// A failure here means the constraint system is contradictory, which for the
	// constraints Layout builds should be impossible. Panicking matches
	// ratatui's split, which expects the solve to succeed.
	mustAdd := func(err error) {
		if err != nil {
			panic(fmt.Sprintf("catatui: failed to split layout %s in area %+v: %v", l, area, err))
		}
	}

	mustAdd(solver.AddConstraints(
		kasuari.Relate(kasuari.FromVariable(areaElem.start), kasuari.Equal, kasuari.FromConstant(areaStart), kasuari.Required),
		kasuari.Relate(kasuari.FromVariable(areaElem.end), kasuari.Equal, kasuari.FromConstant(areaEnd), kasuari.Required),
	))

	// Every boundary lies inside the area, and boundaries are in order.
	for _, v := range variables {
		mustAdd(solver.AddConstraints(
			kasuari.Relate(kasuari.FromVariable(v), kasuari.GreaterOrEqual, kasuari.FromVariable(areaElem.start), kasuari.Required),
			kasuari.Relate(kasuari.FromVariable(v), kasuari.LessOrEqual, kasuari.FromVariable(areaElem.end), kasuari.Required),
		))
	}
	for i := 1; i+1 < len(variables); i += 2 {
		mustAdd(solver.AddConstraint(kasuari.Relate(
			kasuari.FromVariable(variables[i]), kasuari.LessOrEqual,
			kasuari.FromVariable(variables[i+1]), kasuari.Required)))
	}

	l.configureFlexConstraints(solver, mustAdd, areaElem, spacerElems, spacing)
	l.configureConstraints(solver, mustAdd, areaElem, segmentElems)
	l.configureFillConstraints(solver, mustAdd, segmentElems)

	// Outside Legacy mode, segments prefer to be equal to each other. This is
	// the weakest constraint in the system and only breaks ties.
	if !l.flex.isLegacy() {
		for i := 0; i+1 < len(segmentElems); i++ {
			mustAdd(solver.AddConstraint(kasuari.Relate(
				segmentElems[i].size(), kasuari.Equal,
				segmentElems[i+1].size(), allSegmentGrow)))
		}
	}

	segments = elementsToRects(solver, segmentElems, inner, l.direction)
	spacers = elementsToRects(solver, spacerElems, inner, l.direction)
	return segments, spacers
}

// element is one span between two boundary variables: a segment or a spacer.
type element struct {
	start kasuari.Variable
	end   kasuari.Variable
}

func (e element) size() kasuari.Expression {
	return kasuari.FromVariable(e.end).SubVariable(e.start)
}

func (e element) hasMaxSize(size uint16, strength kasuari.Strength) *kasuari.Constraint {
	return kasuari.Relate(e.size(), kasuari.LessOrEqual,
		kasuari.FromConstant(float64(size)*floatPrecisionMultiplier), strength)
}

func (e element) hasMinSize(size int16, strength kasuari.Strength) *kasuari.Constraint {
	return kasuari.Relate(e.size(), kasuari.GreaterOrEqual,
		kasuari.FromConstant(float64(size)*floatPrecisionMultiplier), strength)
}

func (e element) hasIntSize(size uint16, strength kasuari.Strength) *kasuari.Constraint {
	return kasuari.Relate(e.size(), kasuari.Equal,
		kasuari.FromConstant(float64(size)*floatPrecisionMultiplier), strength)
}

func (e element) hasSize(size kasuari.Expression, strength kasuari.Strength) *kasuari.Constraint {
	return kasuari.Relate(e.size(), kasuari.Equal, size, strength)
}

func (e element) hasDoubleSize(size kasuari.Expression, strength kasuari.Strength) *kasuari.Constraint {
	return kasuari.Relate(e.size(), kasuari.Equal, size.Mul(2), strength)
}

func (e element) isEmpty() *kasuari.Constraint {
	return kasuari.Relate(e.size(), kasuari.Equal, kasuari.FromConstant(0),
		kasuari.Required-kasuari.Weak)
}

// The strength ladder. These fixed relative strengths are what make constraints
// resolve in a predictable priority order when they conflict, and their exact
// values are load bearing: changing one changes which constraint gives way.
const (
	spacerSizeEq     = kasuari.Required / 10
	minSizeGE        = kasuari.Strong * 100
	maxSizeLE        = kasuari.Strong * 100
	lengthSizeEq     = kasuari.Strong * 10
	percentageSizeEq = kasuari.Strong
	ratioSizeEq      = kasuari.Strong / 10
	minSizeEq        = kasuari.Medium * 10
	maxSizeEq        = kasuari.Medium * 10
	fillGrow         = kasuari.Medium
	grow             = kasuari.Medium / 10
	spaceGrow        = kasuari.Weak * 10
	allSegmentGrow   = kasuari.Weak
)

func (l Layout) configureConstraints(solver *kasuari.Solver, mustAdd func(error), area element, segments []element) {
	for i, c := range l.constraints {
		if i >= len(segments) {
			break
		}
		seg := segments[i]
		switch c.kind {
		case constraintMax:
			mustAdd(solver.AddConstraint(seg.hasMaxSize(uint16(c.a), maxSizeLE)))
			mustAdd(solver.AddConstraint(seg.hasIntSize(uint16(c.a), maxSizeEq)))
		case constraintMin:
			mustAdd(solver.AddConstraint(seg.hasMinSize(int16(c.a), minSizeGE)))
			if l.flex.isLegacy() {
				mustAdd(solver.AddConstraint(seg.hasIntSize(uint16(c.a), minSizeEq)))
			} else {
				mustAdd(solver.AddConstraint(seg.hasSize(area.size(), fillGrow)))
			}
		case constraintLength:
			mustAdd(solver.AddConstraint(seg.hasIntSize(uint16(c.a), lengthSizeEq)))
		case constraintPercentage:
			size := area.size().Mul(float64(c.a)).Div(100)
			mustAdd(solver.AddConstraint(seg.hasSize(size, percentageSizeEq)))
		case constraintRatio:
			// A zero denominator would divide by zero; ratatui treats it as one.
			den := max(c.b, 1)
			size := area.size().Mul(float64(c.a)).Div(float64(den))
			mustAdd(solver.AddConstraint(seg.hasSize(size, ratioSizeEq)))
		case constraintFill:
			// With nothing else to hold it back, this grows as far as it can.
			mustAdd(solver.AddConstraint(seg.hasSize(area.size(), fillGrow)))
		}
	}
}

func (l Layout) configureFlexConstraints(solver *kasuari.Solver, mustAdd func(error), area element, spacers []element, spacing int16) {
	var middle []element
	if len(spacers) > 2 {
		middle = spacers[1 : len(spacers)-1]
	}
	spacingF := float64(spacing) * floatPrecisionMultiplier

	first, last, hasEnds := element{}, element{}, false
	if len(spacers) > 0 {
		first, last, hasEnds = spacers[0], spacers[len(spacers)-1], true
	}

	switch l.flex {
	case FlexLegacy:
		for _, s := range middle {
			mustAdd(solver.AddConstraint(s.hasSize(kasuari.FromConstant(spacingF), spacerSizeEq)))
		}
		if hasEnds {
			mustAdd(solver.AddConstraint(first.isEmpty()))
			mustAdd(solver.AddConstraint(last.isEmpty()))
		}

	case FlexSpaceAround:
		if len(spacers) <= 2 {
			// With so few spacers there is no distinction from SpaceEvenly.
			for i := range spacers {
				for j := i + 1; j < len(spacers); j++ {
					mustAdd(solver.AddConstraint(spacers[i].hasSize(spacers[j].size(), spacerSizeEq)))
				}
			}
			for _, s := range spacers {
				mustAdd(solver.AddConstraint(s.hasMinSize(spacing, spacerSizeEq)))
				mustAdd(solver.AddConstraint(s.hasSize(area.size(), spaceGrow)))
			}
			break
		}
		// The gaps between segments are all equal, and the two ends are half
		// that size.
		for i := range middle {
			for j := i + 1; j < len(middle); j++ {
				mustAdd(solver.AddConstraint(middle[i].hasSize(middle[j].size(), spacerSizeEq)))
			}
		}
		if len(middle) > 0 {
			mustAdd(solver.AddConstraint(middle[0].hasDoubleSize(first.size(), spacerSizeEq)))
			mustAdd(solver.AddConstraint(middle[0].hasDoubleSize(last.size(), spacerSizeEq)))
		}
		for _, s := range spacers {
			mustAdd(solver.AddConstraint(s.hasMinSize(spacing, spacerSizeEq)))
			mustAdd(solver.AddConstraint(s.hasSize(area.size(), spaceGrow)))
		}

	case FlexSpaceEvenly:
		for i := range spacers {
			for j := i + 1; j < len(spacers); j++ {
				mustAdd(solver.AddConstraint(spacers[i].hasSize(spacers[j].size(), spacerSizeEq)))
			}
		}
		for _, s := range spacers {
			mustAdd(solver.AddConstraint(s.hasMinSize(spacing, spacerSizeEq)))
			mustAdd(solver.AddConstraint(s.hasSize(area.size(), spaceGrow)))
		}

	case FlexSpaceBetween:
		for i := range middle {
			for j := i + 1; j < len(middle); j++ {
				mustAdd(solver.AddConstraint(middle[i].hasSize(middle[j].size(), spacerSizeEq)))
			}
		}
		for _, s := range middle {
			mustAdd(solver.AddConstraint(s.hasMinSize(spacing, spacerSizeEq)))
			mustAdd(solver.AddConstraint(s.hasSize(area.size(), spaceGrow)))
		}
		if hasEnds {
			mustAdd(solver.AddConstraint(first.isEmpty()))
			mustAdd(solver.AddConstraint(last.isEmpty()))
		}

	case FlexStart:
		for _, s := range middle {
			mustAdd(solver.AddConstraint(s.hasSize(kasuari.FromConstant(spacingF), spacerSizeEq)))
		}
		if hasEnds {
			mustAdd(solver.AddConstraint(first.isEmpty()))
			mustAdd(solver.AddConstraint(last.hasSize(area.size(), grow)))
		}

	case FlexCenter:
		for _, s := range middle {
			mustAdd(solver.AddConstraint(s.hasSize(kasuari.FromConstant(spacingF), spacerSizeEq)))
		}
		if hasEnds {
			mustAdd(solver.AddConstraint(first.hasSize(area.size(), grow)))
			mustAdd(solver.AddConstraint(last.hasSize(area.size(), grow)))
			mustAdd(solver.AddConstraint(first.hasSize(last.size(), spacerSizeEq)))
		}

	case FlexEnd:
		for _, s := range middle {
			mustAdd(solver.AddConstraint(s.hasSize(kasuari.FromConstant(spacingF), spacerSizeEq)))
		}
		if hasEnds {
			mustAdd(solver.AddConstraint(last.isEmpty()))
			mustAdd(solver.AddConstraint(first.hasSize(area.size(), grow)))
		}
	}
}

// configureFillConstraints ties the growing segments to each other in proportion
// to their scaling factors, so that Fill(1) and Fill(2) split leftover space one
// to two. Outside Legacy mode, Min participates too, with a factor of one.
func (l Layout) configureFillConstraints(solver *kasuari.Solver, mustAdd func(error), segments []element) {
	type growing struct {
		seg   element
		scale float64
	}
	var growers []growing
	for i, c := range l.constraints {
		if i >= len(segments) {
			break
		}
		switch {
		case c.isFill():
			growers = append(growers, growing{segments[i], max(float64(c.a), 1e-6)})
		case !l.flex.isLegacy() && c.isMin():
			growers = append(growers, growing{segments[i], 1.0})
		}
	}
	for i := range growers {
		for j := i + 1; j < len(growers); j++ {
			left, right := growers[i], growers[j]
			mustAdd(solver.AddConstraint(kasuari.Relate(
				left.seg.size().Mul(right.scale), kasuari.Equal,
				right.seg.size().Mul(left.scale), grow)))
		}
	}
}

// elementsToRects reads the solved boundaries back out and turns them into
// Rects. The double rounding matches ratatui: once to undo the precision
// multiplier's floating point error, once to land on a whole cell.
func elementsToRects(solver *kasuari.Solver, elements []element, area Rect, direction Direction) []Rect {
	rects := make([]Rect, len(elements))
	for i, e := range elements {
		start := uint16(math.Round(math.Round(solver.GetValue(e.start)) / floatPrecisionMultiplier))
		end := uint16(math.Round(math.Round(solver.GetValue(e.end)) / floatPrecisionMultiplier))
		size := satSub(end, start)
		if direction == Horizontal {
			rects[i] = Rect{X: start, Y: area.Y, Width: size, Height: area.Height}
		} else {
			rects[i] = Rect{X: area.X, Y: start, Width: area.Width, Height: size}
		}
	}
	return rects
}
