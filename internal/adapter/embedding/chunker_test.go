package embedding

import (
	"testing"

	"github.com/zk-org/zk/internal/util/test/assert"
)

func TestChunkTextReturnsEmptyForInvalidInputs(t *testing.T) {
	assert.Equal(t, ChunkText("", 5, 1, 10), []string{})
	assert.Equal(t, ChunkText("abc", 0, 1, 10), []string{})
	assert.Equal(t, ChunkText("abc", 5, 1, 0), []string{})
}

func TestChunkTextReturnsSingleChunkWhenShorterThanSize(t *testing.T) {
	assert.Equal(t, ChunkText("abc", 5, 1, 10), []string{"abc"})
}

func TestChunkTextUsesOverlapAndRespectsMaxChunks(t *testing.T) {
	assert.Equal(t,
		ChunkText("abcdefghij", 5, 2, 10),
		[]string{"abcde", "defgh", "ghij"},
	)

	assert.Equal(t,
		ChunkText("abcdefghij", 5, 2, 2),
		[]string{"abcde", "defgh"},
	)
}

func TestChunkTextIsRuneSafe(t *testing.T) {
	assert.Equal(t,
		ChunkText("a🙂b🙂c", 2, 1, 10),
		[]string{"a🙂", "🙂b", "b🙂", "🙂c"},
	)
}
