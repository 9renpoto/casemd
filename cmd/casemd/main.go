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
	"github.com/9renpoto/casemd/internal/interfaces/web"
)

var version = "dev"

type coreParserAdapter struct{}

func (p *coreParserAdapter) Parse(r io.Reader) ([]domain.Case, error) {
	return parser.Parse(r)
}

func (p *coreParserAdapter) ParseWithDiagnostics(source string, r io.Reader) ([]domain.Case, []domain.Diagnostic, error) {
	return parser.ParseWithDiagnostics(source, r)
}

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string, stdout, stderr io.Writer) error {
	if len(args) == 1 && (args[0] == "--version" || args[0] == "version") {
		_, err := fmt.Fprintln(stdout, version)
		return err
	}

	parserAdapter := &coreParserAdapter{}
	validator := app.NewMarkdownValidator(parserAdapter)
	csvConverter := app.NewMarkdownToCSV(parserAdapter)
	spreadsheetConverter := app.NewMarkdownToSpreadsheet(parserAdapter)
	var googleConverter cli.GoogleSpreadsheetCreator

	if token := os.Getenv("GOOGLE_SHEETS_ACCESS_TOKEN"); token != "" {
		if sheetsService, err := googleapi.NewSheetsService(nil, token); err != nil {
			fmt.Fprintf(stderr, "warning: google sheets support disabled: %v\n", err)
		} else {
			googleConverter = app.NewMarkdownToGoogleSpreadsheet(parserAdapter, sheetsService)
		}
	}

	if len(args) > 0 && args[0] == "serve" {
		addr := os.Getenv("CASEMD_WEB_ADDR")
		if len(args) > 1 {
			addr = args[1]
		}
		if addr == "" {
			addr = ":3000"
		}

		server := web.NewServer(csvConverter)
		fmt.Fprintf(stdout, "Starting casemd web UI on %s\n", addr)
		return server.Listen(addr)
	}

	tool := cli.New(stdout, stderr, validator, csvConverter, spreadsheetConverter, googleConverter)
	application := app.New(tool)

	return application.Run(args)
}
