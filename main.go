package caddy_esbuild_plugin

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"github.com/evanw/esbuild/pkg/api"
	"io/ioutil"
	"mime"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"go.uber.org/zap"
	"net/http"
)

type Esbuild struct {
	Source     string `json:"source,omitempty"`
	Target     string `json:"target,omitempty"`
	AutoReload bool   `json:"auto_reload,omitempty"`
	Sass       bool   `json:"sass,omitempty"`

	logger     *zap.Logger
	esbuild    *api.BuildResult
	hashes     map[string]string
	globalQuit chan struct{}
}

func (m *Esbuild) Cleanup() error {
	close(m.globalQuit)
	return nil
}

func init() {
	caddy.RegisterModule(Esbuild{})
}

// CaddyModule returns the Caddy module information.
func (Esbuild) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.esbuild",
		New: func() caddy.Module { return new(Esbuild) },
	}
}

func isJsFile(source string) bool {
	if strings.HasSuffix(source, ".js") {
		return true
	}
	if strings.HasSuffix(source, ".jsx") {
		return true
	}
	if strings.HasSuffix(source, ".ts") {
		return true
	}
	if strings.HasSuffix(source, ".tsx") {
		return true
	}

	return false
}

func (m *Esbuild) Provision(ctx caddy.Context) error {
	m.logger = ctx.Logger(m)
	m.hashes = make(map[string]string)
	m.globalQuit = make(chan struct{})
	var inject []string
	var plugins []api.Plugin

	if m.AutoReload && isJsFile(m.Source) {
		name, err := m.createAutoloadShimFile()
		if err != nil {
			m.logger.Error("Failed to create autoload shim", zap.Error(err))
		} else {
			inject = append(inject, name)
		}
	}

	if m.Sass {
		sassPlugin := m.createSassPlugin()
		if sassPlugin == nil {
			m.logger.Error("Failed to enable sass plugin, caddy must be compiled with CGO enabled!")
		} else {
			plugins = append(plugins, *sassPlugin)
		}
	}

	start := time.Now()
	result := api.Build(api.BuildOptions{
		EntryPoints: []string{m.Source},
		Sourcemap:   api.SourceMapLinked,
		Outdir:      m.Target,
		PublicPath:  m.Target,
		Metafile:    true,
		Write:       false,
		Bundle:      true,
		Inject:      inject,
		JSXMode:     api.JSXModeTransform,
		Plugins:     plugins,
		Incremental: true,
		Loader: map[string]api.Loader{
			".png": api.LoaderFile,
			".svg": api.LoaderFile,
			".js":  api.LoaderJSX,
		},
		Watch: &api.WatchMode{
			OnRebuild: func(result api.BuildResult) {
				m.logger.Debug("Rebuild completed!")
				m.onBuild(result, 0)
			},
		},
	})
	duration := time.Now().Sub(start)
	m.onBuild(result, duration)

	return nil
}

func (m *Esbuild) createAutoloadShimFile() (string, error) {
	livereload := "(() => {const es = new EventSource('%s/__livereload'); es.addEventListener('message', e => e.data === 'reload' && (es.close() || location.reload()))})()"
	file, err := ioutil.TempFile(os.TempDir(), "caddy-esbuild-shim-*.js")
	if err != nil {
		return "", fmt.Errorf("autoload: failed to create tmpfile: %s", err)
	}
	_, err = file.Write([]byte(fmt.Sprintf(livereload, m.Target)))
	if err != nil {
		return "", fmt.Errorf("autoload: failed to write shim: %s", err)
	}
	name := file.Name()
	return name, nil
}

func (m *Esbuild) onBuild(result api.BuildResult, duration time.Duration) {

	for _, err := range result.Errors {
		m.logger.Error(err.Text)
	}

	for _, f := range result.OutputFiles {
		m.logger.Debug("Built file", zap.String("file", f.Path))
		hasher := sha1.New()
		hasher.Write(f.Contents)
		m.hashes[f.Path] = hex.EncodeToString(hasher.Sum(nil))
	}
	if len(result.Errors) > 0 {
		m.logger.Error(fmt.Sprintf("watch build failed: %d errors\n", len(result.Errors)))
		return
	} else {
		m.logger.Info(fmt.Sprintf("watch build succeeded in %dms: %d warnings\n", duration.Milliseconds(), len(result.Warnings)))
	}
	m.esbuild = &result
}

// Validate implements caddy.Validator.
func (m *Esbuild) Validate() error {
	if m.Source == "" {
		return fmt.Errorf("no source file")
	}
	if m.Target == "" {
		return fmt.Errorf("no target file")
	}
	return nil
}

func (m *Esbuild) ServeHTTP(w http.ResponseWriter, r *http.Request, h caddyhttp.Handler) error {

	if r.Method != "GET" {
		return h.ServeHTTP(w, r)
	}

	if r.RequestURI == m.Target+"/__livereload" {
		_ = m.handleLiveReload(w, r)
		return nil
	}

	for _, f := range m.esbuild.OutputFiles {
		if strings.Index(r.RequestURI, f.Path) == 0 {
			return m.handleAsset(w, r, f)
		}
	}

	return h.ServeHTTP(w, r)
}

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

func (m *Esbuild) handleLiveReload(w http.ResponseWriter, r *http.Request) error {
	// Add headers needed for server-sent events (SSE):
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	flusher, ok := w.(http.Flusher)
	if !ok {
		m.logger.Debug("Your browser does not support server-sent events (SSE).")
		return nil
	} else {
		m.logger.Debug("LiveReload started")
	}

	var lastPointer = m.esbuild
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	for {
		compareTimeout := time.After(20 * time.Millisecond)
		pingTimeout := time.After(10 * time.Second)
		select {
		case <-r.Context().Done():
			return nil
		case <-m.globalQuit:
			return nil
		case <-sigs:
			return nil
		case <-compareTimeout:
			var currentPointer = m.esbuild
			if lastPointer != currentPointer {
				_, _ = fmt.Fprintf(w, "data: reload\n\n")
				flusher.Flush()
				lastPointer = m.esbuild
			}
		case <-pingTimeout:
			_, _ = fmt.Fprintf(w, "data: p\n\n")
			flusher.Flush()
		}
	}
}

func (m *Esbuild) Rebuild() {
	if m.esbuild != nil {
		start := time.Now()
		result := m.esbuild.Rebuild()
		duration := time.Now().Sub(start)
		m.onBuild(result, duration)
	}
}

func guessContentType(path string) string {
	ext := filepath.Ext(path)
	contentType := mime.TypeByExtension(ext)

	if contentType != "" {
		return contentType
	}

	return "application/javascript"
}

// Interface guards
var (
	_ caddy.Provisioner           = (*Esbuild)(nil)
	_ caddy.Validator             = (*Esbuild)(nil)
	_ caddy.CleanerUpper          = (*Esbuild)(nil)
	_ caddyhttp.MiddlewareHandler = (*Esbuild)(nil)
)
