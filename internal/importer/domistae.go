package importer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"poe2campain/internal/campaign"
)

// domistaeSourceName is the display name for the Domistae source.
const domistaeSourceName = "domistae/poe2-leveling"

// domistaeFile maps act number to the HTML filename in data/sources/domistae/.
var domistaeFiles = map[int]string{
	1: "poe2_act1_guide.html",
	2: "poe2_act2_guide.html",
	3: "poe2_act3_guide.html",
	4: "poe2_act4_guide.html",
	5: "poe2_interludes_guide.html",
}

// domistaeActName maps act number to the human-readable act name.
var domistaeActName = map[int]string{
	1: "Act 1",
	2: "Act 2",
	3: "Act 3",
	4: "Act 4",
	5: "Interlude",
}

// --- Parsed HTML structures (defined in domistae_parse.go) ---

// --- Zone header parsing ---

type parsedZoneHeader struct {
	Name       string
	LevelRange string
	LevelMin   *float64
	LevelMax   *float64
	IsTown     bool
	IsWaypoint bool
	IsHub      bool
}

var lvlRe = regexp.MustCompile(`Lvl\s+(\d+(?:[.–−\-]\d+)?)`)
var wpTextRe = regexp.MustCompile(`\bWAYPOINT\b`)
var townTextRe = regexp.MustCompile(`\bTOWN\b`)
var hubTextRe = regexp.MustCompile(`\bHUB\b`)
var fragMatikiRe = regexp.MustCompile(`\s*\(?\s*(?:FRAG|MATIC?KI?)\s*/\s*(?:FRAG|MATIC?KI?)\?\s*\)?\s*`)

func parseZoneHeader(raw string) parsedZoneHeader {
	h := parsedZoneHeader{}

	// Extract level range
	lvlMatch := lvlRe.FindStringSubmatch(raw)
	if len(lvlMatch) > 1 {
		h.LevelRange = lvlMatch[1]
		rangeStr := lvlMatch[1]
		parts := regexp.MustCompile(`[.–−\-]`).Split(rangeStr, -1)
		if len(parts) >= 1 {
			if v, err := strconv.ParseFloat(parts[0], 64); err == nil {
				h.LevelMin = &v
				h.LevelMax = &v
			}
		}
		if len(parts) >= 2 {
			if v, err := strconv.ParseFloat(parts[1], 64); err == nil {
				h.LevelMax = &v
			}
		}
	}

	// Detect flags
	h.IsWaypoint = wpTextRe.MatchString(raw)
	h.IsTown = townTextRe.MatchString(raw)
	h.IsHub = hubTextRe.MatchString(raw)

	// Build name by stripping known tag spans
	name := raw
	name = regexp.MustCompile(`\s*⚠\s*`).ReplaceAllString(name, " ")
	name = fragMatikiRe.ReplaceAllString(name, " ")
	name = regexp.MustCompile(`\s+MAP\b`).ReplaceAllString(name, " ")
	name = wpTextRe.ReplaceAllString(name, "")
	name = townTextRe.ReplaceAllString(name, "")
	name = hubTextRe.ReplaceAllString(name, "")
	name = lvlRe.ReplaceAllString(name, "")
	name = regexp.MustCompile(`\s+`).ReplaceAllString(name, " ")
	name = strings.TrimSpace(name)

	h.Name = name
	return h
}

// --- Step type detection ---

func domistaeStepType(step domistaeStep) string {
	text := step.RawText
	hasOpt := false
	for _, sp := range step.Spans {
		if strings.Contains(sp.Classes, "skip") {
			hasOpt = true
		}
	}

	lower := strings.ToLower(text)

	switch {
	case strings.Contains(lower, "talk to"), strings.Contains(lower, "talk ") && !strings.Contains(lower, "tp"):
		return "talk"
	case strings.Contains(lower, "accept quest"):
		return "talk"
	case hasOpt && strings.Contains(lower, "league mech"):
		return "optional"
	case strings.Contains(lower, "kill ") || strings.Contains(lower, "slay "):
		return "boss"
	case strings.Contains(lower, "enter ") || strings.Contains(lower, "exit to "):
		return "exit"
	case strings.Contains(lower, "tp ") || strings.Contains(lower, "tp to ") || strings.Contains(lower, "tp back") || strings.Contains(lower, "travel to "):
		return "travel"
	case strings.Contains(lower, "take waypoint") || strings.Contains(lower, "take wp") || strings.Contains(lower, "→ take wp"):
		return "waypoint"
	case strings.Contains(lower, "find") && strings.Contains(lower, "waypoint"):
		return "waypoint"
	case hasOpt:
		return "optional"
	default:
		return "objective"
	}
}

