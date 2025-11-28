package main

import (
	"bytes"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"go-rag-demo/rag"
)

type FakeEmbedder struct{}

func (f *FakeEmbedder) Embed(text string) []float64 {
	// deterministic embedding
	return []float64{0.1, 0.2, 0.3}
}

type fakePDFReader struct {
	text string
}

func (f *fakePDFReader) GetPlainText() (io.Reader, error) {
	return strings.NewReader(f.text), nil
}

func captureLogs(t *testing.T, fn func()) string {
	t.Helper()
	var buf bytes.Buffer
	old := log.Writer()
	log.SetOutput(&buf)
	defer log.SetOutput(old)

	fn()
	return buf.String()
}

func newTestServer() *Server {
	return NewServerWithEmbedder(&FakeEmbedder{})
}

func TestHealthHandler(t *testing.T) {
	srv := newTestServer()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	logs := captureLogs(t, func() {
		srv.healthHandler(w, req)
	})

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	// health has no logs
	if strings.TrimSpace(logs) != "" {
		t.Fatalf("expected no logs, got %q", logs)
	}
}

func TestUploadHandler(t *testing.T) {
	srv := newTestServer()

	longText := strings.Repeat("This is a sentence that will be chunked. ", 500)

	tests := []struct {
		name           string
		method         string
		body           string
		wantStatusCode int
		wantLogSubstr  string // if empty, we assert there are no logs
	}{
		{
			name:           "success_small_text",
			method:         http.MethodPost,
			body:           "This is a valid small text",
			wantStatusCode: http.StatusOK,
			wantLogSubstr:  "upload_text=",
		},
		{
			name:           "method_not_allowed",
			method:         http.MethodGet,
			body:           "whatever",
			wantStatusCode: http.StatusMethodNotAllowed,
			wantLogSubstr:  "",
		},
		{
			name:           "empty_body",
			method:         http.MethodPost,
			body:           "",
			wantStatusCode: http.StatusBadRequest,
			wantLogSubstr:  "",
		},
		{
			name:           "too_many_chunks",
			method:         http.MethodPost,
			body:           longText,
			wantStatusCode: http.StatusBadRequest,
			wantLogSubstr:  "error - text too big",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			var bodyReader io.Reader
			if tc.body != "" {
				bodyReader = strings.NewReader(tc.body)
			}

			req := httptest.NewRequest(tc.method, "/upload", bodyReader)
			w := httptest.NewRecorder()

			logs := captureLogs(t, func() {
				srv.uploadHandler(w, req)
			})

			if w.Code != tc.wantStatusCode {
				t.Fatalf("expected status %d, got %d", tc.wantStatusCode, w.Code)
			}

			if tc.wantLogSubstr == "" {
				if strings.TrimSpace(logs) != "" {
					t.Fatalf("expected no logs, got %q", logs)
				}
			} else {
				if !strings.Contains(logs, tc.wantLogSubstr) {
					t.Fatalf("expected logs to contain %q, got %q", tc.wantLogSubstr, logs)
				}
			}
		})
	}
}

