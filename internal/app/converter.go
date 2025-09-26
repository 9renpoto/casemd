package app

import (
	"encoding/csv"
	"io"
)

// HeadingExtractor defines the behavior required to derive heading rows from Markdown.
type HeadingExtractor interface {
	ExtractHeadings(r io.Reader) ([]string, error)
}

// MarkdownToCSV orchestrates the conversion of Markdown headings into CSV rows.
type MarkdownToCSV struct {
	extractor HeadingExtractor
}

// NewMarkdownToCSV wires the converter with the provided heading extractor implementation.
func NewMarkdownToCSV(extractor HeadingExtractor) *MarkdownToCSV {
	return &MarkdownToCSV{extractor: extractor}
}

// Convert reads Markdown data and writes a CSV document where each row captures a heading.
func (c *MarkdownToCSV) Convert(input io.Reader, output io.Writer) error {
	headings, err := c.extractor.ExtractHeadings(input)
	if err != nil {
		return err
	}

	writer := csv.NewWriter(output)

	if err := writer.Write([]string{"Heading"}); err != nil {
		return err
	}

	for _, heading := range headings {
		if err := writer.Write([]string{heading}); err != nil {
			return err
		}
	}

	writer.Flush()

	if err := writer.Error(); err != nil {
		return err
	}

	return nil
}
