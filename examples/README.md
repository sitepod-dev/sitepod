# SitePod Examples

This directory contains example projects for testing SitePod deployments.

## Examples Overview

| Example | Description | Test Coverage |
|---------|-------------|---------------|
| `simple-site` | Basic static site | Basic deployment |
| `spa-site` | Single Page Application | SPA fallback routing |
| `unicode-files` | Chinese/Japanese/Korean/emoji filenames | Unicode path handling |
| `special-chars` | Spaces, dashes, dots in filenames | URL encoding |
| `nested-deep` | 12-level deep directories | Deep path handling |
| `binary-assets` | Images, JSON, XML files | Binary file handling |
| `duplicate-content` | Identical files in multiple paths | Content deduplication |
| `large-files` | Files up to 7MB | Large file upload |
| `many-files` | 500+ small files | Batch upload performance |

## Deploy All Examples

```bash
# Start server first
make run

# Deploy each example
for dir in examples/*/; do
  if [ -f "$dir/sitepod.toml" ]; then
    echo "Deploying $dir..."
    (cd "$dir" && ../../bin/sitepod deploy -y)
  fi
done
```

## Individual Examples

### simple-site

Basic static site with HTML, CSS, and JavaScript.

```bash
cd examples/simple-site && sitepod deploy
```

### spa-site

Single Page Application - tests that non-existent paths fall back to index.html.

```bash
cd examples/spa-site && sitepod deploy
# Visit /about, /products - should all render index.html
```

### unicode-files

Tests Unicode filenames: Chinese (中文), Japanese (日本語), Korean (한국어), and emoji.

```bash
cd examples/unicode-files && sitepod deploy
# Visit /中文目录/文档.html
```

### special-chars

Tests special characters in filenames: spaces, dashes, underscores, multiple dots.

```bash
cd examples/special-chars && sitepod deploy
# Visit /dir%20with%20spaces/file%20with%20spaces.html
```

### nested-deep

Tests deeply nested directories (12 levels deep).

```bash
cd examples/nested-deep && sitepod deploy
# Visit /a/b/c/d/e/f/g/h/i/j/k/l/deep.html
```

### binary-assets

Tests binary files: PNG, GIF, SVG, ICO, JSON, XML.

```bash
cd examples/binary-assets && sitepod deploy
```

### duplicate-content

Tests content-addressed deduplication. Contains 10 files but only ~4 unique blobs.

```bash
cd examples/duplicate-content && sitepod deploy
# Check that reuse percentage is high
```

### large-files

Tests large file uploads (~10MB total).

```bash
cd examples/large-files && sitepod deploy
```

### many-files

Tests batch upload with 501 files across 10 directories.

```bash
cd examples/many-files && sitepod deploy
```
