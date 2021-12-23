package caddy_esbuild_plugin

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"go.uber.org/zap"
	"net/http"
	"path/filepath"
)

type Metafile struct {
	Inputs  map[string]Import `json:"inputs"`
	Outputs map[string]Output `json:"outputs"`
}

type Import struct {
	Bytes   int          `json:"bytes"`
	Imports []ImportFile `json:"imports"`
}
type ImportFile struct {
	Path string `json:"path"`
	Kind string `json:"kind"`
}

type Output struct {
	Imports    []interface{}    `json:"imports"`
	Exports    []interface{}    `json:"exports"`
	EntryPoint string           `json:"entryPoint,omitempty"`
	Inputs     map[string]Input `json:"inputs"`
	Bytes      int              `json:"bytes"`
}
type Input struct {
	BytesInOutput int `json:"bytesInOutput"`
}

func (m *Esbuild) handleManifest(w http.ResponseWriter, r *http.Request) error {

	if m.esbuild == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		_, _ = w.Write([]byte("{}"))
		return nil
	}

	sha := sha1.New()
	sha.Write([]byte(m.esbuild.Metafile))
	etag := hex.EncodeToString(sha.Sum(nil))
	cachedETag := r.Header.Get("If-None-Match")
	if cachedETag == etag {
		w.WriteHeader(304) //No change
		return nil
	}
	w.Header().Set("ETag", etag)
	w.Header().Set("Content-Type", "application/json")

	var metafile = Metafile{}
	if err := json.Unmarshal([]byte(m.esbuild.Metafile), &metafile); err != nil {
		m.logger.Error("Failed to build manifest.json", zap.Error(err))
		w.WriteHeader(500)
		return nil
	}

	manifest := make(map[string]string)

	for target, output := range metafile.Outputs {
		source := output.EntryPoint
		if source == "" {
			//target is in form ../../../../_build/index.js
			source, _ = filepath.Abs(target)
		}

		target, _ := filepath.Abs(target)
		manifest[source] = target
	}

	content, err := json.Marshal(manifest)
	if err != nil {
		m.logger.Warn("Failed to build manifest.json", zap.Error(err))
		w.WriteHeader(500)
		return nil
	}
	if _, err = w.Write(content); err != nil {
		m.logger.Warn("Failed to handle manifest.json", zap.Error(err))
		w.WriteHeader(500)
		return nil
	}

	w.WriteHeader(200)
	return nil
}

/*
{
  "build/form_task.css": "http://localhost:8080/build/form_task.css",
  "build/form_task.js": "http://localhost:8080/build/form_task.js",
  "build/search.js": "http://localhost:8080/build/search.js",
}
*/
