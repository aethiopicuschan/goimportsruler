package walker

import "iter"

// Violation represents a single import rule violation.
type Violation struct {
	file        string
	sourcePkg   string
	importPath  string
	ruleName    string
	description string
	line        int
	column      int
}

func (v Violation) File() string       { return v.file }
func (v Violation) SourcePkg() string  { return v.sourcePkg }
func (v Violation) ImportPath() string { return v.importPath }
func (v Violation) RuleName() string {
	if v.ruleName == "" {
		return "<unknown>"
	}
	return v.ruleName
}
func (v Violation) RuleDescription() string {
	if v.description == "" {
		return "<no description>"
	}
	return v.description
}
func (v Violation) Line() int   { return v.line }
func (v Violation) Column() int { return v.column }

// ViolationsError is a collection of violations.
// Note: This is not an "error" in the Go sense unless the caller decides to wrap/return it as one.
// The CLI can print details and then return a sentinel error.
type ViolationsError struct {
	violations []Violation
}

func (ve ViolationsError) Len() int { return len(ve.violations) }

// Violations returns an iterator over violations so callers can `range` over it.
func (ve ViolationsError) Violations() iter.Seq[Violation] {
	return func(yield func(Violation) bool) {
		for _, v := range ve.violations {
			if !yield(v) {
				return
			}
		}
	}
}
