// Port of ratatui-core/src/buffer/cell.rs @ ratatui-v0.30.2

package catatui

// CellDiffOptionKind tags a CellDiffOption.
type CellDiffOptionKind uint8

const (
	// DiffNone applies no special handling.
	DiffNone CellDiffOptionKind = iota
	// DiffSkip omits the cell when the buffer is flushed to the screen.
	DiffSkip
	// DiffAlwaysUpdate redraws the cell even when it is unchanged.
	DiffAlwaysUpdate
	// DiffForcedWidth overrides the width computed from the symbol.
	DiffForcedWidth
)

// CellDiffOption controls how a cell takes part in the diff against the
// previous frame.
//
// It exists for cells whose on-screen appearance the buffer cannot infer from
// the symbol alone: cells covered by terminal graphics, cells another renderer
// draws over, and escape sequences whose printed width differs from their text
// width.
type CellDiffOption struct {
	kind  CellDiffOptionKind
	width uint16
}

// The diff options that carry no payload.
var (
	// CellDiffNone is the default: the cell is diffed normally.
	CellDiffNone = CellDiffOption{kind: DiffNone}
	// CellDiffSkip prevents the buffer from overwriting a cell covered by
	// something outside its control, such as a sixel or kitty image.
	CellDiffSkip = CellDiffOption{kind: DiffSkip}
	// CellDiffAlwaysUpdate bypasses the equality check against the previous
	// buffer, so the cell is redrawn every frame.
	CellDiffAlwaysUpdate = CellDiffOption{kind: DiffAlwaysUpdate}
)

// CellForcedWidth forces the cell to occupy the given number of columns
// regardless of the width of its symbol. The width must be greater than zero;
// CellForcedWidth(0) returns CellDiffNone, matching the NonZeroU16 that
// ratatui's ForcedWidth carries.
func CellForcedWidth(width uint16) CellDiffOption {
	if width == 0 {
		return CellDiffNone
	}
	return CellDiffOption{kind: DiffForcedWidth, width: width}
}

// Kind returns which diff option this is.
func (o CellDiffOption) Kind() CellDiffOptionKind { return o.kind }

// ForcedWidth reports whether the option forces a width, returning it if so.
func (o CellDiffOption) ForcedWidth() (uint16, bool) {
	return o.width, o.kind == DiffForcedWidth
}

// Cell is a single character cell of the terminal: a grapheme cluster plus the
// style it is drawn in.
//
// The zero Cell is a valid empty cell, equivalent to EmptyCell: an empty Symbol
// reads back as a single space, and unset colors read back as ColorReset.
// Unlike Style, a Cell's colors are always concrete, never "inherit".
//
// Do not compare Cells with ==. A cell with no symbol and a cell holding a
// single space are the same cell, and only Equal knows that.
type Cell struct {
	// Symbol is the grapheme cluster drawn in this cell, which may be wider
	// than one column. The empty string means the default symbol, a space.
	Symbol string

	// Fg is the foreground color.
	Fg Color

	// Bg is the background color.
	Bg Color

	// UnderlineColor is the underline color, which only some terminals render
	// separately from the text color.
	UnderlineColor Color

	// Modifier holds the text attributes.
	Modifier Modifier

	// DiffOption controls how the cell takes part in the frame diff.
	DiffOption CellDiffOption
}

// EmptyCell returns a blank cell: no symbol, every color reset, no modifiers.
func EmptyCell() Cell {
	return Cell{
		Fg:             ColorReset,
		Bg:             ColorReset,
		UnderlineColor: ColorReset,
	}
}

// NewCell returns an empty cell displaying the given symbol.
func NewCell(symbol string) Cell {
	c := EmptyCell()
	c.Symbol = symbol
	return c
}

// GetSymbol returns the symbol drawn in the cell, which is a single space when
// no symbol is set.
func (c Cell) GetSymbol() string {
	if c.Symbol == "" {
		return " "
	}
	return c.Symbol
}

// SetSymbol sets the grapheme cluster drawn in the cell.
func (c *Cell) SetSymbol(symbol string) *Cell {
	c.Symbol = symbol
	return c
}

// SetChar sets the cell's symbol to a single rune.
func (c *Cell) SetChar(r rune) *Cell {
	c.Symbol = string(r)
	return c
}

// SetFg sets the foreground color.
func (c *Cell) SetFg(color Color) *Cell {
	c.Fg = color
	return c
}

// SetBg sets the background color.
func (c *Cell) SetBg(color Color) *Cell {
	c.Bg = color
	return c
}

// SetStyle applies a style to the cell. Colors the style leaves unset are kept,
// which is what makes styles layer.
func (c *Cell) SetStyle(s Style) *Cell {
	if s.fg.IsSet() {
		c.Fg = s.fg
	}
	if s.bg.IsSet() {
		c.Bg = s.bg
	}
	if s.underlineColor.IsSet() {
		c.UnderlineColor = s.underlineColor
	}
	c.Modifier = c.Modifier.Insert(s.addModifier).Remove(s.subModifier)
	return c
}

// GetStyle returns the cell's style. Every color is set, since a cell always
// has concrete colors, and no modifiers are subtracted.
func (c Cell) GetStyle() Style {
	return Style{
		fg:             orReset(c.Fg),
		bg:             orReset(c.Bg),
		underlineColor: orReset(c.UnderlineColor),
		addModifier:    c.Modifier,
	}
}

// SetDiffOption sets how the cell takes part in the frame diff.
func (c *Cell) SetDiffOption(o CellDiffOption) *Cell {
	c.DiffOption = o
	return c
}

// SetSkip is shorthand for setting CellDiffSkip or CellDiffNone.
func (c *Cell) SetSkip(skip bool) *Cell {
	if skip {
		c.DiffOption = CellDiffSkip
	} else {
		c.DiffOption = CellDiffNone
	}
	return c
}

// Skip reports whether the cell is skipped when flushed to the screen.
func (c Cell) Skip() bool { return c.DiffOption.kind == DiffSkip }

// Reset returns the cell to the empty state.
func (c *Cell) Reset() { *c = EmptyCell() }

// Equal reports whether two cells would draw identically. A cell with no symbol
// and a cell holding a single space compare equal, as do unset and reset
// colors, so that cells built different ways still diff as unchanged.
func (c Cell) Equal(other Cell) bool {
	return c.GetSymbol() == other.GetSymbol() &&
		orReset(c.Fg) == orReset(other.Fg) &&
		orReset(c.Bg) == orReset(other.Bg) &&
		orReset(c.UnderlineColor) == orReset(other.UnderlineColor) &&
		c.Modifier == other.Modifier &&
		c.DiffOption == other.DiffOption
}

// Width is the number of columns the cell occupies, honouring a forced width.
func (c Cell) Width() uint16 {
	if w, ok := c.DiffOption.ForcedWidth(); ok {
		return w
	}
	return cellWidth(c.GetSymbol())
}

// orReset normalizes an unset color to ColorReset, so that the zero Cell
// behaves exactly like EmptyCell.
func orReset(c Color) Color {
	if c.IsSet() {
		return c
	}
	return ColorReset
}
