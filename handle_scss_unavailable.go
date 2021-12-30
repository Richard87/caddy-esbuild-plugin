//go:build !cgo
// +build !cgo

package caddy_esbuild_plugin

import (
	"github.com/evanw/esbuild/pkg/api"
)

func (m *Esbuild) createSassPlugin() *api.Plugin {
	return nil
}
func (m *Esbuild) hasSassSupport() bool {
	return false
}
