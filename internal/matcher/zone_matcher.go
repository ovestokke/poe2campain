package matcher

import (
	"strings"
	"unicode"

	"poe2campain/internal/campaign"
)

type Matcher struct {
	byID    map[string]int
	byAlias map[string]int
}

func New(data *campaign.CampaignData) Matcher {
	m := Matcher{
		byID:    map[string]int{},
		byAlias: map[string]int{},
	}
	for i, zone := range data.Zones {
		m.byID[strings.ToLower(zone.ID)] = i
		if key := Normalize(zone.Name); key != "" {
			if _, exists := m.byAlias[key]; !exists {
				m.byAlias[key] = i
			}
		}
		for _, alias := range zone.Aliases {
			key := Normalize(alias)
			if key == "" {
				continue
			}
			if _, exists := m.byAlias[key]; !exists {
				m.byAlias[key] = i
			}
		}
	}
	return m
}

func (m Matcher) FindZone(input string, data *campaign.CampaignData) (campaign.Zone, bool) {
	id := strings.ToLower(strings.TrimSpace(input))
	if idx, ok := m.byID[id]; ok {
		return data.Zones[idx], true
	}
	if idx, ok := m.byAlias[Normalize(input)]; ok {
		return data.Zones[idx], true
	}
	return campaign.Zone{}, false
}

func RouteIndexesForZoneID(data *campaign.CampaignData, zoneID string) []int {
	zoneID = strings.ToLower(strings.TrimSpace(zoneID))
	indexes := []int{}
	for i, entry := range data.Route {
		if strings.ToLower(entry.ZoneID) == zoneID {
			indexes = append(indexes, i)
		}
	}
	return indexes
}

func Normalize(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.TrimPrefix(s, "the ")
	var b strings.Builder
	lastSpace := false
	for _, r := range s {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			b.WriteRune(r)
			lastSpace = false
		case r == '_':
			b.WriteRune(r)
			lastSpace = false
		case unicode.IsSpace(r) || unicode.IsPunct(r):
			if !lastSpace {
				b.WriteByte(' ')
				lastSpace = true
			}
		}
	}
	return strings.TrimSpace(b.String())
}
