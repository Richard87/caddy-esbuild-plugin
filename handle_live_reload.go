package caddy_esbuild_plugin

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func (m *Esbuild) handleLiveReload(w http.ResponseWriter, r *http.Request) error {
	// Add headers needed for server-sent events (SSE):
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	flusher, ok := w.(http.Flusher)
	if !ok {
		m.logger.Debug("Your browser does not support server-sent events (SSE).")
		return nil
	} else {
		m.logger.Debug("LiveReload started")
	}

	var lastPointer = m.esbuild
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	for {
		compareTimeout := time.After(20 * time.Millisecond)
		pingTimeout := time.After(10 * time.Second)
		select {
		case <-r.Context().Done():
			return nil
		case <-m.globalQuit:
			return nil
		case <-sigs:
			return nil
		case <-compareTimeout:
			var currentPointer = m.esbuild
			if lastPointer != currentPointer {
				_, _ = fmt.Fprintf(w, "data: reload\n\n")
				flusher.Flush()
				lastPointer = m.esbuild
			}
		case <-pingTimeout:
			_, _ = fmt.Fprintf(w, "data: p\n\n")
			flusher.Flush()
		}
	}
}

func (m *Esbuild) createAutoloadShimFile() (string, error) {
	livereload := "(() => {const es = new EventSource('%s/__livereload'); es.addEventListener('message', e => e.data === 'reload' && (es.close() || location.reload()))})()"
	file, err := ioutil.TempFile(os.TempDir(), "caddy-esbuild-shim-*.js")
	if err != nil {
		return "", fmt.Errorf("autoload: failed to create tmpfile: %s", err)
	}
	_, err = file.Write([]byte(fmt.Sprintf(livereload, m.Target)))
	if err != nil {
		return "", fmt.Errorf("autoload: failed to write shim: %s", err)
	}
	name := file.Name()
	return name, nil
}
