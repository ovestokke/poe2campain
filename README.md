# poe2campain

A fast terminal campaign helper for **Path of Exile 2**.

`poe2campain` watches `Client.txt`, matches your current area, and keeps your leveling route in sync while you push toward maps.

No overlay. No memory reading. No automation. Just a clean terminal guide that stays out of the way.

---

## Why this exists

Most PoE helpers either try to live on top of the game or depend on heavier integrations.

`poe2campain` is intentionally simple:

- follows your campaign route from `Client.txt`
- works offline from bundled campaign data
- auto-detects `Client.txt` from Steam (including secondary library drives)
- lets you move manually when auto-detection is wrong
- keeps your progress between sessions
- fits nicely in a small floating terminal window

---

## Install

### Download a release

Grab the latest release for your platform from **GitHub Releases**.

Each release archive already includes:

- the `poe2campain` binary
- the bundled `data/` directory

### Supported release targets

- Linux `amd64`, `arm64`
- macOS `amd64`, `arm64`
- Windows `amd64`, `arm64`

---

## Terminal requirements

For the UI to render properly, use:

- a **modern terminal emulator** with good Unicode/color support, such as:
  - **Ghostty**
  - **iTerm2**
  - **WezTerm**
  - **kitty**
  - **Alacritty**
  - **Windows Terminal**
  - **GNOME Terminal**
  - **Konsole**
- a **[Nerd Font](https://www.nerdfonts.com)** enabled in that terminal

---

## Quick start

### 1. Launch it

**Linux / macOS**
```sh
./poe2campain
```

**Windows (PowerShell)**
```powershell
.\poe2campain.exe
```

That's it. `Client.txt` is auto-detected from standard Steam and standalone install paths, including secondary Steam library drives. Progress is saved automatically between sessions.

### 2. (Optional) Set a custom `Client.txt` path

If auto-detection doesn't find your install, configure it manually:

```sh
./poe2campain config set-client /path/to/Client.txt
```

Or pass it directly:

```sh
./poe2campain --client /path/to/Client.txt
```

---

## Auto-detection

On startup, `poe2campain` looks for `Client.txt` in this order:

1. **`--client` flag** — explicit path
2. **Config file** — set via `config set-client`
3. **Default paths** — checked automatically:
   - `~/.steam/steam/steamapps/common/Path of Exile 2/logs/Client.txt`
   - `~/.local/share/Steam/steamapps/common/Path of Exile 2/logs/Client.txt`
   - `~/Path of Exile 2/logs/Client.txt` (standalone)
   - All paths discovered from Steam's `libraryfolders.vdf` (secondary drives)

The first existing file is used. If nothing is found, you'll see which paths were checked and how to configure one manually.

---

## Controls

```text
↑/k    step up       ←  zone back
↓/j    step down      →  zone forward
space  done + next / undo
h      toggle help
q      quit
```

Behavior notes:

- area detection only moves the guide **forward**
- the current route, step, and manually completed steps are saved between runs
- if you miss a step in-game, fix it manually

---

## Commands

### Main (TUI mode)

```sh
./poe2campain [--client path] [--config path] [--data path]
```

Flags:
```
  -client string    Path of Exile 2 Client.txt path (auto-detected if empty)
  -config string    config file path (default ~/.config/poe2campain/config.json)
  -data string      normalized campaign data path (default "data/campaign.normalized.json")
  -debug-client     scan Client.txt and show the latest matched area
  -debug-zone string  match and render route entry for a zone or area ID
  -list-zones       list known zones
```

### Config

```sh
./poe2campain config path                  # show config file path
./poe2campain config show                  # show current config
./poe2campain config init [--client path]  # create config
./poe2campain config set-client /path      # set Client.txt path
```

### State

```sh
./poe2campain state path   # show state file path
./poe2campain state show   # show current state
./poe2campain state reset  # delete state file
```

### Data

```sh
./poe2campain update-data    # regenerate normalized data from sources
./poe2campain validate-data  # validate campaign data
```

### Help

```sh
./poe2campain --help
./poe2campain help
./poe2campain -h
./poe2campain config --help
./poe2campain state --help
```

---

## Floating terminal setup

`poe2campain` works especially well in a small pinned terminal beside the game.

### Niri (Wayland)

Based on your Ghostty setup:

```kdl
window-rule {
    match app-id="com.mitchellh.ghostty" title="poe2campain"
    open-floating true
    default-floating-position x=2240 y=1100
    default-column-width { fixed 620; }
    default-window-height { fixed 120; }
    opacity 0.7
}
```

### Hyprland (Wayland)

```ini
windowrulev2 = float, class:^(ghostty)$, title:^(poe2campain)$
windowrulev2 = pin, class:^(ghostty)$, title:^(poe2campain)$
windowrulev2 = size 620 140, class:^(ghostty)$, title:^(poe2campain)$
windowrulev2 = opacity 0.7, class:^(ghostty)$, title:^(poe2campain)$
```

If you use another terminal, swap the app-id/class accordingly.

---

## Build from source

Only needed if you want to build it yourself.

### Requirements

- **Go 1.25+**

### Build

```sh
mkdir -p .build
go test ./...
go build -o .build/poe2campain ./cmd/poe2campain
```

Or use the helper script, which writes a runnable local bundle to `.build/`:

```sh
./build.sh
./.build/poe2campain
```

---

## Data pipeline

Runtime guide data is bundled in:

- `data/campaign.normalized.json`

Source snapshots live in:

- `data/sources/`

### Source roles

**Primary guide text source**
- `domistae/poe2-leveling`

**Area ID matching source**
- `Lailloken/Exile-UI`

Domistae provides the human-readable campaign route and step text.
Exile-UI is used for area IDs and zone matching against `Client.txt`.

---

## Philosophy

`poe2campain` is deliberately narrow in scope.

It will:
- read `Client.txt`
- match zones
- render campaign guidance
- save lightweight progress state

It will **not**:
- read memory
- inject into the game
- automate inputs
- run as an overlay
- depend on Electron or AHK

---

## Credits

- [domistae/poe2-leveling](https://github.com/domistae/poe2-leveling) — campaign guide text
- [Lailloken/Exile-UI](https://github.com/Lailloken/Exile-UI) — area IDs and zone names
- Original fork inspiration: [wiiittttt/poecampain](https://github.com/wiiittttt/poecampain)