// --- Step text enrichment ---

func domistaeStepTags(step domistaeStep, stepType string) []string {
	tags := []string{stepType}
	for _, sp := range step.Spans {
		classes := strings.Fields(sp.Classes)
		for _, cls := range classes {
			switch cls {
			case "npc":
				tags = append(tags, "npc")
			case "boss":
				tags = append(tags, "boss")
			case "loc":
				tags = append(tags, "location")
			case "item":
				tags = append(tags, "item")
			case "skip":
				tags = append(tags, "optional")
			case "wp":
				tags = append(tags, "waypoint")
			}
		}
	}
	return uniqueStrings(tags)
}

func domistaeStepRewards(step domistaeStep) []string {
	var rewards []string
	for _, sp := range step.Spans {
		classes := strings.Fields(sp.Classes)
		for _, cls := range classes {
			if cls == "reward-tag" || strings.HasPrefix(cls, "reward-tag") {
				rewards = append(rewards, sp.Text)
			}
		}
	}
	return uniqueStrings(rewards)
}

// --- Zone name → Exile-UI area ID mapping ---

var zoneNameToAreaID = map[string]string{
	// Act 1
	"Riverbank":                  "g1_1",
	"Clearfell Encampment":       "g1_town",
	"Clearfell":                  "g1_2",
	"The Grelwood":               "g1_4",
	"The Red Vale":               "g1_5",
	"Grim Tangle":                "g1_6",
	"Cemetery of the Eternals":   "g1_7",
	"Mausoleum of the Praetor":   "g1_8",
	"Tomb of the Consort":        "g1_9",
	"Hunting Grounds":            "g1_11",
	"Freythorn":                  "g1_12",
	"Ogham Farmlands":            "g1_13_1",
	"Ogham Village":              "g1_13_2",
	"The Manor Ramparts":         "g1_14",
	"Ogham Manor":                "g1_15",
	"Mud Burrow":                 "g1_3",

	// Act 2
	"Vastiri Outskirts":     "g2_1",
	"The Ardura Caravan":    "g2_town",
	"Ardura Caravan":        "g2_town",
	"Mawdun Quarry":         "g2_10_1",
	"Mawdun Mine":           "g2_10_2",
	"Traitor's Passage":     "g2_2",
	"The Halani Gates":      "g2_3",
	"Halani Gates":          "g2_3",
	"Mastodon Badlands":     "g2_5_1",
	"The Bone Pits":         "g2_5_2",
	"Bone Pits":             "g2_5_2",
	"Keth":                  "g2_4_1",
	"The Lost City":         "g2_4_2",
	"Lost City":             "g2_4_2",
	"Buried Shrines":        "g2_4_3",
	"The Heart of Keth":     "g2_4_3",
	"Valley of the Titans":  "g2_6",
	"The Titan Grotto":      "g2_7",
	"Titan Grotto":          "g2_7",
	"Deshar":                "g2_8",
	"Path of Mourning":      "g2_9_1",
	"The Spires of Deshar":  "g2_9_2",
	"Spires of Deshar":      "g2_9_2",
	"Trial of the Sekhemas": "g2_13",
	"Trial of Sekhemas":     "g2_13",
	"The Dreadnought":       "g2_12_1",
	"Dreadnought":           "g2_12_1",

	// Act 3
	"Sandswept Marsh":          "g3_1",
	"Ziggurat Encampment":      "g3_town",
	"Jungle Ruins":             "g3_3",
	"Infested Barrens":         "g3_2_1",
	"The Venom Crypts":          "g3_4",
	"Venom Crypts":             "g3_4",
	"Chimeral Wetlands":         "g3_5",
	"Trial of Chaos":            "g3_10_airlock",
	"Jiquani's Machinarium":     "g3_6_1",
	"Jiquani's Sanctum":         "g3_6_2",
	"Azak Bog":                  "g3_7",
	"The Matlan Waterways":      "g3_2_2",
	"Matlan Waterways":          "g3_2_2",
	"The Drowned City":          "g3_8",
	"Drowned City":              "g3_8",
	"The Molten Vault":          "g3_9",
	"Molten Vault":              "g3_9",
	"Apex of Filth":             "g3_11",
	"Temple of Kopec":           "g3_12",
	"Utzaal":                    "g3_14",
	"Aggorat":                   "g3_16",
	"The Black Chambers":        "g3_17",
	"Black Chambers":            "g3_17",

	// Act 4
	"Kingsmarch":              "g4_town",
	"Whakapanu Island":         "g4_3_1",
	"Singing Caverns":          "g4_3_2",
	"Shrike Island":            "g4_7",
	"Abandoned Prison":          "g4_5_1",
	"Solitary Confinement":      "g4_5_2",
	"Isle of Kin":              "g4_1_1",
	"Volcanic Warrens":          "g4_1_2",
	"Eye of Hinekora":          "g4_4_1",
	"Halls of the Dead":         "g4_4_2",
	"Trial of the Ancestors":    "g4_4_3",
	"Trial of Ancestors":        "g4_4_3",
	"Kedge Bay":                 "g4_2_1",
	"Journey's End":             "g4_2_2",
	"Arastas":                   "g4_8a",
	"The Excavation":             "g4_10",
	"Excavation":                "g4_10",
	"Ngakanu":                   "g4_11_1a",
	"Heart of the Tribe":         "g4_11_2",
	"Plunder's Point":            "g4_13",

	// Interludes
	"Ogham, The Refuge":     "p1_6",
	"Wolvenhold":            "p1_6",
	"Khari Bazaar":          "p2_town",
	"Khari Crossing":         "p2_1",
	"Mount Kriar, The Glade": "p3_town",
	"The Glade":             "p3_town",
	"Ashen Forest":          "p3_1",
	"Kriar Village":          "p3_2",
	"Glacial Tarn":          "p3_3",
	"Howling Caves":          "p3_4",
	"Kriar Peaks":             "p3_5",
}

