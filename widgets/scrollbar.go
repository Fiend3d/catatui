// Port of ratatui-widgets/src/scrollbar.rs @ ratatui-v0.30.2

package widgets

import (
	"github.com/Fiend3d/catatui"
	"github.com/Fiend3d/catatui/symbols"
)

// ScrollbarOrientation is where a Scrollbar sits around an area and which way
// it scrolls.
//
//	          HorizontalTop
//	            ┌───────┐
//	VerticalLeft│       │VerticalRight
//	            └───────┘
//	         HorizontalBottom
type ScrollbarOrientation uint8

const (
	// ScrollbarVerticalRight puts the scrollbar on the right, scrolling
	// vertically. It is the default.
	ScrollbarVerticalRight ScrollbarOrientation = iota
	// ScrollbarVerticalLeft puts the scrollbar on the left, scrolling vertically.
	ScrollbarVerticalLeft
	// ScrollbarHorizontalBottom puts the scrollbar on the bottom, scrolling
	// horizontally.
	ScrollbarHorizontalBottom
	// ScrollbarHorizontalTop puts the scrollbar on the top, scrolling
	// horizontally.
	ScrollbarHorizontalTop
)

// IsVertical reports whether the scrollbar scrolls vertically.
func (o ScrollbarOrientation) IsVertical() bool {
	return o == ScrollbarVerticalRight || o == ScrollbarVerticalLeft
}

// IsHorizontal reports whether the scrollbar scrolls horizontally.
func (o ScrollbarOrientation) IsHorizontal() bool {
	return o == ScrollbarHorizontalBottom || o == ScrollbarHorizontalTop
}

// String names the orientation the way ratatui's Display does.
func (o ScrollbarOrientation) String() string {
	switch o {
	case ScrollbarVerticalRight:
		return "VerticalRight"
	case ScrollbarVerticalLeft:
		return "VerticalLeft"
	case ScrollbarHorizontalBottom:
		return "HorizontalBottom"
	case ScrollbarHorizontalTop:
		return "HorizontalTop"
	}
	return "Unknown"
}

// ScrollDirection is which way ScrollbarState.Scroll moves. It is useful when
// an application wants to store which way to scroll and apply it later.
type ScrollDirection uint8

const (
	// ScrollForward scrolls down or right, and is the default.
	ScrollForward ScrollDirection = iota
	// ScrollBackward scrolls up or left.
	ScrollBackward
)

// String names the direction the way ratatui's Display does.
func (d ScrollDirection) String() string {
	switch d {
	case ScrollForward:
		return "Forward"
	case ScrollBackward:
		return "Backward"
	}
	return "Unknown"
}

// ScrollbarState is the caller-owned state of a Scrollbar: how long the content
// is, where in it the view currently is, and how much of it the view shows.
//
// The content length must be set, or the scrollbar draws nothing. For a list of
// four items where the view shows two of them:
//
//	state := widgets.NewScrollbarState(4).Position(0).ViewportContentLength(2)
//
//	┌───────────────┐
//	│1. this is a   █
//	│   single item █
//	│2. this is a   ║
//	│   second item ║
//	└───────────────┘
//
// When every item is one row high the viewport length can be left at zero, and
// the track length is used instead.
type ScrollbarState struct {
	contentLength         int
	position              int
	viewportContentLength int
}

// NewScrollbarState returns a state for content of the given length, positioned
// at the start.
func NewScrollbarState(contentLength int) ScrollbarState {
	return ScrollbarState{contentLength: contentLength}
}

// Position returns a copy of s scrolled to the given item.
func (s ScrollbarState) Position(position int) ScrollbarState { s.position = position; return s }

// ContentLength returns a copy of s with the number of scrollable items set.
func (s ScrollbarState) ContentLength(n int) ScrollbarState { s.contentLength = n; return s }

// ViewportContentLength returns a copy of s with the number of items the view
// shows at once set.
func (s ScrollbarState) ViewportContentLength(n int) ScrollbarState {
	s.viewportContentLength = n
	return s
}

// GetPosition returns the current position within the content.
func (s ScrollbarState) GetPosition() int { return s.position }

// Prev moves the position back by one, stopping at zero.
func (s *ScrollbarState) Prev() { s.position = max(s.position-1, 0) }

