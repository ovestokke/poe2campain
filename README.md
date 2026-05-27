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

**You do not need Go installed to run release builds.**

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

### 1. Configure your `Client.txt`

**Linux**
```sh
./poe2campain config set-client ~/.steam/steam/steamapps/common/'Path of Exile 2'/logs/Client.txt
```

**macOS**
```sh
./poe2campain config set-client ~/Library/Application\ Support/Steam/steamapps/common/'Path of Exile 2'/logs/Client.txt
```

**Windows (PowerShell)**
```powershell
.\poe2campain config set-client 'C:\Program Files (x86)\Steam\steamapps\common\Path of Exile 2\logs\Client.txt'
```

### 2. Launch it

**Linux / macOS**
```sh
./poe2campain
```

**Windows (PowerShell)**
```powershell
.\poe2campain.exe
```

If no `Client.txt` is configured yet, the app will tell you what command to run.

---

## Default `Client.txt` paths

| OS | Path |
|---|---|
| Linux | `~/.steam/steam/steamapps/common/Path of Exile 2/logs/Client.txt` |
| macOS | `~/Library/Application Support/Steam/steamapps/common/Path of Exile 2/logs/Client.txt` |
| Windows | `C:\Program Files (x86)\Steam\steamapps\common\Path of Exile 2\logs\Client.txt` |

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

## Useful commands

### Show config path
```sh
./poe2campain config path
```

### Show current config
```sh
./poe2campain config show
```

### Inspect saved state
```sh
./poe2campain state show
```

### Reset saved state
```sh
./poe2campain state reset
```

### Debug latest detected area from `Client.txt`
```sh
./poe2campain --debug-client
```

### Inspect a specific zone or generated area ID
```sh
./poe2campain --debug-zone G1_13_2
```

### List all known zones
```sh
./poe2campain --list-zones
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
go test ./...
go build ./cmd/poe2campain
```

Or use the helper script:

```sh
./build.sh
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
