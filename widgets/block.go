// Package widgets holds catatui's widget library: the ready-made things you
// render into a Frame.
//
// Port of ratatui-widgets @ ratatui-v0.30.2
package widgets

import (
	"github.com/Fiend3d/catatui"
	"github.com/Fiend3d/catatui/symbols"
)

// Borders selects which sides of a Block are drawn. Combine them with |.
type Borders uint8

// The bit values match ratatui's Borders bitflags.
const (
	// BordersNone draws no border.
	BordersNone Borders = 0
	// BordersTop draws the top edge.
	BordersTop Borders = 1 << 0
	// BordersRight draws the right edge.
	BordersRight Borders = 1 << 1
	// BordersBottom draws the bottom edge.
	BordersBottom Borders = 1 << 2
	// BordersLeft draws the left edge.
	BordersLeft Borders = 1 << 3
)

// BordersAll draws every edge.
const BordersAll = BordersTop | BordersRight | BordersBottom | BordersLeft

// Contains reports whether every side in other is drawn.
func (b Borders) Contains(other Borders) bool { return b&other == other }

// Intersects reports whether any side in other is drawn.
func (b Borders) Intersects(other Borders) bool { return b&other != 0 }

// BorderType selects the characters a Block's border is drawn with.
type BorderType uint8

const (
	// BorderPlain is a single-width line, and the default.
	BorderPlain BorderType = iota
	// BorderRounded is a single-width line with curved corners.
	BorderRounded
	// BorderDouble is a double line.
	BorderDouble
	// BorderThick is a heavy line.
	BorderThick
	// BorderQuadrantOutside draws in the outer half of each cell.
	BorderQuadrantOutside
	// BorderQuadrantInside draws in the inner half of each cell.
	BorderQuadrantInside
)

// Set returns the characters for a border type.
func (t BorderType) Set() symbols.BorderSet {
	switch t {
	case BorderRounded:
		return symbols.BorderRounded
	case BorderDouble:
		return symbols.BorderDouble
	case BorderThick:
		return symbols.BorderThick
	case BorderQuadrantOutside:
		return symbols.BorderQuadrantOutside
	case BorderQuadrantInside:
		return symbols.BorderQuadrantInside
	default:
		return symbols.BorderPlain
	}
}

// Padding is space reserved inside a Block's border, between the border and
// whatever is drawn inside it.
type Padding struct {
	Left   uint16
	Right  uint16
	Top    uint16
	Bottom uint16
}

// NewPadding returns a Padding with each side set individually.
func NewPadding(left, right, top, bottom uint16) Padding {
	return Padding{Left: left, Right: right, Top: top, Bottom: bottom}
}

// UniformPadding returns the same padding on every side.
func UniformPadding(n uint16) Padding { return Padding{n, n, n, n} }

// HorizontalPadding returns padding on the left and right only.
func HorizontalPadding(n uint16) Padding { return Padding{Left: n, Right: n} }

// VerticalPadding returns padding on the top and bottom only.
func VerticalPadding(n uint16) Padding { return Padding{Top: n, Bottom: n} }

// TitlePosition is whether a title sits on the top or bottom border.
type TitlePosition uint8

const (
	// TitleTop puts the title on the top border, which is the default.
	TitleTop TitlePosition = iota
	// TitleBottom puts the title on the bottom border.
	TitleBottom
)

type blockTitle struct {
	line     catatui.Line
	position TitlePosition
	// explicitPosition distinguishes a title that chose its position from one
	// that inherits the block's.
	explicitPosition bool
}

// Block is a frame around an area: an optional border on any combination of
// sides, titles on the top and bottom edges, and padding inside.
//
// It is the most-used widget in the library, and mostly it is used for its
// Inner method: draw the block, then draw something else into the area it
// leaves.
//
//	block := widgets.NewBlock().Borders(widgets.BordersAll).Title("Files")
//	inner := block.Inner(area)
//	f.RenderWidget(block, area)
//	f.RenderWidget(list, inner)
type Block struct {
	titles          []blockTitle
	titlesStyle     catatui.Style
	titlesAlignment catatui.Alignment
	titlesPosition  TitlePosition
	borders         Borders
	borderStyle     catatui.Style
	borderSet       symbols.BorderSet
	borderSetIsSet  bool
	style           catatui.Style
	padding         Padding
	shadow          Shadow
	hasShadow       bool
}

// NewBlock returns a block with no borders and no titles.
func NewBlock() Block { return Block{} }

// Bordered returns a block with a border on every side, which is the common
// case and saves writing NewBlock().Borders(BordersAll).
func Bordered() Block { return Block{borders: BordersAll} }