func TestUploadPDFHandler(t *testing.T) {
	srv := newTestServer()

	originalOpenPDF := openPDF
	defer func() { openPDF = originalOpenPDF }()

	t.Run("success", func(t *testing.T) {
		openPDF = func(path string) (*os.File, PDFReader, error) {
			return nil, &fakePDFReader{text: "Text extracted from PDF"}, nil
		}

		var buf bytes.Buffer
		writer := multipart.NewWriter(&buf)

		part, err := writer.CreateFormFile("file", "test.pdf")
		if err != nil {
			t.Fatalf("failed to create form file: %v", err)
		}
		part.Write([]byte("dummy pdf bytes"))
		writer.Close()

		req := httptest.NewRequest(http.MethodPost, "/upload-pdf", &buf)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		w := httptest.NewRecorder()

		logs := captureLogs(t, func() {
			srv.uploadPDFHandler(w, req)
		})

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}

		if !strings.Contains(logs, "upload_pdf=") {
			t.Fatalf("expected upload pdf log, got %q", logs)
		}
	})

	t.Run("method_not_allowed", func(t *testing.T) {
		// openPDF not used in this path
		req := httptest.NewRequest(http.MethodGet, "/upload-pdf", nil)
		w := httptest.NewRecorder()

		logs := captureLogs(t, func() {
			srv.uploadPDFHandler(w, req)
		})

		if w.Code != http.StatusMethodNotAllowed {
			t.Fatalf("expected 405, got %d", w.Code)
		}

		if strings.TrimSpace(logs) != "" {
			t.Fatalf("expected no logs, got %q", logs)
		}
	})

	t.Run("missing_file_field", func(t *testing.T) {
		// openPDF not used in this path
		var buf bytes.Buffer
		writer := multipart.NewWriter(&buf)
		// no file field added here
		if err := writer.Close(); err != nil {
			t.Fatalf("failed to close writer: %v", err)
		}

		req := httptest.NewRequest(http.MethodPost, "/upload-pdf", &buf)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		w := httptest.NewRecorder()

		logs := captureLogs(t, func() {
			srv.uploadPDFHandler(w, req)
		})

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 for missing file, got %d", w.Code)
		}

		if strings.TrimSpace(logs) != "" {
			t.Fatalf("expected no logs for missing file, got %q", logs)
		}
	})

	t.Run("no_text_extracted", func(t *testing.T) {
		openPDF = func(path string) (*os.File, PDFReader, error) {
			return nil, &fakePDFReader{text: ""}, nil
		}

		var buf bytes.Buffer
		writer := multipart.NewWriter(&buf)
		part, err := writer.CreateFormFile("file", "empty.pdf")
		if err != nil {
			t.Fatalf("failed to create form file: %v", err)
		}
		part.Write([]byte("dummy"))
		writer.Close()

		req := httptest.NewRequest(http.MethodPost, "/upload-pdf", &buf)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		w := httptest.NewRecorder()

		logs := captureLogs(t, func() {
			srv.uploadPDFHandler(w, req)
		})

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 for no text extracted, got %d", w.Code)
		}

		// handler does not log anything in this path
		if strings.TrimSpace(logs) != "" {
			t.Fatalf("expected no logs for no-text case, got %q", logs)
		}
	})

	t.Run("pdf_too_big", func(t *testing.T) {
		longText := strings.Repeat("This is a sentence that will be chunked from pdf. ", 500)

		openPDF = func(path string) (*os.File, PDFReader, error) {
			return nil, &fakePDFReader{text: longText}, nil
		}

		var buf bytes.Buffer
		writer := multipart.NewWriter(&buf)
		part, err := writer.CreateFormFile("file", "big.pdf")
		if err != nil {
			t.Fatalf("failed to create form file: %v", err)
		}
		part.Write([]byte("dummy pdf bytes"))
		writer.Close()

		req := httptest.NewRequest(http.MethodPost, "/upload-pdf", &buf)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		w := httptest.NewRecorder()

		logs := captureLogs(t, func() {
			srv.uploadPDFHandler(w, req)
		})

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 for pdf too big, got %d", w.Code)
		}

		// In this path, we log both upload_pdf and the error
		if !strings.Contains(logs, "upload_pdf=") {
			t.Fatalf("expected upload_pdf log, got %q", logs)
		}
		if !strings.Contains(logs, "error - pdf too big") {
			t.Fatalf("expected 'error - pdf too big' log, got %q", logs)
		}
	})
}

func TestQueryHandler(t *testing.T) {
	srv := newTestServer()

	t.Run("success", func(t *testing.T) {
		// Seed store with one chunk so query returns something
		srv.store.Add(rag.Chunk{
			Content:   "hello world",
			Embedding: srv.embedder.Embed("hello world"),
			Source:    "doc1",
		})

		payload := `{"query":"hello"}`
		req := httptest.NewRequest(http.MethodPost, "/query", strings.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		logs := captureLogs(t, func() {
			srv.queryHandler(w, req)
		})

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}

		if !strings.Contains(logs, `query="hello"`) {
			t.Fatalf("expected query log, got %q", logs)
		}
	})

	t.Run("empty_query", func(t *testing.T) {
		body := `{"query":""}`
		req := httptest.NewRequest(http.MethodPost, "/query", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		logs := captureLogs(t, func() {
			srv.queryHandler(w, req)
		})

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 for empty query, got %d", w.Code)
		}

		// No logs are written for the "empty query" error path
		if strings.TrimSpace(logs) != "" {
			t.Fatalf("expected no logs for empty query error, got %q", logs)
		}
	})
}

func TestResetHandler(t *testing.T) {
	srv := newTestServer()

	t.Run("success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/reset", nil)
		w := httptest.NewRecorder()

		logs := captureLogs(t, func() {
			srv.resetHandler(w, req)
		})

		if w.Code != http.StatusNoContent {
			t.Fatalf("expected 204, got %d", w.Code)
		}

		if strings.TrimSpace(logs) != "" {
			t.Fatalf("expected no logs, got %q", logs)
		}
	})

	t.Run("method_not_allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/reset", nil)
		w := httptest.NewRecorder()

		logs := captureLogs(t, func() {
			srv.resetHandler(w, req)
		})

		if w.Code != http.StatusMethodNotAllowed {
			t.Fatalf("expected 405, got %d", w.Code)
		}

		if strings.TrimSpace(logs) != "" {
			t.Fatalf("expected no logs, got %q", logs)
		}
	})
}
