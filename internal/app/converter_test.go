package app

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/csv"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/9renpoto/casemd/internal/core/domain"
	"github.com/9renpoto/casemd/internal/core/parser"
)

type mockCaseParser struct {
	cases       []domain.Case
	diagnostics []domain.Diagnostic
	err         error
}

func (m *mockCaseParser) ParseWithDiagnostics(_ string, _ io.Reader) ([]domain.Case, []domain.Diagnostic, error) {
	return m.cases, m.diagnostics, m.err
}

type mockGoogleCreator struct {
	spreadsheet GoogleSpreadsheet
	id          string
	err         error
}

func (m *mockGoogleCreator) CreateSpreadsheet(ctx context.Context, spreadsheet GoogleSpreadsheet) (string, error) {
	m.spreadsheet = spreadsheet
	return m.id, m.err
}

func TestMarkdownToCSV_Convert(t *testing.T) {
	mockCases := []domain.Case{
		{
			MajorItem:       "Setup",
			MediumItem:      "Environment",
			MinorItem:       "Dependencies",
			ValidationSteps: []string{"Step 1", "Step 2"},
			Checkpoints:     []string{"* [ ] Check 1", "* [ ] Check 2"},
		},
		{
			MajorItem:       "Execution",
			MediumItem:      "Workflow",
			MinorItem:       "Run",
			ValidationSteps: []string{"Run command"},
			Checkpoints:     []string{"* [ ] Success"},
		},
	}

	parser := &mockCaseParser{cases: mockCases}
	converter := NewMarkdownToCSV(parser)

	sources := []Source{{Name: "checks.md", Reader: strings.NewReader("")}}
	var output bytes.Buffer

	if err := converter.Convert(sources, &output); err != nil {
		t.Fatalf("Convert() returned an unexpected error: %v", err)
	}

	reader := csv.NewReader(&output)
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll() returned an unexpected error: %v", err)
	}

	expectedRecords := [][]string{
		spreadsheetHeaders,
		caseRow(mockCases[0]),
		caseRow(mockCases[1]),
	}

	if !reflect.DeepEqual(records, expectedRecords) {
		t.Fatalf("unexpected CSV records: %#v", records)
	}
}

func TestMarkdownToSpreadsheet_Convert(t *testing.T) {
	mockCases := []domain.Case{
		{
			MajorItem:       "Setup",
			MediumItem:      "Environment",
			MinorItem:       "Dependencies",
			ValidationSteps: []string{"Step 1", "Step 2"},
			Checkpoints:     []string{"* [ ] Check 1", "* [ ] Check 2"},
		},
		{
			MajorItem:       "Execution",
			MediumItem:      "Workflow",
			MinorItem:       "Run",
			ValidationSteps: []string{"Run command"},
			Checkpoints:     []string{"* [ ] Success"},
		},
	}

	parser := &mockCaseParser{cases: mockCases}
	converter := NewMarkdownToSpreadsheet(parser)

	sources := []Source{{Name: "checks.md", Reader: strings.NewReader("")}}
	var output bytes.Buffer

	if err := converter.Convert(sources, &output); err != nil {
		t.Fatalf("Convert() returned an unexpected error: %v", err)
	}

	rows := readSheetRows(t, output.Bytes(), 1)

	expectedRows := [][]string{
		append([]string(nil), spreadsheetHeaders...),
		caseRow(mockCases[0]),
		caseRow(mockCases[1]),
	}

	if !reflect.DeepEqual(rows, expectedRows) {
		t.Fatalf("unexpected rows: %#v", rows)
	}
}

func TestMarkdownToSpreadsheet_ConvertMultipleSources(t *testing.T) {
	parser := &mockCaseParser{cases: []domain.Case{{MinorItem: "Row"}}}
	converter := NewMarkdownToSpreadsheet(parser)

	sources := []Source{
		{Name: "alpha.md", Reader: strings.NewReader("")},
		{Name: "alpha.md", Reader: strings.NewReader("")},
		{Name: "beta.md", Reader: strings.NewReader("")},
	}

	var output bytes.Buffer
	if err := converter.Convert(sources, &output); err != nil {
		t.Fatalf("Convert() returned an unexpected error: %v", err)
	}

	sheetNames := readSheetNames(t, output.Bytes())
	expectedSheets := []string{"alpha", "alpha_2", "beta"}
	if !reflect.DeepEqual(sheetNames, expectedSheets) {
		t.Fatalf("unexpected sheets: %#v", sheetNames)
	}
}

