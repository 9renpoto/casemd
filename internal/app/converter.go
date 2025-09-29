package app

import (
	"encoding/csv"
	"io"
	"strings"

	"github.com/9renpoto/casemd/internal/core/domain"
)

// CaseParser defines the behavior required to parse test cases from Markdown.
type CaseParser interface {
	Parse(r io.Reader) ([]domain.Case, error)
}

// MarkdownToCSV orchestrates the conversion of Markdown test cases into CSV rows.
type MarkdownToCSV struct {
	parser CaseParser
}

// NewMarkdownToCSV wires the converter with the provided parser implementation.
func NewMarkdownToCSV(parser CaseParser) *MarkdownToCSV {
	return &MarkdownToCSV{parser: parser}
}

// Convert reads Markdown data and writes a CSV document.
func (c *MarkdownToCSV) Convert(input io.Reader, output io.Writer) error {
	cases, err := c.parser.Parse(input)
	if err != nil {
		return err
	}

	writer := csv.NewWriter(output)
	defer writer.Flush()

	header := []string{
		"Major Item", "Medium Item", "Minor Item",
		"Validation Steps", "Checkpoints",
		"Result", "Test Date", "Tester", "Notes",
	}
	if err := writer.Write(header); err != nil {
		return err
	}

	for _, aCase := range cases {
		record := []string{
			aCase.MajorItem,
			aCase.MediumItem,
			aCase.MinorItem,
			strings.Join(aCase.ValidationSteps, "\n"),
			strings.Join(aCase.Checkpoints, "\n"),
			"", // Result
			"", // Test Date
			"", // Tester
			"", // Notes
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	return writer.Error()
}