# Caddy Esbuild plugin

This plugins watches and builds the source files continiusly in-memory. It includes a etag to cache in the browser to save bandwidth.

## Features:
- [X] Live reload (v0.1.0)
- [X] SASS (v0.2.2)
- [X] Manifest.json (v0.3.0) `/<target>/manifest.json`, defaults to `/_build/manifest.json`
- [X] Env support (v0.4.0)
- [X] Multiple entrypoint (v0.5.0) (filename must be different)
- [X] Refactor to routes (v0.5.0) (`route /_build {esbuild assets/index.js env auto_reload sass}`)
- [X] Custom loader configuration (v0.6.0)
- [X] Enable hashed files with file_hash property (v0.7.0)
- [X] Enable defines (v0.8.0)
- [X] Ouptut alias (v0.9.0)

## Configuration:
`Caddyfile`:
```
{
  order esbuild first
}

localhost:8080 {
  file_server  {
    root example/public
  }
  esbuild {
    env
    scss
    live_reload
    target /_build
    source ./example/src/index.js
    source ./example/src/index.css global
    loader .png dataurl
    loader .svg file
    loader .json json
    loader .js jsx
    define global window
    node_path ./public/node_modules
    node_path ../../node_modules
  }
}
```
or a simplified version: 
```
{
  order esbuild first
}

localhost:8080 {
  file_server  {
    root example/public
  }
  esbuild ./example/src/index.js live_reload scss env ./example/src/index2.js
}
```

## Documentation:
- Env support: It will scan any `.env`, `.env.<NODE_ENV>`, `.env.local`, `.env.<NODE_ENV>.local`, and the runtime environment for relevant variables.  
  It will however not watch them changes or auto-reload them. 

## Devlopment:

To run: `xcaddy run -watch`

Build caddy with: 
```shell
# To enable sass:
CGO_ENABLED=1 xcaddy build --with github.com/richard87/caddy-esbuild-plugin

#without sass:
xcaddy build --with github.com/richard87/caddy-esbuild-plugin
```