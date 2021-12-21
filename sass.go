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
	"strings"
)

var sassPlugin = &api.Plugin{
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
				//fmt.Printf("CURRENT: %s\n", args.Path)
				//fmt.Printf("CWD: %s\n", cwd)

				globSearch := filepath.Join(cwd, "*", "node_modules")
				nodeModules, err := filepath.Glob(globSearch)

				resolver := func(url string, prev string) (newURL string, body string, resolved bool) {
					if strings.HasPrefix("https://", strings.ToLower(url)) {
						return url, "", true
					}

					var path string
					if prev == "" || prev == "stdin" {
						path = filepath.Dir(args.Path)
					} else {
						path = filepath.Dir(prev)
					}
					//fmt.Printf("***********************************************\n")
					//fmt.Printf("Source: %s\n", args.Path)
					//fmt.Printf("Import %s\n", url)
					//fmt.Printf("Prev path: %s\n", prev)
					//fmt.Printf("Using path: %s\n", path)

					relativeFile := filepath.Join(path, url)
					if _, err := os.Stat(relativeFile); err == nil {
						//fmt.Printf("\nFOUND: %s\n", relativeFile)
						return relativeFile, "", true
					}

					relativeFileWithoutExt := strings.TrimSuffix(relativeFile, filepath.Ext(relativeFile))
					glob, _ := filepath.Glob(relativeFileWithoutExt + ".*ss")
					for _, match := range glob {
						//fmt.Printf("\nFOUND: %s\n", match)
						return match, "", true
					}

					// Look for underscore files
					dir := filepath.Dir(relativeFileWithoutExt)
					filename := filepath.Base(relativeFileWithoutExt)
					glob, _ = filepath.Glob(filepath.Join(dir, "_"+filename) + ".*ss")
					for _, match := range glob {
						//fmt.Printf("\nFOUND: %s\n", match)
						return match, "", true
					}

					for _, match := range nodeModules {

						// Node Modules
						relativeFile = filepath.Join(match, url)
						if _, err := os.Stat(relativeFile); err == nil {
							//fmt.Printf("\nFOUND: %s\n", relativeFile)
							return relativeFile, "", true
						}

						relativeFileWithoutExt = strings.TrimSuffix(relativeFile, filepath.Ext(relativeFile))
						glob, _ = filepath.Glob(relativeFileWithoutExt + ".*ss")
						for _, match := range glob {
							//fmt.Printf("\nFOUND: %s\n", match)
							return match, "", true
						}
					}

					return "", "", false
				}
				err = comp.Option(libsass.ImportsOption(libsass.NewImportsWithResolver(resolver)))

				err = comp.Run()
				if err != nil {
					return api.OnLoadResult{}, fmt.Errorf("sass: unable to compile: %s", err)
				}

				contents := buffer.String()
				return api.OnLoadResult{
					Contents: &contents,
					Loader:   api.LoaderCSS,
				}, nil
			})
	},
}
