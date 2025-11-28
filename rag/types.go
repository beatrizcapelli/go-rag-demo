package rag

// Chunk of a document
type Chunk struct {
	ID        string
	Content   string
	Source    string // filename or doc ID
	Embedding []float64
}

// Simple query result
type SearchResult struct {
	Chunk    Chunk
	Score    float64
}