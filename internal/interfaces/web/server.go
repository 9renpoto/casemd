package web

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	templatehtml "github.com/gofiber/template/html/v2"

	"github.com/9renpoto/casemd/internal/app"
)

// CSVConverter drives Markdown transformations for the web UI.
type CSVConverter interface {
	Convert(sources []app.Source, output io.Writer) error
}

// Server exposes a Fiber application that wraps the Markdown converters for ad-hoc debugging.
type Server struct {
	app          *fiber.App
	csvConverter CSVConverter
	defaults     defaultState
}

// NewServer wires the Fiber instance with the provided converter and registers the base routes.
func NewServer(csvConverter CSVConverter) *Server {
	engine := templatehtml.New("./internal/interfaces/web/templates", ".html")
	engine.AddFunc("json", toJSON)

	fiberApp := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		Views:                 engine,
	})

	server := &Server{
		app:          fiberApp,
		csvConverter: csvConverter,
	}
	server.loadDefaultPreview()
	server.registerRoutes()
	return server
}

// Listen starts serving the Fiber application on the provided address.
func (s *Server) Listen(addr string) error {
	if s.app == nil {
		return errors.New("fiber app is not configured")
	}
	return s.app.Listen(addr)
}

// App returns the underlying Fiber application. Useful for tests.
func (s *Server) App() *fiber.App {
	return s.app
}

func (s *Server) registerRoutes() {
	s.app.Get("/healthz", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	s.app.Get("/", func(c *fiber.Ctx) error {
		return c.Render("index", indexViewModel{
			DefaultState: s.defaults,
		})
	})

	s.app.Post("/api/preview", func(c *fiber.Ctx) error {
		var payload previewRequest
		if err := c.BodyParser(&payload); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("parse request body: %v", err))
		}

		if strings.TrimSpace(payload.Markdown) == "" {
			return fiber.NewError(fiber.StatusBadRequest, "markdown content cannot be empty")
		}

		if s.csvConverter == nil {
			return fiber.NewError(fiber.StatusServiceUnavailable, "csv converter is not configured")
		}

		var buffer bytes.Buffer
		err := s.csvConverter.Convert([]app.Source{
			{Name: payload.Name, Reader: strings.NewReader(payload.Markdown)},
		}, &buffer)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("convert markdown: %v", err))
		}

		return c.JSON(fiber.Map{"csv": buffer.String()})
	})
}

type previewRequest struct {
	Name     string `json:"name"`
	Markdown string `json:"markdown"`
}

func (s *Server) loadDefaultPreview() {
	const defaultMarkdownPath = "notes.md"

	data, err := os.ReadFile(defaultMarkdownPath)
	if err != nil {
		return
	}

	s.defaults.Name = defaultMarkdownPath
	s.defaults.Markdown = string(data)

	if s.csvConverter == nil {
		return
	}

	var buffer bytes.Buffer
	err = s.csvConverter.Convert([]app.Source{
		{Name: defaultMarkdownPath, Reader: bytes.NewReader(data)},
	}, &buffer)
	if err != nil {
		return
	}
	s.defaults.CSV = buffer.String()
}

type defaultState struct {
	Name     string `json:"name"`
	Markdown string `json:"markdown"`
	CSV      string `json:"csv"`
}

type indexViewModel struct {
	DefaultState defaultState
}

func toJSON(value any) template.JS {
	data, err := json.Marshal(value)
	if err != nil {
		return template.JS("{}")
	}
	return template.JS(data)
}
