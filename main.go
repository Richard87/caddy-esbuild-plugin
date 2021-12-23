package caddy_esbuild_plugin

import (
	"fmt"
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"github.com/evanw/esbuild/pkg/api"
	"go.uber.org/zap"
	"net/http"
	"strings"
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

func (m *Esbuild) Provision(ctx caddy.Context) error {
	m.logger = ctx.Logger(m)
	m.hashes = make(map[string]string)
	m.globalQuit = make(chan struct{})

	m.initEsbuild()
	return nil
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
	if r.RequestURI == m.Target+"/manifest.json" {
		_ = m.handleManifest(w, r)
		return nil
	}

	for _, f := range m.esbuild.OutputFiles {
		if strings.Index(r.RequestURI, f.Path) == 0 {
			return m.handleAsset(w, r, f)
		}
	}

	return h.ServeHTTP(w, r)
}

// Interface guards
var (
	_ caddy.Provisioner           = (*Esbuild)(nil)
	_ caddy.Validator             = (*Esbuild)(nil)
	_ caddy.CleanerUpper          = (*Esbuild)(nil)
	_ caddyhttp.MiddlewareHandler = (*Esbuild)(nil)
)
