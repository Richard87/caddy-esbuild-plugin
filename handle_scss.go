//go:build cgo
// +build cgo

package caddy_esbuild_plugin

import (
	"bytes"
	"fmt"
	"github.com/evanw/esbuild/pkg/api"
	libsass "github.com/wellington/go-libsass"
	"os"
	"path/filepath"
	"time"
)

func (m *Esbuild) hasSassSupport() bool {
	return true
}

type result struct {
	sources []string
	mtime   time.Time
	content string
}

func (m *Esbuild) createSassPlugin() *api.Plugin {
	return &api.Plugin{
		Name: "sass",
		Setup: func(build api.PluginBuild) {
			cache := make(map[string]result)
			// Load ".txt" files and return an array of words
			build.OnLoad(api.OnLoadOptions{Filter: `\.scss|sass$`},
				func(args api.OnLoadArgs) (api.OnLoadResult, error) {
					fileHandle, err := os.Open(args.Path)
					if err != nil {
						return api.OnLoadResult{}, fmt.Errorf("sass: unable to open file: %s", err)
					}

					if val, ok := cache[args.Path]; ok {
						latestMTime := getLatestMtime(val.sources)
						if latestMTime.Equal(val.mtime) {
							return api.OnLoadResult{
								Contents: &val.content,
								Loader:   api.LoaderCSS,
							}, nil
						}
					}

					contentBuffer := new(bytes.Buffer)
					comp, err := libsass.New(contentBuffer, fileHandle)
					if err != nil {
						return api.OnLoadResult{}, fmt.Errorf("sass: unable to load libsass: %s", err)
					}

					cwd, _ := os.Getwd()

					globSearch := filepath.Join(cwd, "*", "node_modules")
					nodeModules, err := filepath.Glob(globSearch)

					err = comp.Option(libsass.WithSyntax(libsass.SCSSSyntax), libsass.Path(args.Path), libsass.IncludePaths(append(nodeModules, filepath.Dir(args.Path), cwd)))
					if err != nil {
						return api.OnLoadResult{}, fmt.Errorf("sass: unable to set include path: %s", err)
					}

					err = comp.Run()
					if err != nil {
						return api.OnLoadResult{}, fmt.Errorf("sass: unable to compile: %s", err)
					}
					files := comp.Imports()
					go m.watchFiles(files)

					contents := contentBuffer.String()

					files = append(files, args.Path)
					latestMTime := getLatestMtime(files)

					cache[args.Path] = result{
						sources: files,
						content: contents,
						mtime:   latestMTime,
					}

					return api.OnLoadResult{
						Contents: &contents,
						Loader:   api.LoaderCSS,
					}, nil
				})
		},
	}
}

func getLatestMtime(files []string) time.Time {
	var latestMTime time.Time
	for _, path := range files {
		mtime, _ := os.Stat(path)
		modTime := mtime.ModTime()
		if modTime.After(latestMTime) {
			latestMTime = modTime
		}
	}
	return latestMTime
}
