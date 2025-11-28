package rag

import (
    "strconv"
    "strings"
)

// Very naive chunker by number of sentences
func ChunkText(text, source string, embedder Embedder) []Chunk {
	sentences := strings.Split(text, ".")
	const maxSentencesPerChunk = 3

	var chunks []Chunk
	var buffer []string

	maybeFlush := func() {
		if len(buffer) == 0 {
			return
		}
		content := strings.TrimSpace(strings.Join(buffer, ". ") + ".")
		if content == "." {
			return
		}
		chunks = append(chunks, Chunk{
            ID:        source + "-" + strconv.Itoa(len(chunks)+1),
			Content:   content,
			Source:    source,
			Embedding: embedder.Embed(content),
		})
		buffer = []string{}
	}

	for _, s := range sentences {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		buffer = append(buffer, s)
		if len(buffer) >= maxSentencesPerChunk {
			maybeFlush()
		}
	}
	maybeFlush()

	return chunks
}
