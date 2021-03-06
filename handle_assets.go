package caddy_esbuild_plugin

import (
	"fmt"
	"github.com/evanw/esbuild/pkg/api"
	"go.uber.org/zap"
	"mime"
	"net/http"
	"path/filepath"
)

func (m *Esbuild) handleAsset(w http.ResponseWriter, r *http.Request, f api.OutputFile) error {
	cachedETag := r.Header.Get("If-None-Match")
	if cachedETag == m.hashes[f.Path] {
		w.WriteHeader(304) //No change
		return nil
	}

	w.Header().Set("ETag", m.hashes[f.Path])
	w.Header().Set("Content-type", guessContentType(f.Path))

	if m.FileHash {
		w.Header().Set("Cache-Control", "public,max-age=31536000")
	}

	w.WriteHeader(200)
	_, _ = w.Write(f.Contents)
	m.logger.Debug(fmt.Sprintf("esbuild handled %s", r.RequestURI), zap.String("source", f.Path))
	return nil
}

func guessContentType(path string) string {
	ext := filepath.Ext(path)
	contentType := mime.TypeByExtension(ext)

	if contentType != "" {
		return contentType
	}

	return "application/javascript"
}
