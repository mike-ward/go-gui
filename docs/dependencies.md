# Dependency Updates

This project depends on private Go modules under `github.com/mike-ward/*`,
including `github.com/mike-ward/go-glyph` and
`github.com/mike-ward/go-glyph/backend/sdl2`.

## How dependency resolution works

- [`go.mod`](/Users/mikeward/Documents/github/go-gui/go.mod) declares the module
  requirements for both private modules.
- [`go.sum`](/Users/mikeward/Documents/github/go-gui/go.sum) records checksums
  for the exact dependency contents and their `go.mod` files.
- CI sets `GOPRIVATE` and `GONOSUMDB` so the Go toolchain can fetch the private
  modules directly from GitHub.

## Updating `go-glyph`

`go-glyph` and `go-glyph/backend/sdl2` are now versioned modules. Update both
to the same release tag.

```bash
cd /Users/mikeward/Documents/github/go-gui

GOPRIVATE=github.com/mike-ward/* GONOSUMDB=github.com/mike-ward/* \
go get github.com/mike-ward/go-glyph@v1.0.0 \
  github.com/mike-ward/go-glyph/backend/sdl2@v1.0.0

GOPRIVATE=github.com/mike-ward/* GONOSUMDB=github.com/mike-ward/* \
go mod tidy
```

After that, confirm that [`go.mod`](/Users/mikeward/Documents/github/go-gui/go.mod)
keeps `github.com/mike-ward/go-glyph` and
`github.com/mike-ward/go-glyph/backend/sdl2` pinned to the same release.

## Updating public dependencies

For public modules, the normal Go workflow is enough:

```bash
go get <module>@<version>
go mod tidy
```

## Verification

Run these checks before committing:

```bash
GOPRIVATE=github.com/mike-ward/* GONOSUMDB=github.com/mike-ward/* go test ./gui/...
GOPRIVATE=github.com/mike-ward/* GONOSUMDB=github.com/mike-ward/* go vet ./...
```

Commit both [`go.mod`](/Users/mikeward/Documents/github/go-gui/go.mod) and
[`go.sum`](/Users/mikeward/Documents/github/go-gui/go.sum) whenever dependencies
change.
