//go:build cgo
// +build cgo

package caddy_esbuild_plugin

import (
	"bytes"
	"fmt"
	"github.com/evanw/esbuild/pkg/api"
	"github.com/fsnotify/fsnotify"
	libsass "github.com/wellington/go-libsass"
	"go.uber.org/zap"
	"os"
	"path/filepath"
)

func (m *Esbuild) hasSassSupport() bool {
	return true
}

func (m *Esbuild) createSassPlugin() *api.Plugin {
	return &api.Plugin{
		Name: "sass",
		Setup: func(build api.PluginBuild) {
			// Load ".txt" files and return an array of words
			build.OnLoad(api.OnLoadOptions{Filter: `\.scss|sass$`},
				func(args api.OnLoadArgs) (api.OnLoadResult, error) {
					fileHandle, err := os.Open(args.Path)
					if err != nil {
						return api.OnLoadResult{}, fmt.Errorf("sass: unable to open file: %s", err)
					}

					buffer := new(bytes.Buffer)
					comp, err := libsass.New(buffer, fileHandle)
					if err != nil {
						return api.OnLoadResult{}, fmt.Errorf("sass: unable to load libsass: %s", err)
					}

					cwd, _ := os.Getwd()

					globSearch := filepath.Join(cwd, "*", "node_modules")
					nodeModules, err := filepath.Glob(globSearch)

					err = comp.Option(libsass.Path(args.Path), libsass.IncludePaths(append(nodeModules, filepath.Dir(args.Path), cwd)))
					if err != nil {
						return api.OnLoadResult{}, fmt.Errorf("sass: unable to set include path: %s", err)
					}

					err = comp.Run()
					if err != nil {
						return api.OnLoadResult{}, fmt.Errorf("sass: unable to compile: %s", err)
					}
					files := comp.Imports()
					go m.watchFiles(files)

					contents := buffer.String()
					return api.OnLoadResult{
						Contents: &contents,
						Loader:   api.LoaderCSS,
					}, nil
				})
		},
	}
}

func (m *Esbuild) watchFiles(files []string) {

	// creates a new file watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Println("ERROR", err)
	}

	//
	done := make(chan bool)
	defer watcher.Close()
	defer close(done)

	//
	go func() {
		for {
			select {
			// watch for events
			case event := <-watcher.Events:
				if event.Op == fsnotify.Write {
					done <- true
					m.logger.Debug("sass-file changed, rebuilding", zap.String("filename", event.Name), zap.Stringer("op", event.Op))
					m.Rebuild()
					return
				}
			case err := <-watcher.Errors:
				m.logger.Error("sass: failed to watch files!", zap.Error(err))
			case <-m.globalQuit:
				done <- true
			}
		}
	}()

	for _, file := range files {
		if err := watcher.Add(file); err != nil {
			m.logger.Error("sass: failed to add file", zap.Error(err), zap.String("file", file))
		}
	}

	<-done
}
