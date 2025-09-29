package app

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/9renpoto/casemd/internal/core/domain"
)

type mockCaseParser struct {
	cases []domain.Case
	err   error
}

func (m *mockCaseParser) Parse(r io.Reader) ([]domain.Case, error) {
	return m.cases, m.err
}

func TestMarkdownToCSV_Convert(t *testing.T) {
	mockCases := []domain.Case{
		{
			MajorItem:       "Setup",
			MediumItem:      "Environment",
			MinorItem:       "Dependencies",
			ValidationSteps: []string{"Step 1", "Step 2"},
			Checkpoints:     []string{"* [ ] Check 1", "* [ ] Check 2"},
		},
		{
			MajorItem:       "Execution",
			MediumItem:      "Workflow",
			MinorItem:       "Run",
			ValidationSteps: []string{"Run command"},
			Checkpoints:     []string{"* [ ] Success"},
		},
	}

	parser := &mockCaseParser{cases: mockCases}
	converter := NewMarkdownToCSV(parser)

	input := strings.NewReader("") // Input is ignored by the mock
	var output bytes.Buffer

	err := converter.Convert(input, &output)
	if err != nil {
		t.Fatalf("Convert() returned an unexpected error: %v", err)
	}

	// encoding/csv only quotes fields that contain the separator, a newline, or a quote.
	// The test now reflects this standard behavior.
	expectedCSV := `Major Item,Medium Item,Minor Item,Validation Steps,Checkpoints,Result,Test Date,Tester,Notes
Setup,Environment,Dependencies,"Step 1
Step 2","* [ ] Check 1
* [ ] Check 2",,,,
Execution,Workflow,Run,Run command,* [ ] Success,,,,
`

	if output.String() != expectedCSV {
		t.Errorf("Convert() returned CSV:\n%s\nWant:\n%s", output.String(), expectedCSV)
	}
}
