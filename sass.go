//go:build cgo
// +build cgo

package caddy_esbuild_plugin

import (
	"bytes"
	"fmt"
	"github.com/evanw/esbuild/pkg/api"
	libsass "github.com/wellington/go-libsass"
	"os"
)

var sassPlugin = &api.Plugin{
	Name: "sass",
	Setup: func(build api.PluginBuild) {
		// Load ".txt" files and return an array of words
		build.OnLoad(api.OnLoadOptions{Filter: `\.scss|sass$`},
			func(args api.OnLoadArgs) (api.OnLoadResult, error) {
				f, err := os.Open(args.Path)
				if err != nil {
					return api.OnLoadResult{}, fmt.Errorf("sass: unable to open file: %s", err)
				}

				buf := new(bytes.Buffer)
				comp, err := libsass.New(buf, f)
				if err != nil {
					return api.OnLoadResult{}, fmt.Errorf("sass: unable to load libsass: %s", err)
				}

				err = comp.Run()
				if err != nil {
					return api.OnLoadResult{}, fmt.Errorf("sass: unable to compile: %s", err)
				}

				contents := buf.String()
				return api.OnLoadResult{
					Contents: &contents,
					Loader:   api.LoaderCSS,
				}, nil
			})
	},
}
