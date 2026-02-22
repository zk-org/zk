package embedding

import "unicode/utf8"

// ChunkText splits content with a fixed-size overlapping window.
func ChunkText(text string, size int, overlap int, maxChunks int) []string {
	if text == "" || size <= 0 || maxChunks <= 0 {
		return []string{}
	}

	runes := []rune(text)
	if len(runes) <= size {
		return []string{text}
	}

	step := size - overlap
	if step <= 0 {
		step = 1
	}

	chunks := make([]string, 0, maxChunks)
	for start := 0; start < len(runes) && len(chunks) < maxChunks; start += step {
		end := start + size
		if end > len(runes) {
			end = len(runes)
		}
		chunk := string(runes[start:end])
		if utf8.RuneCountInString(chunk) > 0 {
			chunks = append(chunks, chunk)
		}
		if end == len(runes) {
			break
		}
	}
	return chunks
}
