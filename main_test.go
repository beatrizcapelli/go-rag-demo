package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go-rag-demo/rag"
)

func newTestServer() *Server {
	return &Server{
		store:    rag.NewInMemoryStore(),
		embedder: rag.NewSimpleEmbedder(),
	}
}

func TestHealthHandler_OK(t *testing.T) {
	s := newTestServer()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	s.healthHandler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestUploadHandler_StoresChunks(t *testing.T) {
	s := newTestServer()

	body := "This is a test document. It has two sentences."
	req := httptest.NewRequest(http.MethodPost, "/upload", strings.NewReader(body))
	w := httptest.NewRecorder()

	s.uploadHandler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var data map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		t.Fatalf("failed to decode response json: %v", err)
	}

	val, ok := data["chunks_added"]
	if !ok {
		t.Fatalf("expected 'chunks_added' in response")
	}

	if n, ok := val.(float64); !ok || n < 1 {
		t.Fatalf("expected at least 1 chunk added, got %v", val)
	}
}

func TestUploadHandler_WrongMethod(t *testing.T) {
	s := newTestServer()

	req := httptest.NewRequest(http.MethodGet, "/upload", nil)
	w := httptest.NewRecorder()

	s.uploadHandler(w, req)

	if w.Result().StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405 for GET /upload, got %d", w.Result().StatusCode)
	}
}

func TestQueryHandler_NoBody(t *testing.T) {
	s := newTestServer()

	req := httptest.NewRequest(http.MethodPost, "/query", bytes.NewReader([]byte{}))
	w := httptest.NewRecorder()

	s.queryHandler(w, req)

	if w.Result().StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 for empty body, got %d", w.Result().StatusCode)
	}
}

func TestQueryHandler_ReturnsResults(t *testing.T) {
	s := newTestServer()

	// first upload some text
	uploadBody := "Go is great for concurrent services. It works well for APIs."
	uploadReq := httptest.NewRequest(http.MethodPost, "/upload", strings.NewReader(uploadBody))
	uploadRes := httptest.NewRecorder()
	s.uploadHandler(uploadRes, uploadReq)

	// now query
	q := map[string]string{"query": "concurrent services"}
	buf, _ := json.Marshal(q)

	req := httptest.NewRequest(http.MethodPost, "/query", bytes.NewReader(buf))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.queryHandler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var results []struct {
		Chunk struct {
			Content string `json:"Content"`
		} `json:"Chunk"`
		Score float64 `json:"Score"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		t.Fatalf("failed to decode results: %v", err)
	}

	if len(results) == 0 {
		t.Fatalf("expected at least one result")
	}

	if results[0].Chunk.Content == "" {
		t.Fatalf("expected chunk content to be non-empty")
	}
}
