package cli

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/9renpoto/casemd/internal/app"
	"github.com/9renpoto/casemd/internal/core/domain"
	"github.com/9renpoto/casemd/internal/core/parser"
)

type mockValidator struct {
	diagnostics []domain.Diagnostic
	sources     []app.Source
	err         error
}

func (m *mockValidator) Validate(sources []app.Source) ([]domain.Diagnostic, error) {
	m.sources = sources
	return m.diagnostics, m.err
}

type coreParserAdapter struct{}

func (coreParserAdapter) ParseWithDiagnostics(source string, r io.Reader) ([]domain.Case, []domain.Diagnostic, error) {
	return parser.ParseWithDiagnostics(source, r)
}

type mockConverter struct {
	calls int
}

func (m *mockConverter) Convert(_ []app.Source, _ io.Writer) error {
	m.calls++
	return nil
}

type mockGoogleSpreadsheetCreator struct {
	id      string
	title   string
	sources []app.Source
	err     error
}

func (m *mockGoogleSpreadsheetCreator) Create(ctx context.Context, title string, sources []app.Source) (string, error) {
	m.title = title
	m.sources = sources
	return m.id, m.err
}

func TestToolRunCreatesGoogleSpreadsheet(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	inputPath := filepath.Join(dir, "case.md")
	if err := os.WriteFile(inputPath, []byte("# Case"), 0o644); err != nil {
		t.Fatalf("write input file: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	creator := &mockGoogleSpreadsheetCreator{id: "sheet-id"}

	tool := New(&stdout, &stderr, &mockValidator{}, nil, nil, creator)
	if err := tool.Run([]string{"--input", inputPath, "--google-spreadsheet-title", "Casemd Export"}); err != nil {
		t.Fatalf("Run() returned an unexpected error: %v", err)
	}

	if creator.title != "Casemd Export" {
		t.Fatalf("unexpected title: %s", creator.title)
	}
	if len(creator.sources) != 1 {
		t.Fatalf("expected 1 source, got %d", len(creator.sources))
	}
	if !strings.Contains(stdout.String(), "sheet-id") {
		t.Fatalf("stdout missing sheet id: %s", stdout.String())
	}
}

func TestToolRunRequiresGoogleConverter(t *testing.T) {
	dir := t.TempDir()
	inputPath := filepath.Join(dir, "case.md")
	if err := os.WriteFile(inputPath, []byte("# Case"), 0o644); err != nil {
		t.Fatalf("write input file: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	tool := New(&stdout, &stderr, &mockValidator{}, nil, nil, nil)

	err := tool.Run([]string{"--input", inputPath, "--google-spreadsheet-title", "Casemd Export"})
	if !errors.Is(err, errMissingGoogleConverter) {
		t.Fatalf("expected errMissingGoogleConverter, got %v", err)
	}
}

func TestToolRunValidateMultipleInputs(t *testing.T) {
	dir := t.TempDir()
	firstPath := writeMarkdown(t, dir, "first.md", validMarkdown("First"))
	secondPath := writeMarkdown(t, dir, "second.md", validMarkdown("Second"))

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	validator := app.NewMarkdownValidator(coreParserAdapter{})
	tool := New(&stdout, &stderr, validator, nil, nil, nil)

	if err := tool.Run([]string{"validate", "--input", firstPath, "--input", secondPath}); err != nil {
		t.Fatalf("Run() returned an unexpected error: %v", err)
	}
	if stdout.Len() != 0 {
		t.Fatalf("validate wrote to stdout: %q", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("validate wrote diagnostics for valid input: %q", stderr.String())
	}
}

func TestToolRunValidateReportsActionableDiagnostics(t *testing.T) {
	dir := t.TempDir()
	inputPath := writeMarkdown(t, dir, "invalid.md", "## Setup\n### Environment\n#### Incomplete\n")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	validator := app.NewMarkdownValidator(coreParserAdapter{})
	tool := New(&stdout, &stderr, validator, nil, nil, nil)

	err := tool.Run([]string{"validate", "--input", inputPath})
	var validationErr *app.ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("expected ValidationError, got %v", err)
	}
	if len(validationErr.Diagnostics) != 2 {
		t.Fatalf("diagnostics = %+v, want missing steps and checkpoints", validationErr.Diagnostics)
	}
	for _, want := range []string{
		inputPath + ":3: missing-steps:",
		inputPath + ":3: missing-checkpoints:",
		"suggestion:",
	} {
		if !strings.Contains(stderr.String(), want) {
			t.Fatalf("stderr %q does not contain %q", stderr.String(), want)
		}
	}
	if stdout.Len() != 0 {
		t.Fatalf("validate wrote to stdout: %q", stdout.String())
	}
}

func TestToolRunRejectsInvalidInputBeforeCreatingAnyOutput(t *testing.T) {
	dir := t.TempDir()
	validPath := writeMarkdown(t, dir, "valid.md", validMarkdown("Valid"))
	invalidPath := writeMarkdown(t, dir, "invalid.md", "## Setup\n### Environment\n#### Incomplete\n")
	csvPath := filepath.Join(dir, "outputs", "cases.csv")
	spreadsheetPath := filepath.Join(dir, "outputs", "cases.xlsx")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	validator := app.NewMarkdownValidator(coreParserAdapter{})
	csvConverter := &mockConverter{}
	spreadsheetConverter := &mockConverter{}
	googleCreator := &mockGoogleSpreadsheetCreator{id: "should-not-exist"}
	tool := New(&stdout, &stderr, validator, csvConverter, spreadsheetConverter, googleCreator)

	err := tool.Run([]string{
		"--input", validPath,
		"--input", invalidPath,
		"--csv-output", csvPath,
		"--spreadsheet-output", spreadsheetPath,
		"--google-spreadsheet-title", "Invalid Export",
	})
	var validationErr *app.ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("expected ValidationError, got %v", err)
	}
	if csvConverter.calls != 0 || spreadsheetConverter.calls != 0 {
		t.Fatalf("converters were called before validation completed: CSV=%d spreadsheet=%d", csvConverter.calls, spreadsheetConverter.calls)
	}
	if googleCreator.title != "" {
		t.Fatalf("Google Spreadsheet creator was called before validation completed")
	}
	for _, path := range []string{csvPath, spreadsheetPath} {
		if _, statErr := os.Stat(path); !errors.Is(statErr, os.ErrNotExist) {
			t.Fatalf("output %s exists or returned an unexpected error: %v", path, statErr)
		}
	}
}

func TestToolRunValidateHelp(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	tool := New(&stdout, &stderr, nil, nil, nil, nil)

	if err := tool.Run([]string{"validate", "--help"}); err != nil {
		t.Fatalf("Run() returned an unexpected error: %v", err)
	}
	if !strings.Contains(stderr.String(), "casemd validate --input") {
		t.Fatalf("help does not describe validate usage: %q", stderr.String())
	}
}

func writeMarkdown(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write input file: %v", err)
	}
	return path
}

func validMarkdown(caseName string) string {
	return "## Setup\n### Environment\n#### " + caseName + "\n1. Perform the action\n* [ ] Confirm the result\n"
}
