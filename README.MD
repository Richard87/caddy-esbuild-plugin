# Caddy Esbuild plugin

This plugins watches and builds the source files continiusly in-memory. It includes a etag to cache in the browser to save bandwidth.

## Features:
- [X] Live reload (v.0.1.0)
- [X] SASS (v. 0.2.2)
- [X] Manifest.json (v. 0.3.0) `/<target>/manifest.json`, defaults to `/_build/manifest.json`
- [X] Env support (v.0.4.0)
- [X] Multiple entrypoint (v.0.5.0) (filename must be different)
- [X] Refactor to routes (v.0.5.0) (`route /_build {esbuild assets/index.js env auto_reload sass}`)
- [ ] Custom loader configuration?

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
    sass
    live_reload
    target /_build
    source ./example/src/index.js
    source ./example/src/index2.js
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
  esbuild ./example/src/index.js live_reload sass env ./example/src/index2.js
}
```

## Documentation:
- Env support: esbuild will include all environment variables starting with `REACT_APP_` and bundle it with your application,
  **including unused items**. It will scan any `.env`, `.env.<NODE_ENV>`, `.env.local`, `.env.<NODE_ENV>.local`, and the runtime environment for relevant variables.  
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