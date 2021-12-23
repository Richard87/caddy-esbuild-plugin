package caddy_esbuild_plugin

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"go.uber.org/zap"
)

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
					m.logger.Debug("File changed, rebuilding", zap.String("filename", event.Name))
					m.Rebuild()
					return
				}
			case err := <-watcher.Errors:
				m.logger.Error("Failed to watch!", zap.Error(err))
			case <-m.globalQuit:
				done <- true
			}
		}
	}()

	for _, file := range files {
		if err := watcher.Add(file); err != nil {
			m.logger.Error("Failed to watch file", zap.Error(err), zap.String("file", file))
		}
	}

	<-done
}
