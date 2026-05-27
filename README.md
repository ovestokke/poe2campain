# poe2campain

Terminal-only Path of Exile 2 campaign helper.

## Scope

`poe2campain` will:

- read only Path of Exile 2 `Client.txt`
- detect generated area IDs from verified PoE2 log lines
- match those IDs to normalized campaign data
- render campaign route/step guidance in a terminal UI
- support manual route/step navigation when automatic matching is wrong
- work offline at runtime from `data/campaign.normalized.json`

It will not use overlays, memory reading, process hooks, input automation, Electron, AHK, or game integration beyond reading `Client.txt`.

## Requirements

- **Go 1.25+** (see `go.mod`)
- **A [Nerd Font](https://www.nerdfonts.com)** installed and active in your terminal (used for direction arrows and step icons)
- **Path of Exile 2** `Client.txt` log file accessible on disk

## Source data

Current source snapshots live under `data/sources/`.

Primary MIT source:

- `Lailloken/Exile-UI`
- commit `5f3185dd58672baa2859f7357c0704afc18ee7af`
- files copied under `data/sources/exile-ui/`

Optional MIT enrichment source:

- `domistae/poe2-leveling`
- commit `743f0934c246253801a8463c398322952025ab41`
- files copied under `data/sources/domistae/`

See `data/sources/README.md` and `data/sources/sources.json`.

## Build

```sh
go test ./...
go build ./cmd/poe2campain
```

Or use the helper script (outputs to `dist/`):

```sh
./build.sh
```

### Running on each platform

Build produces a single static binary. Place it alongside the `data/` directory.

**Linux:**

```sh
go build -ldflags='-s -w' -trimpath -o poe2campain ./cmd/poe2campain
./poe2campain config set-client ~/.steam/steam/steamapps/common/'Path of Exile 2'/logs/Client.txt
./poe2campain
```

**macOS:**

```sh
go build -ldflags='-s -w' -trimpath -o poe2campain ./cmd/poe2campain
./poe2campain config set-client ~/Library/Application\ Support/Steam/steamapps/common/'Path of Exile 2'/logs/Client.txt
./poe2campain
```

**Windows (PowerShell):**

```powershell
go build -ldflags='-s -w' -trimpath -o poe2campain.exe ./cmd/poe2campain
.\poe2campain config set-client 'C:\Program Files (x86)\Steam\steamapps\common\Path of Exile 2\logs\Client.txt'
.\poe2campain
```

Default `Client.txt` paths per OS:

| OS      | Default path                                                                                          |
|---------|-------------------------------------------------------------------------------------------------------|
| Linux   | `~/.steam/steam/steamapps/common/Path of Exile 2/logs/Client.txt`                                   |
| macOS   | `~/Library/Application Support/Steam/steamapps/common/Path of Exile 2/logs/Client.txt`               |
| Windows | `C:\Program Files (x86)\Steam\steamapps\common\Path of Exile 2\logs\Client.txt` |

## Current CLI

Regenerate the normalized offline runtime data:

```sh
go run ./cmd/poe2campain update-data
```

Validate the normalized data:

```sh
go run ./cmd/poe2campain validate-data
```

List known zone IDs:

```sh
go run ./cmd/poe2campain --list-zones
```

Inspect guide state for a generated area ID or zone:

```sh
go run ./cmd/poe2campain --debug-zone G1_13_2
```

Configure your `Client.txt` path:

```sh
go run ./cmd/poe2campain config set-client '/path/to/Path of Exile 2/logs/Client.txt'
go run ./cmd/poe2campain config show
```

The default config path is:

```sh
go run ./cmd/poe2campain config path
```

A local `config.json` is gitignored if you want to use `--config config.json` during development.

The user config intentionally only stores user-specific settings like `client_txt`. The bundled campaign data is found automatically from the working directory during development or next to the executable in release builds. `--data` remains available as a developer override.

Scan `Client.txt` once and show the latest detected area:

```sh
go run ./cmd/poe2campain --debug-client
```

Run the live terminal UI:

```sh
go run ./cmd/poe2campain
```

Live mode watches `Client.txt` and follows the ordered route from `data/campaign.normalized.json`. Area detection can move the guide forward to the next matching route entry, but it does not rewind automatically. It saves the current route/step under your user state directory.

Progress state commands:

```sh
go run ./cmd/poe2campain state path
go run ./cmd/poe2campain state show
go run ./cmd/poe2campain state reset
```

```text
↑/k  step up       ←  zone back
↓/j  step down      →  zone forward
h    toggle help
q    quit
```

You can still override the config from the command line:

```sh
go run ./cmd/poe2campain --debug-client --client '/path/to/Path of Exile 2/logs/Client.txt'
```

Act 4 follows the imported route order like the rest of the campaign.

## Window manager rules

Pin poe2campain as a small floating overlay alongside the game.

### Niri (Wayland)

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

Adjust `x`, `y`, width, and height for your monitor. Replace `com.mitchellh.ghostty` with your terminal's app-id if you don't use Ghostty.

### Hyprland (Wayland)

```ini
windowrulev2 = float, class:^(ghostty)$, title:^(poe2campain)$
windowrulev2 = pin, class:^(ghostty)$, title:^(poe2campain)$
windowrulev2 = size 620 140, class:^(ghostty)$, title:^(poe2campain)$
windowrulev2 = opacity 0.7, class:^(ghostty)$, title:^(poe2campain)$
```

Replace `ghostty` with your terminal's class if different (`wezterm`, `Alacritty`, etc.).

## Credits

- [Lailloken/Exile-UI](https://github.com/Lailloken/Exile-UI) — area IDs and zone names (MIT)
- [domistae/poe2-leveling](https://github.com/domistae/poe2-leveling) — campaign guide text (MIT)
- Original fork: [wiiittttt/poecampain](https://github.com/wiiittttt/poecampain)