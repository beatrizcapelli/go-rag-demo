package rag

import (
	"context"
	"log"
	"os"

	openai "github.com/sashabaranov/go-openai"
)

// Embedder is an interface so later you can swap implementation.
type Embedder interface {
	Embed(text string) []float64
}

// ---- Simple fake embedder (old one, kept for reference/testing) ----

type SimpleEmbedder struct{}

func NewSimpleEmbedder() *SimpleEmbedder {
	return &SimpleEmbedder{}
}

func (e *SimpleEmbedder) Embed(text string) []float64 {
	// Fake 4D vector: length, vowels, consonants, spaces.
	var length, vowels, consonants, spaces float64
	for _, r := range text {
		length++
		switch {
		case r == 'a' || r == 'e' || r == 'i' || r == 'o' || r == 'u' ||
			r == 'A' || r == 'E' || r == 'I' || r == 'O' || r == 'U':
			vowels++
		case r == ' ':
			spaces++
		default:
			consonants++
		}
	}
	return []float64{length, vowels, consonants, spaces}
}

// ---- Real OpenAI embedder ----

type OpenAIEmbedder struct {
	client *openai.Client
	model  openai.EmbeddingModel
}

// NewOpenAIEmbedder uses OPENAI_API_KEY from the environment
// and the text-embedding-3-small model.
func NewOpenAIEmbedder() *OpenAIEmbedder {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Println("[OpenAIEmbedder] WARNING: OPENAI_API_KEY not set; Embed() will fail")
	}

	client := openai.NewClient(apiKey)

	return &OpenAIEmbedder{
		client: client,
		model:  openai.SmallEmbedding3, // "text-embedding-3-small"
	}
}

func (e *OpenAIEmbedder) Embed(text string) []float64 {
	if text == "" {
		return nil
	}

	req := openai.EmbeddingRequestStrings{
		Input: []string{text},
		Model: e.model,
	}

	resp, err := e.client.CreateEmbeddings(context.Background(), req)
	if err != nil {
		log.Printf("[OpenAIEmbedder] error creating embedding: %v\n", err)
		return nil
	}

	if len(resp.Data) == 0 {
		return nil
	}

	embedding := resp.Data[0].Embedding // []float32
	out := make([]float64, len(embedding))
	for i, v := range embedding {
		out[i] = float64(v)
	}
	return out
}
