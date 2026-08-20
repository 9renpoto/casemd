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

func TestValidate(t *testing.T) {
	tests := []struct {
		name      string
		markdown  string
		want      []domain.Diagnostic
		wantCases int
	}{
		{
			name: "valid document with optional title",
			markdown: `# Inspection Sheet

## Setup
### Environment
#### Dependencies
1. Install required packages
* [ ] Packages installed successfully
`,
			wantCases: 1,
		},
		{
			name: "valid document without title",
			markdown: `## Setup
### Environment
#### Dependencies
1. Install required packages
* [x] Packages installed successfully
`,
			wantCases: 1,
		},
		{
			name: "headings without required parents",
			markdown: `### Environment
#### Dependencies
1. Install required packages
* [ ] Packages installed successfully
`,
			want: []domain.Diagnostic{
				{Line: 1, Rule: domain.RuleHeadingHierarchy},
				{Line: 2, Rule: domain.RuleHeadingHierarchy},
			},
			wantCases: 1,
		},
		{
			name: "empty headings and incomplete case",
			markdown: `##
###
####
`,
			want: []domain.Diagnostic{
				{Line: 1, Rule: domain.RuleEmptyHeading},
				{Line: 2, Rule: domain.RuleEmptyHeading},
				{Line: 3, Rule: domain.RuleEmptyHeading},
				{Line: 3, Rule: domain.RuleMissingSteps},
				{Line: 3, Rule: domain.RuleMissingCheckpoints},
			},
			wantCases: 1,
		},
		{
			name: "multiple incomplete cases use source order",
			markdown: `## Setup
### Environment
#### Missing steps
* [ ] Packages installed successfully
#### Missing checkpoints
1. Install required packages
`,
			want: []domain.Diagnostic{
				{Line: 3, Rule: domain.RuleMissingSteps},
				{Line: 5, Rule: domain.RuleMissingCheckpoints},
			},
			wantCases: 2,
		},
		{
			name:     "document without hierarchy",
			markdown: "# Inspection Sheet\n",
			want: []domain.Diagnostic{
				{Line: 1, Rule: domain.RuleHeadingHierarchy},
			},
		},
		{
			name: "incomplete hierarchy",
			markdown: `## Setup
### Environment
`,
			want: []domain.Diagnostic{
				{Line: 2, Rule: domain.RuleHeadingHierarchy},
			},
		},
		{
			name: "major item without medium item",
			markdown: `## Setup
`,
			want: []domain.Diagnostic{
				{Line: 1, Rule: domain.RuleHeadingHierarchy},
			},
		},
		{
			name: "unsupported heading level",
			markdown: `## Setup
### Environment
##### Detail
`,
			want: []domain.Diagnostic{
				{Line: 2, Rule: domain.RuleHeadingHierarchy},
				{Line: 3, Rule: domain.RuleUnsupportedHeading},
			},
		},
		{
			name: "duplicate and misplaced document titles",
			markdown: `# Inspection Sheet
# Duplicate
## Setup
### Environment
#### Dependencies
1. Install required packages
* [ ] Packages installed successfully
# Misplaced
`,
			want: []domain.Diagnostic{
				{Line: 2, Rule: domain.RuleDocumentTitle},
				{Line: 8, Rule: domain.RuleDocumentTitle},
			},
			wantCases: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			const source = "scenario.md"

			cases, diagnostics, err := ParseWithDiagnostics(source, strings.NewReader(tt.markdown))
			if err != nil {
				t.Fatalf("ParseWithDiagnostics() returned an unexpected error: %v", err)
			}
			if len(cases) != tt.wantCases {
				t.Errorf("ParseWithDiagnostics() returned %d cases, want %d", len(cases), tt.wantCases)
			}
			if len(diagnostics) != len(tt.want) {
				t.Fatalf("ParseWithDiagnostics() diagnostics = %+v, want %+v", diagnostics, tt.want)
			}

			for index, want := range tt.want {
				got := diagnostics[index]
				if got.Source != source || got.Line != want.Line || got.Rule != want.Rule {
					t.Errorf("diagnostic %d = %+v, want source %q, line %d, rule %q", index, got, source, want.Line, want.Rule)
				}
				if got.Message == "" {
					t.Errorf("diagnostic %d has an empty message", index)
				}
				if got.Suggestion == "" {
					t.Errorf("diagnostic %d has an empty suggestion", index)
				}
			}

			validated, err := Validate(source, strings.NewReader(tt.markdown))
			if err != nil {
				t.Fatalf("Validate() returned an unexpected error: %v", err)
			}
			if !reflect.DeepEqual(validated, diagnostics) {
				t.Errorf("Validate() returned %+v, want %+v", validated, diagnostics)
			}
		})
	}
}
