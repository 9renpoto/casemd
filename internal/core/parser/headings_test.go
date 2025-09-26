package parser

import (
	"strings"
	"testing"
)

func TestHeadings(t *testing.T) {
	t.Run("extracts top level headings", func(t *testing.T) {
		input := "# Title\n\nParagraph text.\n\n## Subtitle\n# Another Title\n"
		headings, err := Headings(strings.NewReader(input))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		want := []string{"Title", "Another Title"}
		if len(headings) != len(want) {
			t.Fatalf("expected %d headings, got %d", len(want), len(headings))
		}

		for i := range want {
			if headings[i] != want[i] {
				t.Fatalf("expected headings[%d] = %q, got %q", i, want[i], headings[i])
			}
		}
	})

	t.Run("returns empty slice when no headings found", func(t *testing.T) {
		headings, err := Headings(strings.NewReader("No headings here"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(headings) != 0 {
			t.Fatalf("expected no headings, got %v", headings)
		}
	})
}
