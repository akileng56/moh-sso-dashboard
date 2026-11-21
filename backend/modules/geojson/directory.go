package geojson

import (
	"context"
	"embed"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

//go:embed maps/*.geojson
var bundledMaps embed.FS

// BundledAssets supplies the reviewed map files shipped inside this module.
// No dashboard asset directory, working-directory convention or download is
// needed. Each request owns an independent, seekable stream.
func BundledAssets() AssetReader {
	maps, err := fs.Sub(bundledMaps, "maps")
	if err != nil {
		panic(err)
	} // The compile-time embed pattern guarantees maps.
	return filesystemReader{files: maps}
}

// NewDirectoryReader is an optional operator-selected override. BasePath must
// be trusted deployment configuration, never a request parameter.
func NewDirectoryReader(basePath string) AssetReader {
	return filesystemReader{files: os.DirFS(basePath)}
}

type filesystemReader struct{ files fs.FS }

func (r filesystemReader) Open(ctx context.Context, filename string) (Asset, error) {
	if err := ctx.Err(); err != nil {
		return Asset{}, err
	}
	if filename == "" || filename == "." || filename == ".." || filepath.Base(filename) != filename || strings.ContainsAny(filename, "/\\") {
		return Asset{}, errors.New("asset filename must be a single path component")
	}
	file, err := r.files.Open(filename)
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
	stream, ok := file.(io.ReadSeekCloser)
	if !ok {
		_ = file.Close()
		return Asset{}, errors.New("asset must support seek")
	}
	return Asset{Content: stream, ModifiedAt: stat.ModTime()}, nil
}
