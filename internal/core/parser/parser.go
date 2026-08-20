package parser

import (
	"bufio"
	"io"
	"regexp"
	"sort"
	"strings"

	"github.com/9renpoto/casemd/internal/core/domain"
)

var (
	orderedListRegex = regexp.MustCompile(`^\d+\.\s+(.*)`)
	taskListRegex    = regexp.MustCompile(`^\*\s+\[[ x]\]\s+(.*)`)
	headingRegex     = regexp.MustCompile(`^(#{1,6})(?:[ \t]+(.*))?$`)
)

// Parse extracts test cases from a Markdown reader.
func Parse(r io.Reader) ([]domain.Case, error) {
	cases, _, err := ParseWithDiagnostics("", r)
	return cases, err
}

// Validate checks a Markdown reader against the v1 structure without returning cases.
func Validate(source string, r io.Reader) ([]domain.Diagnostic, error) {
	_, diagnostics, err := ParseWithDiagnostics(source, r)
	return diagnostics, err
}

// ParseWithDiagnostics extracts test cases and reports all v1 structure violations.
func ParseWithDiagnostics(source string, r io.Reader) ([]domain.Case, []domain.Diagnostic, error) {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)

	var cases []domain.Case
	var diagnostics []domain.Diagnostic
	var currentCase *domain.Case
	var currentCaseLine int
	var majorItem, mediumItem string
	var hasMajor, hasMedium bool
	var majorLine, mediumLine int
	var majorHasMedium, mediumHasCase bool
	var sawHierarchy bool
	var sawDocumentTitle bool

	finishCurrentCase := func() {
		if currentCase == nil {
			return
		}

		if len(currentCase.ValidationSteps) == 0 {
			diagnostics = append(diagnostics, domain.Diagnostic{
				Source:     source,
				Line:       currentCaseLine,
				Rule:       domain.RuleMissingSteps,
				Message:    "test case must contain at least one ordered step",
				Suggestion: "add an ordered list item such as `1. Perform the action`",
			})
		}
		if len(currentCase.Checkpoints) == 0 {
			diagnostics = append(diagnostics, domain.Diagnostic{
				Source:     source,
				Line:       currentCaseLine,
				Rule:       domain.RuleMissingCheckpoints,
				Message:    "test case must contain at least one task-list checkpoint",
				Suggestion: "add a task-list item such as `* [ ] Confirm the result`",
			})
		}

		cases = append(cases, *currentCase)
		currentCase = nil
	}

	finishMedium := func() {
		if hasMedium && !mediumHasCase {
			diagnostics = append(diagnostics, domain.Diagnostic{
				Source:     source,
				Line:       mediumLine,
				Rule:       domain.RuleHeadingHierarchy,
				Message:    "level 3 heading must contain at least one level 4 test case",
				Suggestion: "add a `####` test case under this medium item",
			})
		}
		hasMedium = false
		mediumHasCase = false
	}

	finishMajor := func() {
		finishMedium()
		if hasMajor && !majorHasMedium {
			diagnostics = append(diagnostics, domain.Diagnostic{
				Source:     source,
				Line:       majorLine,
				Rule:       domain.RuleHeadingHierarchy,
				Message:    "level 2 heading must contain at least one level 3 heading",
				Suggestion: "add a `###` medium item under this major item",
			})
		}
		hasMajor = false
		majorHasMedium = false
	}

	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)

		if matches := headingRegex.FindStringSubmatch(line); len(matches) > 0 {
			level := len(matches[1])
			title := strings.TrimSpace(matches[2])

			if title == "" {
				diagnostics = append(diagnostics, domain.Diagnostic{
					Source:     source,
					Line:       lineNumber,
					Rule:       domain.RuleEmptyHeading,
					Message:    "heading must not be empty",
					Suggestion: "add a descriptive title after the heading markers",
				})
			}

			switch level {
			case 1:
				if sawDocumentTitle || sawHierarchy {
					diagnostics = append(diagnostics, domain.Diagnostic{
						Source:     source,
						Line:       lineNumber,
						Rule:       domain.RuleDocumentTitle,
						Message:    "document title must appear at most once and before the heading hierarchy",
						Suggestion: "keep a single `#` document title before the first `##` major item",
					})
				}
				sawDocumentTitle = true
			case 2:
				finishCurrentCase()
				finishMajor()
				majorItem = title
				mediumItem = ""
				hasMajor = true
				majorLine = lineNumber
				sawHierarchy = true
			case 3:
				finishCurrentCase()
				finishMedium()
				if !hasMajor {
					diagnostics = append(diagnostics, domain.Diagnostic{
						Source:     source,
						Line:       lineNumber,
						Rule:       domain.RuleHeadingHierarchy,
						Message:    "level 3 heading requires a preceding level 2 heading",
						Suggestion: "add a `##` major item before this heading",
					})
				}
				mediumItem = title
				hasMedium = true
				mediumLine = lineNumber
				mediumHasCase = false
				if hasMajor {
					majorHasMedium = true
				}
				sawHierarchy = true
			case 4:
				finishCurrentCase()
				if !hasMajor || !hasMedium {
					diagnostics = append(diagnostics, domain.Diagnostic{
						Source:     source,
						Line:       lineNumber,
						Rule:       domain.RuleHeadingHierarchy,
						Message:    "level 4 test case requires preceding level 2 and level 3 headings",
						Suggestion: "add `##` and `###` parent headings before this test case",
					})
				}
				if hasMedium {
					mediumHasCase = true
				}
				currentCase = &domain.Case{
					MajorItem:  majorItem,
					MediumItem: mediumItem,
					MinorItem:  title,
				}
				currentCaseLine = lineNumber
				sawHierarchy = true
			default:
				diagnostics = append(diagnostics, domain.Diagnostic{
					Source:     source,
					Line:       lineNumber,
					Rule:       domain.RuleUnsupportedHeading,
					Message:    "v1 Markdown supports heading levels 1 through 4 only",
					Suggestion: "use `####` for a test case or move this text into the test case body",
				})
			}
		} else if matches := orderedListRegex.FindStringSubmatch(trimmedLine); len(matches) > 1 && currentCase != nil {
			if step := strings.TrimSpace(matches[1]); step != "" {
				currentCase.ValidationSteps = append(currentCase.ValidationSteps, step)
			}
		} else if matches := taskListRegex.FindStringSubmatch(trimmedLine); len(matches) > 1 && currentCase != nil {
			if checkpoint := strings.TrimSpace(matches[1]); checkpoint != "" {
				currentCase.Checkpoints = append(currentCase.Checkpoints, trimmedLine)
			}
		}
	}

	finishCurrentCase()
	finishMajor()
	if !sawHierarchy {
		diagnostics = append(diagnostics, domain.Diagnostic{
			Source:     source,
			Line:       1,
			Rule:       domain.RuleHeadingHierarchy,
			Message:    "document must contain a level 2, level 3, and level 4 heading hierarchy",
			Suggestion: "add `##` and `###` parent headings followed by a `####` test case",
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, nil, err
	}

	sortDiagnostics(diagnostics)
	return cases, diagnostics, nil
}

func sortDiagnostics(diagnostics []domain.Diagnostic) {
	ruleOrder := map[domain.ValidationRule]int{
		domain.RuleEmptyHeading:       0,
		domain.RuleDocumentTitle:      1,
		domain.RuleHeadingHierarchy:   2,
		domain.RuleUnsupportedHeading: 3,
		domain.RuleMissingSteps:       4,
		domain.RuleMissingCheckpoints: 5,
	}

	sort.SliceStable(diagnostics, func(i, j int) bool {
		if diagnostics[i].Line != diagnostics[j].Line {
			return diagnostics[i].Line < diagnostics[j].Line
		}
		return ruleOrder[diagnostics[i].Rule] < ruleOrder[diagnostics[j].Rule]
	})
}
