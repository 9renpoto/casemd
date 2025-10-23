package cli

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/9renpoto/casemd/internal/app"
)

var (
	errMissingInput                = errors.New("missing required flag: --input")
	errMissingOutput               = errors.New("missing required flag: --csv-output, --spreadsheet-output, or --google-spreadsheet-title")
	errMissingCSVConverter         = errors.New("csv output requested but converter is not configured")
	errMissingSpreadsheetConverter = errors.New("spreadsheet output requested but converter is not configured")
	errMissingGoogleConverter      = errors.New("google spreadsheet requested but converter is not configured")
)

// Converter drives Markdown transformations from the CLI layer.
type Converter interface {
	Convert(sources []app.Source, output io.Writer) error
}

// GoogleSpreadsheetCreator drives Google Sheets creation from the CLI layer.
type GoogleSpreadsheetCreator interface {
	Create(ctx context.Context, title string, sources []app.Source) (string, error)
}

// Tool represents the CLI adapter that receives user input and dispatches commands.
type Tool struct {
	stdout               io.Writer
	stderr               io.Writer
	csvConverter         Converter
	spreadsheetConverter Converter
	googleConverter      GoogleSpreadsheetCreator
}

// New creates a CLI tool with the provided output streams and conversion use case.
func New(stdout, stderr io.Writer, csvConverter, spreadsheetConverter Converter, googleConverter GoogleSpreadsheetCreator) *Tool {
	return &Tool{stdout: stdout, stderr: stderr, csvConverter: csvConverter, spreadsheetConverter: spreadsheetConverter, googleConverter: googleConverter}
}

// Run parses CLI arguments, validates required options, and executes the conversion pipeline.
func (t *Tool) Run(args []string) (err error) {
	fs := flag.NewFlagSet("casemd", flag.ContinueOnError)
	fs.SetOutput(t.stderr)

	var inputPaths multiValueFlag
	var csvOutputPath string
	var spreadsheetOutputPath string
	var googleSpreadsheetTitle string

	fs.Var(&inputPaths, "input", "Path to the Markdown source file (repeat flag for multiple files)")
	fs.StringVar(&csvOutputPath, "csv-output", "", "Path to the CSV destination file")
	fs.StringVar(&spreadsheetOutputPath, "spreadsheet-output", "", "Path to the spreadsheet destination file")
	fs.StringVar(&googleSpreadsheetTitle, "google-spreadsheet-title", "", "Title for the Google Spreadsheet to create")

	fs.Usage = func() {
		fmt.Fprintf(t.stderr, "casemd converts Markdown inspection sheets into CSV files, Excel workbooks, and Google Spreadsheets.\n\n")
		fmt.Fprintf(t.stderr, "Usage:\n  casemd [flags]\n\nFlags:\n")
		fs.PrintDefaults()
	}

	if parseErr := fs.Parse(args); parseErr != nil {
		if errors.Is(parseErr, flag.ErrHelp) {
			return nil
		}
		return parseErr
	}

	if fs.NArg() > 0 {
		return fmt.Errorf("unexpected positional arguments: %v", fs.Args())
	}

	if len(inputPaths) == 0 {
		fs.Usage()
		return errMissingInput
	}

	if csvOutputPath == "" && spreadsheetOutputPath == "" && googleSpreadsheetTitle == "" {
		fs.Usage()
		return errMissingOutput
	}

	inputs, readErr := readInputFiles([]string(inputPaths))
	if readErr != nil {
		return readErr
	}

	if csvOutputPath != "" {
		if t.csvConverter == nil {
			return errMissingCSVConverter
		}
		if err := ensureParentDirectory(csvOutputPath); err != nil {
			return err
		}

		csvFile, createErr := os.Create(csvOutputPath)
		if createErr != nil {
			return fmt.Errorf("create CSV output file: %w", createErr)
		}
		if convertErr := t.csvConverter.Convert(inputs.asSources(), csvFile); convertErr != nil {
			if closeErr := csvFile.Close(); closeErr != nil {
				return fmt.Errorf("close CSV output file: %w", closeErr)
			}
			return fmt.Errorf("convert markdown to CSV: %w", convertErr)
		}
		if closeErr := csvFile.Close(); closeErr != nil {
			return fmt.Errorf("close CSV output file: %w", closeErr)
		}
		fmt.Fprintf(t.stdout, "CSV written to %s\n", csvOutputPath)
	}

	if spreadsheetOutputPath != "" {
		if t.spreadsheetConverter == nil {
			return errMissingSpreadsheetConverter
		}
		if err := ensureParentDirectory(spreadsheetOutputPath); err != nil {
			return err
		}

		spreadsheetFile, createErr := os.Create(spreadsheetOutputPath)
		if createErr != nil {
			return fmt.Errorf("create spreadsheet output file: %w", createErr)
		}
		if convertErr := t.spreadsheetConverter.Convert(inputs.asSources(), spreadsheetFile); convertErr != nil {
			if closeErr := spreadsheetFile.Close(); closeErr != nil {
				return fmt.Errorf("close spreadsheet output file: %w", closeErr)
			}
			return fmt.Errorf("convert markdown to spreadsheet: %w", convertErr)
		}
		if closeErr := spreadsheetFile.Close(); closeErr != nil {
			return fmt.Errorf("close spreadsheet output file: %w", closeErr)
		}
		fmt.Fprintf(t.stdout, "Spreadsheet written to %s\n", spreadsheetOutputPath)
	}

	if googleSpreadsheetTitle != "" {
		if t.googleConverter == nil {
			return errMissingGoogleConverter
		}

		id, err := t.googleConverter.Create(context.Background(), googleSpreadsheetTitle, inputs.asSources())
		if err != nil {
			return fmt.Errorf("create google spreadsheet: %w", err)
		}
		fmt.Fprintf(t.stdout, "Google Spreadsheet created with ID %s\n", id)
	}

	return nil
}

type multiValueFlag []string

func (m *multiValueFlag) String() string {
	return strings.Join(*m, ",")
}

func (m *multiValueFlag) Set(value string) error {
	if value == "" {
		return errors.New("input path cannot be empty")
	}
	*m = append(*m, value)
	return nil
}

type inputCollection []inputFile

type inputFile struct {
	name string
	data []byte
}

func readInputFiles(paths []string) (inputCollection, error) {
	inputs := make([]inputFile, 0, len(paths))
	for _, path := range paths {
		content, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("open input file %s: %w", path, err)
		}
		inputs = append(inputs, inputFile{name: path, data: content})
	}
	return inputCollection(inputs), nil
}

func (c inputCollection) asSources() []app.Source {
	sources := make([]app.Source, 0, len(c))
	for _, input := range c {
		sources = append(sources, app.Source{Name: input.name, Reader: bytes.NewReader(input.data)})
	}
	return sources
}

func ensureParentDirectory(path string) error {
	dir := filepath.Dir(path)
	if dir == "." || dir == "" {
		return nil
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}
	return nil
}
