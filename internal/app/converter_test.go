package app

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/csv"
	"encoding/xml"
	"fmt"
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/9renpoto/casemd/internal/core/domain"
)

type mockCaseParser struct {
	cases []domain.Case
	err   error
}

func (m *mockCaseParser) Parse(r io.Reader) ([]domain.Case, error) {
	return m.cases, m.err
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

func readSheetRows(t *testing.T, data []byte, sheetIndex int) [][]string {
	t.Helper()

	reader := bytes.NewReader(data)
	zipReader, err := zip.NewReader(reader, int64(len(data)))
	if err != nil {
		t.Fatalf("open zip: %v", err)
	}

	target := fmt.Sprintf("xl/worksheets/sheet%d.xml", sheetIndex)
	for _, file := range zipReader.File {
		if file.Name != target {
			continue
		}

		rc, err := file.Open()
		if err != nil {
			t.Fatalf("open sheet: %v", err)
		}
		defer rc.Close()

		content, err := io.ReadAll(rc)
		if err != nil {
			t.Fatalf("read sheet: %v", err)
		}

		var ws worksheet
		if err := xml.Unmarshal(content, &ws); err != nil {
			t.Fatalf("unmarshal sheet: %v", err)
		}

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

	t.Fatalf("sheet %d not found", sheetIndex)
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

type worksheet struct {
	XMLName   xml.Name  `xml:"worksheet"`
	SheetData sheetData `xml:"sheetData"`
}

type sheetData struct {
	Rows []sheetRow `xml:"row"`
}

type sheetRow struct {
	Cells []sheetCell `xml:"c"`
}

type sheetCell struct {
	InlineStr inlineString `xml:"is"`
}

type inlineString struct {
	Text string `xml:"t"`
}

func (c sheetCell) Value() string {
	return c.InlineStr.Text
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
