package caddy_esbuild_plugin

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"github.com/evanw/esbuild/pkg/api"
	"go.uber.org/zap"
	"strings"
	"time"
)

type Process struct {
	Env map[string]string `json:"env"`
}

func (m *Esbuild) initEsbuild() {
	var inject []string
	define := make(map[string]string)
	var plugins []api.Plugin

	plugins = append(plugins, m.createTimingPlugin())

	if m.AutoReload {
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
		define["process"] = m.handleEnv()
	}

	start := time.Now()
	result := api.Build(api.BuildOptions{
		EntryPoints: m.Sources,
		Sourcemap:   api.SourceMapLinked,
		Outdir:      m.Target,
		PublicPath:  m.Target,
		Define:      define,
		Metafile:    true,
		Write:       false,
		Bundle:      true,
		Inject:      inject,
		JSXMode:     api.JSXModeTransform,
		Plugins:     plugins,
		Incremental: true,
		Loader: map[string]api.Loader{
			".png": api.LoaderFile,
			".svg": api.LoaderFile,
			".js":  api.LoaderJSX,
		},
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

func isJsFile(source string) bool {
	if strings.HasSuffix(source, ".js") {
		return true
	}
	if strings.HasSuffix(source, ".jsx") {
		return true
	}
	if strings.HasSuffix(source, ".ts") {
		return true
	}
	if strings.HasSuffix(source, ".tsx") {
		return true
	}

	return false
}
