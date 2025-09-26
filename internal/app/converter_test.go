package app

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"
)

type stubExtractor struct {
	headings []string
	err      error
}

func (s stubExtractor) ExtractHeadings(r io.Reader) ([]string, error) {
	return s.headings, s.err
}

func TestMarkdownToCSV(t *testing.T) {
	t.Run("writes headings to csv", func(t *testing.T) {
		extractor := stubExtractor{headings: []string{"Title", "Another"}}
		converter := NewMarkdownToCSV(extractor)

		input := strings.NewReader("irrelevant")
		var output bytes.Buffer

		if err := converter.Convert(input, &output); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		got := output.String()
		want := "Heading\nTitle\nAnother\n"
		if got != want {
			t.Fatalf("expected output %q, got %q", want, got)
		}
	})

	t.Run("returns extractor error", func(t *testing.T) {
		extractor := stubExtractor{err: errors.New("extract failed")}
		converter := NewMarkdownToCSV(extractor)

		err := converter.Convert(strings.NewReader("input"), &bytes.Buffer{})
		if !errors.Is(err, extractor.err) {
			t.Fatalf("expected extractor error, got %v", err)
		}
	})
}
