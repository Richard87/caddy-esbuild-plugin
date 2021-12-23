# Caddy Esbuild plugin

This plugins watches and builds the source files continiusly in-memory. It includes a etag to cache in the browser to save bandwidth.

## TODO:
- [X] Better esbuild/Caddyfile configuration
  - Handle Loader configuration
- include env support
- [X] include sass support (go-libsass) (initial support, compile with `CGO_ENABLED=1 xcaddy build` to enable it 
  - [X] Bug: Importing scss files inside of a sass file fails
  - [X] Bug: live-reloading doesnt work when chaning imported sass/css file
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