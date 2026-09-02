// Port of ratatui-core/src/symbols/merge.rs @ ratatui-v0.30.2

package symbols

import "fmt"

// MergeStrategy decides what happens when one box-drawing character is drawn
// over another.
//
// Drawing two blocks that share an edge normally leaves the second block's
// border on top, so a T-junction is drawn as a plain corner. Merging combines
// the two characters into the one that shows both lines, which is how borders
// are collapsed:
//
//	symbols.MergeReplace.Merge("│", "─") // "─", the second character wins
//	symbols.MergeExact.Merge("│", "─")   // "┼", the two lines together
//	symbols.MergeFuzzy.Merge("┘", "╔")   // "╬", the closest character there is
//
// Not every combination of box-drawing characters exists in Unicode. Dashed
// segments cannot be combined with anything but themselves, rounded corners
// have no junction forms, double and thick lines have no shared junctions, and
// a few double-with-plain junctions are simply missing. MergeExact falls back
// to MergeReplace for those; MergeFuzzy substitutes the closest style that does
// exist.
type MergeStrategy uint8

const (
	// MergeReplace draws the new character over the old one, which is the
	// default and what every widget did before merging existed.
	MergeReplace MergeStrategy = iota
	// MergeExact combines the two characters when a single character shows
	// both, and otherwise replaces.
	MergeExact
	// MergeFuzzy combines the two characters even when no exact character
	// exists, by moving one or both to a nearby line style. It falls back to
	// MergeExact, and so to MergeReplace.
	MergeFuzzy
)

// String returns the strategy's name.
func (s MergeStrategy) String() string {
	switch s {
	case MergeExact:
		return "Exact"
	case MergeFuzzy:
		return "Fuzzy"
	default:
		return "Replace"
	}
}

// ParseMergeStrategy parses "Replace", "Exact" or "Fuzzy".
func ParseMergeStrategy(s string) (MergeStrategy, error) {
	switch s {
	case "Replace":
		return MergeReplace, nil
	case "Exact":
		return MergeExact, nil
	case "Fuzzy":
		return MergeFuzzy, nil
	}
	return MergeReplace, fmt.Errorf("catatui: unknown merge strategy %q", s)
}

// Merge combines prev, the character already on screen, with next, the one
// being drawn over it, and returns the character to draw.
//
// A character outside the Box Drawing block cannot be merged: if next is not a
// box-drawing character it wins outright, and if only prev is not one, it is
// kept — so a merging border does not overwrite the text it runs into. Under
// MergeReplace next always wins.
func (s MergeStrategy) Merge(prev, next string) string {
	// Replace never looks at what is underneath.
	if s == MergeReplace {
		return next
	}

	prevSymbol, prevOK := parseBorderSymbol(prev)
	nextSymbol, nextOK := parseBorderSymbol(next)
	switch {
	case prevOK && nextOK:
		if merged, ok := prevSymbol.merge(nextSymbol, s).symbol(); ok {
			return merged
		}
		return next
	case !prevOK && nextOK:
		// Symbols that are not borders take precedence over the border
		// being drawn, in every strategy but Replace.
		return prev
	default:
		return next
	}
}

// lineStyle is the weight and pattern of one of the four line segments meeting
// in a box-drawing character.
type lineStyle uint8

const (
	// lineNothing is the absence of a segment.
	lineNothing lineStyle = iota
	// linePlain is a single line: ─ │.
	linePlain
	// lineRounded is a curved corner, which exists only in the four corner
	// characters: ╭ ╮ ╯ ╰.
	lineRounded
	// lineDouble is a double line: ═ ║.
	lineDouble
	// lineThick is a heavy line: ━ ┃.
	lineThick
	// lineDoubleDash is a two-dash line: ╌ ╎.
	lineDoubleDash
	// lineDoubleDashThick is a heavy two-dash line: ╍ ╏.
	lineDoubleDashThick
	// lineTripleDash is a three-dash line: ┄ ┆.
	lineTripleDash
	// lineTripleDashThick is a heavy three-dash line: ┅ ┇.
	lineTripleDashThick
	// lineQuadrupleDash is a four-dash line: ┈ ┊.
	lineQuadrupleDash
	// lineQuadrupleDashThick is a heavy four-dash line: ┉ ┋.
	lineQuadrupleDashThick
)

