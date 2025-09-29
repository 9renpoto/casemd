package parser

import (
	"reflect"
	"strings"
	"testing"

	"github.com/9renpoto/casemd/internal/core/domain"
)

func TestParse(t *testing.T) {
	markdown := `# Inspection Sheet

## Setup
### Environment
#### Dependencies
1. Install required packages
2. Confirm default configurations
* [ ] Packages installed successfully
* [ ] Defaults match specification

#### Environment variables
1. Validate required environment variables are set
* [ ] Variables align with deployment checklist

### Configuration
#### CLI defaults
1. Inspect generated CSV path
* [ ] Output file lands in build/
* [ ] Delimiter is comma

## Execution
### Workflow
#### CLI run
1. Run casemd with sample.md
* [ ] Exit code is 0
* [ ] CSV file exists

#### Post-run cleanup
1. Remove temporary files from build/
* [ ] No leftover artifacts

### Validation
#### Error handling
1. Run casemd without --input
* [ ] CLI prints actionable error
* [ ] Exit code is 1
`

	expectedCases := []domain.Case{
		{
			MajorItem:       "Setup",
			MediumItem:      "Environment",
			MinorItem:       "Dependencies",
			ValidationSteps: []string{"Install required packages", "Confirm default configurations"},
			Checkpoints:     []string{"* [ ] Packages installed successfully", "* [ ] Defaults match specification"},
		},
		{
			MajorItem:       "Setup",
			MediumItem:      "Environment",
			MinorItem:       "Environment variables",
			ValidationSteps: []string{"Validate required environment variables are set"},
			Checkpoints:     []string{"* [ ] Variables align with deployment checklist"},
		},
		{
			MajorItem:       "Setup",
			MediumItem:      "Configuration",
			MinorItem:       "CLI defaults",
			ValidationSteps: []string{"Inspect generated CSV path"},
			Checkpoints:     []string{"* [ ] Output file lands in build/", "* [ ] Delimiter is comma"},
		},
		{
			MajorItem:       "Execution",
			MediumItem:      "Workflow",
			MinorItem:       "CLI run",
			ValidationSteps: []string{"Run casemd with sample.md"},
			Checkpoints:     []string{"* [ ] Exit code is 0", "* [ ] CSV file exists"},
		},
		{
			MajorItem:       "Execution",
			MediumItem:      "Workflow",
			MinorItem:       "Post-run cleanup",
			ValidationSteps: []string{"Remove temporary files from build/"},
			Checkpoints:     []string{"* [ ] No leftover artifacts"},
		},
		{
			MajorItem:       "Execution",
			MediumItem:      "Validation",
			MinorItem:       "Error handling",
			ValidationSteps: []string{"Run casemd without --input"},
			Checkpoints:     []string{"* [ ] CLI prints actionable error", "* [ ] Exit code is 1"},
		},
	}

	reader := strings.NewReader(markdown)
	actualCases, err := Parse(reader)
	if err != nil {
		t.Fatalf("Parse() returned an unexpected error: %v", err)
	}

	if !reflect.DeepEqual(actualCases, expectedCases) {
		t.Errorf("Parse() returned %+v, want %+v", actualCases, expectedCases)
	}
}