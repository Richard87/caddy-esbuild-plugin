# Caddy Esbuild plugin

This plugins watches and builds the source files continiusly in-memory. It includes a etag to cache in the browser to save bandwidth.

## Features:
- v0.1.0: Live reload 
- v0.2.2: SASS 
- v0.3.0: Manifest.json `/<target>/manifest.json`, defaults to `/_build/manifest.json` 
- v0.4.0: Env support 
- v0.5.0: Multiple entrypoint (filename must be different) 
- v0.6.0: Custom loader configuration 
- v0.7.0: Enable hashed files with file_hash property 
- v0.8.0: Enable defines 
- v0.9.0: Ouptut alias 
- v.0.10.0: Transparent compilation and NodePath configuration

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
- If target is missing, assets will be available at `/_build`, check `/_build/manifest.json` for details. The source files will be available at the path from Caddyfile. For example `./example/src/global.scss` is available at `https://example.com/example/src/global.scss`, but will return compiled css content. The same with EcmaScript code.
- Env support: It will scan any `.env`, `.env.<NODE_ENV>`, `.env.local`, `.env.<NODE_ENV>.local`, and the runtime environment for relevant variables.  
  It will however not watch them changes or auto-reload them. 
- If no node_paths are specified, I will automatically use all node_modules paths found under current working directory

## Devlopment:

To run: `xcaddy run -watch`

Build caddy with: 
```shell
# To enable sass:
CGO_ENABLED=1 xcaddy build --with github.com/richard87/caddy-esbuild-plugin

#without sass:
xcaddy build --with github.com/richard87/caddy-esbuild-plugin
```