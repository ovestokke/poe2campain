package runtime

import (
	"testing"

	"poe2campain/internal/campaign"
)

func TestSessionAreaDetectionAndStepNavigation(t *testing.T) {
	data := testData()
	s := New(data)

	state, ok := s.OnAreaDetected("G1_2")
	if !ok {
		t.Fatal("expected zone match")
	}
	if state.Zone == nil || state.Zone.ID != "g1_2" {
		t.Fatalf("unexpected zone: %+v", state.Zone)
	}
	if state.Route == nil || state.Route.Order != 20 {
		t.Fatalf("unexpected route: %+v", state.Route)
	}
	if state.Step == nil || state.Step.Text != "first" {
		t.Fatalf("unexpected step: %+v", state.Step)
	}

	state, ok = s.NextStep()
	if !ok || state.Step == nil || state.Step.Text != "second" {
		t.Fatalf("expected second step, got ok=%v state=%+v", ok, state)
	}
	state, ok = s.PrevStep()
	if !ok || state.Step == nil || state.Step.Text != "first" {
		t.Fatalf("expected first step, got ok=%v state=%+v", ok, state)
	}
}

func TestSessionRepeatedAreaDoesNotResetStep(t *testing.T) {
	data := testData()
	s := New(data)
	_, _ = s.OnAreaDetected("g1_2")
	_, _ = s.NextStep()

	state, ok := s.OnAreaDetected("g1_2")
	if !ok {
		t.Fatal("expected repeated zone match")
	}
	if state.Step == nil || state.Step.Text != "second" {
		t.Fatalf("expected repeated area to keep current step, got %+v", state.Step)
	}
}

func TestSessionRepeatedTownUsesRouteOrder(t *testing.T) {
	data := testData()
	s := New(data)
	state, ok := s.SelectRouteOrder(20)
	if !ok || state.Route == nil || state.Route.Order != 20 {
		t.Fatalf("expected selected route 20, got %+v", state.Route)
	}

	state, ok = s.OnAreaDetected("g1_town")
	if !ok {
		t.Fatal("expected town match")
	}
	if state.Route == nil || state.Route.Order != 30 {
		t.Fatalf("expected next town route 30, got %+v", state.Route)
	}
}

func TestSessionAutoDetectionDoesNotMoveBackward(t *testing.T) {
	data := testData()
	s := New(data)
	state, ok := s.SelectRouteOrder(40)
	if !ok || state.Route == nil || state.Route.Order != 40 {
		t.Fatalf("expected selected route 40, got %+v", state.Route)
	}

	state, ok = s.OnAreaDetected("g1_2")
	if !ok {
		t.Fatal("expected old zone match")
	}
	if state.Route == nil || state.Route.Order != 40 {
		t.Fatalf("expected guide to stay at route 40, got %+v", state.Route)
	}
}

func TestSessionManualJumpCanMoveBackward(t *testing.T) {
	data := testData()
	s := New(data)
	_, _ = s.SelectRouteOrder(40)

	state, ok := s.JumpToZone("g1_2")
	if !ok {
		t.Fatal("expected manual old zone match")
	}
	if state.Route == nil || state.Route.Order != 20 {
		t.Fatalf("expected manual jump to route 20, got %+v", state.Route)
	}
}

func TestSessionSnapshotRestore(t *testing.T) {
	data := testData()
	s := New(data)
	_, _ = s.SelectRouteOrder(20)
	_, _ = s.NextStep()

	restored := New(data)
	state := restored.Restore(s.Snapshot())
	if state.Route == nil || state.Route.Order != 20 {
		t.Fatalf("unexpected restored route: %+v", state.Route)
	}
	if state.Step == nil || state.Step.Text != "second" {
		t.Fatalf("unexpected restored step: %+v", state.Step)
	}
}

