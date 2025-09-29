package parser

import (
	"bufio"
	"io"
	"regexp"
	"strings"

	"github.com/9renpoto/casemd/internal/core/domain"
)

var (
	orderedListRegex = regexp.MustCompile(`^\d+\.\s+(.*)`)
	taskListRegex    = regexp.MustCompile(`^\*\s+\[[ x]\]\s+(.*)`)
)

// Parse extracts test cases from a Markdown reader.
func Parse(r io.Reader) ([]domain.Case, error) {
	scanner := bufio.NewScanner(r)
	var cases []domain.Case
	var currentCase *domain.Case
	var majorItem, mediumItem string

	for scanner.Scan() {
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)

		if strings.HasPrefix(line, "## ") {
			majorItem = strings.TrimPrefix(line, "## ")
			mediumItem = "" // Reset on new major item
		} else if strings.HasPrefix(line, "### ") {
			mediumItem = strings.TrimPrefix(line, "### ")
		} else if strings.HasPrefix(line, "#### ") {
			if currentCase != nil {
				cases = append(cases, *currentCase)
			}
			currentCase = &domain.Case{
				MajorItem:  majorItem,
				MediumItem: mediumItem,
				MinorItem:  strings.TrimPrefix(line, "#### "),
			}
		} else if matches := orderedListRegex.FindStringSubmatch(trimmedLine); len(matches) > 1 && currentCase != nil {
			currentCase.ValidationSteps = append(currentCase.ValidationSteps, matches[1])
		} else if matches := taskListRegex.FindStringSubmatch(trimmedLine); len(matches) > 1 && currentCase != nil {
			// The regex captures the content, but the original spec wants the `* [ ]` part.
			// Let's just use the trimmed line for checkpoints to preserve the marker.
			currentCase.Checkpoints = append(currentCase.Checkpoints, trimmedLine)
		}
	}

	if currentCase != nil {
		cases = append(cases, *currentCase)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return cases, nil
}