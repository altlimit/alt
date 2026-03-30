# alt — Acquire Latest Tools

A stateless, zero-config CLI distribution proxy. Install any tool from GitHub Releases with a single command.

```
alt install user/repo
```

No `sudo`. No package manager. No config files. Just one command and the binary is on your `PATH`.

## Install

### Linux & macOS

```bash
curl -fsSL https://raw.githubusercontent.com/altlimit/alt/main/scripts/install.sh | sh
```

### Windows (PowerShell)

```powershell
powershell -Command "iwr https://raw.githubusercontent.com/altlimit/alt/main/scripts/install.ps1 -useb | iex"
```

### Quick Run (No Install Required)

Run any tool instantly — alt is installed automatically on first use:

**Linux & macOS:**
```bash
curl -fsSL https://raw.githubusercontent.com/altlimit/alt/main/scripts/run.sh | sh -s -- user/repo [args...]
```

**Windows (PowerShell):**
```powershell
powershell -Command "$env:ALT_RUN='user/repo [args...]'; iwr https://raw.githubusercontent.com/altlimit/alt/main/scripts/run.ps1 -useb | iex"
```

## Quick Start

```bash
# Install tools (supports multiple at once)
alt install altlimit/sitegen altlimit/taskr

# Install a specific version
alt install altlimit/altclaw@v2026.03.29

# Run a tool without installing it
alt run altlimit/sitegen -serve

# Force re-download
alt install -f altlimit/sitegen

# Create a short alias
alt link altlimit/sitegen sg

# Update everything
alt update

# See what's installed
alt list altlimit

# Find a binary
alt which taskr

# Free up old versions and run cache
alt clean

# Remove an alias
alt unlink sg

# Remove completely
alt purge altlimit/taskr
```

## How It Works

When you run `alt install user/repo`, alt:

1. **Fetches** the latest release from the GitHub Releases API
2. **Scores** each asset for compatibility with your OS and architecture
3. **Downloads** the best match (skips if already cached)
4. **Verifies** the checksum (if the release includes one)
5. **Extracts** the binary (if it's an archive)
6. **Links** it to your `PATH`

That's it. No package manifests, no build steps, no elevated permissions.

## Asset Scoring

alt automatically picks the right binary for your machine using a scoring algorithm:

```
Score = (OS × 100) + (Arch × 100) + Preference
```

| Factor | Score |
|--------|-------|
| OS matches (e.g., `linux`, `darwin`, `macos`) | +100 |
| Architecture matches (e.g., `amd64`, `x86_64`, `arm64`) | +100 |
| Raw binary / `.exe` | +50 |
| Archive (`.tar.gz`, `.zip`, `.tgz`) | +20 |
| System installer (`.msi`, `.pkg`, `.deb`, `.rpm`) | -50 |

Checksum files (`.sha256`, `checksums.txt`, `SHA256SUMS`) are automatically filtered out.

## Commands

| Command | Usage | Description |
|---------|-------|-------------|
| `install` | `alt install [-f] user/repo[@tag] [...]` | Install one or more tools from GitHub |
| `run` | `alt run user/repo[@tag] [args]` | Run a tool without installing it |
| `update` | `alt update [user/repo]` | Update installed tools to latest release |
| `list` | `alt list [user]` | Show installed tools |
| `link` | `alt link user/repo <alias>` | Create a custom command alias |
| `unlink` | `alt unlink <alias>` | Remove an alias without uninstalling |
| `clean` | `alt clean [user/repo\|user]` | Remove old versions and run cache |
| `purge` | `alt purge user/repo\|user [...]` | Remove everything for one or more tools |
| `versions` | `alt versions user/repo` | List locally cached versions |
| `which` | `alt which <command>` | Show binary path |

**Flags:**
- `-f`, `--force` — Force re-download even if cached (install only)
- `-h`, `--help` — Show help
- `-v`, `--version` — Show version

## Storage

All files are stored in user space — no root/admin required.

```
~/.local/share/alt/
├── manifest.json
├── bin/             # symlinks to active binaries
├── internal/        # alt binary itself
├── run/             # cached one-shot binaries (cleaned by alt clean)
└── storage/
    └── github.com/
        └── user/
            └── repo/
                ├── v1.0.0/
                └── v1.1.0/
```

On Windows, `%LOCALAPPDATA%\alt\` is used instead with the same structure.

## Security

- **Checksums**: Validated automatically when the release includes a `checksums.txt` or `.sha256` file
- **User-space**: Never requests `sudo` or admin privileges
- **Transparency**: Downloads directly from official GitHub Release URLs
- **Zero dependencies**: Single static binary, no runtime dependencies

## Environment Variables

| Variable | Description |
|----------|-------------|
| `GITHUB_TOKEN` | GitHub personal access token — raises API rate limit from 60 to 5,000 requests/hour |

## For Maintainers

Your users can install your tool with:

```bash
# One-time setup
curl -fsSL https://raw.githubusercontent.com/altlimit/alt/main/scripts/install.sh | sh

# Install your tool
alt install your-username/your-repo
```

Or let them run it instantly with zero setup:

```bash
curl -fsSL https://raw.githubusercontent.com/altlimit/alt/main/scripts/run.sh | sh -s -- your-username/your-repo
```

Just publish your binaries as GitHub Release assets. alt handles OS/architecture detection automatically.

## License

MIT — see [LICENSE](LICENSE).
