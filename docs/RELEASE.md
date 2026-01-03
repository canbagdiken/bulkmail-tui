# Release Process

This document describes how to create a new release.

## Automatic Release via GitHub Actions

This project uses GitHub Actions to automatically build and release binaries for multiple platforms.

### Supported Platforms

- **Linux**: AMD64, ARM64
- **macOS**: AMD64 (Intel), ARM64 (Apple Silicon)
- **Windows**: AMD64, ARM64

### Creating a Release

1. **Update version in code** (if applicable)

2. **Commit all changes**
   ```bash
   git add .
   git commit -m "Prepare for release vX.Y.Z"
   ```

3. **Create and push a tag**
   ```bash
   git tag -a v1.0.0 -m "Release version 1.0.0"
   git push origin v1.0.0
   ```

4. **GitHub Actions will automatically:**
   - Build binaries for all platforms
   - Create a GitHub Release
   - Upload all binaries to the release
   - Generate release notes from commits

### Tag Naming Convention

Use semantic versioning with a `v` prefix:
- `v1.0.0` - Major release
- `v1.1.0` - Minor release (new features)
- `v1.0.1` - Patch release (bug fixes)
- `v1.0.0-beta.1` - Pre-release

### Manual Release (Alternative)

If you need to build manually:

```bash
# Linux AMD64
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bulkmail-tui-linux-amd64

# macOS AMD64
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o bulkmail-tui-darwin-amd64

# macOS ARM64 (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o bulkmail-tui-darwin-arm64

# Windows AMD64
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o bulkmail-tui-windows-amd64.exe
```

### Build Flags

- `-ldflags="-s -w"` - Strip debug information to reduce binary size
  - `-s` - Omit symbol table
  - `-w` - Omit DWARF symbol table

### Verifying Release

After the release is created:

1. Go to the [Releases page](https://github.com/yourusername/bulkmail-tui/releases)
2. Check that all binaries are uploaded
3. Download and test each binary (at least one per OS)
4. Verify the release notes are correct

### Troubleshooting

**Build fails:**
- Check Go version compatibility (requires Go 1.20+)
- Verify all dependencies are available
- Check for syntax errors in code

**Release not created:**
- Ensure tag follows `v*` pattern
- Check GitHub Actions permissions
- Verify `GITHUB_TOKEN` has write access

**Binary doesn't work:**
- Test on target platform before release
- Check CGO dependencies (should be disabled for cross-compilation)
- Verify GOOS and GOARCH combinations are valid

## Release Checklist

Before creating a release:

- [ ] All tests pass
- [ ] Documentation is updated
- [ ] README.md reflects current version
- [ ] CHANGELOG.md is updated (if exists)
- [ ] No sensitive data in config files
- [ ] config.yaml.example is up to date
- [ ] Build works on your local machine
- [ ] Version number follows semantic versioning

After release:

- [ ] Test downloaded binaries
- [ ] Update any installation instructions
- [ ] Announce release (if applicable)
- [ ] Monitor for issues
