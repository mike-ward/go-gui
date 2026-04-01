# Dependency Updates

This project depends on `github.com/mike-ward/go-glyph` for text
shaping/rendering. The module is public and fetched directly from GitHub.

## How dependency resolution works

- [`go.mod`](../go.mod) declares module requirements.
- [`go.sum`](../go.sum) records checksums for exact dependency contents.

## Updating `go-glyph`

```bash
go get github.com/mike-ward/go-glyph@vX.Y.Z
go mod tidy
```

## Updating other dependencies

```bash
go get <module>@<version>
go mod tidy
```

## Verification

Run these checks before committing:

```bash
go test ./gui/...
go vet ./...
```

Commit both `go.mod` and `go.sum` whenever dependencies change.
