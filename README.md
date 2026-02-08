# Go Imports Ruler

[![License: MIT](https://img.shields.io/badge/License-MIT-brightgreen?style=flat-square)](/LICENSE)
[![Go Reference](https://pkg.go.dev/badge/github.com/aethiopicuschan/goimportsruler.svg)](https://pkg.go.dev/github.com/aethiopicuschan/goimportsruler)
[![GitHub Action](https://img.shields.io/badge/GitHub%20Action-goimportsruler-blue?logo=github-actions)](https://github.com/marketplace/actions/goimportsruler)
[![Go Report Card](https://goreportcard.com/badge/github.com/aethiopicuschan/goimportsruler)](https://goreportcard.com/report/github.com/aethiopicuschan/goimportsruler)
[![CI](https://github.com/aethiopicuschan/goimportsruler/actions/workflows/ci.yaml/badge.svg)](https://github.com/aethiopicuschan/goimportsruler/actions/workflows/ci.yaml)
[![codecov](https://codecov.io/gh/aethiopicuschan/goimportsruler/graph/badge.svg?token=9cPFqNxKJC)](https://codecov.io/gh/aethiopicuschan/goimportsruler)

`goimportsruler` is a tool for defining and enforcing **import rules** in Go projects.

It helps you prevent unwanted dependencies between packages (e.g. `pkg` importing `cmd`, or domain logic depending on infrastructure), keeping your project structure clean, intentional, and maintainable.

You can run `goimportsruler` from **any directory inside a Go module**.  
The tool automatically locates `go.mod` by walking up the directory tree.

---

## CLI

### Installation

```sh
go install github.com/aethiopicuschan/goimportsruler@v1
```

### Usage

```sh
goimportsruler ./...
```

You may run the command from **any subdirectory** of the target Go module.

If any rule is violated, `goimportsruler` exits with a non-zero status code,  
making it suitable for use in CI pipelines.

---

### Example Output

```sh
pkg/lib.go (package pkg)
  - rule: "Ban pkg to cmd" (Disallow imports from pkg to cmd)
    • pkg/lib.go:13:2  ←  cmd/config
    • pkg/lib.go:14:2  ←  cmd/logger
```

---

## GitHub Actions

You can integrate `goimportsruler` into your CI/CD pipeline using GitHub Actions.

### Example Workflow

```yaml
jobs:
  import-check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Go Imports Ruler Check
        uses: aethiopicuschan/goimportsruler@v1
        with:
          args: "."
```

---

## Configuration

To use `goimportsruler`, create a configuration file in the root of your project.
The configuration defines which packages are allowed to import which others, and which packages should be excluded from checks entirely.

A minimal example configuration looks like this:

```json
{
  "rules": [
    {
      "name": "Ban pkg to cmd",
      "description": "Disallow imports from pkg to cmd",
      "sources": ["pkg/**"],
      "disallow": ["cmd/**"]
    },
    {
      "name": "Ban testify",
      "description": "Disallow importing testify",
      "sources": ["**"],
      "disallow": ["github.com/stretchr/testify/**"]
    }
  ],
  "excludes": [
    {
      "name": "Vendor",
      "description": "Vendor directory is excluded from checks",
      "sources": ["vendor/**"]
    }
  ]
}
```

### Configuration Fields

- `rules`: Defines import restrictions.
  - `name`: A descriptive name for the rule (**optional but recommended**).
  - `description`: A brief explanation of the rule (**optional**).
  - `sources`: Source **import path patterns** to which the rule applies.
  - `disallow`: Import path or external module patterns that are not allowed for the given sources.
- `excludes`: Defines source packages that are completely excluded from checks.
  - `name`: A descriptive name for the exclude rule (**optional but recommended**).
  - `description`: A brief explanation of the exclude rule (**optional**).
  - `sources`: Source import path patterns to be excluded.

When a source package matches an exclude rule, **all checks for that package are skipped entirely**.

### Supported Configuration File Names

- Normal files:
  - `goimportsruler.json`
  - `goimportsruler.yaml`
  - `goimportsruler.yml`
- Hidden files:
  - `.goimportsruler.json`
  - `.goimportsruler.yaml`
  - `.goimportsruler.yml`

---

## Pattern Matching

The `sources` and `disallow` fields support glob-like pattern matching to specify **import paths and external module paths**.

Patterns are matched against **import paths**, not directory names, file names, or package declarations.

### Syntax

- `*` — Matches any characters within a single path segment (does not match `/`)
- `**` — Matches zero or more path segments (can cross `/`)

### Examples

Given the following directory structure:

```sh
.
├── cmd
│   ├── app.go
│   └── server
│       └── server.go
├── internal
│   ├── helper
│   │   └── helper.go
│   ├── internal.go
│   └── logger
│       └── logger.go
└── pkg
    └── handler.go
```

The following patterns behave as:

- `cmd/**` → matches `cmd` and its sub-packages (e.g. `cmd/server`)
- `cmd/**/server` → matches `cmd/server`
- `internal/*` → matches `internal` but not `internal/helper`
- `**/logger` → matches `internal/logger`
- `**` → matches all import paths

⚠️ The examples above assume that the directory structure directly corresponds to import paths.

---

## Absolute vs Relative Import Path Patterns

Patterns can be written in **two equivalent forms**:

- Module-relative paths
  - e.g. `cmd/**`, `pkg/**`
- Fully-qualified module paths
  - e.g. `example.com/myapp/cmd/**`

Both forms are supported and matched automatically.

---

## Banning External Modules

When banning an external dependency, it is recommended to use a prefix pattern with `/**` to ensure all its sub-packages are also disallowed.

```json
{
  "sources": ["**"],
  "disallow": ["github.com/stretchr/testify/**"]
}
```

---

## Why Go Imports Ruler?

While tools like `go vet` or `staticcheck` focus on correctness and best practices, `goimportsruler` focuses on **architectural boundaries**.

It allows teams to:

- Enforce layered architectures
- Prevent dependency drift over time
- Make architectural rules explicit and reviewable

This makes it especially useful for medium-to-large Go codebases and long-lived projects.
