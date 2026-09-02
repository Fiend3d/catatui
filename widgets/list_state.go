// Port of ratatui-widgets/src/list/state.rs @ ratatui-v0.30.2

package widgets

import "math"

// ListState is the caller-owned state of a List: which item is selected and
// which item is shown first.
//
// Rendering the list as a stateful widget highlights the selected item and
// adjusts the offset so that it is visible, writing the new offset back into
// the state. The state therefore lives outside the draw function, in the
// application, and is passed to catatui.RenderStatefulWidget on every frame.
//
//	var state widgets.ListState // stored in the application
//
//	state.SetOffset(1) // display the second item and onwards
//	state.Select(3)    // select the fourth item (0-indexed)
//	catatui.RenderStatefulWidget(list, area, buf, &state)
//
// The zero value has nothing selected and shows the list from the top.
//
// Deviation from ratatui: the selection is an Option<usize> in Rust. Here it
// is a pair read with Selected and written with Select and SelectNone. Indices
// are ints and never negative; a negative index passed to Select or SetOffset
// is treated as 0.
type ListState struct {
	offset      int
	selected    int
	hasSelected bool
}

// NewListState returns a state with nothing selected and offset 0. It is the
// same as the zero value.
func NewListState() ListState { return ListState{} }

// WithOffset returns a copy of s showing the list from the given item.
func (s ListState) WithOffset(offset int) ListState { s.offset = max(offset, 0); return s }

// WithSelected returns a copy of s with the given item selected.
func (s ListState) WithSelected(index int) ListState {
	s.selected, s.hasSelected = max(index, 0), true
	return s
}

// Offset is the index of the first item displayed.
func (s ListState) Offset() int { return s.offset }

// SetOffset sets the index of the first item displayed. Rendering corrects it
// if the selected item would otherwise be out of view.
func (s *ListState) SetOffset(offset int) { s.offset = max(offset, 0) }

// Selected returns the index of the selected item, and false if nothing is
// selected.
func (s ListState) Selected() (int, bool) { return s.selected, s.hasSelected }

// Select selects the item at the given index.
//
// Until the list is rendered, the number of items is not known, so an index
// past the end is kept as given and corrected to the last item on rendering.
func (s *ListState) Select(index int) {
	s.selected, s.hasSelected = max(index, 0), true
}

// SelectNone clears the selection. This also resets the offset to 0, as
// ratatui's select(None) does.
func (s *ListState) SelectNone() {
	s.selected, s.hasSelected = 0, false
	s.offset = 0
}

// SelectNext selects the item after the selected one, or the first item if
// nothing is selected.
func (s *ListState) SelectNext() {
	next := 0
	if s.hasSelected {
		next = satAddInt(s.selected, 1)
	}
	s.Select(next)
}

// SelectPrevious selects the item before the selected one, or the last item if
// nothing is selected.
//
// Until the list is rendered, the number of items is not known, so "the last
// item" is represented by math.MaxInt and corrected on rendering.
func (s *ListState) SelectPrevious() {
	previous := math.MaxInt
	if s.hasSelected {
		previous = max(s.selected-1, 0)
	}
	s.Select(previous)
}

// SelectFirst selects the first item.
func (s *ListState) SelectFirst() { s.Select(0) }

// SelectLast selects the last item.
//
// Until the list is rendered, the number of items is not known, so the index is
// set to math.MaxInt and corrected on rendering.
func (s *ListState) SelectLast() { s.Select(math.MaxInt) }

// ScrollDownBy moves the selection down by amount items, starting from the
// first item if nothing is selected. An index past the end selects the last
// item once the list is rendered.
func (s *ListState) ScrollDownBy(amount uint16) {
	selected := 0
	if s.hasSelected {
		selected = s.selected
	}
	s.Select(satAddInt(selected, int(amount)))
}

// ScrollUpBy moves the selection up by amount items, stopping at the first
// item. If nothing is selected, the first item is selected.
func (s *ListState) ScrollUpBy(amount uint16) {
	selected := 0
	if s.hasSelected {
		selected = s.selected
	}
	s.Select(max(selected-int(amount), 0))
}

// satAddInt returns a+b for non-negative operands, clamped at math.MaxInt
// instead of overflowing, mirroring usize::saturating_add.
func satAddInt(a, b int) int {
	if a > math.MaxInt-b {
		return math.MaxInt
	}
	return a + b
}
