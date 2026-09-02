// Port of ratatui-widgets/src/clear.rs @ ratatui-v0.30.2

package widgets

import "github.com/Fiend3d/catatui"

// Clear resets every cell in its area to the empty cell, so that whatever is
// drawn next does not show what was underneath. It is what you render before a
// popup, so the popup's block covers the content behind it rather than mixing
// with it.
//
//	f.RenderWidget(widgets.Clear{}, popupArea)
//	f.RenderWidget(widgets.Bordered().Title("Confirm"), popupArea)
//
// It cannot be used to clear the terminal on the first frame, because the
// terminal already assumes the render area is empty; use Terminal.Clear for
// that.
type Clear struct{}

// Render resets every cell in area that lies within the buffer.
func (Clear) Render(area catatui.Rect, buf *catatui.Buffer) {
	area = area.Intersection(buf.Area)
	if area.IsEmpty() {
		return
	}
	for x := area.Left(); x < area.Right(); x++ {
		for y := area.Top(); y < area.Bottom(); y++ {
			buf.Get(x, y).Reset()
		}
	}
}

var _ catatui.Widget = Clear{}
