package app

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/csv"
	"encoding/xml"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/9renpoto/casemd/internal/core/domain"
)

// CaseParser defines the behavior required to parse test cases from Markdown.
type CaseParser interface {
	Parse(r io.Reader) ([]domain.Case, error)
}

// Source represents a Markdown document and the metadata needed to build a sheet.
type Source struct {
	Name   string
	Reader io.Reader
}

// GoogleSpreadsheetCreator defines the behavior required to create Google Spreadsheets.
type GoogleSpreadsheetCreator interface {
	CreateSpreadsheet(ctx context.Context, spreadsheet GoogleSpreadsheet) (string, error)
}

// GoogleSpreadsheet represents a Google Sheets spreadsheet to create through the API.
type GoogleSpreadsheet struct {
	Title  string
	Sheets []GoogleSpreadsheetSheet
}

// GoogleSpreadsheetSheet describes a single Google Sheets worksheet and its data.
type GoogleSpreadsheetSheet struct {
	Title string
	Rows  [][]string
}

var spreadsheetHeaders = []string{
	"Major Item", "Medium Item", "Minor Item",
	"Validation Steps", "Checkpoints",
	"Result", "Test Date", "Tester", "Notes",
}

func caseRow(aCase domain.Case) []string {
	return []string{
		aCase.MajorItem,
		aCase.MediumItem,
		aCase.MinorItem,
		strings.Join(aCase.ValidationSteps, "\n"),
		strings.Join(aCase.Checkpoints, "\n"),
		"", // Result
		"", // Test Date
		"", // Tester
		"", // Notes
	}
}

// MarkdownToCSV orchestrates the conversion of Markdown test cases into CSV rows.
type MarkdownToCSV struct {
	parser CaseParser
}

// NewMarkdownToCSV wires the converter with the provided parser implementation.
func NewMarkdownToCSV(parser CaseParser) *MarkdownToCSV {
	return &MarkdownToCSV{parser: parser}
}

