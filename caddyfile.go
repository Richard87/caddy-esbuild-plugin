package caddy_esbuild_plugin

import (
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
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
	if !h.NextArg() {
		return nil, h.ArgErr()
	}

	var esbuild Esbuild
	esbuild.AutoReload = false
	esbuild.Target = "/_build"
	esbuild.Sass = false

	// read the prefix to strip
	esbuild.Source = h.Val()

	for nesting := h.Nesting(); h.NextBlock(nesting); {
		switch h.Val() {
		case "auto_reload":
			esbuild.AutoReload = true
		case "sass":
			if sassPlugin == nil {
				return nil, h.Err("sass requires caddy to be compiled with CGO and libsass available")
			}
			esbuild.Sass = true
		case "target":
			if !h.NextArg() {
				return nil, h.ArgErr()
			}
			esbuild.Target = h.Val()
		}
	}

	return &esbuild, nil
}