// Next moves the position forward by one, stopping at the last item.
func (s *ScrollbarState) Next() { s.position = min(s.position+1, max(s.contentLength-1, 0)) }

// First moves the position to the start of the content.
func (s *ScrollbarState) First() { s.position = 0 }

// Last moves the position to the end of the content.
func (s *ScrollbarState) Last() { s.position = max(s.contentLength-1, 0) }

// Scroll moves the position one step in the given direction.
func (s *ScrollbarState) Scroll(direction ScrollDirection) {
	switch direction {
	case ScrollForward:
		s.Next()
	case ScrollBackward:
		s.Prev()
	}
}

// Scrollbar draws a scrollbar along one edge of an area, showing where a
// ScrollbarState's position falls within its content.
//
//	<--▮------->
//	^  ^   ^   ^
//	│  │   │   └ end
//	│  │   └──── track
//	│  └──────── thumb
//	└─────────── begin
//
// Every part has its own symbol and style. The track, begin and end symbols
// can also be removed entirely, which leaves whatever was drawn underneath
// showing through. Create one with NewScrollbar; the zero value has no symbols
// at all.
//
//	scrollbar := widgets.NewScrollbar(widgets.ScrollbarVerticalRight).
//		BeginSymbol("↑").
//		EndSymbol("↓")
//	state := widgets.NewScrollbarState(len(items)).Position(scroll)
//	catatui.RenderStatefulWidget(scrollbar, area, buf, &state)
type Scrollbar struct {
	orientation    ScrollbarOrientation
	thumbStyle     catatui.Style
	thumbSymbol    string
	trackStyle     catatui.Style
	trackSymbol    string
	hasTrackSymbol bool
	beginSymbol    string
	hasBeginSymbol bool
	beginStyle     catatui.Style
	endSymbol      string
	hasEndSymbol   bool
	endStyle       catatui.Style
}

// NewScrollbar returns a scrollbar with the given orientation, drawn with the
// double-line symbol set that matches it.
func NewScrollbar(orientation ScrollbarOrientation) Scrollbar {
	return Scrollbar{orientation: orientation}.symbolsForOrientation()
}

func (s Scrollbar) symbolsForOrientation() Scrollbar {
	set := symbols.ScrollbarDoubleHorizontal
	if s.orientation.IsVertical() {
		set = symbols.ScrollbarDoubleVertical
	}
	s.thumbSymbol = set.Thumb
	s.trackSymbol, s.hasTrackSymbol = set.Track, true
	s.beginSymbol, s.hasBeginSymbol = set.Begin, true
	s.endSymbol, s.hasEndSymbol = set.End, true
	return s
}

// Orientation returns a copy of s with the given orientation, and its symbols
// reset to the double-line set matching it.
func (s Scrollbar) Orientation(orientation ScrollbarOrientation) Scrollbar {
	s.orientation = orientation
	set := symbols.ScrollbarDoubleHorizontal
	if s.orientation.IsVertical() {
		set = symbols.ScrollbarDoubleVertical
	}
	return s.Symbols(set)
}

// OrientationAndSymbol returns a copy of s with the given orientation and
// symbol set, the same as calling Orientation then Symbols.
func (s Scrollbar) OrientationAndSymbol(orientation ScrollbarOrientation, set symbols.ScrollbarSet) Scrollbar {
	s.orientation = orientation
	return s.Symbols(set)
}

// ThumbSymbol returns a copy of s with the thumb drawn as the given symbol.
func (s Scrollbar) ThumbSymbol(sym string) Scrollbar { s.thumbSymbol = sym; return s }

// ThumbStyle returns a copy of s with the thumb drawn in the given style.
func (s Scrollbar) ThumbStyle(style catatui.Style) Scrollbar { s.thumbStyle = style; return s }

// TrackSymbol returns a copy of s with the track drawn as the given symbol.
func (s Scrollbar) TrackSymbol(sym string) Scrollbar {
	s.trackSymbol, s.hasTrackSymbol = sym, true
	return s
}

// TrackSymbolNone returns a copy of s with no track symbol, so the cells
// between the thumb and the arrows are left untouched.
func (s Scrollbar) TrackSymbolNone() Scrollbar {
	s.trackSymbol, s.hasTrackSymbol = "", false
	return s
}

