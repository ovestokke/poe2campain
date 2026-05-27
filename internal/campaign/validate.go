package campaign

import (
	"fmt"
	"strings"
)

type ValidationReport struct {
	Zones         int
	RouteEntries  int
	Steps         int
	FallbackSteps int
	Warnings      []string
}

func Validate(data *CampaignData) (ValidationReport, error) {
	report := ValidationReport{}
	if data == nil {
		return report, fmt.Errorf("campaign data is nil")
	}
	if data.SchemaVersion < 1 || data.SchemaVersion > 2 {
		return report, fmt.Errorf("unsupported schema_version %d", data.SchemaVersion)
	}
	if data.Game != "poe2" {
		return report, fmt.Errorf("unsupported game %q", data.Game)
	}
	if len(data.Zones) == 0 {
		return report, fmt.Errorf("no zones")
	}
	if len(data.Route) == 0 {
		return report, fmt.Errorf("no route entries")
	}

	report.Zones = len(data.Zones)
	report.RouteEntries = len(data.Route)

	zoneIDs := make(map[string]bool, len(data.Zones))
	for _, zone := range data.Zones {
		if strings.TrimSpace(zone.ID) == "" {
			return report, fmt.Errorf("zone with empty id")
		}
		if strings.TrimSpace(zone.Name) == "" {
			return report, fmt.Errorf("zone %q has empty name", zone.ID)
		}
		id := strings.ToLower(zone.ID)
		if zoneIDs[id] {
			report.Warnings = append(report.Warnings, fmt.Sprintf("duplicate zone id %q", zone.ID))
		}
		zoneIDs[id] = true
	}

	for _, entry := range data.Route {
		if strings.TrimSpace(entry.ZoneID) == "" && strings.TrimSpace(entry.ZoneName) == "" {
			return report, fmt.Errorf("route entry %d has no zone_id or zone_name", entry.Order)
		}
		if entry.ZoneID != "" && !zoneIDs[strings.ToLower(entry.ZoneID)] {
			report.Warnings = append(report.Warnings, fmt.Sprintf("route entry %d references unknown zone_id %q", entry.Order, entry.ZoneID))
		}
		if len(entry.Steps) == 0 && len(entry.Notes) == 0 {
			return report, fmt.Errorf("route entry %d has no steps or notes", entry.Order)
		}

		for _, step := range entry.Steps {
			report.Steps++
			if strings.TrimSpace(step.Text) == "" {
				return report, fmt.Errorf("route entry %d step %d has empty text", entry.Order, step.Order)
			}
			if step.TargetZoneID != nil && *step.TargetZoneID != "" && !zoneIDs[strings.ToLower(*step.TargetZoneID)] {
				report.Warnings = append(report.Warnings, fmt.Sprintf("route entry %d step %d references unknown target_zone_id %q", entry.Order, step.Order, *step.TargetZoneID))
			}
		}
		for _, nextID := range entry.NextZoneIDs {
			if nextID != "" && !zoneIDs[strings.ToLower(nextID)] {
				report.Warnings = append(report.Warnings, fmt.Sprintf("route entry %d references unknown next_zone_id %q", entry.Order, nextID))
			}
		}
	}

	return report, nil
}