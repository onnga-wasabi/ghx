# ghx

A keyboard-driven TUI for GitHub — PRs, Issues, Actions, and Notifications in one place.

```
╭─ 1:PRs  2:Actions ──────────────────────────────────────────╮
│ Open Closed  ║  All │ Mine │ Review │ Involved               │
│╭─ Pull Requests · Open · All ───────╮╭─ Preview ────────────╮│
││▸ ● #96 staging deploy fix    user  ││ Author:  user        ││
││  ● #95 add health check      user  ││ Branch:  feat → main ││
││  ◇ #94 WIP: new feature      user  ││ State:   OPEN        ││
│╰────────────────────────────────────╯╰──────────────────────╯│
│ ? help │ q quit │ tab next │ / filter          owner/repo    │
╰──────────────────────────────────────────────────────────────╯
```

## Features

- **Unified Dashboard** — PRs, Issues, Actions, and Notifications in one TUI
- **Consistent Filter Bar** — Every view has a top filter bar with state toggle (`s`) and scope/type cycling (`←`/`→`)
- **Rich PR Sidebar** — Checks with duration, summary bar (✓ 3  ✗ 1  ⏳ 2), per-file diff stats with visual +/- bars
- **Actions Deep Dive** — Workflows → Runs → Jobs → Logs in 4 horizontal panes, status filter, inline job expansion via `Enter`
- **Floating Help** — Press `?` for a centered keybinding overlay
- **Configurable Tabs** — Show only the tabs you need, in your preferred order
- **Keyboard First** — Vim-style navigation (`j`/`k`/`g`/`G`), numbered tab switching (`1`-`4`)
- **Smart Layout** — Focused pane auto-expands, configurable via `smartLayout`
- **Custom Keybindings** — Run arbitrary shell commands from within the TUI

## Install

### Go

```bash
go install github.com/onnga-wasabi/ghx/cmd/main.go@latest
```

### Homebrew

```bash
brew install onnga-wasabi/tap/ghx
```

### From Source

```bash
git clone https://github.com/onnga-wasabi/ghx.git
cd ghx
make install
```

### Nix Flake

```nix
{
  inputs.ghx.url = "github:onnga-wasabi/ghx";
  # Then add ghx.packages.${system}.default to your packages
}
```

## Prerequisites

- [GitHub CLI (`gh`)](https://cli.github.com/) must be installed and authenticated (`gh auth login`)
- `ghx` uses `gh auth token` for API authentication — no separate token setup needed

## Configuration

Config file: `~/.config/ghx/config.yml` (or `$XDG_CONFIG_HOME/ghx/config.yml`)

```yaml
defaults:
  prsLimit: 20
  issuesLimit: 20
  view: prs                          # Default tab: prs, issues, actions, notifications
  tabs:                              # Customize which tabs appear and their order
    - prs
    - actions
    # - issues
    # - notifications
  smartLayout: true                  # Auto-expand focused pane
  preview:
    open: true
    width: 0.45

keybindings:
  universal:
    - key: "b"
      name: "open lazygit"
      command: "lazygit"

theme:
  colors:
    primary: "#7aa2f7"
    success: "#9ece6a"
    warning: "#e0af68"
    error: "#f7768e"
```

## Keybindings

### Global

| Key | Action |
|-----|--------|
| `?` | Toggle help overlay |
| `q` / `Ctrl+C` | Quit (or close help) |
| `Tab` / `Shift+Tab` | Next / Previous tab |
| `1`-`9` | Jump to tab by number |
| `j`/`k` or `↑`/`↓` | Navigate up/down |
| `h`/`l` or `←`/`→` | Switch scope (PRs/Issues) or pane (Actions) |
| `g` / `G` | First / Last item |
| `Enter` | Toggle sidebar (PRs/Issues) or expand run (Actions) |
| `o` | Open in browser |
| `R` | Refresh |
| `/` | Filter |

### PRs / Issues

| Key | Action |
|-----|--------|
| `s` | Toggle Open / Closed |
| `←`/`→` | Cycle scope: All → Mine → Review → Involved |
| `e` | Jump to CI checks (Actions tab) |
| `d` | Open diff in browser |
| `a` | Approve PR |
| `m` | Merge PR |
| `x` | Close |
| `C` | Comment (opens `gh` CLI) |

### Actions

| Key | Action |
|-----|--------|
| `s` | Cycle status filter: All → Success → Failed → Running |
| `Enter` | Toggle run expansion (show/hide jobs inline) |
| `r` | Rerun workflow |
| `Ctrl+R` | Rerun failed jobs |
| `c` | Cancel run |
| `t` | Trigger workflow |
| `f` | Fullscreen log viewer |
| `w` | Toggle word wrap |
| `n`/`N` | Next/Prev search match |

### Notifications

| Key | Action |
|-----|--------|
| `s` | Toggle Unread / All |
| `←`/`→` | Cycle type: All → PR → Issue → Release → CI |
| `r` | Mark read |
| `d` | Mark done |

## Design Philosophy

`ghx` is a **command center**, not a replacement for every GitHub feature:

- **Triage & Act** — Quickly scan PRs/Issues, take action (approve, merge, rerun), move on
- **Delegate to Specialists** — Detailed code review belongs in the browser; `o` opens GitHub, `d` opens the diff page
- **Terminal Native** — Actions log viewer with ANSI color, search, and word wrap
- **Composable** — Custom keybindings let you launch `lazygit`, `gh pr diff`, or any tool

## Acknowledgements

This project is inspired by and pays respect to:

- [gh-dash](https://github.com/dlvhdr/gh-dash) — The original GitHub dashboard TUI for `gh`. `ghx` was born from wanting a unified Actions + PR experience.
- [lazygit](https://github.com/jesseduffield/lazygit) — Pioneered the keyboard-driven Git TUI paradigm.
- [Charm](https://charm.sh/) — The Bubble Tea framework and Lip Gloss styling library that power this TUI.

## License

[MIT](LICENSE)