func lookupAreaID(zoneName string) string {
	if id, ok := zoneNameToAreaID[zoneName]; ok {
		return id
	}
	// Strip common suffixes
	trimmed := zoneName
	suffixes := []string{" (Return)", " (End of Act)"}
	for _, suffix := range suffixes {
		trimmed = strings.TrimSuffix(trimmed, suffix)
		if id, ok := zoneNameToAreaID[trimmed]; ok {
			return id
		}
	}
	// Strip leading "The "
	trimmed = strings.TrimPrefix(zoneName, "The ")
	if id, ok := zoneNameToAreaID[trimmed]; ok {
		return id
	}
	return ""
}

// --- Build normalized data ---

type domistaeActData struct {
	ActNum   int
	ActName  string
	FileName string
	Zones    []domistaeZone
}

func buildCampaignFromDomistae(domZones []domistaeActData, exileZones []campaign.Zone, exileZoneByID map[string]campaign.Zone) (*campaign.CampaignData, error) {
	// Build a fast lookup for Exile-UI zones by name
	exileZoneByName := make(map[string]campaign.Zone)
	for _, z := range exileZones {
		for _, alias := range z.Aliases {
			exileZoneByName[strings.ToLower(alias)] = z
		}
		exileZoneByName[strings.ToLower(z.Name)] = z
	}

	// Merge zones: start from Exile-UI zones (for area IDs + Client.txt detection),
	// then add synthetic zones for Domistae entries that lack Exile-UI matches.
	zones := make([]campaign.Zone, len(exileZones))
	copy(zones, exileZones)

	// Track which zone IDs we have
	seenZoneIDs := make(map[string]bool)
	for _, z := range zones {
		seenZoneIDs[z.ID] = true
	}

	// Route entries built from Domistae data
	var route []campaign.RouteEntry
	order := 10

	for _, act := range domZones {
		for _, z := range act.Zones {
			header := parseZoneHeader(z.RawHeader)
			areaID := lookupAreaID(header.Name)

			// If no direct mapping, try matching against Exile-UI zone by name
			if areaID == "" {
				if ez, ok := exileZoneByName[strings.ToLower(header.Name)]; ok {
					areaID = ez.ID
				}
			}

			// Find the Exile-UI zone for level data
			var exileZone campaign.Zone
			if areaID != "" {
				if ez, ok := exileZoneByID[areaID]; ok {
					exileZone = ez
				}
			}

			displayName := header.Name
			if displayName == "" {
				displayName = z.RawHeader
			}

			// Level range
			levelRange := header.LevelRange
			levelMin := header.LevelMin
			levelMax := header.LevelMax
			if levelMin == nil && exileZone.LevelMin != nil {
				levelMin = exileZone.LevelMin
				levelMax = exileZone.LevelMax
				if levelRange == "" && exileZone.LevelRecommendation != "" {
					levelRange = exileZone.LevelRecommendation
				}
			}

			// Flags
			flags := []string{}
			if header.IsTown || exileZone.Kind == "town" {
				flags = append(flags, "town")
			}
			if header.IsWaypoint {
				flags = append(flags, "waypoint")
			}
			if header.IsHub {
				flags = append(flags, "hub")
			}

			// Steps
			steps := make([]campaign.Step, 0, len(z.Steps))
			for i, s := range z.Steps {
				text := cleanStepText(s.RawText)
				if text == "" {
					continue
				}

				stepType := domistaeStepType(s)
				optional := false
				for _, sp := range s.Spans {
					if strings.Contains(sp.Classes, "skip") {
						optional = true
					}
				}
				if strings.HasPrefix(s.RawText, "(Opt)") || strings.HasPrefix(s.RawText, "(Alt)") {
					optional = true
				}

				// Detect target zone from <span class="loc">
				var targetID, targetName *string
				for _, sp := range s.Spans {
					if strings.Contains(sp.Classes, "loc") {
						locName := sp.Text
						locID := lookupAreaID(locName)
						if locID == "" {
							if ez, ok := exileZoneByName[strings.ToLower(locName)]; ok {
								locID = ez.ID
							}
						}
						if locID != "" {
							targetID = &locID
							dn := locName
							targetName = &dn
						}
					}
				}

				tags := domistaeStepTags(s, stepType)
				rewards := domistaeStepRewards(s)


				steps = append(steps, campaign.Step{
					Order:        (i + 1) * 10,
					Type:         stepType,
					Text:         text,
					Optional:     optional,
					TargetZoneID: targetID,
					TargetZone:   targetName,
					Tags:         tags,
					Rewards:      rewards,
					Source:       domistaeSourceName,
				})
			}

			// Notes
			var notes []string
			for _, n := range z.Notes {
				text := cleanStepText(n.RawText)
				if text != "" {
					notes = append(notes, text)
				}
			}

			// Next zone IDs from exit steps
			var nextZoneIDs, nextZones []string
			for _, s := range steps {
				if s.TargetZoneID != nil && s.Type == "exit" {
					found := false
					for _, nz := range nextZoneIDs {
						if nz == *s.TargetZoneID {
							found = true
							break
						}
					}
					if !found {
						nextZoneIDs = append(nextZoneIDs, *s.TargetZoneID)
						nextZones = append(nextZones, *s.TargetZone)
					}
				}
			}

			zoneID := areaID
			if zoneID == "" {
				zoneID = slugify(header.Name)
			}

			// If this zone doesn't exist in the zones list, add a synthetic zone
			if !seenZoneIDs[zoneID] && zoneID != "" {
				syntheticZone := campaign.Zone{
					ID:      zoneID,
					Name:    header.Name,
					Aliases: []string{zoneID, strings.ToLower(header.Name)},
					Act:     act.ActName,
					Kind:                zoneKindFromHeader(header),
					Flags:               flags,
					LevelRecommendation: levelRange,
					LevelMin:            levelMin,
					LevelMax:            levelMax,
				}
				zones = append(zones, syntheticZone)
				seenZoneIDs[zoneID] = true
			}

			// Zone name from Exile-UI if available
			zoneName := header.Name
			if exileZone.ID != "" {
				zoneName = exileZone.Name
			}

			aliases := []string{zoneID, zoneName}
			if areaID != "" && areaID != zoneID {
				aliases = append(aliases, areaID)
			}
			if header.Name != zoneName {
				aliases = append(aliases, header.Name)
			}
			if exileZone.ID != "" {
				aliases = append(aliases, exileZone.Aliases...)
			}

			entry := campaign.RouteEntry{
				Order:       order,
				Act:         act.ActName,
				ZoneID:      zoneID,
				ZoneName:    header.Name,
				DisplayName: header.Name,
				LevelRange:  levelRange,
				LevelMin:    levelMin,
				LevelMax:    levelMax,
				Flags:       flags,
				Steps:       steps,
				Notes:       uniqueStrings(notes),
				NextZoneIDs: nextZoneIDs,
				NextZones:   nextZones,
			}
			route = append(route, entry)
			order += 10
		}
	}

	return &campaign.CampaignData{
		SchemaVersion: 2,
		Game:          "poe2",
		GeneratedAt:   time.Now().UTC().Format(time.RFC3339),
		Sources:       buildDomistaeSources(),
		Zones:         zones,
		Route:         route,
	}, nil
}

