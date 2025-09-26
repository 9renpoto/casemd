package parser

import (
	"bufio"
	"io"
	"strings"
)

// Headings extracts top-level Markdown headings from the provided reader.
func Headings(r io.Reader) ([]string, error) {
	scanner := bufio.NewScanner(r)
	headings := make([]string, 0)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "# ") {
			headings = append(headings, strings.TrimSpace(line[2:]))
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return headings, nil
}
