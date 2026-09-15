# pickbin

**What:** Picks the right binary from any GitHub release for your machine.

## Install & run

```bash
cd pickbin && go build -o pickbin.exe .
.\pickbin.exe https://github.com/px0-ai/px0
```

Result:

```
✓ Detected: windows / Amd64
Recommended: https://github.com/.../px0-0.1.4-windows-amd64.exe
Asset:    px0-0.1.4-windows-amd64.exe   Size: 11.7 MB   Downloads: 48
```

## Usage

```
pickbin <github-repo-or-release-url>
```

| Input | What happens |
|---|---|
| `github.com/user/repo` | Picks **latest** release automatically |
| URL ending `/tag/vX.Y.Z` | Fetches that specific version's assets |
| GitHub API URL (`api.github.com/repos/...`) | Same as tag format or plain repo |
| No arguments | Prints usage and exits |
| Bad URL | Prints error and exits |

## How it picks

```
Parse URL → GET release JSON → filter non-binaries → score by OS+arch → tiebreak downloads
```

### Matching keywords

| Concept | Recognizes |
|---|---|
| OS | `linux`, `windows`, `win`, `darwin`, `macos` |
| 64-bit CPU | `x86_64`, `amd64`, `x64` |
| ARM 64-bit | `aarch64`, `arm64` |

### Rejected silently

Checksums (`.sha`, `.md5`), signatures (`.asc`, `.sig`), packages (`.deb`, `.rpm`, `.msi`), source tarballs. Everything else passes through.

## Current state

Working:

- Parses GitHub release URLs AND plain repo URLs (auto-picks latest)
- Filters by MIME type, then matches OS + arch via keyword scoring
- Tiebreaks tied scores by download count
- Works on Windows with `.exe` and `.zip` bundle formats

Not done yet:

1. Browser opener (`start`/`open`/`xdg-open`) — prints URL only
2. Response caching — every invocation hits the GitHub API
3. Output truncation — shows all assets when nothing matches; could show top-3 candidates instead

## License

MIT
