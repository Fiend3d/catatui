// Port of kasuari/src/row.rs @ v0.4.11

package kasuari

import (
	"maps"
	"slices"
)

// SymbolKind distinguishes the roles a symbol plays in the tableau.
type SymbolKind uint8

const (
	// SymbolInvalid is the "no symbol" sentinel.
	SymbolInvalid SymbolKind = iota
	// SymbolExternal stands for a user-supplied Variable.
	SymbolExternal
	// SymbolSlack takes up the difference in an inequality.
	SymbolSlack
	// SymbolError absorbs the violation of a non-required constraint, and is
	// what the objective function minimizes.
	SymbolError
	// SymbolDummy marks a required equality, and never leaves the basis.
	SymbolDummy
)

// symbol is an entry in the tableau. Symbols are numbered in creation order,
// which is what makes the deterministic scans below stable.
type symbol struct {
	id   int
	kind SymbolKind
}

func newSymbol(id int, kind SymbolKind) symbol { return symbol{id: id, kind: kind} }

func invalidSymbol() symbol { return symbol{id: 0, kind: SymbolInvalid} }

// compareSymbols orders symbols by id, then kind.
func compareSymbols(a, b symbol) int {
	if a.id != b.id {
		return a.id - b.id
	}
	return int(a.kind) - int(b.kind)
}

const epsilon = 1e-8

func nearZero(v float64) bool {
	if v < 0 {
		return -v < epsilon
	}
	return v < epsilon
}

// row is one equation of the tableau: a set of symbol coefficients plus a
// constant.
type row struct {
	cells    map[symbol]float64
	constant float64
}

func newRow(constant float64) *row {
	return &row{cells: make(map[symbol]float64), constant: constant}
}

func (r *row) clone() *row {
	cells := make(map[symbol]float64, len(r.cells))
	maps.Copy(cells, r.cells)
	return &row{cells: cells, constant: r.constant}
}

// sortedSymbols returns the row's symbols in id order.
//
// Every scan that picks "the first symbol satisfying X" goes through this. The
// Rust original scans a HashMap, where the pick is arbitrary but at least
// stable within a run; Go map order changes every run, so without this a
// degenerate system would lay out differently each time the program started.
func (r *row) sortedSymbols() []symbol {
	out := make([]symbol, 0, len(r.cells))
	for s := range r.cells {
		out = append(out, s)
	}
	slices.SortFunc(out, compareSymbols)
	return out
}

func (r *row) insertSymbol(s symbol, coefficient float64) {
	existing, ok := r.cells[s]
	if !ok {
		if !nearZero(coefficient) {
			r.cells[s] = coefficient
		}
		return
	}
	sum := existing + coefficient
	if nearZero(sum) {
		delete(r.cells, s)
	} else {
		r.cells[s] = sum
	}
}

// insertRow adds other scaled by coefficient into r, reporting whether the
// constant changed.
func (r *row) insertRow(other *row, coefficient float64) bool {
	constantDiff := other.constant * coefficient
	r.constant += constantDiff
	for _, s := range other.sortedSymbols() {
		r.insertSymbol(s, other.cells[s]*coefficient)
	}
	return constantDiff != 0
}

func (r *row) remove(s symbol) { delete(r.cells, s) }

func (r *row) reverseSign() {
	r.constant = -r.constant
	for s, v := range r.cells {
		r.cells[s] = -v
	}
}

// solveForSymbol rearranges the row to express s in terms of the others.
func (r *row) solveForSymbol(s symbol) {
	v, ok := r.cells[s]
	if !ok {
		panic("kasuari: solveForSymbol called with a symbol not in the row")
	}
	delete(r.cells, s)
	coeff := -1 / v
	r.constant *= coeff
	for k, val := range r.cells {
		r.cells[k] = val * coeff
	}
}

// solveForSymbols rearranges the row from "lhs = ..." to "rhs = ...".
func (r *row) solveForSymbols(lhs, rhs symbol) {
	r.insertSymbol(lhs, -1)
	r.solveForSymbol(rhs)
}

func (r *row) coefficientFor(s symbol) float64 { return r.cells[s] }

// substitute replaces s with the given row, reporting whether the constant
// changed.
func (r *row) substitute(s symbol, other *row) bool {
	coeff, ok := r.cells[s]
	if !ok {
		return false
	}
	delete(r.cells, s)
	return r.insertRow(other, coeff)
}
