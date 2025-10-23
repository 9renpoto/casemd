package cli

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/9renpoto/casemd/internal/app"
)

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

	tool := New(&stdout, &stderr, nil, nil, creator)
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
	tool := New(&stdout, &stderr, nil, nil, nil)

	err := tool.Run([]string{"--input", inputPath, "--google-spreadsheet-title", "Casemd Export"})
	if !errors.Is(err, errMissingGoogleConverter) {
		t.Fatalf("expected errMissingGoogleConverter, got %v", err)
	}
}
