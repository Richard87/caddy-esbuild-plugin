package caddy_esbuild_plugin

import (
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"github.com/evanw/esbuild/pkg/api"
	"path/filepath"
	"strings"
)

func init() {
	httpcaddyfile.RegisterHandlerDirective("esbuild", parseCaddyfileEsbuild)
}

// parseCaddyfileEsbuild sets up a basic rewrite handler from Caddyfile tokens. Syntax:
//
//     esbuild [source]
//     esbuild ./assets/index.js {
//        auto_reload
//        sass
//        target /_build
//     }
//
//     sass requires cgo to work
//
// Only URI components which are given in <to> will be set in the resulting URI.
// See the docs for the rewrite handler for more information.
func parseCaddyfileEsbuild(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
	if !h.Next() {
		return nil, h.ArgErr()
	}

	var esbuild Esbuild

	esbuild.Target = "/_build"
	esbuild.LiveReload = false
	esbuild.Scss = false
	esbuild.FileHash = false
	esbuild.Loader = make(map[string]string)
	esbuild.Defines = make(map[string]string)
	esbuild.Loader[".png"] = "file"
	esbuild.Loader[".svg"] = "file"
	esbuild.Loader[".js"] = "jsx"

	for h.NextArg() {
		val := h.Val()
		switch val {
		case "file_hash":
			esbuild.FileHash = true
		case "live_reload":
			esbuild.LiveReload = true
		case "scss":
			if esbuild.hasSassSupport() == false {
				return nil, h.Err("sass requires caddy to be compiled with CGO and libsass available")
			}
			esbuild.Scss = true
		case "env":
			esbuild.Env = true
		default:
			alias := parseSourceName(val)

			esbuild.Sources = append(esbuild.Sources, api.EntryPoint{
				OutputPath: alias,
				InputPath:  val,
			})
		}
	}

	for nesting := h.Nesting(); h.NextBlock(nesting); {
		switch h.Val() {
		case "file_hash":
			esbuild.FileHash = true
		case "live_reload":
			esbuild.LiveReload = true
		case "scss":
			if esbuild.hasSassSupport() == false {
				return nil, h.Err("sass requires caddy to be compiled with CGO and libsass available")
			}
			esbuild.Scss = true
		case "env":
			esbuild.Env = true
		case "source":
			if !h.NextArg() {
				return nil, h.Err("source requires asset filename: source ./src/index.js")
			}

			source := h.Val()
			alias := parseSourceName(source)
			if h.NextArg() {
				alias = h.Val()
			}

			esbuild.Sources = append(esbuild.Sources, api.EntryPoint{
				OutputPath: alias,
				InputPath:  source,
			})
		case "target":
			if !h.NextArg() {
				return nil, h.Err("source requires path: target /build")
			}

			target := h.Val()
			if !strings.HasPrefix(target, "/") {
				target = "/" + target
			}
			if strings.HasSuffix(target, "/") {
				target = strings.TrimSuffix(target, "/")
			}

			esbuild.Target = target
		case "loader":
			if !h.NextArg() {
				return nil, h.Err("loader require filetype and loader: loader .svg text")
			}
			filetype := h.Val()

			if !h.NextArg() {
				return nil, h.Err("loader require filetype and loader: loader .svg text")
			}
			loaderValue := h.Val()

			esbuild.Loader[filetype] = loaderValue
		case "define":
			if !h.NextArg() {
				return nil, h.Err("loader require filetype and loader: loader .svg text")
			}
			define := h.Val()

			if !h.NextArg() {
				return nil, h.Err("loader require filetype and loader: loader .svg text")
			}
			value := h.Val()

			esbuild.Defines[define] = value
		}

	}

	return &esbuild, nil
}

func parseSourceName(source string) string {
	alias := filepath.Base(source)
	alias = strings.TrimSuffix(alias, filepath.Ext(alias))
	return alias
}
