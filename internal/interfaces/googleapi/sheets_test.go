package googleapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/9renpoto/casemd/internal/app"
)

func TestBuildSpreadsheetPayload(t *testing.T) {
	spreadsheet := app.GoogleSpreadsheet{
		Title: "Casemd Export",
		Sheets: []app.GoogleSpreadsheetSheet{
			{
				Title: "alpha",
				Rows:  [][]string{{"A", "B"}, {"1", "2"}},
			},
		},
	}

	payload, err := buildSpreadsheetPayload(spreadsheet)
	if err != nil {
		t.Fatalf("buildSpreadsheetPayload() error = %v", err)
	}

	if payload.Properties.Title != "Casemd Export" {
		t.Fatalf("unexpected title: %s", payload.Properties.Title)
	}

	if len(payload.Sheets) != 1 {
		t.Fatalf("expected 1 sheet, got %d", len(payload.Sheets))
	}

	values := [][]string{}
	for _, row := range payload.Sheets[0].Data[0].RowData {
		rowValues := make([]string, len(row.Values))
		for i, cell := range row.Values {
			if cell.UserEnteredValue == nil {
				rowValues[i] = ""
				continue
			}
			rowValues[i] = cell.UserEnteredValue.StringValue
		}
		values = append(values, rowValues)
	}

	expected := [][]string{{"A", "B"}, {"1", "2"}}
	if !reflect.DeepEqual(values, expected) {
		t.Fatalf("unexpected cell values: %#v", values)
	}
}

func TestBuildSpreadsheetPayloadRequiresTitle(t *testing.T) {
	_, err := buildSpreadsheetPayload(app.GoogleSpreadsheet{})
	if err == nil {
		t.Fatalf("expected error for missing title")
	}
}

func TestCreateSpreadsheet(t *testing.T) {
	var capturedRequest spreadsheetPayload
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer token" {
			t.Fatalf("unexpected authorization header: %s", got)
		}
		if err := json.NewDecoder(r.Body).Decode(&capturedRequest); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		_ = r.Body.Close()
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"spreadsheetId":"sheet-id"}`))
	}))
	defer server.Close()

	service, err := NewSheetsService(server.Client(), "token")
	if err != nil {
		t.Fatalf("NewSheetsService() error = %v", err)
	}
	service.endpoint = server.URL

	spreadsheet := app.GoogleSpreadsheet{
		Title:  "Casemd Export",
		Sheets: []app.GoogleSpreadsheetSheet{{Title: "alpha", Rows: [][]string{{"A"}}}},
	}

	id, err := service.CreateSpreadsheet(context.Background(), spreadsheet)
	if err != nil {
		t.Fatalf("CreateSpreadsheet() error = %v", err)
	}
	if id != "sheet-id" {
		t.Fatalf("unexpected spreadsheet id: %s", id)
	}
	if capturedRequest.Properties.Title != "Casemd Export" {
		t.Fatalf("unexpected request title: %s", capturedRequest.Properties.Title)
	}
}

func TestCreateSpreadsheetHandlesErrorResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte("forbidden"))
	}))
	defer server.Close()

	service, err := NewSheetsService(server.Client(), "token")
	if err != nil {
		t.Fatalf("NewSheetsService() error = %v", err)
	}
	service.endpoint = server.URL

	spreadsheet := app.GoogleSpreadsheet{Title: "Casemd Export"}
	_, err = service.CreateSpreadsheet(context.Background(), spreadsheet)
	if err == nil {
		t.Fatalf("expected error from API response")
	}
}