// TrackStyle returns a copy of s with the track drawn in the given style.
func (s Scrollbar) TrackStyle(style catatui.Style) Scrollbar { s.trackStyle = style; return s }

// BeginSymbol returns a copy of s with the given symbol at the start of the
// scrollbar.
func (s Scrollbar) BeginSymbol(sym string) Scrollbar {
	s.beginSymbol, s.hasBeginSymbol = sym, true
	return s
}

// BeginSymbolNone returns a copy of s with no begin symbol, so the track runs
// to the edge of the area.
func (s Scrollbar) BeginSymbolNone() Scrollbar {
	s.beginSymbol, s.hasBeginSymbol = "", false
	return s
}

// BeginStyle returns a copy of s with the begin symbol drawn in the given style.
func (s Scrollbar) BeginStyle(style catatui.Style) Scrollbar { s.beginStyle = style; return s }

// EndSymbol returns a copy of s with the given symbol at the end of the
// scrollbar.
func (s Scrollbar) EndSymbol(sym string) Scrollbar {
	s.endSymbol, s.hasEndSymbol = sym, true
	return s
}

// EndSymbolNone returns a copy of s with no end symbol, so the track runs to
// the edge of the area.
func (s Scrollbar) EndSymbolNone() Scrollbar {
	s.endSymbol, s.hasEndSymbol = "", false
	return s
}

// EndStyle returns a copy of s with the end symbol drawn in the given style.
func (s Scrollbar) EndStyle(style catatui.Style) Scrollbar { s.endStyle = style; return s }

// Symbols returns a copy of s drawn with the given symbol set.
//
// The track, begin and end symbols are only taken from the set if the
// scrollbar currently has them: a part removed with TrackSymbolNone,
// BeginSymbolNone or EndSymbolNone stays removed. Use those parts' own
// setters to bring them back.
func (s Scrollbar) Symbols(set symbols.ScrollbarSet) Scrollbar {
	s.thumbSymbol = set.Thumb
	if s.hasTrackSymbol {
		s.trackSymbol = set.Track
	}
	if s.hasBeginSymbol {
		s.beginSymbol = set.Begin
	}
	if s.hasEndSymbol {
		s.endSymbol = set.End
	}
	return s
}

// Style returns a copy of s with every part drawn in the given style.
func (s Scrollbar) Style(style catatui.Style) Scrollbar {
	s.trackStyle = style
	s.thumbStyle = style
	s.beginStyle = style
	s.endStyle = style
	return s
}

// RenderStateful draws the scrollbar along the edge of area selected by its
// orientation. Nothing is drawn when the content length is zero or the area
// leaves no room for a track once the arrows are taken out.
func (s Scrollbar) RenderStateful(area catatui.Rect, buf *catatui.Buffer, state *ScrollbarState) {
	area = area.Intersection(buf.Area)
	if area.IsEmpty() {
		return
	}
	if state.contentLength == 0 || s.trackLengthExcludingArrowHeads(area) == 0 {
		return
	}
	bar, ok := s.scrollbarArea(area)
	if !ok {
		return
	}
	var cells []catatui.Rect
	for _, col := range bar.Columns() {
		cells = append(cells, col.Rows()...)
	}
	for i, part := range s.barSymbols(bar, state) {
		if i >= len(cells) {
			break
		}
		if part.draw {
			buf.SetString(cells[i].X, cells[i].Y, part.symbol, part.style)
		}
	}
}

// scrollbarPart is one cell of the bar: what to draw there, or nothing when
// the track has no symbol.
type scrollbarPart struct {
	symbol string
	style  catatui.Style
	draw   bool
}

