package campaign

import "testing"

func TestValidateValidData(t *testing.T) {
	data := &CampaignData{
		SchemaVersion: 2,
		Game:          "poe2",
		Zones: []Zone{
			{ID: "g1_1", Name: "The Riverbank", Aliases: []string{"g1_1", "riverbank"}, Act: "Act 1", Kind: "area"},
			{ID: "g1_town", Name: "Clearfell Encampment", Aliases: []string{"g1_town", "clearfell encampment"}, Kind: "town", Flags: []string{"town"}},
		},
		Route: []RouteEntry{
			{
				Order:    10,
				Act:      "Act 1",
				ZoneID:   "g1_1",
				ZoneName: "The Riverbank",
				Steps: []Step{
					{Order: 10, Type: "boss", Text: "Kill The Bloated Miller", Source: "test"},
					{Order: 20, Type: "exit", Text: "Enter Clearfell Encampment", TargetZoneID: strPtr("g1_town"), TargetZone: strPtr("Clearfell Encampment"), Source: "test"},
				},
			},
		},
	}
	report, err := Validate(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.Zones != 2 {
		t.Errorf("expected 2 zones, got %d", report.Zones)
	}
	if report.Steps != 2 {
		t.Errorf("expected 2 steps, got %d", report.Steps)
	}
	if report.RouteEntries != 1 {
		t.Errorf("expected 1 route entry, got %d", report.RouteEntries)
	}
}

func TestValidateRejectsBadSchema(t *testing.T) {
	data := &CampaignData{SchemaVersion: 99, Game: "poe2"}
	_, err := Validate(data)
	if err == nil {
		t.Fatal("expected error for bad schema version")
	}
}

func TestValidateWarnsUnknownTargets(t *testing.T) {
	data := &CampaignData{
		SchemaVersion: 2,
		Game:          "poe2",
		Zones: []Zone{
			{ID: "g1_1", Name: "The Riverbank", Aliases: []string{"g1_1"}, Kind: "area"},
		},
		Route: []RouteEntry{
			{
				Order:    10,
				ZoneID:   "g1_1",
				ZoneName: "The Riverbank",
				Steps: []Step{
					{Order: 10, Type: "exit", Text: "Go somewhere", TargetZoneID: strPtr("unknown_zone"), Source: "test"},
				},
			},
		},
	}
	report, err := Validate(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(report.Warnings) == 0 {
		t.Error("expected warnings for unknown target zone")
	}
}

func strPtr(s string) *string { return &s }