func TestMarkdownToSpreadsheet_HumanFirstFormatting(t *testing.T) {
	fixturePath := filepath.Join("..", "..", "testdata", "human-first-scenario.md")
	fixture, err := os.Open(fixturePath)
	if err != nil {
		t.Fatalf("open fixture: %v", err)
	}
	defer fixture.Close()

	cases, diagnostics, err := parser.ParseWithDiagnostics(fixturePath, fixture)
	if err != nil {
		t.Fatalf("parse fixture: %v", err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("fixture diagnostics = %+v, want none", diagnostics)
	}

	var output bytes.Buffer
	converter := NewMarkdownToSpreadsheet(&mockCaseParser{cases: cases})
	if err := converter.Convert([]Source{{Name: fixturePath, Reader: strings.NewReader("")}}, &output); err != nil {
		t.Fatalf("Convert() returned an unexpected error: %v", err)
	}

	worksheet := readWorksheet(t, output.Bytes(), 1)
	if worksheet.SheetViews.Views[0].Pane.State != "frozen" || worksheet.SheetViews.Views[0].Pane.YSplit != 1 || worksheet.SheetViews.Views[0].Pane.TopLeftCell != "A2" {
		t.Fatalf("unexpected frozen pane: %+v", worksheet.SheetViews.Views[0].Pane)
	}
	if worksheet.AutoFilter.Ref != "A1:I4" {
		t.Fatalf("auto filter ref = %q, want %q", worksheet.AutoFilter.Ref, "A1:I4")
	}

	if len(worksheet.Columns.Columns) != len(spreadsheetColumnWidths) {
		t.Fatalf("column count = %d, want %d", len(worksheet.Columns.Columns), len(spreadsheetColumnWidths))
	}
	for index, column := range worksheet.Columns.Columns {
		if column.Min != index+1 || column.Max != index+1 || column.Width != spreadsheetColumnWidths[index] || column.CustomWidth != 1 {
			t.Errorf("column %d = %+v, want width %.1f", index+1, column, spreadsheetColumnWidths[index])
		}
	}

	if len(worksheet.SheetData.Rows) != 4 {
		t.Fatalf("row count = %d, want 4", len(worksheet.SheetData.Rows))
	}
	headerRow := worksheet.SheetData.Rows[0]
	if headerRow.Height != 30 || headerRow.CustomHeight != 1 {
		t.Fatalf("header row presentation = %+v", headerRow)
	}
	for _, cell := range headerRow.Cells {
		if cell.Style != 1 {
			t.Errorf("header cell %s style = %d, want 1", cell.Reference, cell.Style)
		}
	}

	dataRow := worksheet.SheetData.Rows[1]
	if dataRow.Height <= 36 {
		t.Fatalf("long-content row height = %.1f, want greater than 36", dataRow.Height)
	}
	wantStyles := []int{2, 2, 2, 2, 2, 3, 4, 3, 3}
	for index, want := range wantStyles {
		if got := dataRow.Cells[index].Style; got != want {
			t.Errorf("data cell %s style = %d, want %d", dataRow.Cells[index].Reference, got, want)
		}
	}
	if got := dataRow.Cells[3].Value(); !strings.Contains(got, "\n") || !strings.Contains(got, "Save the profile") {
		t.Errorf("validation steps did not preserve multiline content: %q", got)
	}
	if got := dataRow.Cells[4].Value(); !strings.Contains(got, "\n") || !strings.Contains(got, "* [ ]") {
		t.Errorf("checkpoints did not preserve task-list content: %q", got)
	}

	styles := readStyles(t, output.Bytes())
	if len(styles.CellXfs.Xfs) != 5 {
		t.Fatalf("cell style count = %d, want 5", len(styles.CellXfs.Xfs))
	}
	headerStyle := styles.CellXfs.Xfs[1]
	if headerStyle.FontID != 1 || headerStyle.FillID != 2 || !headerStyle.Alignment.WrapText || headerStyle.Alignment.Vertical != "center" {
		t.Fatalf("unexpected header style: %+v", headerStyle)
	}
	definitionStyle := styles.CellXfs.Xfs[2]
	if definitionStyle.FillID != 0 || !definitionStyle.Alignment.WrapText || definitionStyle.Alignment.Vertical != "top" {
		t.Fatalf("unexpected definition style: %+v", definitionStyle)
	}
	executionStyle := styles.CellXfs.Xfs[3]
	if executionStyle.FillID == definitionStyle.FillID || !executionStyle.Alignment.WrapText || executionStyle.Alignment.Vertical != "top" {
		t.Fatalf("unexpected execution style: %+v", executionStyle)
	}
	if dateStyle := styles.CellXfs.Xfs[4]; dateStyle.NumFmtID != 14 || dateStyle.FillID != executionStyle.FillID {
		t.Fatalf("unexpected test-date style: %+v", dateStyle)
	}

	contentTypes := readZipFile(t, output.Bytes(), "[Content_Types].xml")
	for _, part := range []string{`/docProps/core.xml`, `/docProps/app.xml`, `/xl/styles.xml`} {
		if !bytes.Contains(contentTypes, []byte(part)) {
			t.Errorf("content types do not include %s", part)
		}
	}
	workbookRelationships := readZipFile(t, output.Bytes(), "xl/_rels/workbook.xml.rels")
	if !bytes.Contains(workbookRelationships, []byte(`relationships/styles`)) {
		t.Fatalf("workbook relationships do not include styles.xml")
	}
}

func readSheetRows(t *testing.T, data []byte, sheetIndex int) [][]string {
	t.Helper()

	ws := readWorksheet(t, data, sheetIndex)
	rows := make([][]string, len(ws.SheetData.Rows))
	for i, row := range ws.SheetData.Rows {
		values := make([]string, len(row.Cells))
		for j, cell := range row.Cells {
			values[j] = cell.Value()
		}
		rows[i] = values
	}
	return rows
}

func readWorksheet(t *testing.T, data []byte, sheetIndex int) worksheet {
	t.Helper()

	content := readZipFile(t, data, fmt.Sprintf("xl/worksheets/sheet%d.xml", sheetIndex))
	var ws worksheet
	if err := xml.Unmarshal(content, &ws); err != nil {
		t.Fatalf("unmarshal sheet: %v", err)
	}
	return ws
}

func readStyles(t *testing.T, data []byte) styleSheet {
	t.Helper()

	content := readZipFile(t, data, "xl/styles.xml")
	var styles styleSheet
	if err := xml.Unmarshal(content, &styles); err != nil {
		t.Fatalf("unmarshal styles: %v", err)
	}
	return styles
}

func readZipFile(t *testing.T, data []byte, target string) []byte {
	t.Helper()

	reader := bytes.NewReader(data)
	zipReader, err := zip.NewReader(reader, int64(len(data)))
	if err != nil {
		t.Fatalf("open zip: %v", err)
	}

	for _, file := range zipReader.File {
		if file.Name != target {
			continue
		}
		rc, err := file.Open()
		if err != nil {
			t.Fatalf("open %s: %v", target, err)
		}
		defer rc.Close()
		content, err := io.ReadAll(rc)
		if err != nil {
			t.Fatalf("read %s: %v", target, err)
		}
		return content
	}

	t.Fatalf("%s not found", target)
	return nil
}

func readSheetNames(t *testing.T, data []byte) []string {
	t.Helper()

	reader := bytes.NewReader(data)
	zipReader, err := zip.NewReader(reader, int64(len(data)))
	if err != nil {
		t.Fatalf("open zip: %v", err)
	}

	for _, file := range zipReader.File {
		if file.Name != "xl/workbook.xml" {
			continue
		}

		rc, err := file.Open()
		if err != nil {
			t.Fatalf("open workbook: %v", err)
		}
		defer rc.Close()

		content, err := io.ReadAll(rc)
		if err != nil {
			t.Fatalf("read workbook: %v", err)
		}

		var wb workbookFile
		if err := xml.Unmarshal(content, &wb); err != nil {
			t.Fatalf("unmarshal workbook: %v", err)
		}

		names := make([]string, len(wb.Sheets.Sheets))
		for i, sheet := range wb.Sheets.Sheets {
			names[i] = sheet.Name
		}
		return names
	}

	t.Fatalf("workbook metadata not found")
	return nil
}

func TestMarkdownToGoogleSpreadsheet_Create(t *testing.T) {
	parser := &mockCaseParser{cases: []domain.Case{{MinorItem: "One"}}}
	creator := &mockGoogleCreator{id: "spreadsheet-id"}
	converter := NewMarkdownToGoogleSpreadsheet(parser, creator)

	sources := []Source{{Name: "alpha.md", Reader: strings.NewReader("")}}
	id, err := converter.Create(context.Background(), "Casemd Export", sources)
	if err != nil {
		t.Fatalf("Create() returned an unexpected error: %v", err)
	}
	if id != "spreadsheet-id" {
		t.Fatalf("unexpected spreadsheet id: %s", id)
	}

	if creator.spreadsheet.Title != "Casemd Export" {
		t.Fatalf("unexpected spreadsheet title: %s", creator.spreadsheet.Title)
	}
	if len(creator.spreadsheet.Sheets) != 1 {
		t.Fatalf("expected 1 sheet, got %d", len(creator.spreadsheet.Sheets))
	}
	sheet := creator.spreadsheet.Sheets[0]
	if sheet.Title != "alpha" {
		t.Fatalf("unexpected sheet title: %s", sheet.Title)
	}
	expectedRows := [][]string{
		append([]string(nil), spreadsheetHeaders...),
		{"", "", "One", "", "", "", "", "", ""},
	}
	if !reflect.DeepEqual(sheet.Rows, expectedRows) {
		t.Fatalf("unexpected rows: %#v", sheet.Rows)
	}
}

func TestMarkdownToGoogleSpreadsheet_CreatePropagatesParserError(t *testing.T) {
	parser := &mockCaseParser{err: fmt.Errorf("parse error")}
	creator := &mockGoogleCreator{}
	converter := NewMarkdownToGoogleSpreadsheet(parser, creator)

	sources := []Source{{Name: "alpha.md", Reader: strings.NewReader("")}}
	_, err := converter.Create(context.Background(), "Casemd Export", sources)
	if err == nil {
		t.Fatalf("Create() expected error but got nil")
	}
	if creator.spreadsheet.Title != "" {
		t.Fatalf("creator should not have been invoked")
	}
}

func TestMarkdownToGoogleSpreadsheet_CreateRequiresSources(t *testing.T) {
	parser := &mockCaseParser{}
	creator := &mockGoogleCreator{}
	converter := NewMarkdownToGoogleSpreadsheet(parser, creator)

	_, err := converter.Create(context.Background(), "Casemd Export", nil)
	if err == nil {
		t.Fatalf("expected error for missing sources")
	}
}

func TestMarkdownToGoogleSpreadsheet_CreateRequiresTitle(t *testing.T) {
	parser := &mockCaseParser{}
	creator := &mockGoogleCreator{}
	converter := NewMarkdownToGoogleSpreadsheet(parser, creator)

	sources := []Source{{Name: "alpha.md", Reader: strings.NewReader("")}}
	_, err := converter.Create(context.Background(), "", sources)
	if err == nil {
		t.Fatalf("expected error for missing title")
	}
}

func TestConvertersRejectDiagnosticsBeforeWriting(t *testing.T) {
	diagnostic := domain.Diagnostic{
		Source:  "invalid.md",
		Line:    3,
		Rule:    domain.RuleMissingSteps,
		Message: "test case must contain at least one ordered step",
	}
	parser := &mockCaseParser{diagnostics: []domain.Diagnostic{diagnostic}}
	sources := []Source{{Name: "invalid.md", Reader: strings.NewReader("")}}

	t.Run("CSV", func(t *testing.T) {
		var output bytes.Buffer
		err := NewMarkdownToCSV(parser).Convert(sources, &output)
		assertValidationError(t, err, diagnostic)
		if output.Len() != 0 {
			t.Fatalf("CSV output was written before validation completed: %q", output.String())
		}
	})

	t.Run("spreadsheet", func(t *testing.T) {
		var output bytes.Buffer
		err := NewMarkdownToSpreadsheet(parser).Convert(sources, &output)
		assertValidationError(t, err, diagnostic)
		if output.Len() != 0 {
			t.Fatalf("spreadsheet output was written before validation completed")
		}
	})

	t.Run("Google Spreadsheet", func(t *testing.T) {
		creator := &mockGoogleCreator{}
		_, err := NewMarkdownToGoogleSpreadsheet(parser, creator).Create(context.Background(), "Export", sources)
		assertValidationError(t, err, diagnostic)
		if creator.spreadsheet.Title != "" {
			t.Fatalf("Google Spreadsheet creator was invoked before validation completed")
		}
	})
}

func assertValidationError(t *testing.T, err error, want domain.Diagnostic) {
	t.Helper()
	var validationErr *ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("expected ValidationError, got %v", err)
	}
	if !reflect.DeepEqual(validationErr.Diagnostics, []domain.Diagnostic{want}) {
		t.Fatalf("diagnostics = %+v, want %+v", validationErr.Diagnostics, []domain.Diagnostic{want})
	}
}

