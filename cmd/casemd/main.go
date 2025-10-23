package main

import (
	"fmt"
	"io"
	"os"

	"github.com/9renpoto/casemd/internal/app"
	"github.com/9renpoto/casemd/internal/core/domain"
	"github.com/9renpoto/casemd/internal/core/parser"
	"github.com/9renpoto/casemd/internal/interfaces/cli"
	"github.com/9renpoto/casemd/internal/interfaces/googleapi"
)

type coreParserAdapter struct{}

func (p *coreParserAdapter) Parse(r io.Reader) ([]domain.Case, error) {
	return parser.Parse(r)
}

func main() {
	parserAdapter := &coreParserAdapter{}
	csvConverter := app.NewMarkdownToCSV(parserAdapter)
	spreadsheetConverter := app.NewMarkdownToSpreadsheet(parserAdapter)
	var googleConverter cli.GoogleSpreadsheetCreator

	if token := os.Getenv("GOOGLE_SHEETS_ACCESS_TOKEN"); token != "" {
		if sheetsService, err := googleapi.NewSheetsService(nil, token); err != nil {
			fmt.Fprintf(os.Stderr, "warning: google sheets support disabled: %v\n", err)
		} else {
			googleConverter = app.NewMarkdownToGoogleSpreadsheet(parserAdapter, sheetsService)
		}
	}

	tool := cli.New(os.Stdout, os.Stderr, csvConverter, spreadsheetConverter, googleConverter)
	application := app.New(tool)

	if err := application.Run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
