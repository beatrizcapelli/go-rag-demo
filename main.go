package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"go-rag-demo/rag"
	"os"
	"bytes"
	"github.com/ledongthuc/pdf"

)

type Server struct {
	store    *rag.InMemoryStore
	embedder rag.Embedder
}

func NewServer() *Server {
	return &Server{
		store:    rag.NewInMemoryStore(),
		embedder: rag.NewSimpleEmbedder(),
	}
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "ok")
}

// POST /upload  (body: raw text for now)
func (s *Server) uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}
	text := string(body)
	if text == "" {
		http.Error(w, "empty body", http.StatusBadRequest)
		return
	}

	chunks := rag.ChunkText(text, "doc1", s.embedder)
	s.store.Add(chunks...)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"chunks_added": len(chunks),
	})
}

func (s *Server) uploadPDFHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Max 10MB for safety
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "failed to parse form", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "missing file field", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Save to temp file because pdf library works with file paths
	tmp, err := os.CreateTemp("", "upload-*.pdf")
	if err != nil {
		http.Error(w, "failed to create temp file", http.StatusInternalServerError)
		return
	}
	defer os.Remove(tmp.Name())
	defer tmp.Close()

	if _, err := io.Copy(tmp, file); err != nil {
		http.Error(w, "failed to save temp pdf", http.StatusInternalServerError)
		return
	}

	f, rdr, err := pdf.Open(tmp.Name())
	if err != nil {
		http.Error(w, "failed to open pdf", http.StatusInternalServerError)
		return
	}
	defer f.Close()

	var buf bytes.Buffer
	b, err := rdr.GetPlainText()
	if err != nil {
		http.Error(w, "failed to read pdf text", http.StatusInternalServerError)
		return
	}
	if _, err := io.Copy(&buf, b); err != nil {
		http.Error(w, "failed to read pdf buffer", http.StatusInternalServerError)
		return
	}

	text := buf.String()
	if text == "" {
		http.Error(w, "no text extracted from pdf", http.StatusBadRequest)
		return
	}

	source := header.Filename
	chunks := rag.ChunkText(text, source, s.embedder)
	s.store.Add(chunks...)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"chunks_added": len(chunks),
		"filename":     source,
	})
}


type queryRequest struct {
	Query string `json:"query"`
}

// POST /query  { "query": "your question" }
func (s *Server) queryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req queryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if req.Query == "" {
		http.Error(w, "query is required", http.StatusBadRequest)
		return
	}

	qEmbedding := s.embedder.Embed(req.Query)
	results := s.store.Search(qEmbedding, 3)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func main() {
	srv := NewServer()

	http.HandleFunc("/health", srv.healthHandler)
	http.HandleFunc("/upload", srv.uploadHandler)
	http.HandleFunc("/query", srv.queryHandler)
	http.HandleFunc("/upload-pdf", srv.uploadPDFHandler)


	fs := http.FileServer(http.Dir("./frontend"))
    http.Handle("/", fs)

	fmt.Println("Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
