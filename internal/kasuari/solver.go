// Port of kasuari/src/solver.rs @ v0.4.11

package kasuari

import (
	"errors"
	"math"
	"slices"
)

// Errors the solver can return.
var (
	// ErrUnsatisfiableConstraint means a Required constraint contradicts the
	// ones already added.
	ErrUnsatisfiableConstraint = errors.New("kasuari: unsatisfiable constraint")
	// ErrDuplicateConstraint means the same Constraint was added twice.
	ErrDuplicateConstraint = errors.New("kasuari: duplicate constraint")
	// ErrObjectiveUnbounded means the objective function has no minimum, which
	// indicates a malformed system.
	ErrObjectiveUnbounded = errors.New("kasuari: the objective is unbounded")
)

// tag records the marker and error symbols introduced for one constraint.
type tag struct {
	marker symbol
	other  symbol
}

// Solver is an incremental Cassowary solver.
//
// Build one, add constraints, then read variable values back with GetValue. The
// solver re-optimizes after each constraint, so values are always current.
//
// A Solver is not safe for concurrent use.
type Solver struct {
	constraints   map[*Constraint]tag
	varSymbol     map[Variable]symbol
	varForSymbol  map[symbol]Variable
	rows          map[symbol]*row
	objective     *row
	artificial    *row
	hasArtificial bool
	idTick        int
}

// NewSolver returns an empty solver.
func NewSolver() *Solver {
	return &Solver{
		constraints:  make(map[*Constraint]tag),
		varSymbol:    make(map[Variable]symbol),
		varForSymbol: make(map[symbol]Variable),
		rows:         make(map[symbol]*row),
		objective:    newRow(0),
		idTick:       1,
	}
}

// sortedRowSymbols returns the basic symbols in id order, so that scans over the
// tableau are reproducible. See row.sortedSymbols for why this matters.
func (s *Solver) sortedRowSymbols() []symbol {
	out := make([]symbol, 0, len(s.rows))
	for sym := range s.rows {
		out = append(out, sym)
	}
	slices.SortFunc(out, compareSymbols)
	return out
}

// AddConstraints adds each constraint in turn, stopping at the first error.
func (s *Solver) AddConstraints(constraints ...*Constraint) error {
	for _, c := range constraints {
		if err := s.AddConstraint(c); err != nil {
			return err
		}
	}
	return nil
}

// AddConstraint adds one constraint and re-optimizes.
func (s *Solver) AddConstraint(constraint *Constraint) error {
	if _, ok := s.constraints[constraint]; ok {
		return ErrDuplicateConstraint
	}

	r, t := s.createRow(constraint)
	subject := chooseSubject(r, t)

	// If no entering symbol was found but the row is all dummies, a zero
	// constant means the constraint is merely redundant and the dummy marker
	// can enter the basis; a non-zero one means it cannot be satisfied.
	if subject.kind == SymbolInvalid && allDummies(r) {
		if !nearZero(r.constant) {
			return ErrUnsatisfiableConstraint
		}
		subject = t.marker
	}

	if subject.kind == SymbolInvalid {
		satisfiable, err := s.addWithArtificialVariable(r)
		if err != nil {
			return err
		}
		if !satisfiable {
			return ErrUnsatisfiableConstraint
		}
	} else {
		r.solveForSymbol(subject)
		s.substitute(subject, r)
		s.rows[subject] = r
	}

	s.constraints[constraint] = t

	// Optimizing after every constraint keeps the average system small and
	// leaves the solver in a consistent state throughout.
	return s.optimize(s.objective)
}

// createRow turns a constraint into a tableau row, introducing the slack, error
// and dummy symbols the relation needs, and registering error symbols in the
// objective so that violating them has a cost.
func (s *Solver) createRow(constraint *Constraint) (*row, tag) {
	expr := constraint.Expr()
	r := newRow(expr.Constant)

	for _, term := range expr.Terms {
		if nearZero(term.Coefficient) {
			continue
		}
		sym := s.getVarSymbol(term.Variable)
		if other, ok := s.rows[sym]; ok {
			r.insertRow(other, term.Coefficient)
		} else {
			r.insertSymbol(sym, term.Coefficient)
		}
	}

	var t tag
	switch constraint.Op() {
	case GreaterOrEqual, LessOrEqual:
		coeff := -1.0
		if constraint.Op() == LessOrEqual {
			coeff = 1.0
		}
		slack := s.newSymbol(SymbolSlack)
		r.insertSymbol(slack, coeff)
		if constraint.Strength() < Required {
			errSym := s.newSymbol(SymbolError)
			r.insertSymbol(errSym, -coeff)
			s.objective.insertSymbol(errSym, float64(constraint.Strength()))
			t = tag{marker: slack, other: errSym}
		} else {
			t = tag{marker: slack, other: invalidSymbol()}
		}
	case Equal:
		if constraint.Strength() < Required {
			// v = errPlus - errMinus, so the objective can pay to move v in
			// either direction.
			errPlus := s.newSymbol(SymbolError)
			errMinus := s.newSymbol(SymbolError)
			r.insertSymbol(errPlus, -1)
			r.insertSymbol(errMinus, 1)
			s.objective.insertSymbol(errPlus, float64(constraint.Strength()))
			s.objective.insertSymbol(errMinus, float64(constraint.Strength()))
			t = tag{marker: errPlus, other: errMinus}
		} else {
			dummy := s.newSymbol(SymbolDummy)
			r.insertSymbol(dummy, 1)
			t = tag{marker: dummy, other: invalidSymbol()}
		}
	}

	if r.constant < 0 {
		r.reverseSign()
	}
	return r, t
}