// Convert reads Markdown sources and writes a CSV document containing every parsed case.
func (c *MarkdownToCSV) Convert(sources []Source, output io.Writer) error {
	if len(sources) == 0 {
		return fmt.Errorf("no sources provided")
	}

	writer := csv.NewWriter(output)
	if err := writer.Write(spreadsheetHeaders); err != nil {
		return fmt.Errorf("write csv header: %w", err)
	}

	for _, source := range sources {
		cases, err := c.parser.Parse(source.Reader)
		if err != nil {
			writer.Flush()
			return fmt.Errorf("parse %s: %w", source.Name, err)
		}

		for _, aCase := range cases {
			if err := writer.Write(caseRow(aCase)); err != nil {
				writer.Flush()
				return fmt.Errorf("write csv row: %w", err)
			}
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return fmt.Errorf("flush csv: %w", err)
	}

	return nil
}

// MarkdownToSpreadsheet orchestrates the conversion of Markdown test cases into spreadsheet sheets.
type MarkdownToSpreadsheet struct {
	parser CaseParser
}

// NewMarkdownToSpreadsheet wires the converter with the provided parser implementation.
func NewMarkdownToSpreadsheet(parser CaseParser) *MarkdownToSpreadsheet {
	return &MarkdownToSpreadsheet{parser: parser}
}

// Convert reads Markdown data and writes a spreadsheet workbook with one sheet per Markdown file.
func (c *MarkdownToSpreadsheet) Convert(sources []Source, output io.Writer) error {
	if len(sources) == 0 {
		return fmt.Errorf("no sources provided")
	}

	sheets := make([]workbookSheet, 0, len(sources))
	nameUsage := make(map[string]int)
	finalNames := make(map[string]struct{})

	for index, source := range sources {
		sheetBase := deriveSheetName(source.Name, index)
		sheetName := ensureUniqueSheetName(sheetBase, nameUsage, finalNames)

		cases, err := c.parser.Parse(source.Reader)
		if err != nil {
			return fmt.Errorf("parse %s: %w", sheetName, err)
		}

		rows := make([][]string, 0, len(cases)+1)
		rows = append(rows, append([]string(nil), spreadsheetHeaders...))

		for _, aCase := range cases {
			rows = append(rows, caseRow(aCase))
		}

		sheets = append(sheets, workbookSheet{Name: sheetName, Rows: rows})
	}

	return writeWorkbook(output, sheets)
}

// MarkdownToGoogleSpreadsheet orchestrates the conversion of Markdown cases into Google Sheets.
type MarkdownToGoogleSpreadsheet struct {
	parser  CaseParser
	creator GoogleSpreadsheetCreator
}

// NewMarkdownToGoogleSpreadsheet wires the Google Sheets converter with the provided dependencies.
func NewMarkdownToGoogleSpreadsheet(parser CaseParser, creator GoogleSpreadsheetCreator) *MarkdownToGoogleSpreadsheet {
	return &MarkdownToGoogleSpreadsheet{parser: parser, creator: creator}
}

// Create parses sources and creates a Google Spreadsheet using the configured creator.
func (c *MarkdownToGoogleSpreadsheet) Create(ctx context.Context, title string, sources []Source) (string, error) {
	if title == "" {
		return "", fmt.Errorf("spreadsheet title cannot be empty")
	}
	if len(sources) == 0 {
		return "", fmt.Errorf("no sources provided")
	}

	sheets := make([]GoogleSpreadsheetSheet, 0, len(sources))
	nameUsage := make(map[string]int)
	finalNames := make(map[string]struct{})

	for index, source := range sources {
		sheetBase := deriveSheetName(source.Name, index)
		sheetName := ensureUniqueSheetName(sheetBase, nameUsage, finalNames)

		cases, err := c.parser.Parse(source.Reader)
		if err != nil {
			return "", fmt.Errorf("parse %s: %w", sheetName, err)
		}

		rows := make([][]string, 0, len(cases)+1)
		rows = append(rows, append([]string(nil), spreadsheetHeaders...))

		for _, aCase := range cases {
			rows = append(rows, caseRow(aCase))
		}

		sheets = append(sheets, GoogleSpreadsheetSheet{Title: sheetName, Rows: rows})
	}

	spreadsheet := GoogleSpreadsheet{Title: title, Sheets: sheets}
	spreadsheetID, err := c.creator.CreateSpreadsheet(ctx, spreadsheet)
	if err != nil {
		return "", fmt.Errorf("create google spreadsheet: %w", err)
	}

	return spreadsheetID, nil
}

type workbookSheet struct {
	Name string
	Rows [][]string
}

func writeWorkbook(w io.Writer, sheets []workbookSheet) error {
	zipWriter := zip.NewWriter(w)

	if err := writeZipFile(zipWriter, "[Content_Types].xml", buildContentTypes(sheets)); err != nil {
		zipWriter.Close()
		return err
	}

	if err := writeZipFile(zipWriter, "_rels/.rels", rootRelationships); err != nil {
		zipWriter.Close()
		return err
	}

	if err := writeZipFile(zipWriter, "docProps/app.xml", appProperties); err != nil {
		zipWriter.Close()
		return err
	}

	if err := writeZipFile(zipWriter, "docProps/core.xml", coreProperties); err != nil {
		zipWriter.Close()
		return err
	}

	if err := writeZipFile(zipWriter, "xl/workbook.xml", buildWorkbookXML(sheets)); err != nil {
		zipWriter.Close()
		return err
	}

	if err := writeZipFile(zipWriter, "xl/_rels/workbook.xml.rels", buildWorkbookRelationships(sheets)); err != nil {
		zipWriter.Close()
		return err
	}

	for i, sheet := range sheets {
		path := fmt.Sprintf("xl/worksheets/sheet%d.xml", i+1)
		if err := writeZipFile(zipWriter, path, buildWorksheetXML(sheet.Rows)); err != nil {
			zipWriter.Close()
			return err
		}
	}

	return zipWriter.Close()
}

func writeZipFile(zipWriter *zip.Writer, name, content string) error {
	writer, err := zipWriter.Create(name)
	if err != nil {
		return fmt.Errorf("create %s: %w", name, err)
	}
	if _, err := writer.Write([]byte(content)); err != nil {
		return fmt.Errorf("write %s: %w", name, err)
	}
	return nil
}

func buildContentTypes(sheets []workbookSheet) string {
	var builder strings.Builder
	builder.WriteString(xml.Header)
	builder.WriteString(`<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">`)
	builder.WriteString(`<Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>`)
	builder.WriteString(`<Default Extension="xml" ContentType="application/xml"/>`)
	builder.WriteString(`<Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/>`)
	for i := range sheets {
		builder.WriteString(fmt.Sprintf(`<Override PartName="/xl/worksheets/sheet%d.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/>`, i+1))
	}
	builder.WriteString(`</Types>`)
	return builder.String()
}

func buildWorkbookXML(sheets []workbookSheet) string {
	var builder strings.Builder
	builder.WriteString(xml.Header)
	builder.WriteString(`<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">`)
	builder.WriteString(`<sheets>`)
	for i, sheet := range sheets {
		builder.WriteString(fmt.Sprintf(`<sheet name="%s" sheetId="%d" r:id="rId%d"/>`, xmlEscapeAttr(sheet.Name), i+1, i+1))
	}
	builder.WriteString(`</sheets></workbook>`)
	return builder.String()
}

func buildWorkbookRelationships(sheets []workbookSheet) string {
	var builder strings.Builder
	builder.WriteString(xml.Header)
	builder.WriteString(`<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">`)
	for i := range sheets {
		builder.WriteString(fmt.Sprintf(`<Relationship Id="rId%d" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet%d.xml"/>`, i+1, i+1))
	}
	builder.WriteString(`</Relationships>`)
	return builder.String()
}

func buildWorksheetXML(rows [][]string) string {
	var builder strings.Builder
	builder.WriteString(xml.Header)
	builder.WriteString(`<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">`)

	if len(rows) > 0 && len(rows[0]) > 0 {
		lastCol := columnName(len(rows[0]))
		builder.WriteString(fmt.Sprintf(`<dimension ref="A1:%s%d"/>`, lastCol, len(rows)))
	}

	builder.WriteString(`<sheetData>`)
	for i, row := range rows {
		rowIndex := i + 1
		builder.WriteString(fmt.Sprintf(`<row r="%d">`, rowIndex))
		for j, value := range row {
			cellRef := fmt.Sprintf("%s%d", columnName(j+1), rowIndex)
			if value == "" {
				builder.WriteString(fmt.Sprintf(`<c r="%s"/>`, cellRef))
				continue
			}
			builder.WriteString(fmt.Sprintf(`<c r="%s" t="inlineStr"><is><t>%s</t></is></c>`, cellRef, escapeCellText(value)))
		}
		builder.WriteString(`</row>`)
	}
	builder.WriteString(`</sheetData></worksheet>`)
	return builder.String()
}

func escapeCellText(value string) string {
	var buf bytes.Buffer
	if err := xml.EscapeText(&buf, []byte(value)); err != nil {
		return ""
	}
	escaped := buf.String()
	return strings.ReplaceAll(escaped, "\n", "&#10;")
}

func xmlEscapeAttr(value string) string {
	var buf bytes.Buffer
	if err := xml.EscapeText(&buf, []byte(value)); err != nil {
		return ""
	}
	return buf.String()
}

func columnName(n int) string {
	const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	if n <= 0 {
		return "A"
	}
	result := ""
	for n > 0 {
		n--
		result = string(letters[n%26]) + result
		n /= 26
	}
	return result
}

func deriveSheetName(name string, index int) string {
	if name == "" {
		return fmt.Sprintf("Sheet%d", index+1)
	}

	base := filepath.Base(name)
	if ext := filepath.Ext(base); ext != "" {
		base = strings.TrimSuffix(base, ext)
	}

	sanitized := sanitizeSheetName(base)
	if sanitized == "" {
		return fmt.Sprintf("Sheet%d", index+1)
	}

	return sanitized
}

func ensureUniqueSheetName(base string, usage map[string]int, final map[string]struct{}) string {
	name := base
	count := usage[base]

	if _, exists := final[name]; !exists {
		usage[base] = max(count, 1)
		final[name] = struct{}{}
		return name
	}

	for {
		count++
		candidate := buildSheetNameWithSuffix(base, count)
		if _, exists := final[candidate]; exists {
			continue
		}
		usage[base] = count
		final[candidate] = struct{}{}
		return candidate
	}
}

func buildSheetNameWithSuffix(base string, count int) string {
	suffix := fmt.Sprintf("_%d", count)
	maxBaseRunes := 31 - utf8.RuneCountInString(suffix)
	if maxBaseRunes < 1 {
		maxBaseRunes = 1
	}
	trimmedBase := trimRunes(base, maxBaseRunes)
	return trimmedBase + suffix
}

func trimRunes(value string, limit int) string {
	if utf8.RuneCountInString(value) <= limit {
		return value
	}
	var builder strings.Builder
	builder.Grow(limit)
	count := 0
	for _, r := range value {
		builder.WriteRune(r)
		count++
		if count == limit {
			break
		}
	}
	return builder.String()
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func sanitizeSheetName(name string) string {
	cleaned := invalidSheetNameChars.Replace(name)
	cleaned = strings.Trim(cleaned, " ")
	if cleaned == "" {
		return ""
	}

	if utf8.RuneCountInString(cleaned) <= 31 {
		return cleaned
	}

	var builder strings.Builder
	builder.Grow(31)
	count := 0
	for _, r := range cleaned {
		builder.WriteRune(r)
		count++
		if count == 31 {
			break
		}
	}
	return builder.String()
}

var invalidSheetNameChars = strings.NewReplacer(
	"*", "_",
	":", "_",
	"?", "_",
	"[", "_",
	"]", "_",
	"/", "_",
	"\\", "_",
)

const rootRelationships = `<?xml version="1.0" encoding="UTF-8"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="xl/workbook.xml"/>
  <Relationship Id="rId2" Type="http://schemas.openxmlformats.org/package/2006/relationships/metadata/core-properties" Target="docProps/core.xml"/>
  <Relationship Id="rId3" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/extended-properties" Target="docProps/app.xml"/>
</Relationships>`

const appProperties = `<?xml version="1.0" encoding="UTF-8"?>
<Properties xmlns="http://schemas.openxmlformats.org/officeDocument/2006/extended-properties" xmlns:vt="http://schemas.openxmlformats.org/officeDocument/2006/docPropsVTypes">
  <Application>casemd</Application>
</Properties>`

const coreProperties = `<?xml version="1.0" encoding="UTF-8"?>
<cp:coreProperties xmlns:cp="http://schemas.openxmlformats.org/package/2006/metadata/core-properties" xmlns:dc="http://purl.org/dc/elements/1.1/" xmlns:dcterms="http://purl.org/dc/terms/" xmlns:dcmitype="http://purl.org/dc/dcmitype/" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
  <dc:creator>casemd</dc:creator>
  <cp:lastModifiedBy>casemd</cp:lastModifiedBy>
</cp:coreProperties>`