// merge keeps the segment being drawn unless it has nothing to draw there.
func (s lineStyle) merge(other lineStyle) lineStyle {
	if other == lineNothing {
		return s
	}
	return other
}

// borderSymbol is a box-drawing character taken apart into the four segments
// that meet in it, which is what makes merging a per-direction decision.
type borderSymbol struct {
	right lineStyle
	up    lineStyle
	left  lineStyle
	down  lineStyle
}

// merge combines two characters under a strategy. The exact result keeps every
// segment of both; fuzzy then moves it to the closest character that exists.
func (b borderSymbol) merge(other borderSymbol, strategy MergeStrategy) borderSymbol {
	exact := borderSymbol{
		right: b.right.merge(other.right),
		up:    b.up.merge(other.up),
		left:  b.left.merge(other.left),
		down:  b.down.merge(other.down),
	}
	switch strategy {
	case MergeFuzzy:
		return exact.fuzzy(other)
	case MergeExact:
		return exact
	default:
		return other
	}
}

// fuzzy walks the symbol towards one that Unicode actually has, in four steps.
// Where a choice has to be made, other — the character being drawn — decides,
// so the block drawn last keeps its own look.
func (b borderSymbol) fuzzy(other borderSymbol) borderSymbol {
	// Dashes exist only as plain vertical and horizontal lines.
	if !b.isStraight() {
		b = b.replace(lineDoubleDash, linePlain).
			replace(lineTripleDash, linePlain).
			replace(lineQuadrupleDash, linePlain).
			replace(lineDoubleDashThick, lineThick).
			replace(lineTripleDashThick, lineThick).
			replace(lineQuadrupleDashThick, lineThick)
	}

	// Rounded exists only as corners.
	if !b.isCorner() {
		b = b.replace(lineRounded, linePlain)
	}

	// There are no double-with-thick characters.
	if b.contains(lineDouble) && b.contains(lineThick) {
		if other.contains(lineDouble) {
			b = b.replace(lineThick, lineDouble)
		} else {
			b = b.replace(lineDouble, lineThick)
		}
	}

	// Some double-with-plain characters are missing.
	if _, ok := b.symbol(); !ok {
		if other.contains(lineDouble) {
			b = b.replace(linePlain, lineDouble)
		} else {
			b = b.replace(lineDouble, linePlain)
		}
	}
	return b
}

// isStraight reports whether the symbol is a single line, both halves of it in
// the same style.
func (b borderSymbol) isStraight() bool {
	return b.up == b.down && b.left == b.right &&
		(b.up == lineNothing || b.left == lineNothing)
}

// isCorner reports whether the symbol is a corner, both arms of it in the same
// style.
func (b borderSymbol) isCorner() bool {
	switch {
	case b.down == lineNothing && b.left == lineNothing:
		return b.up == b.right
	case b.up == lineNothing && b.left == lineNothing:
		return b.right == b.down
	case b.up == lineNothing && b.right == lineNothing:
		return b.down == b.left
	case b.right == lineNothing && b.down == lineNothing:
		return b.up == b.left
	default:
		return false
	}
}

// contains reports whether any of the four segments is drawn in style.
func (b borderSymbol) contains(style lineStyle) bool {
	return b.up == style || b.right == style || b.down == style || b.left == style
}

// replace swaps every segment drawn in from for one drawn in to.
func (b borderSymbol) replace(from, to lineStyle) borderSymbol {
	if b.up == from {
		b.up = to
	}
	if b.right == from {
		b.right = to
	}
	if b.down == from {
		b.down = to
	}
	if b.left == from {
		b.left = to
	}
	return b
}

// symbol returns the character for this combination of segments, if there is
// one. Most combinations have none: Unicode's Box Drawing block covers 125 of
// the 11^4 the type can express.
func (b borderSymbol) symbol() (string, bool) {
	s, ok := symbolByBorder[b]
	return s, ok
}

// parseBorderSymbol takes a box-drawing character apart into its segments,
// reporting false for anything outside the Box Drawing block.
func parseBorderSymbol(s string) (borderSymbol, bool) {
	b, ok := borderBySymbol[s]
	return b, ok
}

