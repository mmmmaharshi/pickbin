# pickbin

**What:** Picks the right binary from any GitHub release for your machine.

## Install

| Platform | Command |
|---|---|
| macOS / Linux | `brew install mmmmaharshi/tap/pickbin` |
| Windows | `winget install mmmmaharshi.pickbin` |
| Go (any) | `go install github.com/mmmmaharshi/pickbin@latest` |

Or [download a release](https://github.com/mmmmaharshi/pickbin/releases) directly — single binary, no deps.

## Run

```bash
pickbin https://github.com/px0-ai/px0
```

Result:

```
✓ Detected: windows / Amd64
Recommended: https://github.com/.../px0-0.1.4-windows-amd64.exe
Asset:    px0-0.1.4-windows-amd64.exe   Size: 11.7 MB   Downloads: 48
✓ Opening release page in browser...
```

The browser opens the release page — see changelog, click your asset.

## Usage

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
| 64-bit x86 | `x86_64`, `amd64`, `x64` |
| 32-bit x86 | `i386`, `i686`, `x86`, `ia32` |
| 64-bit ARM | `aarch64`, `arm64` |
| 32-bit ARM | `armv6`, `armv7`, `armhf` |

### Rejected silently

Checksums (`.sha`, `.md5`), signatures (`.asc`, `.sig`), packages (`.deb`, `.rpm`, `.msi`), source tarballs. Everything else passes through.

## License

MIT
