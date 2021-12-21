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
//     esbuild [source] <target>
//     esbuild /build/index.js
//     esbuild ./assets/index.js /build/index.js
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

	// read the prefix to strip
	esbuild.Source = h.Val()

	if h.NextArg() {
		esbuild.Target = h.Val()
	} else {
		esbuild.Target = "/_build"
	}

	return &esbuild, nil
}
