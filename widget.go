// Port of ratatui-core/src/widgets/widget.rs and stateful_widget.rs @ ratatui-v0.30.2

package catatui

// Widget is anything that can draw itself into a region of a Buffer.
//
// Widgets are immediate mode: they hold no state between frames and are built
// fresh each time, so a widget value is cheap and disposable. Implement Render
// on a value receiver, and treat the widget as consumed by the call, as ratatui
// does.
//
//	type Banner struct{ Text string }
//
//	func (b Banner) Render(area catatui.Rect, buf *catatui.Buffer) {
//		buf.SetStringn(area.X, area.Y, b.Text, area.Width, catatui.NewStyle())
//	}
//
// Render must not draw outside area. Buffer.Get panics on out-of-bounds
// coordinates precisely so that this shows up immediately rather than as
// corruption elsewhere on screen.
type Widget interface {
	Render(area Rect, buf *Buffer)
}

// StatefulWidget is a Widget that reads and updates caller-owned state, such as
// which row of a list is selected or how far a view is scrolled.
//
// The state lives with the caller and outlives the widget, which is what lets a
// list remember its selection while the widget itself is rebuilt each frame.
//
// Deviation from ratatui: Rust declares the state as an associated type on the
// trait. Go has no associated types, so the interface is generic over the state
// and rendering goes through the free function RenderStatefulWidget rather than
// a method on Frame — Go methods cannot have their own type parameters.
type StatefulWidget[S any] interface {
	RenderStateful(area Rect, buf *Buffer, state *S)
}

// WidgetFunc adapts a plain function to the Widget interface.
type WidgetFunc func(area Rect, buf *Buffer)

// Render calls f.
func (f WidgetFunc) Render(area Rect, buf *Buffer) { f(area, buf) }

// RenderWidget draws a widget into a buffer at the given area, clipped to the
// buffer. Widgets normally reach a buffer through Frame.RenderWidget; this is
// the direct form, useful for composing one widget inside another.
func RenderWidget(w Widget, area Rect, buf *Buffer) {
	w.Render(buf.Area.Intersection(area), buf)
}

// RenderStatefulWidget draws a stateful widget into a buffer at the given area,
// clipped to the buffer, letting it update state.
func RenderStatefulWidget[S any](w StatefulWidget[S], area Rect, buf *Buffer, state *S) {
	w.RenderStateful(buf.Area.Intersection(area), buf, state)
}
