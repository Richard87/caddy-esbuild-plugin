package caddy_esbuild_plugin

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"go.uber.org/zap"
	"reflect"
	"unsafe"
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
				if !isChanClosed(done) {
					done <- true
				}
				return
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

func isChanClosed(ch interface{}) bool {
	if reflect.TypeOf(ch).Kind() != reflect.Chan {
		panic("only channels!")
	}

	// get interface value pointer, from cgo_export
	// typedef struct { void *t; void *v; } GoInterface;
	// then get channel real pointer
	cptr := *(*uintptr)(unsafe.Pointer(
		unsafe.Pointer(uintptr(unsafe.Pointer(&ch)) + unsafe.Sizeof(uint(0))),
	))

	// this function will return true if chan.closed > 0
	// see hchan on https://github.com/golang/go/blob/master/src/runtime/chan.go
	// type hchan struct {
	// qcount   uint           // total data in the queue
	// dataqsiz uint           // size of the circular queue
	// buf      unsafe.Pointer // points to an array of dataqsiz elements
	// elemsize uint16
	// closed   uint32
	// **

	cptr += unsafe.Sizeof(uint(0)) * 2
	cptr += unsafe.Sizeof(unsafe.Pointer(uintptr(0)))
	cptr += unsafe.Sizeof(uint16(0))
	return *(*uint32)(unsafe.Pointer(cptr)) > 0
}
