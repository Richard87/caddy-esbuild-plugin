# Caddy Esbuild plugin

This plugins watches and builds the source files continiusly in-memory. It includes a etag to cache in the browser to save bandwidth.

## TODO:
- Better esbuild/Caddyfile configuration
  - Handle Loader configuration
  - Support esbuild config flags
  - Remove `order esbuild first` config requirement
- include env support
- include sass support (go-libsass)
- Hot reloading?! :O 

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
  esbuild ./example/src/index.js /build/index.js
}
```

Run with xcaddy run