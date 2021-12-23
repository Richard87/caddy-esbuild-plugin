# Caddy Esbuild plugin

This plugins watches and builds the source files continiusly in-memory. It includes a etag to cache in the browser to save bandwidth.

## Features:
- [X] Live reload (v.0.1.0)
- [X] SASS (v. 0.2.2)
- [ ] Manifest.json (v. 0.3.0)
- [ ] Env support (v.0.4.0)
- [ ] Refactor to routes (v.0.5.0) (`route /_build {esbuild assets/index.js auto_reload sass}`)
- [ ] Custom loader configuration?

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

## Devlopment:

To run: `xcaddy run -watch`

Build caddy with: 
```shell
# To enable sass:
CGO_ENABLED=1 xcaddy build --with github.com/richard87/caddy-esbuild-plugin

#without sass:
xcaddy build --with github.com/richard87/caddy-esbuild-plugin
```