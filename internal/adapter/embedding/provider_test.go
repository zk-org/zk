package embedding

import (
	"testing"

	"github.com/zk-org/zk/internal/core"
	"github.com/zk-org/zk/internal/util/test/assert"
)

func TestNewProviderDisabledReturnsNil(t *testing.T) {
	provider, err := NewProvider(core.EmbeddingConfig{
		Enabled: false,
	})

	assert.Nil(t, err)
	assert.Nil(t, provider)
}

func TestNewProviderValidatesProvider(t *testing.T) {
	_, err := NewProvider(core.EmbeddingConfig{
		Enabled:  true,
		Provider: "unknown",
	})

	assert.Err(t, err, "unknown embedding.provider")
}

func TestNewProviderLocalRequiresEndpoint(t *testing.T) {
	_, err := NewProvider(core.EmbeddingConfig{
		Enabled:  true,
		Provider: "local",
		Model:    "text-embedding-3-small",
	})

	assert.Err(t, err, "embedding.endpoint is required when embedding.provider = \"local\"")
}

func TestNewProviderOpenAIRequiresAPIKeyEnvByDefault(t *testing.T) {
	t.Setenv("ZK_EMBEDDING_OPENAI_API_KEY", "")

	_, err := NewProvider(core.EmbeddingConfig{
		Enabled:  true,
		Provider: "openai",
		Model:    "text-embedding-3-small",
	})

	assert.Err(t, err, "embedding API key env ZK_EMBEDDING_OPENAI_API_KEY is not set")
}

func TestNewProviderGoogleAIRequiresAPIKeyEnvByDefault(t *testing.T) {
	t.Setenv("ZK_EMBEDDING_GOOGLE_API_KEY", "")

	_, err := NewProvider(core.EmbeddingConfig{
		Enabled:  true,
		Provider: "googleai",
		Model:    "text-embedding-004",
	})

	assert.Err(t, err, "embedding API key env ZK_EMBEDDING_GOOGLE_API_KEY is not set")
}

func TestNewProviderLocalBuildsOpenAICompatibleClient(t *testing.T) {
	provider, err := NewProvider(core.EmbeddingConfig{
		Enabled:  true,
		Provider: "local",
		Model:    "text-embedding-3-small",
		Endpoint: "http://localhost:11434/v1/",
	})

	assert.Nil(t, err)
	assert.NotNil(t, provider)

	openAI, ok := provider.(*openAIProvider)
	assert.True(t, ok)
	assert.Equal(t, openAI.endpoint, "http://localhost:11434/v1")
}
