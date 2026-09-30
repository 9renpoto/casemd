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
	"github.com/9renpoto/casemd/internal/core/domain"
)

var (
	errMissingInput                = errors.New("missing required flag: --input; specify --input <file.md> (repeat for multiple files)")
	errMissingOutput               = errors.New("missing output destination; specify --csv-output <file.csv>, --spreadsheet-output <file.xlsx>, or --google-spreadsheet-title <title>; to check the format only, run casemd validate --input <file.md>")
	errMissingValidator            = errors.New("validation requested but validator is not configured")
	errMissingCSVConverter         = errors.New("csv output requested but converter is not configured")
	errMissingSpreadsheetConverter = errors.New("spreadsheet output requested but converter is not configured")
	errMissingGoogleConverter      = errors.New("Google Spreadsheet creation is unavailable; set GOOGLE_SHEETS_ACCESS_TOKEN to an OAuth access token with the https://www.googleapis.com/auth/spreadsheets scope and retry; check any initialization warning on standard error")
)

// Converter drives Markdown transformations from the CLI layer.
type Converter interface {
	Convert(sources []app.Source, output io.Writer) error
}

// Validator checks Markdown sources without producing artifacts.
type Validator interface {
	Validate(sources []app.Source) ([]domain.Diagnostic, error)
}

// GoogleSpreadsheetCreator drives Google Sheets creation from the CLI layer.
type GoogleSpreadsheetCreator interface {
	Create(ctx context.Context, title string, sources []app.Source) (string, error)
}

// Tool represents the CLI adapter that receives user input and dispatches commands.
type Tool struct {
	stdout               io.Writer
	stderr               io.Writer
	validator            Validator
	csvConverter         Converter
	spreadsheetConverter Converter
	googleConverter      GoogleSpreadsheetCreator
}

// New creates a CLI tool with the provided output streams and conversion use case.
func New(stdout, stderr io.Writer, validator Validator, csvConverter, spreadsheetConverter Converter, googleConverter GoogleSpreadsheetCreator) *Tool {
	return &Tool{stdout: stdout, stderr: stderr, validator: validator, csvConverter: csvConverter, spreadsheetConverter: spreadsheetConverter, googleConverter: googleConverter}
}

// Run parses CLI arguments, validates required options, and executes the conversion pipeline.
func (t *Tool) Run(args []string) error {
	if len(args) > 0 && args[0] == "validate" {
		return t.runValidate(args[1:])
	}
	return t.runConvert(args)
}

func (t *Tool) runValidate(args []string) error {
	fs := flag.NewFlagSet("casemd validate", flag.ContinueOnError)
	fs.SetOutput(t.stderr)

	var inputPaths multiValueFlag
	fs.Var(&inputPaths, "input", "Path to the Markdown source file (repeat flag for multiple files)")
	fs.Usage = func() {
		fmt.Fprintf(t.stderr, "Validate Markdown inspection sheets without generating output artifacts.\n\n")
		fmt.Fprintf(t.stderr, "Usage:\n  casemd validate --input <file> [--input <file> ...]\n\nFlags:\n")
		fs.PrintDefaults()
		printValidationHelp(t.stderr)
	}

	help, err := parseFlagSet(fs, args)
	if err != nil {
		return err
	}
	if help {
		return nil
	}
	if len(inputPaths) == 0 {
		fs.Usage()
		return errMissingInput
	}

	inputs, err := readInputFiles([]string(inputPaths))
	if err != nil {
		return err
	}
	if err := t.validateInputs(inputs); err != nil {
		return err
	}
	_, err = fmt.Fprintf(t.stdout, "Validation passed: %d input file(s) conform to the casemd v1 format.\n", len(inputs))
	return err
}

func (t *Tool) runConvert(args []string) (err error) {
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
		fmt.Fprintf(t.stderr, "Usage:\n  casemd --input <file> [--input <file> ...] <output flags>\n  casemd validate --input <file> [--input <file> ...]\n  casemd serve [address]\n  casemd --version\n\nFlags:\n")
		fs.PrintDefaults()
		fmt.Fprint(t.stderr, "\nConversion:\n  Specify at least one output flag; multiple output formats may be requested together.\n  Each input becomes a separate XLSX worksheet. Existing output files are overwritten.\n  Google Sheets requires GOOGLE_SHEETS_ACCESS_TOKEN with the spreadsheets OAuth scope.\n\nExamples:\n  casemd --input notes.md --csv-output build/notes.csv\n  casemd --input notes.md --spreadsheet-output build/notes.xlsx\n  casemd validate --input notes.md\n\nWeb UI:\n  casemd serve defaults to :3000; set CASEMD_WEB_ADDR or pass an address.\n")
		printValidationHelp(t.stderr)
	}

	help, parseErr := parseFlagSet(fs, args)
	if parseErr != nil {
		return parseErr
	}
	if help {
		return nil
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
	if validationErr := t.validateInputs(inputs); validationErr != nil {
		return validationErr
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

func parseFlagSet(fs *flag.FlagSet, args []string) (bool, error) {
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return true, nil
		}
		return false, err
	}
	if fs.NArg() > 0 {
		return false, fmt.Errorf("unexpected positional arguments: %v; pass Markdown paths with --input <file.md>; run %s --help for usage", fs.Args(), fs.Name())
	}
	return false, nil
}

func printValidationHelp(output io.Writer) {
	fmt.Fprint(output, "\nInput format (casemd v1):\n  Optional # document title, once before the hierarchy.\n  ## major item > ### medium item > #### test case; heading titles must be nonempty.\n  Each major item needs a medium item; each medium item needs a test case.\n  Each test case needs an ordered step (1. Action) and a checkpoint (* [ ] Expected result).\n  Heading levels 5 and 6 are unsupported.\n\nValidation:\n  casemd validate --input notes.md --input follow-up.md\n  Success is reported on standard output. Diagnostics go to standard error as\n  file:line: rule: message, followed by a suggested correction when available.\n  Fix all diagnostics, then run validation again. Validation generates no artifacts.\n  Conversion checks the same rules before creating output files or Google Sheets.\n\nExit codes:\n  0  Command succeeded (including help and version).\n  1  Invalid input, invalid arguments, or an execution error.\n")
}

func (t *Tool) validateInputs(inputs inputCollection) error {
	if t.validator == nil {
		return errMissingValidator
	}
	diagnostics, err := t.validator.Validate(inputs.asSources())
	if err != nil {
		return err
	}
	if len(diagnostics) == 0 {
		return nil
	}

	for _, diagnostic := range diagnostics {
		fmt.Fprintf(t.stderr, "%s:%d: %s: %s\n", diagnostic.Source, diagnostic.Line, diagnostic.Rule, diagnostic.Message)
		if diagnostic.Suggestion != "" {
			fmt.Fprintf(t.stderr, "  suggestion: %s\n", diagnostic.Suggestion)
		}
	}
	return &app.ValidationError{Diagnostics: diagnostics}
}

type multiValueFlag []string

func (m *multiValueFlag) String() string {
	return strings.Join(*m, ",")
}

func (m *multiValueFlag) Set(value string) error {
	if value == "" {
		return errors.New("input path cannot be empty; specify --input <file.md>")
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
