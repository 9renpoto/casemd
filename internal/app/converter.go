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
	parser MarkdownParser
}

// NewMarkdownToCSV wires the converter with the provided parser implementation.
func NewMarkdownToCSV(parser MarkdownParser) *MarkdownToCSV {
	return &MarkdownToCSV{parser: parser}
}

// Convert reads Markdown sources and writes a CSV document containing every parsed case.
func (c *MarkdownToCSV) Convert(sources []Source, output io.Writer) error {
	parsedSources, err := requireValidSources(c.parser, sources)
	if err != nil {
		return err
	}

	writer := csv.NewWriter(output)
	if err := writer.Write(spreadsheetHeaders); err != nil {
		return fmt.Errorf("write csv header: %w", err)
	}

	for _, source := range parsedSources {
		for _, aCase := range source.Cases {
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
	parser MarkdownParser
}

// NewMarkdownToSpreadsheet wires the converter with the provided parser implementation.
func NewMarkdownToSpreadsheet(parser MarkdownParser) *MarkdownToSpreadsheet {
	return &MarkdownToSpreadsheet{parser: parser}
}

// Convert reads Markdown data and writes a spreadsheet workbook with one sheet per Markdown file.
func (c *MarkdownToSpreadsheet) Convert(sources []Source, output io.Writer) error {
	parsedSources, err := requireValidSources(c.parser, sources)
	if err != nil {
		return err
	}

	sheets := make([]workbookSheet, 0, len(parsedSources))
	nameUsage := make(map[string]int)
	finalNames := make(map[string]struct{})

	for index, source := range parsedSources {
		sheetBase := deriveSheetName(source.Name, index)
		sheetName := ensureUniqueSheetName(sheetBase, nameUsage, finalNames)

		rows := make([][]string, 0, len(source.Cases)+1)
		rows = append(rows, append([]string(nil), spreadsheetHeaders...))

		for _, aCase := range source.Cases {
			rows = append(rows, caseRow(aCase))
		}

		sheets = append(sheets, workbookSheet{Name: sheetName, Rows: rows})
	}

	return writeWorkbook(output, sheets)
}

// MarkdownToGoogleSpreadsheet orchestrates the conversion of Markdown cases into Google Sheets.
type MarkdownToGoogleSpreadsheet struct {
	parser  MarkdownParser
	creator GoogleSpreadsheetCreator
}

// NewMarkdownToGoogleSpreadsheet wires the Google Sheets converter with the provided dependencies.
func NewMarkdownToGoogleSpreadsheet(parser MarkdownParser, creator GoogleSpreadsheetCreator) *MarkdownToGoogleSpreadsheet {
	return &MarkdownToGoogleSpreadsheet{parser: parser, creator: creator}
}

// Create parses sources and creates a Google Spreadsheet using the configured creator.
func (c *MarkdownToGoogleSpreadsheet) Create(ctx context.Context, title string, sources []Source) (string, error) {
	if title == "" {
		return "", fmt.Errorf("spreadsheet title cannot be empty")
	}
	parsedSources, err := requireValidSources(c.parser, sources)
	if err != nil {
		return "", err
	}

	sheets := make([]GoogleSpreadsheetSheet, 0, len(parsedSources))
	nameUsage := make(map[string]int)
	finalNames := make(map[string]struct{})

	for index, source := range parsedSources {
		sheetBase := deriveSheetName(source.Name, index)
		sheetName := ensureUniqueSheetName(sheetBase, nameUsage, finalNames)

		rows := make([][]string, 0, len(source.Cases)+1)
		rows = append(rows, append([]string(nil), spreadsheetHeaders...))

		for _, aCase := range source.Cases {
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

	if err := writeZipFile(zipWriter, "xl/styles.xml", workbookStyles); err != nil {
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
	builder.WriteString(`<Override PartName="/docProps/core.xml" ContentType="application/vnd.openxmlformats-package.core-properties+xml"/>`)
	builder.WriteString(`<Override PartName="/docProps/app.xml" ContentType="application/vnd.openxmlformats-officedocument.extended-properties+xml"/>`)
	builder.WriteString(`<Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/>`)
	for i := range sheets {
		builder.WriteString(fmt.Sprintf(`<Override PartName="/xl/worksheets/sheet%d.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/>`, i+1))
	}
	builder.WriteString(`<Override PartName="/xl/styles.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.styles+xml"/>`)
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
	builder.WriteString(fmt.Sprintf(`<Relationship Id="rId%d" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles" Target="styles.xml"/>`, len(sheets)+1))
	builder.WriteString(`</Relationships>`)
	return builder.String()
}

func buildWorksheetXML(rows [][]string) string {
	var builder strings.Builder
	builder.WriteString(xml.Header)
	builder.WriteString(`<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">`)

	columnCount := maxColumnCount(rows)
	if len(rows) > 0 && columnCount > 0 {
		lastCol := columnName(columnCount)
		builder.WriteString(fmt.Sprintf(`<dimension ref="A1:%s%d"/>`, lastCol, len(rows)))
	}

	builder.WriteString(`<sheetViews><sheetView workbookViewId="0"><pane ySplit="1" topLeftCell="A2" activePane="bottomLeft" state="frozen"/><selection pane="bottomLeft" activeCell="A2" sqref="A2"/></sheetView></sheetViews>`)
	builder.WriteString(`<sheetFormatPr defaultRowHeight="18"/>`)
	builder.WriteString(`<cols>`)
	for index, width := range spreadsheetColumnWidths {
		columnIndex := index + 1
		builder.WriteString(fmt.Sprintf(`<col min="%d" max="%d" width="%.1f" customWidth="1"/>`, columnIndex, columnIndex, width))
	}
	builder.WriteString(`</cols>`)

	builder.WriteString(`<sheetData>`)
	for i, row := range rows {
		rowIndex := i + 1
		rowHeight := spreadsheetRowHeight(rowIndex, row)
		builder.WriteString(fmt.Sprintf(`<row r="%d" ht="%.1f" customHeight="1">`, rowIndex, rowHeight))
		for j, value := range row {
			cellRef := fmt.Sprintf("%s%d", columnName(j+1), rowIndex)
			styleIndex := spreadsheetStyleIndex(rowIndex, j+1)
			if value == "" {
				builder.WriteString(fmt.Sprintf(`<c r="%s" s="%d"/>`, cellRef, styleIndex))
				continue
			}
			builder.WriteString(fmt.Sprintf(`<c r="%s" s="%d" t="inlineStr"><is><t>%s</t></is></c>`, cellRef, styleIndex, escapeCellText(value)))
		}
		builder.WriteString(`</row>`)
	}
	builder.WriteString(`</sheetData>`)
	if len(rows) > 0 && columnCount > 0 {
		builder.WriteString(fmt.Sprintf(`<autoFilter ref="A1:%s%d"/>`, columnName(columnCount), len(rows)))
	}
	builder.WriteString(`</worksheet>`)
	return builder.String()
}

var spreadsheetColumnWidths = []float64{18, 18, 24, 42, 42, 14, 13, 16, 32}

func maxColumnCount(rows [][]string) int {
	count := 0
	for _, row := range rows {
		if len(row) > count {
			count = len(row)
		}
	}
	return count
}

func spreadsheetStyleIndex(row, column int) int {
	if row == 1 {
		return 1
	}
	if column <= 5 {
		return 2
	}
	if column == 7 {
		return 4
	}
	return 3
}

func spreadsheetRowHeight(rowIndex int, row []string) float64 {
	if rowIndex == 1 {
		return 30
	}

	maxLines := 1
	for index, value := range row {
		width := 18
		if index < len(spreadsheetColumnWidths) {
			width = int(spreadsheetColumnWidths[index])
		}

		lines := 0
		for _, segment := range strings.Split(value, "\n") {
			lineCount := (utf8.RuneCountInString(segment) + width - 1) / width
			if lineCount < 1 {
				lineCount = 1
			}
			lines += lineCount
		}
		if lines > maxLines {
			maxLines = lines
		}
	}

	height := float64(maxLines*15 + 6)
	if height < 36 {
		return 36
	}
	if height > 180 {
		return 180
	}
	return height
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

const workbookStyles = `<?xml version="1.0" encoding="UTF-8"?>
<styleSheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">
  <fonts count="2">
    <font><sz val="11"/><color theme="1"/><name val="Calibri"/><family val="2"/><scheme val="minor"/></font>
    <font><b/><sz val="11"/><color rgb="FFFFFFFF"/><name val="Calibri"/><family val="2"/><scheme val="minor"/></font>
  </fonts>
  <fills count="4">
    <fill><patternFill patternType="none"/></fill>
    <fill><patternFill patternType="gray125"/></fill>
    <fill><patternFill patternType="solid"><fgColor rgb="FF1F4E78"/><bgColor indexed="64"/></patternFill></fill>
    <fill><patternFill patternType="solid"><fgColor rgb="FFFFF2CC"/><bgColor indexed="64"/></patternFill></fill>
  </fills>
  <borders count="2">
    <border><left/><right/><top/><bottom/><diagonal/></border>
    <border><left style="thin"><color rgb="FFD9E2F3"/></left><right style="thin"><color rgb="FFD9E2F3"/></right><top style="thin"><color rgb="FFD9E2F3"/></top><bottom style="thin"><color rgb="FFD9E2F3"/></bottom><diagonal/></border>
  </borders>
  <cellStyleXfs count="1"><xf numFmtId="0" fontId="0" fillId="0" borderId="0"/></cellStyleXfs>
  <cellXfs count="5">
    <xf numFmtId="0" fontId="0" fillId="0" borderId="0" xfId="0"/>
    <xf numFmtId="0" fontId="1" fillId="2" borderId="1" xfId="0" applyFont="1" applyFill="1" applyBorder="1" applyAlignment="1"><alignment horizontal="center" vertical="center" wrapText="1"/></xf>
    <xf numFmtId="0" fontId="0" fillId="0" borderId="1" xfId="0" applyBorder="1" applyAlignment="1"><alignment vertical="top" wrapText="1"/></xf>
    <xf numFmtId="0" fontId="0" fillId="3" borderId="1" xfId="0" applyFill="1" applyBorder="1" applyAlignment="1"><alignment vertical="top" wrapText="1"/></xf>
    <xf numFmtId="14" fontId="0" fillId="3" borderId="1" xfId="0" applyNumberFormat="1" applyFill="1" applyBorder="1" applyAlignment="1"><alignment vertical="top" wrapText="1"/></xf>
  </cellXfs>
  <cellStyles count="1"><cellStyle name="Normal" xfId="0" builtinId="0"/></cellStyles>
  <tableStyles count="0" defaultTableStyle="TableStyleMedium2" defaultPivotStyle="PivotStyleLight16"/>
</styleSheet>`
