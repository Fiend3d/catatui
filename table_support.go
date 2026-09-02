// Support for widgets/table.go: the one Constraint accessor the Table widget
// needs and ratatui gets from pattern matching.

package catatui

// IsPercentage reports whether c is a Percentage constraint, and if so what
// percentage it asks for.
func (c Constraint) IsPercentage() (uint16, bool) {
	return uint16(c.a), c.kind == constraintPercentage
}
