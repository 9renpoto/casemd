package googleapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/9renpoto/casemd/internal/app"
)

const defaultSheetsEndpoint = "https://sheets.googleapis.com/v4/spreadsheets"

// SheetsService adapts HTTP interactions with the Google Sheets API.
type SheetsService struct {
	client      *http.Client
	endpoint    string
	accessToken string
}

// NewSheetsService builds a SheetsService with the provided HTTP client and OAuth token.
// The OAuth token must grant the `https://www.googleapis.com/auth/spreadsheets` scope.
func NewSheetsService(client *http.Client, accessToken string) (*SheetsService, error) {
	if accessToken == "" {
		return nil, fmt.Errorf("missing Google Sheets access token")
	}
	if client == nil {
		client = http.DefaultClient
	}
	return &SheetsService{
		client:      client,
		endpoint:    defaultSheetsEndpoint,
		accessToken: accessToken,
	}, nil
}

// CreateSpreadsheet converts the domain spreadsheet into a Google Sheets API call.
func (s *SheetsService) CreateSpreadsheet(ctx context.Context, spreadsheet app.GoogleSpreadsheet) (string, error) {
	payload, err := buildSpreadsheetPayload(spreadsheet)
	if err != nil {
		return "", err
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal spreadsheet payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.endpoint, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+s.accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("call google sheets api: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		message, readErr := readErrorMessage(resp.Body)
		if readErr != nil {
			return "", fmt.Errorf("google sheets api error (%d)", resp.StatusCode)
		}
		return "", fmt.Errorf("google sheets api error (%d): %s", resp.StatusCode, message)
	}

	var result struct {
		SpreadsheetID string `json:"spreadsheetId"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode google sheets response: %w", err)
	}
	if result.SpreadsheetID == "" {
		return "", fmt.Errorf("google sheets response missing spreadsheetId")
	}
	return result.SpreadsheetID, nil
}

func readErrorMessage(body io.Reader) (string, error) {
	data, err := io.ReadAll(io.LimitReader(body, 4096))
	if err != nil {
		return "", err
	}
	trimmed := strings.TrimSpace(string(data))
	if trimmed == "" {
		return "", fmt.Errorf("empty error body")
	}
	return trimmed, nil
}

func buildSpreadsheetPayload(spreadsheet app.GoogleSpreadsheet) (*spreadsheetPayload, error) {
	if spreadsheet.Title == "" {
		return nil, fmt.Errorf("spreadsheet title cannot be empty")
	}
	payload := &spreadsheetPayload{
		Properties: spreadsheetProperties{Title: spreadsheet.Title},
		Sheets:     make([]sheetPayload, 0, len(spreadsheet.Sheets)),
	}

	for _, sheet := range spreadsheet.Sheets {
		rows := make([]rowData, 0, len(sheet.Rows))
		for _, row := range sheet.Rows {
			cells := make([]cellData, 0, len(row))
			for _, value := range row {
				cell := cellData{}
				if value != "" {
					cell.UserEnteredValue = &extendedValue{StringValue: value}
				}
				cells = append(cells, cell)
			}
			rows = append(rows, rowData{Values: cells})
		}

		payload.Sheets = append(payload.Sheets, sheetPayload{
			Properties: sheetProperties{Title: sheet.Title},
			Data:       []gridData{{RowData: rows}},
		})
	}

	return payload, nil
}

type spreadsheetPayload struct {
	Properties spreadsheetProperties `json:"properties"`
	Sheets     []sheetPayload        `json:"sheets"`
}

type spreadsheetProperties struct {
	Title string `json:"title"`
}

type sheetPayload struct {
	Properties sheetProperties `json:"properties"`
	Data       []gridData      `json:"data"`
}

type sheetProperties struct {
	Title string `json:"title"`
}

type gridData struct {
	RowData []rowData `json:"rowData"`
}

type rowData struct {
	Values []cellData `json:"values"`
}

type cellData struct {
	UserEnteredValue *extendedValue `json:"userEnteredValue,omitempty"`
}

type extendedValue struct {
	StringValue string `json:"stringValue,omitempty"`
}