// barSymbols returns the bar cell by cell, from the begin symbol to the end
// symbol, with the parts that are absent left out.
func (s Scrollbar) barSymbols(area catatui.Rect, state *ScrollbarState) []scrollbarPart {
	trackStart, thumbLen, trackEnd := s.partLengths(area, state)

	track := scrollbarPart{symbol: s.trackSymbol, style: s.trackStyle, draw: s.hasTrackSymbol}
	thumb := scrollbarPart{symbol: s.thumbSymbol, style: s.thumbStyle, draw: true}

	parts := make([]scrollbarPart, 0, trackStart+thumbLen+trackEnd+2)
	if s.hasBeginSymbol {
		parts = append(parts, scrollbarPart{symbol: s.beginSymbol, style: s.beginStyle, draw: true})
	}
	for range trackStart {
		parts = append(parts, track)
	}
	for range thumbLen {
		parts = append(parts, thumb)
	}
	for range trackEnd {
		parts = append(parts, track)
	}
	if s.hasEndSymbol {
		parts = append(parts, scrollbarPart{symbol: s.endSymbol, style: s.endStyle, draw: true})
	}
	return parts
}

// partLengths returns the lengths of the three parts of the track:
//
//	<═══█████═══════>   full scrollbar
//	 ═══                track start
//	    █████           thumb
//	         ═══════    track end
func (s Scrollbar) partLengths(area catatui.Rect, state *ScrollbarState) (trackStart, thumbLen, trackEnd int) {
	// roundingDivide rounds to the nearest integer rather than down.
	roundingDivide := func(numerator, denominator int) int {
		return (numerator + denominator/2) / denominator
	}

	trackLength := int(s.trackLengthExcludingArrowHeads(area))
	if trackLength == 0 {
		return 0, 0, 0
	}

	viewportLength := s.viewportLength(state, area)

	maxPosition := max(state.contentLength-1, 0)
	startPosition := min(max(state.position, 0), maxPosition)
	maxViewportPosition := maxPosition + viewportLength

	if maxViewportPosition == 0 {
		// Just in case, to prevent division by zero.
		return 0, trackLength, 0
	}

	thumbLength := roundingDivide(viewportLength*trackLength, maxViewportPosition)
	thumbLength = min(max(thumbLength, 1), trackLength)

	// Clamp so the thumb always fits within the track. Clamping to
	// trackLength-1 instead let a large thumb overrun the track at the end,
	// pushing the end symbol out of the rendered area (ratatui issue #2582).
	thumbStart := roundingDivide(startPosition*trackLength, maxViewportPosition)
	thumbStart = min(max(thumbStart, 0), max(trackLength-thumbLength, 0))

	trackEnd = max(trackLength-(thumbStart+thumbLength), 0)
	return thumbStart, thumbLength, trackEnd
}

// scrollbarArea returns the single column or row of area the bar occupies.
func (s Scrollbar) scrollbarArea(area catatui.Rect) (catatui.Rect, bool) {
	var cells []catatui.Rect
	switch s.orientation {
	case ScrollbarVerticalLeft, ScrollbarVerticalRight:
		cells = area.Columns()
	default:
		cells = area.Rows()
	}
	if len(cells) == 0 {
		return catatui.Rect{}, false
	}
	switch s.orientation {
	case ScrollbarVerticalLeft, ScrollbarHorizontalTop:
		return cells[0], true
	default:
		return cells[len(cells)-1], true
	}
}

// trackLengthExcludingArrowHeads is the length of the bar left once the begin
// and end symbols have taken their cells.
//
//	       ┌────────── track length
//	 vvvvvvvvvvvvvvv
//	<═══█████═══════>
func (s Scrollbar) trackLengthExcludingArrowHeads(area catatui.Rect) uint16 {
	var startLen, endLen uint16
	if s.hasBeginSymbol {
		startLen = symbolWidth(s.beginSymbol)
	}
	if s.hasEndSymbol {
		endLen = symbolWidth(s.endSymbol)
	}
	arrowsLen := catatui.SatAdd(startLen, endLen)
	if s.orientation.IsVertical() {
		return catatui.SatSub(area.Height, arrowsLen)
	}
	return catatui.SatSub(area.Width, arrowsLen)
}

// viewportLength is the state's viewport length, or the track's own length
// when the state does not set one.
func (s Scrollbar) viewportLength(state *ScrollbarState, area catatui.Rect) int {
	if state.viewportContentLength != 0 {
		return state.viewportContentLength
	}
	if s.orientation.IsVertical() {
		return int(area.Height)
	}
	return int(area.Width)
}

func symbolWidth(sym string) uint16 {
	return uint16(min(catatui.StringWidth(sym), 0xFFFF))
}

var _ catatui.StatefulWidget[ScrollbarState] = Scrollbar{}