func TestSessionToggleCurrentStepDoneAndRestore(t *testing.T) {
	data := testData()
	s := New(data)
	state, ok := s.SelectRouteOrder(20)
	if !ok || state.Route == nil {
		t.Fatalf("expected selected route, got ok=%v state=%+v", ok, state)
	}

	state, ok = s.ToggleCurrentStepDone()
	if !ok {
		t.Fatal("expected toggle to succeed")
	}
	if !state.CompletedStepIndexes[0] {
		t.Fatalf("expected first step to be completed, got %+v", state.CompletedStepIndexes)
	}

	_, _ = s.NextStep()
	state, ok = s.ToggleCurrentStepDone()
	if !ok {
		t.Fatal("expected second toggle to succeed")
	}
	if !state.CompletedStepIndexes[1] {
		t.Fatalf("expected second step to be completed, got %+v", state.CompletedStepIndexes)
	}

	restored := New(data)
	state = restored.Restore(s.Snapshot())
	if !state.CompletedStepIndexes[0] || !state.CompletedStepIndexes[1] {
		t.Fatalf("expected completed steps after restore, got %+v", state.CompletedStepIndexes)
	}

	state, ok = restored.ToggleCurrentStepDone()
	if !ok {
		t.Fatal("expected untoggle to succeed")
	}
	if state.CompletedStepIndexes[1] {
		t.Fatalf("expected current step to be cleared, got %+v", state.CompletedStepIndexes)
	}
	if !state.CompletedStepIndexes[0] {
		t.Fatalf("expected first step to remain completed, got %+v", state.CompletedStepIndexes)
	}
}

func TestSessionToggleCurrentStepDoneAdvance(t *testing.T) {
	data := testData()
	s := New(data)
	state, ok := s.SelectRouteOrder(20)
	if !ok || state.Step == nil || state.Step.Text != "first" {
		t.Fatalf("expected first step, got ok=%v state=%+v", ok, state)
	}

	state, ok = s.ToggleCurrentStepDoneAdvance()
	if !ok {
		t.Fatal("expected mark-and-advance to succeed")
	}
	if !state.CompletedStepIndexes[0] {
		t.Fatalf("expected first step done, got %+v", state.CompletedStepIndexes)
	}
	if state.Step == nil || state.Step.Text != "second" {
		t.Fatalf("expected auto-advance to second step, got %+v", state.Step)
	}

	state, ok = s.ToggleCurrentStepDoneAdvance()
	if !ok {
		t.Fatal("expected second mark-and-advance to succeed")
	}
	if !state.CompletedStepIndexes[1] {
		t.Fatalf("expected second step done, got %+v", state.CompletedStepIndexes)
	}
	if state.Step == nil || state.Step.Text != "second" {
		t.Fatalf("expected to stay on last step, got %+v", state.Step)
	}

	state, ok = s.ToggleCurrentStepDoneAdvance()
	if !ok {
		t.Fatal("expected undo on done step to succeed")
	}
	if state.CompletedStepIndexes[1] {
		t.Fatalf("expected undo to clear current step, got %+v", state.CompletedStepIndexes)
	}
	if state.Step == nil || state.Step.Text != "second" {
		t.Fatalf("expected undo to stay on same step, got %+v", state.Step)
	}
}

func testData() *campaign.CampaignData {
	return &campaign.CampaignData{
		Zones: []campaign.Zone{
			{ID: "g1_town", Name: "Town", Act: "Act 1", Kind: "town"},
			{ID: "g1_2", Name: "Clearfell", Act: "Act 1", Kind: "area"},
			{ID: "g2_1", Name: "Act 2 Start", Act: "Act 2", Kind: "area"},
		},
		Route: []campaign.RouteEntry{
			{Order: 10, Act: "Act 1", ZoneID: "g1_town", ZoneName: "Town", DisplayName: "Town", Steps: []campaign.Step{{Order: 10, Text: "leave town"}}},
			{Order: 20, Act: "Act 1", ZoneID: "g1_2", ZoneName: "Clearfell", DisplayName: "Clearfell", Steps: []campaign.Step{{Order: 10, Text: "first"}, {Order: 20, Text: "second"}}},
			{Order: 30, Act: "Act 1", ZoneID: "g1_town", ZoneName: "Town", DisplayName: "Town", Steps: []campaign.Step{{Order: 10, Text: "back in town"}}},
			{Order: 40, Act: "Act 2", ZoneID: "g2_1", ZoneName: "Act 2 Start", DisplayName: "Act 2 Start", Steps: []campaign.Step{{Order: 10, Text: "later route"}}},
		},
	}
}
