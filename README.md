# gog — Recursive Git Repository Manager

`gog` discovers, registers, clones, and operates on every git repository under a
workspace root. A **workspace** is a directory containing a `gog.toml` manifest;
all repo paths are relative to it. Put the manifest at `~/git` (or `~`) to manage
all the repos on your system.

Batch operations run in parallel and report results as an aligned table, a quiet
failure list, or JSON.

## Install

```sh
go build -o gog .
# or
go install .
```

Requires `git` on PATH.

## Manifest (`gog.toml`)

```toml
version = 1
base_url = "https://github.com/jlcoulter"

[[repo]]
path = "momus"                                # relative, forward slashes

[[repo]]
path = "vendored/other"
url = "https://gitlab.com/foo/other"          # optional; omit → base_url + "/" + basename
```

`gog` locates the manifest by walking up from the current directory, or via
`--config <path>`.

## Commands

| Command | Description |
|---|---|
| `gog init` | Create an empty `gog.toml` in the current directory |
| `gog scan [--dry-run]` | Discover repos under the workspace and register them (records origin URLs) |
| `gog add <name-or-url>...` | Declare repos for later cloning |
| `gog remove <path>...` | Remove entries from the manifest (never touches disk) |
| `gog clone [path...]` | Clone manifest repos missing from disk |
| `gog list [--missing]` | List manifest entries with URL and on-disk status |
| `gog status [path...]` | Branch + working-tree state per repo |
| `gog pull [--rebase] [path...]` | Fast-forward each repo to its upstream |
| `gog fetch [path...]` | Fetch all remotes in each repo |

## Global flags

- `--config <path>` — explicit manifest path
- `--json` — JSON output (for scripting)
- `--no-color` — disable colors (also honors `NO_COLOR`)

Batch commands (`clone`, `status`, `pull`, `fetch`) also accept `-p/--parallel`
(default 4) and `-q/--quiet` (print failures only).

## Exit codes

- `0` — success (warnings/skips don't fail)
- `1` — any repo operation failed, or a usage/config error

## Notes

- `pull` defaults to `--ff-only` to avoid surprise merges in bulk; use
  `--rebase` to rebase instead.
- `scan` prunes below discovered repos (nested repos inside a repo are not
  registered) and skips the workspace root even if it is itself a repo.
- `scan` records each repo's origin URL so the manifest is a self-contained,
  rebuildable inventory.
