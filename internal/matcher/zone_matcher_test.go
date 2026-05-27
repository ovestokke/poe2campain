package matcher

import (
	"testing"

	"poe2campain/internal/campaign"
)

func TestFindZoneByGeneratedAreaID(t *testing.T) {
	data := &campaign.CampaignData{Zones: []campaign.Zone{{ID: "g1_13_2", Name: "Ogham Village"}}}
	m := New(data)
	zone, ok := m.FindZone("G1_13_2", data)
	if !ok {
		t.Fatal("expected match")
	}
	if zone.Name != "Ogham Village" {
		t.Fatalf("unexpected zone: %+v", zone)
	}
}

func TestFindZoneByAlias(t *testing.T) {
	data := &campaign.CampaignData{Zones: []campaign.Zone{{ID: "g1_4", Name: "The Grelwood", Aliases: []string{"Grelwood"}}}}
	m := New(data)
	zone, ok := m.FindZone("grelwood", data)
	if !ok {
		t.Fatal("expected match")
	}
	if zone.ID != "g1_4" {
		t.Fatalf("unexpected zone: %+v", zone)
	}
}

func TestDuplicateAliasKeepsFirstZone(t *testing.T) {
	data := &campaign.CampaignData{Zones: []campaign.Zone{
		{ID: "g4_11_1a", Name: "Ngakanu", Aliases: []string{"ngakanu"}},
		{ID: "g4_11_1b", Name: "Ngakanu (Hostile)", Aliases: []string{"ngakanu"}},
	}}
	m := New(data)
	zone, ok := m.FindZone("ngakanu", data)
	if !ok {
		t.Fatal("expected match")
	}
	if zone.ID != "g4_11_1a" {
		t.Fatalf("unexpected zone: %+v", zone)
	}
}

func TestRouteIndexesForZoneID(t *testing.T) {
	data := &campaign.CampaignData{Route: []campaign.RouteEntry{{ZoneID: "g1_1"}, {ZoneID: "g1_2"}, {ZoneID: "g1_1"}}}
	indexes := RouteIndexesForZoneID(data, "G1_1")
	if len(indexes) != 2 || indexes[0] != 0 || indexes[1] != 2 {
		t.Fatalf("unexpected indexes: %+v", indexes)
	}
}
