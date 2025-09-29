package main

import (
	"fmt"
	"io"
	"os"

	"github.com/9renpoto/casemd/internal/app"
	"github.com/9renpoto/casemd/internal/core/domain"
	"github.com/9renpoto/casemd/internal/core/parser"
	"github.com/9renpoto/casemd/internal/interfaces/cli"
)

type coreParserAdapter struct{}

func (p *coreParserAdapter) Parse(r io.Reader) ([]domain.Case, error) {
	return parser.Parse(r)
}

func main() {
	parserAdapter := &coreParserAdapter{}
	converter := app.NewMarkdownToCSV(parserAdapter)
	tool := cli.New(os.Stdout, os.Stderr, converter)
	application := app.New(tool)

	if err := application.Run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}