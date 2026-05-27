# Source data snapshot

This directory contains offline snapshots of reusable upstream data for `poe2campain`.

Only MIT-licensed sources are copied here. Sources with unclear licensing are intentionally excluded.

## Primary source

### domistae/poe2-leveling

- Repo: <https://github.com/domistae/poe2-leveling>
- Commit: `743f0934c246253801a8463c398322952025ab41`
- License: MIT, copied at `domistae/LICENSE`
- Role: **primary** — human-readable campaign guide text, step structure, and zone progression

Copied files:

```text
domistae/poe2_act1_guide.html
domistae/poe2_act2_guide.html
domistae/poe2_act3_guide.html
domistae/poe2_act4_guide.html
domistae/poe2_interludes_guide.html
domistae/LICENSE
```

The Domistae HTML provides:
- Zone headers with level ranges, town/waypoint flags
- Step-by-step instructions with semantic `<span>` annotations (NPC, boss, location, item, reward type, skip-marker)
- Optional/alternative step markers from `class="skip"` spans
- Notes/tips with class markers (new, tip, warning, success)
- Progress tracking via `data-step` attributes
- 327 steps across 92 zone route entries, 0 fallback

Key HTML classes used in parsing:
- `zone-header` — zone name + level range + waypoint/town flags
- `step-content` — step text with semantic spans
- `span.npc` — NPC name
- `span.boss` — boss name
- `span.loc` — location/zone name
- `span.item` — in-game item
- `span.skip` — "(Opt)" or "(Alt)" marker
- `span.reward-tag` + subtypes (`gem`, `craft`, `perm`, `asc`, `optional`, `choice`, `map`, `frag`, `quest`) — reward markers
- `span.wp` — waypoint marker, `span.wp.town` — town marker
- `span.action` — action marker
- `div.note` with subtypes (`new`, `tip`, `warning`, `success`) — tips and patch notes

## Area ID source

### Lailloken/Exile-UI

- Repo: <https://github.com/Lailloken/Exile-UI>
- Commit: `5f3185dd58672baa2859f7357c0704afc18ee7af`
- License: MIT, copied at `exile-ui/LICENSE.md`
- Role: area IDs for PoE2 Client.txt zone matching (not used for guide text)

Copied files:

```text
exile-ui/areas_2.json
exile-ui/LICENSE.md
```

The Exile-UI guide data (`default_guide_2.json`) is **not used** for step text because its DSL markup (`(img:skill)`, `(color:cc99ff)league`, `<dryad>`, etc.) cannot be reliably interpreted into human-readable text by automated translation. Exile-UI area IDs and names are used solely for mapping Domistae zone names to the Client.txt area ID matching system.

## Why Domistae replaced Exile-UI as the primary source

The Exile-UI `default_guide_2.json` uses a custom AHK rendering DSL that encodes visual instructions as markup tokens:

- `(img:skill)` → tiny inline icon (no text alternative)
- `(color:cc99ff)league` → lavender-colored label meaning "League mechanic encounter"
- `(img:exa) (color:cc99ff)league || (img:support) <dryad> || (img:skill) ritual` → three alternative optional rewards
- `<dryad>` → class tag for boss name styling

When flattened to plain text, this produces: `"Exalted Orb League or Support Gem dryad or Skill Gem ritual"` — incomprehensible.

In contrast, the same zone in Domistae reads: `"(Opt) League mech Exalted Orb"` — clear, human-readable, zero ambiguity.

## Excluded sources

- Mobalytics pages: reference only; not copied/imported because reusable license is not confirmed.
- `nicolasbagatello/poe2-helper`: not copied/imported because no clear open-source license was found during inspection.

## Verification commands

```sh
go run ./cmd/poe2campain update-data
go run ./cmd/poe2campain validate-data
```

Expected output: 327 steps, 0 fallback, 109 zones (98 Exile-UI + 11 synthetic from Domistae), 92 route entries.