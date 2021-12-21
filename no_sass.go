//go:build !cgo
// +build !cgo

package caddy_esbuild_plugin

import (
	"github.com/evanw/esbuild/pkg/api"
)

var sassPlugin = (*api.Plugin)(nil)
