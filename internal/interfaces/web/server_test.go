package web_test

import (
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"

	"github.com/9renpoto/casemd/internal/app"
	"github.com/9renpoto/casemd/internal/interfaces/web"
)

func TestHealthEndpoint(t *testing.T) {
	server := web.NewServer(&stubConverter{})

	req := httptest.NewRequest(fiber.MethodGet, "/healthz", nil)
	resp, err := server.App().Test(req, -1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.StatusCode != 200 {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestPreviewEndpoint(t *testing.T) {
	converter := &stubConverter{output: "header1,header2\nvalue1,value2\n"}
	server := web.NewServer(converter)

	body := strings.NewReader(`{"name":"sample.md","markdown":"# heading"}`)
	req := httptest.NewRequest(fiber.MethodPost, "/api/preview", body)
	req.Header.Set("Content-Type", "application/json")

	resp, err := server.App().Test(req, -1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.StatusCode != 200 {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	if len(converter.sources) != 1 {
		t.Fatalf("expected converter to receive 1 source, got %d", len(converter.sources))
	}
	if converter.sources[0].Name != "sample.md" {
		t.Fatalf("expected source name 'sample.md', got %q", converter.sources[0].Name)
	}
}

func TestPreviewEndpointMissingMarkdown(t *testing.T) {
	server := web.NewServer(&stubConverter{})

	body := strings.NewReader(`{"name":"sample.md","markdown":"   "}`)
	req := httptest.NewRequest(fiber.MethodPost, "/api/preview", body)
	req.Header.Set("Content-Type", "application/json")

	resp, err := server.App().Test(req, -1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", fiber.StatusBadRequest, resp.StatusCode)
	}
}

type stubConverter struct {
	output  string
	err     error
	sources []app.Source
}

func (s *stubConverter) Convert(sources []app.Source, writer io.Writer) error {
	s.sources = sources
	if s.err != nil {
		return s.err
	}
	if s.output == "" {
		return nil
	}
	_, err := io.WriteString(writer, s.output)
	return err
}
