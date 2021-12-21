# Caddy Esbuild plugin

This plugins watches and builds the source files continiusly in-memory. It includes a etag to cache in the browser to save bandwidth.

## TODO:
- [X] Better esbuild/Caddyfile configuration
  - Handle Loader configuration
  - Support esbuild config flags
  - Remove `order esbuild first` config requirement
  - Test multiple entrypoints
- include env support
- [X] include sass support (go-libsass) (initial support, compile with `CGO_ENABLED=1 xcaddy build` to enable it 
- [X] Live reload when code changes

It works right now in my ReactApp example, its also the only one I have tested so far :)

## Configuration:
`Caddyfile`:
```
{
	order esbuild first
	debug
}

localhost:8080 {
  file_server  {
    root example/public
  }
  esbuild ./example/src/index.js {
    auto_reload
    target /_build
    sass
  }
}
```

Devlopment_ `xcaddy run -watch`

Build caddy with: 
```shell
# To enable sass:
CGO_ENABLED=1 xcaddy build --with github.com/richard87/caddy-esbuild-plugin

#without sass:
xcaddy build --with github.com/richard87/caddy-esbuild-plugin
```