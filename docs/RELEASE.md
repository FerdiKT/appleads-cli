# Release Guide

## Local dist build

```bash
make dist VERSION=1.1.0
ls dist
```

This produces:

- darwin/linux binaries for `amd64` and `arm64`
- `dist/checksums.txt`

## GitHub release

Tag and push:

```bash
git tag v1.1.0
git push origin v1.1.0
```

Workflow: `.github/workflows/release.yml`

It will:

1. Build `dist/*`
2. Create/update the GitHub release
3. Upload all artifacts + checksums