func (s *Solver) newSymbol(kind SymbolKind) symbol {
	sym := newSymbol(s.idTick, kind)
	s.idTick++
	return sym
}

func (s *Solver) getVarSymbol(v Variable) symbol {
	if sym, ok := s.varSymbol[v]; ok {
		return sym
	}
	sym := s.newSymbol(SymbolExternal)
	s.varSymbol[v] = sym
	s.varForSymbol[sym] = v
	return sym
}

// chooseSubject picks the symbol to solve the row for: an external variable if
// there is one, otherwise a slack or error symbol with a negative coefficient.
func chooseSubject(r *row, t tag) symbol {
	for _, sym := range r.sortedSymbols() {
		if sym.kind == SymbolExternal {
			return sym
		}
	}
	if (t.marker.kind == SymbolSlack || t.marker.kind == SymbolError) && r.coefficientFor(t.marker) < 0 {
		return t.marker
	}
	if (t.other.kind == SymbolSlack || t.other.kind == SymbolError) && r.coefficientFor(t.other) < 0 {
		return t.other
	}
	return invalidSymbol()
}

// addWithArtificialVariable handles a row with no natural entering symbol, by
// minimizing an artificial objective. It reports whether the row is satisfiable.
func (s *Solver) addWithArtificialVariable(r *row) (bool, error) {
	art := s.newSymbol(SymbolSlack)
	s.rows[art] = r.clone()
	s.artificial = r.clone()
	s.hasArtificial = true

	if err := s.optimize(s.artificial); err != nil {
		s.artificial, s.hasArtificial = nil, false
		return false, err
	}
	success := nearZero(s.artificial.constant)
	s.artificial, s.hasArtificial = nil, false

	if artRow, ok := s.rows[art]; ok {
		delete(s.rows, art)
		if len(artRow.cells) == 0 {
			return success, nil
		}
		entering := anyPivotableSymbol(artRow)
		if entering.kind == SymbolInvalid {
			return false, nil
		}
		artRow.solveForSymbols(art, entering)
		s.substitute(entering, artRow)
		s.rows[entering] = artRow
	}

	for _, r := range s.rows {
		r.remove(art)
	}
	s.objective.remove(art)
	return success, nil
}

// substitute replaces sym throughout the tableau, the objective and any
// artificial objective.
//
// The Rust original also records rows whose constant went negative, for
// dual_optimize to repair. That path only runs from remove_constraint and
// suggest_value, neither of which is ported, so the bookkeeping is left out.
func (s *Solver) substitute(sym symbol, r *row) {
	for _, other := range s.sortedRowSymbols() {
		s.rows[other].substitute(sym, r)
	}
	s.objective.substitute(sym, r)
	if s.hasArtificial {
		s.artificial.substitute(sym, r)
	}
}

// optimize runs simplex until the objective has no negative coefficient left.
func (s *Solver) optimize(objective *row) error {
	for {
		entering := getEnteringSymbol(objective)
		if entering.kind == SymbolInvalid {
			return nil
		}
		leaving, r, ok := s.getLeavingRow(entering)
		if !ok {
			return ErrObjectiveUnbounded
		}
		r.solveForSymbols(leaving, entering)
		s.substitute(entering, r)
		s.rows[entering] = r
	}
}

func getEnteringSymbol(objective *row) symbol {
	for _, sym := range objective.sortedSymbols() {
		if sym.kind != SymbolDummy && objective.cells[sym] < 0 {
			return sym
		}
	}
	return invalidSymbol()
}

func (s *Solver) getDualEnteringSymbol(r *row) symbol {
	entering := invalidSymbol()
	ratio := math.Inf(1)
	for _, sym := range r.sortedSymbols() {
		v := r.cells[sym]
		if v > 0 && sym.kind != SymbolDummy {
			got := s.objective.coefficientFor(sym) / v
			if got < ratio {
				ratio = got
				entering = sym
			}
		}
	}
	return entering
}

func anyPivotableSymbol(r *row) symbol {
	for _, sym := range r.sortedSymbols() {
		if sym.kind == SymbolSlack || sym.kind == SymbolError {
			return sym
		}
	}
	return invalidSymbol()
}

// getLeavingRow picks the row that limits how far the entering symbol can move,
// removing and returning it.
func (s *Solver) getLeavingRow(entering symbol) (symbol, *row, bool) {
	ratio := math.Inf(1)
	found := invalidSymbol()
	ok := false
	for _, sym := range s.sortedRowSymbols() {
		if sym.kind == SymbolExternal {
			continue
		}
		coeff := s.rows[sym].coefficientFor(entering)
		if coeff < 0 {
			got := -s.rows[sym].constant / coeff
			if got < ratio {
				ratio = got
				found = sym
				ok = true
			}
		}
	}
	if !ok {
		return invalidSymbol(), nil, false
	}
	r := s.rows[found]
	delete(s.rows, found)
	return found, r, true
}

func allDummies(r *row) bool {
	for sym := range r.cells {
		if sym.kind != SymbolDummy {
			return false
		}
	}
	return true
}

// GetValue returns the current value of a variable, or zero if the variable does
// not appear in any constraint.
func (s *Solver) GetValue(v Variable) float64 {
	sym, ok := s.varSymbol[v]
	if !ok {
		return 0
	}
	r, ok := s.rows[sym]
	if !ok {
		return 0
	}
	return r.constant
}