// Title returns a copy of b with a title added, using the block's default
// alignment and position.
func (b Block) Title(title string) Block {
	return b.TitleLine(catatui.LineFromString(title))
}

// TitleLine returns a copy of b with a styled title added. A title that sets
// its own alignment overrides the block's.
func (b Block) TitleLine(line catatui.Line) Block {
	b.titles = append(append([]blockTitle(nil), b.titles...), blockTitle{line: line})
	return b
}

// TitleTop returns a copy of b with a title added to the top border.
func (b Block) TitleTop(line catatui.Line) Block {
	b.titles = append(append([]blockTitle(nil), b.titles...),
		blockTitle{line: line, position: TitleTop, explicitPosition: true})
	return b
}

// TitleBottom returns a copy of b with a title added to the bottom border.
func (b Block) TitleBottom(line catatui.Line) Block {
	b.titles = append(append([]blockTitle(nil), b.titles...),
		blockTitle{line: line, position: TitleBottom, explicitPosition: true})
	return b
}

// TitleStyle returns a copy of b with the style applied beneath every title's
// own style.
func (b Block) TitleStyle(s catatui.Style) Block { b.titlesStyle = s; return b }

// TitleAlignment returns a copy of b with the default alignment for titles that
// do not set their own.
func (b Block) TitleAlignment(a catatui.Alignment) Block { b.titlesAlignment = a; return b }

// TitlePosition returns a copy of b with the default position for titles that
// do not set their own.
func (b Block) TitlePosition(p TitlePosition) Block { b.titlesPosition = p; return b }

// Borders returns a copy of b with the given sides drawn.
func (b Block) Borders(sides Borders) Block { b.borders = sides; return b }

// BorderStyle returns a copy of b with the border drawn in the given style.
func (b Block) BorderStyle(s catatui.Style) Block { b.borderStyle = s; return b }

// BorderType returns a copy of b with the border drawn in the given characters.
func (b Block) BorderType(t BorderType) Block {
	b.borderSet, b.borderSetIsSet = t.Set(), true
	return b
}

// BorderSet returns a copy of b with an explicit set of border characters.
func (b Block) BorderSet(s symbols.BorderSet) Block {
	b.borderSet, b.borderSetIsSet = s, true
	return b
}

// Style returns a copy of b with a style applied to the whole area, beneath
// everything the block draws and anything drawn inside it.
func (b Block) Style(s catatui.Style) Block { b.style = s; return b }

// Padding returns a copy of b with space reserved inside the border.
func (b Block) Padding(p Padding) Block { b.padding = p; return b }

// Shadow returns a copy of b with a shadow drawn behind it.
//
// The shadow is rendered using the block area plus the shadow's configured
// offset.
//
//	block := widgets.Bordered().Title("Popup").Shadow(
//		widgets.ShadowDarkShade().
//			Style(catatui.NewStyle().Fg(catatui.ColorBlack).Bg(catatui.ColorWhite)).
//			Offset(catatui.Offset{X: 2, Y: 1}))
func (b Block) Shadow(s Shadow) Block { b.shadow, b.hasShadow = s, true; return b }

func (b Block) set() symbols.BorderSet {
	if b.borderSetIsSet {
		return b.borderSet
	}
	return symbols.BorderPlain
}

// hasTitleAt reports whether any title lands on the given edge, which matters
// because a title reserves a row even where there is no border to sit on.
func (b Block) hasTitleAt(p TitlePosition) bool {
	for _, t := range b.titles {
		pos := b.titlesPosition
		if t.explicitPosition {
			pos = t.position
		}
		if pos == p {
			return true
		}
	}
	return false
}

// Inner returns the area left for content once the border, the title rows and
// the padding are taken out.
//
// This is the method most code actually wants: render the block into area, then
// render the content into Inner(area).
func (b Block) Inner(area catatui.Rect) catatui.Rect {
	inner := area
	if b.borders.Intersects(BordersLeft) {
		inner.X = catatui.MinU16(catatui.SatAdd(inner.X, 1), inner.Right())
		inner.Width = catatui.SatSub(inner.Width, 1)
	}
	if b.borders.Intersects(BordersTop) || b.hasTitleAt(TitleTop) {
		inner.Y = catatui.MinU16(catatui.SatAdd(inner.Y, 1), inner.Bottom())
		inner.Height = catatui.SatSub(inner.Height, 1)
	}
	if b.borders.Intersects(BordersRight) {
		inner.Width = catatui.SatSub(inner.Width, 1)
	}
	if b.borders.Intersects(BordersBottom) || b.hasTitleAt(TitleBottom) {
		inner.Height = catatui.SatSub(inner.Height, 1)
	}

	inner.X = catatui.SatAdd(inner.X, b.padding.Left)
	inner.Y = catatui.SatAdd(inner.Y, b.padding.Top)
	inner.Width = catatui.SatSub(inner.Width, catatui.SatAdd(b.padding.Left, b.padding.Right))
	inner.Height = catatui.SatSub(inner.Height, catatui.SatAdd(b.padding.Top, b.padding.Bottom))
	return inner
}

