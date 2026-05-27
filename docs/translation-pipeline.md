# Translation Pipeline

## Architecture (Schema v2)

The campaign data pipeline uses **two sources**:

1. **Domistae/poe2-leveling** (primary) — human-written HTML campaign guides with semantic markup
2. **Lailloken/Exile-UI** (area IDs only) — zone area IDs for Client.txt matching

### Why Domistae is the primary source

Exile-UI's `default_guide_2.json` uses a custom AHK rendering DSL that encodes visual instructions as markup tokens (e.g., `(img:skill)`, `(color:cc99ff)league`, `<dryad>`). When flattened to plain text, this produces incomprehensible output like `"Exalted Orb League or Support Gem dryad or Skill Gem ritual"`.

The same zone in Domistae reads: `"(Opt) League mech Exalted Orb"`.

The Exile-UI guide data is retained solely for area IDs — the mapping between zone names and the `G1_11`-style IDs that appear in PoE2's `Client.txt` log.

## Pipeline flow

```
data/sources/domistae/*.html  ──→  internal/importer/domistae_parse.go  ──→  domistaeZone
data/sources/exile-ui/areas_2.json  ──→  internal/importer/exileui.go  ──→  campaign.Zone
                                    ↓
                    internal/importer/domistae.go  ──→  campaign.CampaignData
                                    ↓
                    data/campaign.normalized.json
```

### Step 1: Parse Domistae HTML

The HTML is parsed using `golang.org/x/net/html`. Key elements extracted:

| HTML element | CSS class | Extracted data |
|---|---|---|
| `<div>` | `zone-header` | Zone name, level range, town/waypoint flags |
| `<div>` | `step-content` | Step text with semantic `<span>` annotations |
| `<div>` | `note new` / `note tip` / `note warning` | Tips and patch notes |
| `<span>` | `npc` | NPC name |
| `<span>` | `boss` | Boss name |
| `<span>` | `loc` | Location/zone name (used for target detection) |
| `<span>` | `item` | In-game item name |
| `<span>` | `skip` | "(Opt)" / "(Alt)" optional marker |
| `<span>` | `reward-tag` (with subtypes) | Reward markers (gem, craft, perm, asc, etc.) |
| `<span>` | `wp` | Waypoint marker |
| `<span>` | `wp town` | Town marker |
| `<span>` | `action` | Action marker |
| `<span>` | `label` | Note type prefix (New, Pathing, etc.) |

### Step 2: Match zone names to Exile-UI area IDs

The `zoneNameToAreaID` map in `domistae.go` provides hand-curated mappings between Domistae zone names and Exile-UI area IDs. For zones that exist in both sources, the ID map provides the `Client.txt` match key. For zones that only exist in Domistae (e.g., "Act 4 Finale", "Kingsmarch — Hideout"), synthetic zone entries are generated with slug-based IDs.

### Step 3: Generate normalized JSON

The `GenerateNormalizedFromDomistae()` function:
1. Loads Exile-UI zones (area IDs + level ranges)
2. Parses all 5 Domistae HTML files (Acts 1–4 + Interludes)
3. Builds route entries from Domistae steps with human-readable text
4. Detects step types from text patterns and semantic classes
5. Detects target zones from `<span class="loc">` tags
6. Adds synthetic zone entries for Domistae-only zones
7. Writes `data/campaign.normalized.json`

## Step type detection

Step types are inferred from Domistae text patterns and semantic classes:

| Type | Detection |
|---|---|
| `talk` | Contains "talk to" or `<span class="npc">` |
| `boss` | Contains "kill" or "slay" |
| `exit` | Contains "enter" with a `<span class="loc">` target |
| `travel` | Contains "TP", "TP to", "TP back", or "travel to" |
| `waypoint` | Contains "take waypoint" or "take WP" |
| `optional` | Has `<span class="skip">` or starts with "(Opt)" |
| `objective` | Anything else |

## Updating the data

```sh
go run ./cmd/poe2campain update-data
```

This reads from `data/sources/` and writes to `data/campaign.normalized.json`.

## Schema v2 vs v1

Schema v2 uses Domistae as the primary source. The key differences:

- Route step text is human-readable (written by humans for humans)
- Step types are inferred from text patterns rather than DSL markup interpretation
- 327 steps across 92 route entries, 0 fallback (vs 474 steps with many fallback-quality texts)
- Notes/tips are included from Domistae
- Zone names use Domistae naming (more descriptive for players)

## Source file reference

| File | Purpose |
|---|---|
| `internal/importer/domistae.go` | Zone matching, step type detection, campaign data builder |
| `internal/importer/domistae_parse.go` | HTML parser using golang.org/x/net/html |
| `internal/importer/domistae_types.go` | HTML data structures |
| `internal/importer/exileui.go` | Exile-UI area ID loader |
| `data/sources/domistae/*.html` | Source HTML (5 files, MIT licensed) |
| `data/sources/exile-ui/areas_2.json` | Area ID registry |
| `data/sources/sources.json` | Source metadata |