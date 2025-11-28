# Go RAG Demo

A minimal Retrieval-Augmented Generation (RAG) system built in **Go**, featuring:

* ✅ Go backend API
* ✅ In-memory vector store
* ✅ Text & PDF upload
* ✅ Dark-mode frontend UI
* ✅ Unit tests
* ✅ Docker & Docker Compose support

---

## ✨ Features

* Upload raw text and PDF files
* Chunking + embedding pipeline
* Cosine similarity search
* Query interface with similarity scores
* Responsive dark-mode UI
* Fully testable architecture

---

## 🗂 Project Structure

```
go-rag-demo/
├── frontend/        # HTML UI (served by Go)
├── rag/             # Core RAG logic (chunking, store, embedding)
├── main.go          # HTTP server and handlers
├── go.mod
├── go.sum
├── Dockerfile
├── docker-compose.yml
└── README.md
```

---

## 🧪 Running tests

Run all tests with coverage:

```bash
go test -v -cover ./...
```

---

## 🧱 Run with Docker Compose (recommended)

### Start

```bash
docker compose up
```

### Stop

```bash
docker compose down
```

---

## 🔌 API Endpoints

### POST /upload

Upload raw text

```bash
curl -X POST http://localhost:8080/upload -d "your text here"
```

### POST /upload-pdf

Upload a PDF file

```bash
curl -X POST http://localhost:8080/upload-pdf \
  -F "file=@document.pdf"
```

### POST /query

Query indexed content

```bash
curl -X POST http://localhost:8080/query \
  -H "Content-Type: application/json" \
  -d '{"query": "your question"}'
```

---

## 🧠 Architecture Overview

```
Client
  ↓
Frontend (HTML)
  ↓
Go API
  ↓
Chunking → Embedding → InMemory Vector Store
  ↓
Cosine Similarity Search
```

Current embedder is a deterministic demo implementation. It can be replaced with real embeddings (OpenAI, Cohere, etc) without changing the rest of the architecture.

---

## ⚙️ Deployment

The project is using Google Cloud Platform - Cloud Run.

It has automatic deploy through Github Actions whenever there is a merge to "main" branch.

It uses Github Action Secrets - **GCP_PROJECT_ID, GCP_REGION, GCP_SA_KEY**.

---

## 👤 Author

Built by **Beatriz Capelli** as a portfolio AI backend project.

---

Feel free to fork, clone, and adapt this project for your own experiments 🚀

