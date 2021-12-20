package caddy_esbuild_plugin

import (
	"fmt"
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"github.com/evanw/esbuild/pkg/api"
	"strings"

	"go.uber.org/zap"
	"net/http"
)

type Esbuild struct {
	Source string `json:"source,omitempty"`
	Target string `json:"target,omitempty"`

	logger  *zap.Logger
	esbuild api.BuildResult
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

	result := api.Build(api.BuildOptions{
		EntryPoints: []string{m.Source},
		Outfile:     m.Target,
		Write:       false,
		Watch: &api.WatchMode{
			OnRebuild: m.onBuild,
		},
	})
	m.onBuild(result)

	return nil
}

func (m *Esbuild) onBuild(result api.BuildResult) {
	for _, f := range result.OutputFiles {
		m.logger.Debug("Built file", zap.String("file", f.Path))
	}

	if len(result.Errors) > 0 {
		m.logger.Error(fmt.Sprintf("watch build failed: %d errors\n", len(result.Errors)))
	} else {
		m.logger.Info(fmt.Sprintf("watch build succeeded: %d warnings\n", len(result.Warnings)))
	}
	m.esbuild = result
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

	for _, f := range m.esbuild.OutputFiles {
		if strings.Index(r.RequestURI, f.Path) == 0 {
			w.Header().Set("Content-type", "application/javascript")
			w.WriteHeader(200)
			_, _ = w.Write(f.Contents)
			m.logger.Debug(fmt.Sprintf("esbuild handled %s", f.Path))
			return nil
		}
	}

	return h.ServeHTTP(w, r)
}

// Interface guards
var (
	_ caddy.Provisioner           = (*Esbuild)(nil)
	_ caddy.Validator             = (*Esbuild)(nil)
	_ caddyhttp.MiddlewareHandler = (*Esbuild)(nil)
)
