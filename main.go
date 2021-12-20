package caddy_esbuild_plugin

import (
	"fmt"
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"io/ioutil"
	"strings"

	"go.uber.org/zap"
	"net/http"
)

type Esbuild struct {
	Source string `json:"source,omitempty"`
	Target string `json:"target,omitempty"`

	logger *zap.Logger
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
	m.logger = ctx.Logger(m) // g.logger is a *zap.Logger

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
	m.logger.Debug(fmt.Sprint(m))
	return nil
}

func (m *Esbuild) ServeHTTP(w http.ResponseWriter, r *http.Request, h caddyhttp.Handler) error {

	if r.Method != "GET" {
		return h.ServeHTTP(w, r)
	}

	if strings.Index(r.RequestURI, m.Target) != 0 {
		return h.ServeHTTP(w, r)
	}

	fileBytes, err := ioutil.ReadFile(m.Source)
	if err != nil {
		m.logger.Error("Could not read file: " + err.Error())
		w.WriteHeader(500)
		_, _ = w.Write([]byte("Server error"))
		return nil
	}
	w.Header().Set("Content-type", "application/javascript")
	w.WriteHeader(200)
	sentBytes, _ := w.Write(fileBytes)

	m.logger.Debug(fmt.Sprintf("Sent %d", sentBytes))

	return nil
}

// Interface guards
var (
	_ caddy.Provisioner           = (*Esbuild)(nil)
	_ caddy.Validator             = (*Esbuild)(nil)
	_ caddyhttp.MiddlewareHandler = (*Esbuild)(nil)
)
