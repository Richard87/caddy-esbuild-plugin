package caddy_esbuild_plugin

import (
	"fmt"
	"github.com/evanw/esbuild/pkg/api"
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
	w.WriteHeader(200)
	_, _ = w.Write(f.Contents)
	m.logger.Debug(fmt.Sprintf("esbuild handled %s", f.Path))
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
