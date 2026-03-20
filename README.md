# ghx

**Your GitHub command center, right in the terminal.**

ghx is a keyboard-driven TUI that brings PRs, Issues, Actions, and Notifications into a single, unified dashboard. It's built for developers who live in the terminal and want to triage their GitHub workflow without context-switching to the browser.

<!-- TODO: Replace with a real screenshot or GIF demo -->
<!-- ![ghx demo](https://raw.githubusercontent.com/onnga-wasabi/ghx/main/docs/demo.gif) -->

```
 1:PRs  2:Issues  3:Actions  4:Notifications
 Open Closed  ║  All │ Mine │ Review │ Involved      s:state ←→:scope
╭─ Pull Requests · Open · All ──────────╮╭─ #96 staging deploy fix ────────────╮
│▸ ● #96 staging deploy fix    alice  3/│ Author:  alice                       │
│  ● #95 add health check      bob   2/│ Branch:  staging-fix → main          │
│  ◇ #94 WIP: refactor auth    carol   │ Review:  APPROVED                    │
│  ◆ #93 fix rate limiter      alice   │                                      │
│  ○ #92 remove legacy API     dave    │ ── Checks ──                         │
│                                       │   ✓ 3  ✗ 1  ⏳ 0  (3/4 passed)      │
│                                       │   ✓ Build (1m42s)                    │
│                                       │   ✓ Lint (32s)                       │
│                                       │   ✓ Unit Tests (2m14s)              │
│                                       │   ✗ E2E Tests (4m01s)               │
│                                       │                                      │
│                                       │ ── Diff Stats ──                     │
│                                       │   4 files  +56 -12                   │
│                                       │   +++--····· +28 -6  src/api/client  │
│                                       │   ++········ +12 -0  src/api/auth    │
│                                       │   +--······· +8  -4  tests/api_test  │
│                                       │   +-········ +8  -2  README.md       │
╰───────────────────────────────────────╯╰─────────────────────────────────────╯
 ? help  q quit  s open/closed  ←→ scope  o open  a approve       owner/repo
```

## Why ghx?

**The problem you know too well:**

You're in the terminal. A CI check fails. You open the browser, find the PR, click through to Actions, scroll through logs, find the error, come back to the terminal to fix it, push, then switch *back* to the browser to watch CI again. Repeat this 20 times a day.

**ghx keeps you in flow:**

See your PRs → jump to their CI checks → read the logs → rerun the workflow → approve and merge — all without touching the mouse or leaving the terminal.

## Philosophy

### What we built

A **command center**. You see everything at a glance, triage with a keystroke, and move on. Every view follows the same pattern: filter bar at the top, list on the left, detail on the right, consistent shortcuts everywhere. Once you learn one view, you know them all.

### What we chose not to build

A full GitHub client. Detailed code review — diffs with syntax highlighting, inline comments, commit-by-commit walkthrough — belongs in the browser or dedicated tools. That's not a limitation; it's a deliberate choice. The browser does those things better than any TUI ever will.

Instead, we made the handoff seamless: press `o` to open in the browser, `d` to jump straight to the diff, `e` to inspect CI. ghx gets you to the right place faster.

### What this means in practice

| ghx does this well | Use the browser for this |
|---|---|
| Scan 20 PRs in seconds | Read a 500-line diff |
| See which checks failed and why | Leave inline review comments |
| Approve, merge, close with one key | Resolve merge conflicts |
| Rerun CI, cancel runs, trigger workflows | Configure workflow YAML |
| Triage notifications by type | Manage repository settings |

## Features

```
╭─ 1:PRs  2:Issues  3:Actions  4:Notifications ──────────────────────────────╮
│                                                                              │
│  ┌─ Filter Bar ──────────────────────────────────────────────────────────┐   │
│  │ Open Closed  ║  All │ Mine │ Review │ Involved       s:state ←→:scope│   │
│  └───────────────────────────────────────────────────────────────────────┘   │
│                                                                              │
│  Every view. Same pattern. Same shortcuts.                                   │
│                                                                              │
╰──────────────────────────────────────────────────────────────────────────────╯
```

- **Unified Filter Bar** — Every view has the same interaction pattern: `s` toggles state, `←`/`→` cycles scope. PRs, Issues, Actions, Notifications — all consistent.

- **Rich PR Sidebar** — Checks with pass/fail/pending summary, individual durations, and per-file diff stats with visual `+`/`-` bars. Enough context to decide without opening the browser.

- **Actions Explorer** — Four-pane layout: Workflows → Runs → Jobs → Logs. Expand runs inline with `Enter`, filter by status, toggle the log pane with `L`, go fullscreen with `f`.

```
 All │ ✓ Success │ ✗ Failed │ ⏳ Running                                s:status
╭ Workflows ╮╭ Runs ────────────────────╮╭ Jobs ────╮╭ Logs ───────────────────╮
│All Workflow││▾ ✗ #142 fix deploy  main ││ ✓ Build  ││ Step 3: Run tests       │
│CI          ││   ✗ E2E Tests           ││ ✓ Lint   ││ --- FAIL: TestAuth      │
│Deploy      ││   ✓ Build               ││ ✗ E2E    ││ Expected 200, got 401   │
│Release     ││▸ ✓ #141 add cache  main ││          ││                         │
╰────────────╯╰──────────────────────────╯╰──────────╯╰─────────────────────────╯
```

- **Floating Help** — Press `?` and a keybinding overlay appears over your current view. No page switching, no context loss.

- **Configurable Tabs** — Only care about PRs and Actions? Set `tabs: [prs, actions]` in your config. Tabs reorder, number keys adapt.

- **Custom Keybindings** — Bind any key to a shell command. Launch `lazygit`, run `gh pr diff`, open `nvim` — ghx gets out of your way.

## Quick Start

```bash
# Install (pick one)
brew install onnga-wasabi/tap/ghx       # Homebrew
go install github.com/onnga-wasabi/ghx/cmd/ghx@latest  # Go
nix profile install github:onnga-wasabi/ghx                 # Nix

# Prerequisite: GitHub CLI must be authenticated
gh auth login

# Run in any git repository
cd your-repo
ghx
```

That's it. No tokens to configure, no config files to create. ghx reads your `gh` auth automatically.

## Install

### Homebrew

```bash
brew install onnga-wasabi/tap/ghx
```

### Go

```bash
go install github.com/onnga-wasabi/ghx/cmd/ghx@latest
```

### From Source

```bash
git clone https://github.com/onnga-wasabi/ghx.git
cd ghx
make install
```

### Nix

```bash
# One-off
nix profile install github:onnga-wasabi/ghx

# Home Manager (overlay)
# In your flake.nix inputs:
#   ghx.url = "github:onnga-wasabi/ghx";
# In your nixpkgs overlays:
#   ghx.overlays.default
# Then:
#   home.packages = [ pkgs.ghx ];
```

<details>
<summary>Full Nix Home Manager example</summary>

```nix
{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    home-manager.url = "github:nix-community/home-manager";
    ghx.url = "github:onnga-wasabi/ghx";
  };

  outputs = { nixpkgs, home-manager, ghx, ... }: {
    homeConfigurations."you" = home-manager.lib.homeManagerConfiguration {
      pkgs = import nixpkgs {
        system = "aarch64-darwin";
        overlays = [ ghx.overlays.default ];
      };
      modules = [{
        home.packages = [ pkgs.ghx ];
      }];
    };
  };
}
```

</details>

## Configuration

Config file: `~/.config/ghx/config.yml`

```yaml
defaults:
  prsLimit: 20
  issuesLimit: 20
  view: prs                    # Start on this tab
  tabs:                        # Which tabs, in what order
    - prs
    - actions
  smartLayout: true            # Focused pane auto-expands
  preview:
    open: true
    width: 0.45

keybindings:
  universal:
    - key: "b"
      name: "lazygit"
      command: "lazygit"

theme:
  colors:
    primary: "#7aa2f7"
    secondary: "#bb9af7"
    success: "#9ece6a"
    warning: "#e0af68"
    error: "#f7768e"
```

## Keybindings

Every view shares the same foundation. Learn once, use everywhere.

### Global

| Key | Action |
|-----|--------|
| `?` | Toggle help overlay |
| `q` | Quit |
| `Tab` / `Shift+Tab` | Next / Previous tab |
| `1`–`9` | Jump to tab by number |
| `j`/`k` | Navigate up / down |
| `g` / `G` | First / Last item |
| `Enter` | Toggle sidebar or expand |
| `o` | Open in browser |
| `R` | Refresh |
| `s` | Toggle state filter |
| `←`/`→` | Cycle scope or switch pane |

<details>
<summary>PRs / Issues</summary>

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

</details>

<details>
<summary>Actions</summary>

| Key | Action |
|-----|--------|
| `s` | Cycle status: All → Success → Failed → Running |
| `Enter` | Expand/collapse run (show jobs inline) |
| `L` | Toggle log pane |
| `f` | Fullscreen log viewer |
| `r` | Rerun workflow |
| `Ctrl+R` | Rerun failed jobs |
| `c` | Cancel run |
| `t` | Trigger workflow |
| `w` | Toggle word wrap |
| `n`/`N` | Next / Prev search match |

</details>

<details>
<summary>Notifications</summary>

| Key | Action |
|-----|--------|
| `s` | Toggle Unread / All |
| `←`/`→` | Cycle type: All → PR → Issue → Release → CI |
| `r` | Mark read |
| `d` | Mark done |

</details>

## Acknowledgements

ghx stands on the shoulders of great tools:

- **[gh-dash](https://github.com/dlvhdr/gh-dash)** — The original GitHub dashboard TUI. ghx was born from wanting to add a deep Actions integration to the dashboard experience.
- **[lazygit](https://github.com/jesseduffield/lazygit)** — Proved that complex Git workflows can be fully keyboard-driven. A major design inspiration.
- **[Charm](https://charm.sh/)** — [Bubble Tea](https://github.com/charmbracelet/bubbletea) and [Lip Gloss](https://github.com/charmbracelet/lipgloss) make building beautiful TUIs in Go a joy.

## License

[MIT](LICENSE)

## Disclaimer

This project is provided "AS IS". See [DISCLAIMER.md](DISCLAIMER.md) for details (including limitation of liability).
