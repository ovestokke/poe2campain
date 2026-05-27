package importer

import (
	"encoding/json"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"poe2campain/internal/campaign"
)

type rawArea struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Recommendation string `json:"recommendation"`
}

func readAreas(path string) ([][]rawArea, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var areas [][]rawArea
	if err := json.Unmarshal(b, &areas); err != nil {
		return nil, err
	}
	return areas, nil
}

func buildZones(areaGroups [][]rawArea) ([]campaign.Zone, map[string]campaign.Zone) {
	zones := make([]campaign.Zone, 0, len(areaGroups))
	byID := make(map[string]campaign.Zone)

	for groupIndex, group := range areaGroups {
		act := actName(groupIndex)
		for _, area := range group {
			id := strings.ToLower(strings.TrimSpace(area.ID))
			name := titleName(area.Name)
			levelMin, levelMax := parseLevelRecommendation(area.Recommendation)
			zone := campaign.Zone{
				ID:                  id,
				Name:                name,
				Aliases:             uniqueStrings([]string{id, area.ID, area.Name, name, withoutLeadingThe(area.Name), withoutLeadingThe(name)}),
				Act:                 act,
				LevelRecommendation: area.Recommendation,
				LevelMin:            levelMin,
				LevelMax:            levelMax,
				Kind:                zoneKind(id),
				Flags:               zoneFlags(id, name),
			}
			zones = append(zones, zone)
			byID[id] = zone
		}
	}

	return zones, byID
}

func actName(groupIndex int) string {
	switch groupIndex {
	case 0:
		return "Act 1"
	case 1:
		return "Act 2"
	case 2:
		return "Act 3"
	case 3:
		return "Act 4"
	case 4:
		return "Interlude 1"
	case 5:
		return "Interlude 2"
	case 6:
		return "Interlude 3"
	default:
		return "Other"
	}
}

var levelNumberRe = regexp.MustCompile(`\d+(?:\.\d+)?`)

func parseLevelRecommendation(recommendation string) (*float64, *float64) {
	numbers := levelNumberRe.FindAllString(recommendation, -1)
	if len(numbers) == 0 {
		return nil, nil
	}
	min, err := strconv.ParseFloat(numbers[0], 64)
	if err != nil {
		return nil, nil
	}
	max := min
	if len(numbers) > 1 {
		if parsedMax, err := strconv.ParseFloat(numbers[1], 64); err == nil {
			max = parsedMax
		}
	}
	return &min, &max
}

func zoneKind(id string) string {
	if strings.Contains(id, "town") {
		return "town"
	}
	return "area"
}

func zoneFlags(id, name string) []string {
	flags := []string{}
	if strings.Contains(id, "town") {
		flags = append(flags, "town")
	}
	if strings.Contains(strings.ToLower(name), "hideout") {
		flags = append(flags, "hideout")
	}
	return flags
}

func titleName(s string) string {
	words := strings.Fields(strings.ReplaceAll(s, "_", " "))
	for i, word := range words {
		lower := strings.ToLower(word)
		if i > 0 && smallTitleWord(lower) {
			words[i] = lower
			continue
		}
		words[i] = capitalize(word)
	}
	return strings.Join(words, " ")
}

func smallTitleWord(word string) bool {
	switch word {
	case "a", "an", "and", "at", "for", "from", "in", "of", "on", "or", "the", "to", "with":
		return true
	default:
		return false
	}
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	runes := []rune(s)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

func withoutLeadingThe(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(strings.ToLower(s), "the ") {
		return strings.TrimSpace(s[4:])
	}
	return s
}

func uniqueStrings(values []string) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		key := strings.ToLower(value)
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}