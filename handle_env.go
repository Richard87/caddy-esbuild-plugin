package caddy_esbuild_plugin

import (
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"os"
	"strings"
)

func (m *Esbuild) handleEnv() {
	process := Process{Env: map[string]string{}}
	currentEnv := os.Getenv("NODE_ENV")
	if currentEnv == "" {
		currentEnv = "development"
	}

	_ = godotenv.Load(".env." + currentEnv + ".local")
	_ = godotenv.Load(".env.local")
	_ = godotenv.Load(".env." + currentEnv)
	err := godotenv.Load()

	process.Env["NODE_ENV"] = os.Getenv("NODE_ENV")
	if err != nil {
		m.logger.Error("Failed to load env", zap.Error(err))
	} else {
		for _, pair := range os.Environ() {
			item := strings.SplitN(pair, "=", 2)
			key := item[0]
			val := item[1]

			val = strings.ReplaceAll(val, "\"", "\\\"")
			m.Defines["process.env."+key] = "\"" + val + "\""
		}
	}
}
