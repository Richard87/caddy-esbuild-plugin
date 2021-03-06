package caddy_esbuild_plugin

import (
	"fmt"
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"github.com/evanw/esbuild/pkg/api"
	"go.uber.org/zap"
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

type Esbuild struct {
	Target     string            `json:"target,omitempty"`
	LiveReload bool              `json:"auto_reload,omitempty"`
	Scss       bool              `json:"scss,omitempty"`
	Env        bool              `json:"env,omitempty"`
	Loader     map[string]string `json:"loader,omitempty"`
	FileHash   bool              `json:"file_hash,omitempty"`
	Defines    map[string]string `json:"defines,omitempty"`
	Sources    []api.EntryPoint  `json:"source,omitempty"`
	NodePaths  []string          `json:"n_ode_paths,omitempty"`

	logger       *zap.Logger
	esbuild      *api.BuildResult
	hashes       map[string]string
	globalQuit   chan struct{}
	lastDuration *time.Duration
	metafile     *Metafile
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
	m.Defines = make(map[string]string)
	m.initEsbuild()

	var sources []string
	for _, s := range m.Sources {
		sources = append(sources, s.InputPath)
	}
	var loaders []string
	for ext, l := range m.Loader {
		loaders = append(loaders, ext+"="+l)
	}

	m.logger.Info("Initialized esbuild",
		zap.String("target", m.Target),
		zap.Strings("sources", sources),
		zap.Strings("loaders", loaders),
		zap.Bool("sass", m.Scss),
		zap.Bool("env", m.Env),
		zap.Bool("live_reload", m.LiveReload),
		zap.Strings("node_path", m.NodePaths))
	return nil
}

// Validate implements caddy.Validator.
func (m *Esbuild) Validate() error {
	if len(m.Sources) == 0 {
		return fmt.Errorf("no source file")
	}

	for _, l := range m.Loader {
		_, err := ParseLoader(l)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *Esbuild) ServeHTTP(w http.ResponseWriter, r *http.Request, h caddyhttp.Handler) error {
	if r.Method != "GET" {
		return h.ServeHTTP(w, r)
	}

	outdir := m.Target
	if outdir == "" {
		outdir = "/_build"
	}

	file := r.RequestURI
	if index := strings.Index(file, "?"); index > 1 {
		file = file[:index-1]
	}

	if file == outdir+"/__livereload" {
		_ = m.handleLiveReload(w, r)
		return nil
	}
	if file == outdir+"/manifest.json" {
		_ = m.handleManifest(w, r)
		return nil
	}

	if m.esbuild != nil {
		for _, f := range m.esbuild.OutputFiles {
			if file == f.Path {
				return m.handleAsset(w, r, f)
			}
		}

		if m.Target == "" {
			for target, output := range m.metafile.Outputs {
				entrypoint := output.EntryPoint
				if !strings.HasPrefix(entrypoint, "/") {
					entrypoint = "/" + entrypoint
				}
				target, _ = filepath.Abs(target)
				m.logger.Debug("Handling output files",
					zap.String("file", file),
					zap.String("entrypoint", entrypoint),
					zap.String("target", target))

				if file == entrypoint && file != "/" {
					for _, f := range m.esbuild.OutputFiles {
						if target == f.Path {
							return m.handleAsset(w, r, f)
						}
					}
				}
			}
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
