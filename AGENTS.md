# Agent Instructions

Instructions for AI agents working on this project.

---

## Versioning

Current version is tracked via git tags (`v1.0.0`, `v1.0.1`, ...). The latest tag is the source of truth.

### When to prompt for a new version

After completing a set of changes, **ask the user** if they want to tag a new version when:

- a `feat` or `fix` commit was made
- changes affect runtime behavior (not just docs/chore)
- multiple commits have accumulated since the last tag

### How to tag

```sh
# Prompt user for version bump type (patch/minor/major)
# Then tag and push:
git tag v<version>
git push origin v<version>
```

The GitHub Actions release workflow triggers on `v*` tags and builds binaries for Linux and Windows.

### Don't auto-tag

Never tag a release without asking. The user decides when to release.

---

## Build

Local build:
```sh
mkdir -p .build
go build -o .build/poe2campain ./cmd/poe2campain
cp -r data .build/data
```

Run tests before committing:
```sh
go test ./...
```

---

## Commits

Use conventional commits (`feat`, `fix`, `docs`, `chore`, etc.).

---

## Data pipeline

When source data changes:
```sh
./poe2campain update-data
./poe2campain validate-data
```

---

## Platforms

Only Linux and Windows. PoE2 is not on macOS — do not add macOS build targets.