// Render draws the block's background, border and titles.
func (b Block) Render(area catatui.Rect, buf *catatui.Buffer) {
	area = area.Intersection(buf.Area)
	if area.IsEmpty() {
		return
	}
	buf.SetStyle(area, b.style)
	b.renderBorders(area, buf)
	b.renderTitles(area, buf)
	b.renderShadow(area, buf)
}

func (b Block) renderBorders(area catatui.Rect, buf *catatui.Buffer) {
	set := b.set()
	left, top := area.Left(), area.Top()
	// Right() and Bottom() are one past the edge.
	right, bottom := catatui.SatSub(area.Right(), 1), catatui.SatSub(area.Bottom(), 1)

	if b.borders.Contains(BordersLeft) {
		for y := top; y <= bottom; y++ {
			buf.Get(left, y).SetSymbol(set.VerticalLeft).SetStyle(b.borderStyle)
		}
	}
	if b.borders.Contains(BordersRight) {
		for y := top; y <= bottom; y++ {
			buf.Get(right, y).SetSymbol(set.VerticalRight).SetStyle(b.borderStyle)
		}
	}
	if b.borders.Contains(BordersTop) {
		for x := left; x <= right; x++ {
			buf.Get(x, top).SetSymbol(set.HorizontalTop).SetStyle(b.borderStyle)
		}
	}
	if b.borders.Contains(BordersBottom) {
		for x := left; x <= right; x++ {
			buf.Get(x, bottom).SetSymbol(set.HorizontalBottom).SetStyle(b.borderStyle)
		}
	}

	// Corners go on last, so they overwrite the sides that ran through them.
	corners := []struct {
		need Borders
		x, y uint16
		sym  string
	}{
		{BordersRight | BordersBottom, right, bottom, set.BottomRight},
		{BordersRight | BordersTop, right, top, set.TopRight},
		{BordersLeft | BordersBottom, left, bottom, set.BottomLeft},
		{BordersLeft | BordersTop, left, top, set.TopLeft},
	}
	for _, c := range corners {
		if b.borders.Contains(c.need) {
			buf.Get(c.x, c.y).SetSymbol(c.sym).SetStyle(b.borderStyle)
		}
	}
}

// titlesArea is the strip of one border row that titles may use, excluding the
// corners.
func (b Block) titlesArea(area catatui.Rect, p TitlePosition) catatui.Rect {
	var leftBorder, rightBorder uint16
	if b.borders.Contains(BordersLeft) {
		leftBorder = 1
	}
	if b.borders.Contains(BordersRight) {
		rightBorder = 1
	}
	y := area.Top()
	if p == TitleBottom {
		y = catatui.SatSub(area.Bottom(), 1)
	}
	return catatui.Rect{
		X:      catatui.SatAdd(area.Left(), leftBorder),
		Y:      y,
		Width:  catatui.SatSub(catatui.SatSub(area.Width, leftBorder), rightBorder),
		Height: 1,
	}
}

func (b Block) renderTitles(area catatui.Rect, buf *catatui.Buffer) {
	for _, pos := range []TitlePosition{TitleTop, TitleBottom} {
		// The order matters: later groups draw over earlier ones where they
		// collide, and ratatui resolves overlap left, then center, then right.
		b.renderTitleGroup(pos, catatui.AlignmentLeft, area, buf)
		b.renderTitleGroup(pos, catatui.AlignmentCenter, area, buf)
		b.renderTitleGroup(pos, catatui.AlignmentRight, area, buf)
	}
}

// filteredTitles returns the titles at a given edge with a given alignment,
// resolving each title's inherited position and alignment.
func (b Block) filteredTitles(p TitlePosition, a catatui.Alignment) []catatui.Line {
	var out []catatui.Line
	for _, t := range b.titles {
		pos := b.titlesPosition
		if t.explicitPosition {
			pos = t.position
		}
		if pos != p {
			continue
		}
		align := t.line.GetAlignment()
		if align == catatui.AlignmentNone {
			align = b.titlesAlignment
		}
		if align == catatui.AlignmentNone {
			align = catatui.AlignmentLeft
		}
		if align == a {
			out = append(out, t.line)
		}
	}
	return out
}

