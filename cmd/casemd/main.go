package main

import (
	"fmt"
	"io"
	"os"

	"github.com/9renpoto/casemd/internal/app"
	"github.com/9renpoto/casemd/internal/core/parser"
	"github.com/9renpoto/casemd/internal/interfaces/cli"
)

type markdownHeadingAdapter struct{}

func (markdownHeadingAdapter) ExtractHeadings(r io.Reader) ([]string, error) {
	return parser.Headings(r)
}

func main() {
	converter := app.NewMarkdownToCSV(markdownHeadingAdapter{})
	tool := cli.New(os.Stdout, os.Stderr, converter)
	application := app.New(tool)

	if err := application.Run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
