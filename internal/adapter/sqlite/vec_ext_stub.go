//go:build !vec

package sqlite

import "fmt"

func registerVecExtension() {}

func vecBuildEnabled() bool { return false }

func serializeEmbedding(vec []float32) ([]byte, error) {
	return nil, fmt.Errorf("semantic search requires building zk with -tags vec")
}