// borderSymbolEntry pairs a character with the segments that meet in it.
type borderSymbolEntry struct {
	symbol string
	border borderSymbol
}

var (
	borderBySymbol = func() map[string]borderSymbol {
		m := make(map[string]borderSymbol, len(borderSymbolTable))
		for _, e := range borderSymbolTable {
			m[e.symbol] = e.border
		}
		return m
	}()

	symbolByBorder = func() map[borderSymbol]string {
		m := make(map[borderSymbol]string, len(borderSymbolTable))
		for _, e := range borderSymbolTable {
			m[e.border] = e.symbol
		}
		return m
	}()
)

// borderSymbolTable is every character of Unicode's Box Drawing block that
// catatui can merge, with the four segments that make it up, in the order
// right, up, left, down.
var borderSymbolTable = []borderSymbolEntry{
	{"─", borderSymbol{linePlain, lineNothing, linePlain, lineNothing}},
	{"━", borderSymbol{lineThick, lineNothing, lineThick, lineNothing}},
	{"│", borderSymbol{lineNothing, linePlain, lineNothing, linePlain}},
	{"┃", borderSymbol{lineNothing, lineThick, lineNothing, lineThick}},
	{"┄", borderSymbol{lineTripleDash, lineNothing, lineTripleDash, lineNothing}},
	{"┅", borderSymbol{lineTripleDashThick, lineNothing, lineTripleDashThick, lineNothing}},
	{"┆", borderSymbol{lineNothing, lineTripleDash, lineNothing, lineTripleDash}},
	{"┇", borderSymbol{lineNothing, lineTripleDashThick, lineNothing, lineTripleDashThick}},
	{"┈", borderSymbol{lineQuadrupleDash, lineNothing, lineQuadrupleDash, lineNothing}},
	{"┉", borderSymbol{lineQuadrupleDashThick, lineNothing, lineQuadrupleDashThick, lineNothing}},
	{"┊", borderSymbol{lineNothing, lineQuadrupleDash, lineNothing, lineQuadrupleDash}},
	{"┋", borderSymbol{lineNothing, lineQuadrupleDashThick, lineNothing, lineQuadrupleDashThick}},
	{"┌", borderSymbol{linePlain, lineNothing, lineNothing, linePlain}},
	{"┍", borderSymbol{lineThick, lineNothing, lineNothing, linePlain}},
	{"┎", borderSymbol{linePlain, lineNothing, lineNothing, lineThick}},
	{"┏", borderSymbol{lineThick, lineNothing, lineNothing, lineThick}},
	{"┐", borderSymbol{lineNothing, lineNothing, linePlain, linePlain}},
	{"┑", borderSymbol{lineNothing, lineNothing, lineThick, linePlain}},
	{"┒", borderSymbol{lineNothing, lineNothing, linePlain, lineThick}},
	{"┓", borderSymbol{lineNothing, lineNothing, lineThick, lineThick}},
	{"└", borderSymbol{linePlain, linePlain, lineNothing, lineNothing}},
	{"┕", borderSymbol{lineThick, linePlain, lineNothing, lineNothing}},
	{"┖", borderSymbol{linePlain, lineThick, lineNothing, lineNothing}},
	{"┗", borderSymbol{lineThick, lineThick, lineNothing, lineNothing}},
	{"┘", borderSymbol{lineNothing, linePlain, linePlain, lineNothing}},
	{"┙", borderSymbol{lineNothing, linePlain, lineThick, lineNothing}},
	{"┚", borderSymbol{lineNothing, lineThick, linePlain, lineNothing}},
	{"┛", borderSymbol{lineNothing, lineThick, lineThick, lineNothing}},
	{"├", borderSymbol{linePlain, linePlain, lineNothing, linePlain}},
	{"┝", borderSymbol{lineThick, linePlain, lineNothing, linePlain}},
	{"┞", borderSymbol{linePlain, lineThick, lineNothing, linePlain}},
	{"┟", borderSymbol{linePlain, linePlain, lineNothing, lineThick}},
	{"┠", borderSymbol{linePlain, lineThick, lineNothing, lineThick}},
	{"┡", borderSymbol{lineThick, lineThick, lineNothing, linePlain}},
	{"┢", borderSymbol{lineThick, linePlain, lineNothing, lineThick}},
	{"┣", borderSymbol{lineThick, lineThick, lineNothing, lineThick}},
	{"┤", borderSymbol{lineNothing, linePlain, linePlain, linePlain}},
	{"┥", borderSymbol{lineNothing, linePlain, lineThick, linePlain}},
	{"┦", borderSymbol{lineNothing, lineThick, linePlain, linePlain}},
	{"┧", borderSymbol{lineNothing, linePlain, linePlain, lineThick}},
	{"┨", borderSymbol{lineNothing, lineThick, linePlain, lineThick}},
	{"┩", borderSymbol{lineNothing, lineThick, lineThick, linePlain}},
	{"┪", borderSymbol{lineNothing, linePlain, lineThick, lineThick}},
	{"┫", borderSymbol{lineNothing, lineThick, lineThick, lineThick}},
	{"┬", borderSymbol{linePlain, lineNothing, linePlain, linePlain}},
	{"┭", borderSymbol{linePlain, lineNothing, lineThick, linePlain}},
	{"┮", borderSymbol{lineThick, lineNothing, linePlain, linePlain}},
	{"┯", borderSymbol{lineThick, lineNothing, lineThick, linePlain}},
	{"┰", borderSymbol{linePlain, lineNothing, linePlain, lineThick}},
	{"┱", borderSymbol{linePlain, lineNothing, lineThick, lineThick}},
	{"┲", borderSymbol{lineThick, lineNothing, linePlain, lineThick}},
	{"┳", borderSymbol{lineThick, lineNothing, lineThick, lineThick}},
	{"┴", borderSymbol{linePlain, linePlain, linePlain, lineNothing}},
	{"┵", borderSymbol{linePlain, linePlain, lineThick, lineNothing}},
	{"┶", borderSymbol{lineThick, linePlain, linePlain, lineNothing}},
	{"┷", borderSymbol{lineThick, linePlain, lineThick, lineNothing}},
	{"┸", borderSymbol{linePlain, lineThick, linePlain, lineNothing}},
	{"┹", borderSymbol{linePlain, lineThick, lineThick, lineNothing}},
	{"┺", borderSymbol{lineThick, lineThick, linePlain, lineNothing}},
	{"┻", borderSymbol{lineThick, lineThick, lineThick, lineNothing}},
	{"┼", borderSymbol{linePlain, linePlain, linePlain, linePlain}},
	{"┽", borderSymbol{linePlain, linePlain, lineThick, linePlain}},
	{"┾", borderSymbol{lineThick, linePlain, linePlain, linePlain}},
	{"┿", borderSymbol{lineThick, linePlain, lineThick, linePlain}},
	{"╀", borderSymbol{linePlain, lineThick, linePlain, linePlain}},
	{"╁", borderSymbol{linePlain, linePlain, linePlain, lineThick}},
	{"╂", borderSymbol{linePlain, lineThick, linePlain, lineThick}},
	{"╃", borderSymbol{linePlain, lineThick, lineThick, linePlain}},
	{"╄", borderSymbol{lineThick, lineThick, linePlain, linePlain}},
	{"╅", borderSymbol{linePlain, linePlain, lineThick, lineThick}},
	{"╆", borderSymbol{lineThick, linePlain, linePlain, lineThick}},
	{"╇", borderSymbol{lineThick, lineThick, lineThick, linePlain}},
	{"╈", borderSymbol{lineThick, linePlain, lineThick, lineThick}},
	{"╉", borderSymbol{linePlain, lineThick, lineThick, lineThick}},
	{"╊", borderSymbol{lineThick, lineThick, linePlain, lineThick}},
	{"╋", borderSymbol{lineThick, lineThick, lineThick, lineThick}},
	{"╌", borderSymbol{lineDoubleDash, lineNothing, lineDoubleDash, lineNothing}},
	{"╍", borderSymbol{lineDoubleDashThick, lineNothing, lineDoubleDashThick, lineNothing}},
	{"╎", borderSymbol{lineNothing, lineDoubleDash, lineNothing, lineDoubleDash}},
	{"╏", borderSymbol{lineNothing, lineDoubleDashThick, lineNothing, lineDoubleDashThick}},
	{"═", borderSymbol{lineDouble, lineNothing, lineDouble, lineNothing}},
	{"║", borderSymbol{lineNothing, lineDouble, lineNothing, lineDouble}},
	{"╒", borderSymbol{lineDouble, lineNothing, lineNothing, linePlain}},
	{"╓", borderSymbol{linePlain, lineNothing, lineNothing, lineDouble}},
	{"╔", borderSymbol{lineDouble, lineNothing, lineNothing, lineDouble}},
	{"╕", borderSymbol{lineNothing, lineNothing, lineDouble, linePlain}},
	{"╖", borderSymbol{lineNothing, lineNothing, linePlain, lineDouble}},
	{"╗", borderSymbol{lineNothing, lineNothing, lineDouble, lineDouble}},
	{"╘", borderSymbol{lineDouble, linePlain, lineNothing, lineNothing}},
	{"╙", borderSymbol{linePlain, lineDouble, lineNothing, lineNothing}},
	{"╚", borderSymbol{lineDouble, lineDouble, lineNothing, lineNothing}},
	{"╛", borderSymbol{lineNothing, linePlain, lineDouble, lineNothing}},
	{"╜", borderSymbol{lineNothing, lineDouble, linePlain, lineNothing}},
	{"╝", borderSymbol{lineNothing, lineDouble, lineDouble, lineNothing}},
	{"╞", borderSymbol{lineDouble, linePlain, lineNothing, linePlain}},
	{"╟", borderSymbol{linePlain, lineDouble, lineNothing, lineDouble}},
	{"╠", borderSymbol{lineDouble, lineDouble, lineNothing, lineDouble}},
	{"╡", borderSymbol{lineNothing, linePlain, lineDouble, linePlain}},
	{"╢", borderSymbol{lineNothing, lineDouble, linePlain, lineDouble}},
	{"╣", borderSymbol{lineNothing, lineDouble, lineDouble, lineDouble}},
	{"╤", borderSymbol{lineDouble, lineNothing, lineDouble, linePlain}},
	{"╥", borderSymbol{linePlain, lineNothing, linePlain, lineDouble}},
	{"╦", borderSymbol{lineDouble, lineNothing, lineDouble, lineDouble}},
	{"╧", borderSymbol{lineDouble, linePlain, lineDouble, lineNothing}},
	{"╨", borderSymbol{linePlain, lineDouble, linePlain, lineNothing}},
	{"╩", borderSymbol{lineDouble, lineDouble, lineDouble, lineNothing}},
	{"╪", borderSymbol{lineDouble, linePlain, lineDouble, linePlain}},
	{"╫", borderSymbol{linePlain, lineDouble, linePlain, lineDouble}},
	{"╬", borderSymbol{lineDouble, lineDouble, lineDouble, lineDouble}},
	{"╭", borderSymbol{lineRounded, lineNothing, lineNothing, lineRounded}},
	{"╮", borderSymbol{lineNothing, lineNothing, lineRounded, lineRounded}},
	{"╯", borderSymbol{lineNothing, lineRounded, lineRounded, lineNothing}},
	{"╰", borderSymbol{lineRounded, lineRounded, lineNothing, lineNothing}},
	{"╴", borderSymbol{lineNothing, lineNothing, linePlain, lineNothing}},
	{"╵", borderSymbol{lineNothing, linePlain, lineNothing, lineNothing}},
	{"╶", borderSymbol{linePlain, lineNothing, lineNothing, lineNothing}},
	{"╷", borderSymbol{lineNothing, lineNothing, lineNothing, linePlain}},
	{"╸", borderSymbol{lineNothing, lineNothing, lineThick, lineNothing}},
	{"╹", borderSymbol{lineNothing, lineThick, lineNothing, lineNothing}},
	{"╺", borderSymbol{lineThick, lineNothing, lineNothing, lineNothing}},
	{"╻", borderSymbol{lineNothing, lineNothing, lineNothing, lineThick}},
	{"╼", borderSymbol{lineThick, lineNothing, linePlain, lineNothing}},
	{"╽", borderSymbol{lineNothing, linePlain, lineNothing, lineThick}},
	{"╾", borderSymbol{linePlain, lineNothing, lineThick, lineNothing}},
	{"╿", borderSymbol{lineNothing, lineThick, lineNothing, linePlain}},
}
