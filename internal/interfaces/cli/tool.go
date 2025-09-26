package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

var (
	errMissingInput  = errors.New("missing required flag: --input")
	errMissingOutput = errors.New("missing required flag: --output")
)

// Converter drives the Markdown to CSV transformation from the CLI layer.
type Converter interface {
	Convert(input io.Reader, output io.Writer) error
}

// Tool represents the CLI adapter that receives user input and dispatches commands.
type Tool struct {
	stdout    io.Writer
	stderr    io.Writer
	converter Converter
}

// New creates a CLI tool with the provided output streams and conversion use case.
func New(stdout, stderr io.Writer, converter Converter) *Tool {
	return &Tool{stdout: stdout, stderr: stderr, converter: converter}
}

// Run parses CLI arguments, validates required options, and executes the conversion pipeline.
func (t *Tool) Run(args []string) (err error) {
	fs := flag.NewFlagSet("casemd", flag.ContinueOnError)
	fs.SetOutput(t.stderr)

	var inputPath string
	var outputPath string

	fs.StringVar(&inputPath, "input", "", "Path to the Markdown source file")
	fs.StringVar(&outputPath, "output", "", "Path to the CSV destination file")

	fs.Usage = func() {
		fmt.Fprintf(t.stderr, "casemd converts Markdown headings into CSV rows suitable for spreadsheet ingestion.\n\n")
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

	if inputPath == "" {
		fs.Usage()
		return errMissingInput
	}

	if outputPath == "" {
		fs.Usage()
		return errMissingOutput
	}

	inputFile, openErr := os.Open(inputPath)
	if openErr != nil {
		return fmt.Errorf("open input file: %w", openErr)
	}
	defer inputFile.Close()

	if err := ensureParentDirectory(outputPath); err != nil {
		return err
	}

	outputFile, createErr := os.Create(outputPath)
	if createErr != nil {
		return fmt.Errorf("create output file: %w", createErr)
	}
	defer func() {
		closeErr := outputFile.Close()
		if err == nil && closeErr != nil {
			err = fmt.Errorf("close output file: %w", closeErr)
		}
	}()

	if err := t.converter.Convert(inputFile, outputFile); err != nil {
		return fmt.Errorf("convert markdown: %w", err)
	}

	fmt.Fprintf(t.stdout, "CSV written to %s\n", outputPath)
	return nil
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