type worksheet struct {
	XMLName    xml.Name     `xml:"worksheet"`
	SheetViews sheetViews   `xml:"sheetViews"`
	Columns    sheetColumns `xml:"cols"`
	SheetData  sheetData    `xml:"sheetData"`
	AutoFilter autoFilter   `xml:"autoFilter"`
}

type sheetViews struct {
	Views []sheetView `xml:"sheetView"`
}

type sheetView struct {
	Pane sheetPane `xml:"pane"`
}

type sheetPane struct {
	YSplit      float64 `xml:"ySplit,attr"`
	TopLeftCell string  `xml:"topLeftCell,attr"`
	ActivePane  string  `xml:"activePane,attr"`
	State       string  `xml:"state,attr"`
}

type sheetColumns struct {
	Columns []sheetColumn `xml:"col"`
}

type sheetColumn struct {
	Min         int     `xml:"min,attr"`
	Max         int     `xml:"max,attr"`
	Width       float64 `xml:"width,attr"`
	CustomWidth int     `xml:"customWidth,attr"`
}

type autoFilter struct {
	Ref string `xml:"ref,attr"`
}

type sheetData struct {
	Rows []sheetRow `xml:"row"`
}

type sheetRow struct {
	Reference    int         `xml:"r,attr"`
	Height       float64     `xml:"ht,attr"`
	CustomHeight int         `xml:"customHeight,attr"`
	Cells        []sheetCell `xml:"c"`
}

