package caddy_esbuild_plugin

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"github.com/evanw/esbuild/pkg/api"
	"go.uber.org/zap"
	"os"
	"path/filepath"
	"time"
)

type Process struct {
	Env map[string]string `json:"env"`
}

func (m *Esbuild) initEsbuild() {
	var inject []string
	var plugins []api.Plugin

	plugins = append(plugins, m.createTimingPlugin())

	if m.LiveReload {
		name, err := m.createAutoloadShimFile()
		if err != nil {
			m.logger.Error("Failed to create autoload shim", zap.Error(err))
		} else {
			inject = append(inject, name)
		}
	}

	if m.Sass {
		sassPlugin := m.createSassPlugin()
		if sassPlugin == nil {
			m.logger.Error("Failed to enable sass plugin, caddy must be compiled with CGO enabled!")
		} else {
			plugins = append(plugins, *sassPlugin)
		}
	}

	if m.Env {
		m.Defines["process"] = m.handleEnv()
	}

	start := time.Now()
	loader := map[string]api.Loader{}
	for ext, l := range m.Loader {
		parseLoader, _ := ParseLoader(l)
		loader[ext] = parseLoader
	}

	entryName := "[name]"
	if m.FileHash {
		entryName = "[name]-[hash]"
	}

	cwd, _ := os.Getwd()
	pattern := filepath.Join(cwd, "*", "node_modules")
	nodePaths, _ := filepath.Glob(pattern)

	result := api.Build(api.BuildOptions{
		EntryPoints: m.Sources,
		NodePaths:   nodePaths,
		Sourcemap:   api.SourceMapLinked,
		Outdir:      m.Target,
		EntryNames:  entryName,
		PublicPath:  m.Target,
		Define:      m.Defines,
		Metafile:    true,
		Write:       false,
		Bundle:      true,
		Inject:      inject,
		JSXMode:     api.JSXModeTransform,
		Plugins:     plugins,
		Incremental: true,
		Loader:      loader,
		Watch: &api.WatchMode{
			OnRebuild: func(result api.BuildResult) {
				m.logger.Debug("Rebuild completed!")
				m.onBuild(result, m.lastDuration)
			},
		},
	})
	duration := time.Now().Sub(start)
	m.onBuild(result, &duration)
}

func (m *Esbuild) onBuild(result api.BuildResult, duration *time.Duration) {

	m.esbuild = &result
	for _, err := range result.Errors {
		m.logger.Error(err.Text)
	}

	for _, f := range result.OutputFiles {
		m.logger.Debug("Built file", zap.String("file", f.Path))
		hasher := sha1.New()
		hasher.Write(f.Contents)
		m.hashes[f.Path] = hex.EncodeToString(hasher.Sum(nil))
	}
	if len(result.Errors) > 0 {
		m.logger.Error(fmt.Sprintf("watch build failed: %d errors\n", len(result.Errors)))
		return
	} else {
		m.logger.Info(fmt.Sprintf("watch build succeeded in %dms: %d warnings\n", duration.Milliseconds(), len(result.Warnings)))
	}
	m.esbuild = &result
}

func (m *Esbuild) Rebuild() {
	if m.esbuild != nil {
		start := time.Now()
		result := m.esbuild.Rebuild()
		duration := time.Now().Sub(start)
		m.onBuild(result, &duration)
	}
}

func ParseLoader(text string) (api.Loader, error) {
	switch text {
	case "js":
		return api.LoaderJS, nil
	case "jsx":
		return api.LoaderJSX, nil
	case "ts":
		return api.LoaderTS, nil
	case "tsx":
		return api.LoaderTSX, nil
	case "css":
		return api.LoaderCSS, nil
	case "json":
		return api.LoaderJSON, nil
	case "text":
		return api.LoaderText, nil
	case "base64":
		return api.LoaderBase64, nil
	case "dataurl":
		return api.LoaderDataURL, nil
	case "file":
		return api.LoaderFile, nil
	case "binary":
		return api.LoaderBinary, nil
	case "default":
		return api.LoaderDefault, nil
	default:
		return api.LoaderNone, fmt.Errorf("Invalid loader value: %q", text)
	}
}