func (b Block) renderTitleGroup(p TitlePosition, a catatui.Alignment, area catatui.Rect, buf *catatui.Buffer) {
	titles := b.filteredTitles(p, a)
	if len(titles) == 0 {
		return
	}
	titlesArea := b.titlesArea(area, p)

	switch a {
	case catatui.AlignmentRight:
		// Drawn right to left so the last title ends up at the right edge.
		for i := len(titles) - 1; i >= 0; i-- {
			if titlesArea.IsEmpty() {
				return
			}
			w := lineWidth(titles[i])
			titleArea := titlesArea
			titleArea.X = catatui.MaxU16(catatui.SatSub(titlesArea.Right(), w), titlesArea.Left())
			titleArea.Width = catatui.MinU16(w, titlesArea.Width)
			b.drawTitle(titles[i], titleArea, buf)
			titlesArea.Width = catatui.SatSub(catatui.SatSub(titlesArea.Width, w), 1)
		}

	case catatui.AlignmentCenter:
		// Titles are separated by one space, and the group is centered as a whole.
		var total uint16
		for _, t := range titles {
			total = catatui.SatAdd(total, catatui.SatAdd(lineWidth(t), 1))
		}
		total = catatui.SatSub(total, 1)

		if total > titlesArea.Width {
			b.renderCenteredTitlesTruncated(titles, total, titlesArea, buf)
			return
		}
		cur := titlesArea
		cur.X = catatui.SatAdd(titlesArea.Left(), (titlesArea.Width-total)/2)
		cur.Width = catatui.SatSub(titlesArea.Right(), cur.X)
		for _, t := range titles {
			if cur.IsEmpty() {
				return
			}
			w := lineWidth(t)
			titleArea := cur
			titleArea.Width = catatui.MinU16(w, cur.Width)
			b.drawTitle(t, titleArea, buf)
			advance := catatui.SatAdd(w, 1)
			cur.X = catatui.SatAdd(cur.X, advance)
			cur.Width = catatui.SatSub(cur.Width, advance)
		}

	default: // left
		cur := titlesArea
		for _, t := range titles {
			if cur.IsEmpty() {
				return
			}
			w := lineWidth(t)
			titleArea := cur
			titleArea.Width = catatui.MinU16(w, cur.Width)
			b.drawTitle(t, titleArea, buf)
			advance := catatui.SatAdd(w, 1)
			cur.X = catatui.SatAdd(cur.X, advance)
			cur.Width = catatui.SatSub(cur.Width, advance)
		}
	}
}

// renderCenteredTitlesTruncated lays out centered titles that are collectively
// too wide for the area.
//
// The group overflows equally on both sides, so half the excess is cut from the
// left. That is done by right-aligning the titles that fall in the cut, which
// makes Line's own truncation drop their leading columns; once the offset is
// used up, the remaining titles are left-aligned and truncated on the right by
// the area edge.
func (b Block) renderCenteredTitlesTruncated(titles []catatui.Line, total uint16, area catatui.Rect, buf *catatui.Buffer) {
	offset := catatui.SatSub(total, area.Width) / 2
	for _, t := range titles {
		if area.IsEmpty() {
			return
		}
		width := catatui.SatSub(catatui.MinU16(area.Width, lineWidth(t)), offset)
		titleArea := area
		titleArea.Width = width
		buf.SetStyle(titleArea, b.titlesStyle)
		if offset > 0 {
			t.Right().Render(titleArea, buf)
			offset = catatui.SatSub(catatui.SatSub(offset, width), 1)
		} else {
			t.Left().Render(titleArea, buf)
		}
		advance := catatui.SatAdd(width, 1)
		area.X = catatui.SatAdd(area.X, advance)
		area.Width = catatui.SatSub(area.Width, advance)
	}
}

func (b Block) renderShadow(baseArea catatui.Rect, buf *catatui.Buffer) {
	if b.hasShadow {
		b.shadow.Render(baseArea, buf)
	}
}

func (b Block) drawTitle(line catatui.Line, area catatui.Rect, buf *catatui.Buffer) {
	buf.SetStyle(area, b.titlesStyle)
	line.Render(area, buf)
}

func lineWidth(l catatui.Line) uint16 {
	return uint16(min(l.Width(), 0xFFFF))
}

var _ catatui.Widget = Block{}