func cleanStepText(text string) string {
	text = regexp.MustCompile(`^\(Opt\)\s*`).ReplaceAllString(text, "")
	text = regexp.MustCompile(`^\(Alt\)\s*`).ReplaceAllString(text, "")
	text = regexp.MustCompile(`^\(1st char\)\s*`).ReplaceAllString(text, "")
	text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")
	return strings.TrimSpace(text)
}

func slugify(name string) string {
	name = strings.ToLower(name)
	name = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(name, "_")
	name = regexp.MustCompile(`^_|_$`).ReplaceAllString(name, "")
	return name
}

func zoneKindFromHeader(h parsedZoneHeader) string {
	if h.IsTown {
		return "town"
	}
	if h.IsHub {
		return "hub"
	}
	return "area"
}

func buildDomistaeSources() []campaign.Source {
	return []campaign.Source{
		{
			Name:    "domistae/poe2-leveling",
			Repo:    "https://github.com/domistae/poe2-leveling",
			Commit:  "743f0934c246253801a8463c398322952025ab41",
			License: "MIT",
			Files: []string{
				"domistae/poe2_act1_guide.html",
				"domistae/poe2_act2_guide.html",
				"domistae/poe2_act3_guide.html",
				"domistae/poe2_act4_guide.html",
				"domistae/poe2_interludes_guide.html",
			},
		},
		{
			Name:    "Lailloken/Exile-UI",
			Repo:    "https://github.com/Lailloken/Exile-UI",
			Commit:  "5f3185dd58672baa2859f7357c0704afc18ee7af",
			License: "MIT",
			Files: []string{
				"exile-ui/areas_2.json",
			},
		},
	}
}

