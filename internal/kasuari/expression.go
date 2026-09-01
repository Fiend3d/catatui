// Package kasuari is a Go port of the kasuari crate @ v0.4.11, the incremental
// Cassowary constraint solver that ratatui's Layout is built on.
//
// Cassowary solves systems of linear constraints where the constraints have
// strengths, so an over-constrained system degrades gracefully instead of
// failing: the weakest constraints are violated first. That is what lets a
// layout with a fixed sidebar, a minimum-width pane and a filler behave sensibly
// at every terminal size.
//
// Two deliberate differences from the Rust original:
//
//   - Iteration order is deterministic. The Rust solver scans HashMaps in
//     arbitrary order in several places where the choice of pivot is a free
//     one; Go randomizes map iteration per run, which would make degenerate
//     layouts (three equal Fill constraints, say) come out differently from run
//     to run. Every such scan here walks symbols in id order instead.
//   - Constraint removal and edit variables are omitted. ratatui's Layout only
//     ever builds a solver, adds constraints and reads values back, and leaving
//     out the rest keeps this package to what is actually exercised.
//
// Port of kasuari/src/{variable,term,expression,relations,strength,constraint}.rs
package kasuari

import "sync/atomic"

// Variable is an unknown in the constraint system. Variables are compared by
// identity, so each call to NewVariable produces a distinct one.
type Variable uint64

var variableID atomic.Uint64

// NewVariable returns a fresh variable.
func NewVariable() Variable { return Variable(variableID.Add(1)) }

// Term is a variable scaled by a coefficient.
type Term struct {
	Variable    Variable
	Coefficient float64
}

// Expression is a linear combination of terms plus a constant.
type Expression struct {
	Terms    []Term
	Constant float64
}

// NewExpression returns an expression with the given constant and terms.
func NewExpression(constant float64, terms ...Term) Expression {
	return Expression{Terms: terms, Constant: constant}
}

// FromConstant returns the constant expression c.
func FromConstant(c float64) Expression { return Expression{Constant: c} }

// FromVariable returns the expression v, with coefficient 1.
func FromVariable(v Variable) Expression {
	return Expression{Terms: []Term{{Variable: v, Coefficient: 1}}}
}

// FromTerm returns the expression holding just the given term.
func FromTerm(t Term) Expression { return Expression{Terms: []Term{t}} }

// Add returns e + other.
func (e Expression) Add(other Expression) Expression {
	terms := make([]Term, 0, len(e.Terms)+len(other.Terms))
	terms = append(terms, e.Terms...)
	terms = append(terms, other.Terms...)
	return Expression{Terms: terms, Constant: e.Constant + other.Constant}
}

// Sub returns e - other.
func (e Expression) Sub(other Expression) Expression { return e.Add(other.Negate()) }

// AddConstant returns e + c.
func (e Expression) AddConstant(c float64) Expression {
	return Expression{Terms: e.Terms, Constant: e.Constant + c}
}

// AddVariable returns e + v.
func (e Expression) AddVariable(v Variable) Expression { return e.Add(FromVariable(v)) }

// SubVariable returns e - v.
func (e Expression) SubVariable(v Variable) Expression { return e.Sub(FromVariable(v)) }

// Negate returns -e.
func (e Expression) Negate() Expression { return e.Mul(-1) }

// Mul returns e scaled by f.
func (e Expression) Mul(f float64) Expression {
	terms := make([]Term, len(e.Terms))
	for i, t := range e.Terms {
		terms[i] = Term{Variable: t.Variable, Coefficient: t.Coefficient * f}
	}
	return Expression{Terms: terms, Constant: e.Constant * f}
}

// Div returns e scaled by 1/f.
func (e Expression) Div(f float64) Expression { return e.Mul(1 / f) }

// --- Strength -------------------------------------------------------------

// Strength is how hard the solver tries to satisfy a constraint. A constraint at
// Required is inviolable; weaker ones are given up in strength order when the
// system cannot satisfy everything.
type Strength float64

// The standard strengths.
const (
	Required Strength = 1_001_001_000
	Strong   Strength = 1_000_000
	Medium   Strength = 1_000
	Weak     Strength = 1
	Zero     Strength = 0
)

// NewStrength clamps a raw value into the valid range.
func NewStrength(v float64) Strength {
	return Strength(min(max(v, 0), float64(Required)))
}

// CreateStrength blends strong, medium and weak components into one strength,
// each clamped to 1000 after scaling by the multiplier. It matches kasuari's
// Strength::create, which ratatui uses to build its layout strength ladder.
func CreateStrength(strong, medium, weak, multiplier float64) Strength {
	s := min(max(strong*multiplier, 0), 1000) * float64(Strong)
	m := min(max(medium*multiplier, 0), 1000) * float64(Medium)
	w := min(max(weak*multiplier, 0), 1000) * float64(Weak)
	return NewStrength(s + m + w)
}

// --- Constraint -----------------------------------------------------------

// RelationalOperator is the comparison in a constraint.
type RelationalOperator uint8

const (
	// LessOrEqual is <=.
	LessOrEqual RelationalOperator = iota
	// Equal is ==.
	Equal
	// GreaterOrEqual is >=.
	GreaterOrEqual
)

// String returns the operator symbol.
func (o RelationalOperator) String() string {
	switch o {
	case LessOrEqual:
		return "<="
	case GreaterOrEqual:
		return ">="
	default:
		return "=="
	}
}

// Constraint is a linear relation the solver should satisfy, at a given
// strength. It holds the relation in the form expr <op> 0.
//
// Constraints are handled by pointer, so two structurally identical constraints
// are still distinct to the solver. This matches the Rust original, where
// Constraint is reference-counted and hashed by pointer.
type Constraint struct {
	expr     Expression
	op       RelationalOperator
	strength Strength
}

// NewConstraint returns the constraint expr <op> 0 at the given strength.
func NewConstraint(expr Expression, op RelationalOperator, strength Strength) *Constraint {
	return &Constraint{expr: expr, op: op, strength: strength}
}

// Relate returns the constraint lhs <op> rhs at the given strength.
func Relate(lhs Expression, op RelationalOperator, rhs Expression, strength Strength) *Constraint {
	return NewConstraint(lhs.Sub(rhs), op, strength)
}

// Expr returns the constraint's expression.
func (c *Constraint) Expr() Expression { return c.expr }

// Op returns the constraint's relational operator.
func (c *Constraint) Op() RelationalOperator { return c.op }

// Strength returns the constraint's strength.
func (c *Constraint) Strength() Strength { return c.strength }