type sheetCell struct {
	Reference string       `xml:"r,attr"`
	Style     int          `xml:"s,attr"`
	InlineStr inlineString `xml:"is"`
}

type inlineString struct {
	Text string `xml:"t"`
}

func (c sheetCell) Value() string {
	return c.InlineStr.Text
}

type styleSheet struct {
	CellXfs styleCellXfs `xml:"cellXfs"`
}

type styleCellXfs struct {
	Count int       `xml:"count,attr"`
	Xfs   []styleXF `xml:"xf"`
}

type styleXF struct {
	NumFmtID  int            `xml:"numFmtId,attr"`
	FontID    int            `xml:"fontId,attr"`
	FillID    int            `xml:"fillId,attr"`
	BorderID  int            `xml:"borderId,attr"`
	Alignment styleAlignment `xml:"alignment"`
}

type styleAlignment struct {
	Horizontal string `xml:"horizontal,attr"`
	Vertical   string `xml:"vertical,attr"`
	WrapText   bool   `xml:"wrapText,attr"`
}

type workbookFile struct {
	XMLName xml.Name    `xml:"workbook"`
	Sheets  workbookSet `xml:"sheets"`
}

type workbookSet struct {
	Sheets []workbookSheetInfo `xml:"sheet"`
}

type workbookSheetInfo struct {
	Name string `xml:"name,attr"`
}
