// Port of ratatui-core/src/buffer/diff.rs @ ratatui-v0.30.2

package catatui

import (
	"fmt"
	"iter"
	"strings"
)

// visibleOnBlank are the modifiers that still show on a blank cell. If the
// previous frame drew a wide character with any of them, replacing it leaves
// visible residue in the columns it used to cover, so those columns have to be
// rewritten even when the buffer says they did not change.
const visibleOnBlank = ModifierReversed | ModifierUnderlined |
	ModifierSlowBlink | ModifierRapidBlink | ModifierCrossedOut

// variationSelector16 requests emoji presentation. Terminals are unreliable
// about clearing the trailing column of such a sequence, so those get explicit
// clears where plain CJK does not.
const variationSelector16 = '️'

// trailingState tracks a run of columns that a wide character used to cover and
// that now have to be re-emitted.
type trailingState struct {
	nextIndex int
	end       int
	force     bool
}

// bufferDiff walks two buffers and yields the cells that must be written to the
// terminal to turn prev into next.
//
// It is not a simple cell-by-cell comparison. A wide character occupies columns
// its neighbours must not overwrite, terminals disagree about who clears the
// trailing column of an emoji sequence, and a cell may opt out of diffing
// entirely; all of that lives here.
type bufferDiff struct {
	prev, next []Cell
	area       Rect
	pos        int
	trailing   *trailingState
}

func newBufferDiff(prev, next *Buffer) *bufferDiff {
	if prev.Area.X != next.Area.X || prev.Area.Y != next.Area.Y || prev.Area.Width != next.Area.Width {
		panic(fmt.Sprintf("catatui: buffer areas must have the same x, y, and width: prev=%+v, next=%+v", prev.Area, next.Area))
	}
	area := prev.Area
	area.Height = MinU16(area.Height, next.Area.Height)
	return &bufferDiff{prev: prev.Content, next: next.Content, area: area}
}

func (d *bufferDiff) posOf(index int) (uint16, uint16) {
	w := int(d.area.Width)
	return uint16(index%w + int(d.area.X)), uint16(index/w + int(d.area.Y))
}

func isSkip(c Cell) bool { return c.DiffOption.kind == DiffSkip }

// nextCell returns the next cell to write, or ok == false when the walk is done.
func (d *bufferDiff) nextCell() (PositionedCell, bool) {
	n := min(len(d.next), len(d.prev))

	// Finish any pending run of trailing columns before resuming the main walk.
	if d.trailing != nil {
		for d.trailing.nextIndex < d.trailing.end {
			j := d.trailing.nextIndex
			// Step over this cell, and over its own trailing column if it is
			// itself wide, so the main loop does not blank it afterwards.
			w := int(max(d.next[j].Width(), 1))
			d.trailing.nextIndex += w
			d.trailing.end = min(max(d.trailing.end, d.trailing.nextIndex), n)

			if !isSkip(d.next[j]) &&
				(d.trailing.force || d.prev[j].GetSymbol() != d.next[j].GetSymbol()) {
				x, y := d.posOf(j)
				return PositionedCell{X: x, Y: y, Cell: d.next[j]}, true
			}
		}
		d.pos = d.trailing.end
		d.trailing = nil
	}

	for d.pos < n {
		i := d.pos
		d.pos++

		current, previous := d.next[i], d.prev[i]
		if isSkip(current) {
			continue
		}

		if w, ok := current.DiffOption.ForcedWidth(); ok {
			d.pos += int(SatSub(w, 1))
			if !current.Equal(previous) {
				x, y := d.posOf(i)
				return PositionedCell{X: x, Y: y, Cell: current}, true
			}
			continue
		}

		// CellDiffNone and CellDiffAlwaysUpdate.
		width := int(current.Width())
		if current.DiffOption.kind == DiffNone && current.Equal(previous) {
			// Unchanged, but still step over the columns it covers.
			d.pos += max(width-1, 0)
			continue
		}
		previousWidth := int(previous.Width())

		switch {
		case width > 1 && strings.ContainsRune(current.GetSymbol(), variationSelector16):
			// Emoji presentation sequence: clear its trailing columns
			// explicitly, since terminals do not do it reliably.
			d.trailing = &trailingState{nextIndex: i + 1, end: min(i+width, n)}
		case width > 1:
			d.pos += width - 1
		case previousWidth > width &&
			(previous.Bg != ColorReset || previous.Modifier.Intersects(visibleOnBlank)):
			// A wide character with a style that shows on blank cells is being
			// replaced by a narrow one. Force-rewrite the columns it used to
			// cover, whether or not the buffer thinks they changed.
			d.trailing = &trailingState{nextIndex: i + 1, end: i + previousWidth, force: true}
		}

		x, y := d.posOf(i)
		return PositionedCell{X: x, Y: y, Cell: current}, true
	}
	return PositionedCell{}, false
}

// DiffSeq yields the cells that must be written to turn prev into next.
//
// The two buffers must share an X, Y and Width; only the height may differ, in
// which case the shorter one wins. Prefer this over Diff to avoid building the
// intermediate slice.
func (b *Buffer) DiffSeq(next *Buffer) iter.Seq[PositionedCell] {
	return func(yield func(PositionedCell) bool) {
		d := newBufferDiff(b, next)
		for {
			pc, ok := d.nextCell()
			if !ok || !yield(pc) {
				return
			}
		}
	}
}

// Diff collects the cells that must be written to turn b into next.
//
// The two buffers must share an X, Y and Width; only the height may differ.
func (b *Buffer) Diff(next *Buffer) []PositionedCell {
	var out []PositionedCell
	for pc := range b.DiffSeq(next) {
		out = append(out, pc)
	}
	return out
}
