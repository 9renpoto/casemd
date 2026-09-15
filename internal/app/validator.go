package app

import (
	"fmt"
	"io"

	"github.com/9renpoto/casemd/internal/core/domain"
)

// MarkdownParser parses cases and reports v1 Markdown diagnostics in one pass.
type MarkdownParser interface {
	ParseWithDiagnostics(source string, r io.Reader) ([]domain.Case, []domain.Diagnostic, error)
}

// ValidationError reports one or more v1 Markdown violations.
type ValidationError struct {
	Diagnostics []domain.Diagnostic
}

// Error summarizes the validation failure.
func (e *ValidationError) Error() string {
	return fmt.Sprintf("markdown validation failed with %d diagnostic(s)", len(e.Diagnostics))
}

// MarkdownValidator validates Markdown sources without producing artifacts.
type MarkdownValidator struct {
	parser MarkdownParser
}

// NewMarkdownValidator wires the validator with the shared Markdown parser.
func NewMarkdownValidator(parser MarkdownParser) *MarkdownValidator {
	return &MarkdownValidator{parser: parser}
}

// Validate checks every source and returns diagnostics in source and line order.
func (v *MarkdownValidator) Validate(sources []Source) ([]domain.Diagnostic, error) {
	_, diagnostics, err := parseSources(v.parser, sources)
	return diagnostics, err
}

type parsedSource struct {
	Name  string
	Cases []domain.Case
}

func parseSources(parser MarkdownParser, sources []Source) ([]parsedSource, []domain.Diagnostic, error) {
	if len(sources) == 0 {
		return nil, nil, fmt.Errorf("no sources provided")
	}

	parsed := make([]parsedSource, 0, len(sources))
	var diagnostics []domain.Diagnostic
	for _, source := range sources {
		cases, sourceDiagnostics, err := parser.ParseWithDiagnostics(source.Name, source.Reader)
		if err != nil {
			return nil, nil, fmt.Errorf("parse %s: %w", source.Name, err)
		}
		parsed = append(parsed, parsedSource{Name: source.Name, Cases: cases})
		diagnostics = append(diagnostics, sourceDiagnostics...)
	}

	return parsed, diagnostics, nil
}

func requireValidSources(parser MarkdownParser, sources []Source) ([]parsedSource, error) {
	parsed, diagnostics, err := parseSources(parser, sources)
	if err != nil {
		return nil, err
	}
	if len(diagnostics) > 0 {
		return nil, &ValidationError{Diagnostics: diagnostics}
	}
	return parsed, nil
}
