package caddy_esbuild_plugin

import (
	"github.com/evanw/esbuild/pkg/api"
	"time"
)

func (m *Esbuild) createTimingPlugin() api.Plugin {
	return api.Plugin{
		Name: "timingPlugin",
		Setup: func(build api.PluginBuild) {
			var start time.Time

			build.OnStart(func() (api.OnStartResult, error) {
				start = time.Now()
				return api.OnStartResult{}, nil
			})
			build.OnEnd(func(result *api.BuildResult) {
				duration := time.Now().Sub(start)
				m.lastDuration = &duration
			})
		},
	}
}
