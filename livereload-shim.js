(() => new EventSource('/__livereload').onmessage = e => e.data === "reload" && location.reload())()