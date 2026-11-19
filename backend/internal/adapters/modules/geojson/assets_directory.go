package geojson

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

type directoryReader struct {
	basePath string
}

// NewDirectoryReader adapts host-owned map files to the asset-reader contract.
func NewDirectoryReader(basePath string) AssetReader {
	return &directoryReader{basePath: basePath}
}

func (r *directoryReader) Open(ctx context.Context, filename string) (Asset, error) {
	if err := ctx.Err(); err != nil {
		return Asset{}, err
	}
	if filename == "" || filename == "." || filename == ".." ||
		filepath.Base(filename) != filename || strings.ContainsAny(filename, "/\\") {
		return Asset{}, errors.New("asset filename must be a single path component")
	}
	file, err := os.Open(filepath.Join(r.basePath, filename))
	if err != nil {
		return Asset{}, err
	}
	stat, err := file.Stat()
	if err != nil || !stat.Mode().IsRegular() {
		_ = file.Close()
		if err != nil {
			return Asset{}, err
		}
		return Asset{}, errors.New("asset must be a regular file")
	}
	return Asset{Content: file, ModifiedAt: stat.ModTime()}, nil
}
