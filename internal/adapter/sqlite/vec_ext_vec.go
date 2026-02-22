//go:build vec

package sqlite

import sqlite_vec "github.com/asg017/sqlite-vec-go-bindings/cgo"

func registerVecExtension() {
	sqlite_vec.Auto()
}

func vecBuildEnabled() bool { return true }

func serializeEmbedding(vec []float32) ([]byte, error) {
	return sqlite_vec.SerializeFloat32(vec)
}