// GenerateNormalizedFromDomistae reads Domistae HTML + Exile-UI areas and writes the normalized runtime JSON.
func GenerateNormalizedFromDomistae(sourceDir, outputPath string) error {
	// Load Exile-UI zones
	exileZones, exileZoneByID, err := loadExileUIZones(sourceDir)
	if err != nil {
		return fmt.Errorf("loading Exile-UI zones: %w", err)
	}

	// Parse all Domistae HTML files
	var domZones []domistaeActData
	for actNum := 1; actNum <= 5; actNum++ {
		fileName, ok := domistaeFiles[actNum]
		if !ok {
			continue
		}
		filePath := filepath.Join(sourceDir, "domistae", fileName)
		zones, err := parseDomistaeHTML(filePath, actNum)
		if err != nil {
			return fmt.Errorf("parsing %s: %w", fileName, err)
		}
		domZones = append(domZones, domistaeActData{
			ActNum:   actNum,
			ActName:  domistaeActName[actNum],
			FileName: fileName,
			Zones:   zones,
		})
	}

	data, err := buildCampaignFromDomistae(domZones, exileZones, exileZoneByID)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return err
	}
	encoded, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	encoded = append(encoded, '\n')
	return os.WriteFile(outputPath, encoded, 0644)
}

// capitalizeFirst capitalizes the first rune of a string.
func capitalizeFirst(s string) string {
	if s == "" {
		return s
	}
	runes := []rune(s)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}