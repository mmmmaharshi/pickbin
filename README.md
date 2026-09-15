# pickbin

One command. Right binary. No guessing.

## What it does

Pass a GitHub release URL. It tells you which asset matches your machine.

```
pickbin https://github.com/user/repo/releases/tag/v1.0
```

Output:

```
✓ Detected: windows / Amd64
Recommended: https://github.com/user/repo/releases/download/v1.0/app-v1.0-windows-amd64.exe
Asset:    app-v1.0-windows-amd64.exe
Size:     11.7 MB
Downloads: 48
```

## Quick start

| Step | Command |
|------|---------|
| Install Go | `winget install GoLang.Go` (or from go.dev) |
| Build | `cd pickbin && go build -o pickbin.exe .` |
| Run | `.\pickbin.exe <release-url>` |

That's it. One executable. No dependencies.

## How it decides

```
URL → parse owner/repo/tag → fetch release JSON → filter non-binaries → score by OS + arch → tiebreak by downloads
```

- **OS keywords**: `windows`, `win`, `linux`, `darwin`, `macos`
- **Arch synonyms**: `x86_64` = `amd64`, `aarch64` = `arm64`
- **Rejected**: checksums (`.sha`), signatures (`.asc`), packages (`.deb`, `.rpm`), source tarballs

## Usage

```
pickbin <github-release-url>
pickbin https://github.com/BurntSushi/ripgrep/releases/tag/14.1.0
pickbin https://api.github.com/repos/BurntSushi/ripgrep/releases/tags/14.1.0  # API URL works too
```

## Known gaps

1. No browser opener — copies URL but doesn't launch it (`start`/`open`/`xdg-open`)
2. No cache — every call hits the GitHub API
3. Shows full asset list if nothing matches — could truncate to top-3 candidates

## License

MIT
