package runtime

import (
	"sort"
	"strconv"
	"strings"

	"poe2campain/internal/campaign"
	"poe2campain/internal/matcher"
)

// Session is deliberately dumb: route[] order is the guide.
// Client.txt detection may move the cursor forward to the next matching route
// entry, but never invents quest completion or rewinds automatically.
type Session struct {
	data                 *campaign.CampaignData
	matcher              matcher.Matcher
	zoneIndexByID        map[string]int
	currentRouteIndex    int
	currentStepIndex     int
	lastDetectedArea     string
	completedStepByRoute map[int]map[int]bool
}

type State struct {
	Zone                 *campaign.Zone
	Route                *campaign.RouteEntry
	Step                 *campaign.Step
	RouteIndex           int
	StepIndex            int
	RouteCandidates      []int
	LastDetectedArea     string
	CompletedStepIndexes map[int]bool
}

type Snapshot struct {
	CurrentRouteOrder int              `json:"current_route_order,omitempty"`
	CurrentStepIndex  int              `json:"current_step_index"`
	CompletedSteps    map[string][]int `json:"completed_steps,omitempty"`
}

func New(data *campaign.CampaignData) *Session {
	s := &Session{
		data:                 data,
		matcher:              matcher.New(data),
		zoneIndexByID:        map[string]int{},
		currentRouteIndex:    -1,
		currentStepIndex:     -1,
		completedStepByRoute: map[int]map[int]bool{},
	}
	for i, zone := range data.Zones {
		s.zoneIndexByID[strings.ToLower(zone.ID)] = i
	}
	return s
}

func (s *Session) Data() *campaign.CampaignData { return s.data }

func (s *Session) Snapshot() Snapshot {
	snapshot := Snapshot{CurrentRouteOrder: s.routeOrder(s.currentRouteIndex), CurrentStepIndex: s.currentStepIndex}
	if len(s.completedStepByRoute) == 0 {
		return snapshot
	}
	snapshot.CompletedSteps = map[string][]int{}
	for routeOrder, completed := range s.completedStepByRoute {
		if routeOrder <= 0 || len(completed) == 0 {
			continue
		}
		stepIndexes := make([]int, 0, len(completed))
		for stepIndex, done := range completed {
			if done {
				stepIndexes = append(stepIndexes, stepIndex)
			}
		}
		if len(stepIndexes) == 0 {
			continue
		}
		sort.Ints(stepIndexes)
		snapshot.CompletedSteps[strconv.Itoa(routeOrder)] = stepIndexes
	}
	if len(snapshot.CompletedSteps) == 0 {
		snapshot.CompletedSteps = nil
	}
	return snapshot
}

func (s *Session) Restore(snapshot Snapshot) State {
	s.completedStepByRoute = map[int]map[int]bool{}
	for routeOrderText, stepIndexes := range snapshot.CompletedSteps {
		routeOrder, err := strconv.Atoi(routeOrderText)
		if err != nil || routeOrder <= 0 {
			continue
		}
		routeIndex := s.routeIndexByOrder(routeOrder)
		if routeIndex < 0 {
			continue
		}
		stepsInRoute := len(s.data.Route[routeIndex].Steps)
		for _, stepIndex := range stepIndexes {
			if stepIndex < 0 || stepIndex >= stepsInRoute {
				continue
			}
			if s.completedStepByRoute[routeOrder] == nil {
				s.completedStepByRoute[routeOrder] = map[int]bool{}
			}
			s.completedStepByRoute[routeOrder][stepIndex] = true
		}
	}

	s.currentRouteIndex = s.routeIndexByOrder(snapshot.CurrentRouteOrder)
	if s.currentRouteIndex >= 0 {
		s.currentStepIndex = clampStep(snapshot.CurrentStepIndex, len(s.data.Route[s.currentRouteIndex].Steps))
	} else {
		s.currentStepIndex = -1
	}
	return s.State()
}

func (s *Session) OnAreaDetected(areaID string) (State, bool) { return s.SetZone(areaID) }

// SetZone follows the ordered route. It selects the first matching route entry
// at or after the current route. If the detected area only exists behind the
// current route, it is ignored.
func (s *Session) SetZone(input string) (State, bool) {
	zone, ok := s.matcher.FindZone(input, s.data)
	if !ok {
		s.lastDetectedArea = input
		return s.State(), false
	}
	s.lastDetectedArea = zone.ID
	candidates := matcher.RouteIndexesForZoneID(s.data, zone.ID)
	if len(candidates) == 0 {
		return s.State(), true
	}

	selected := -1
	for _, candidate := range candidates {
		if candidate == s.currentRouteIndex {
			return s.State(), true
		}
		if candidate > s.currentRouteIndex {
			selected = candidate
			break
		}
	}
	if selected == -1 {
		return s.State(), true
	}
	return s.SelectRouteIndex(selected)
}

// JumpToZone is only a manual/debug escape hatch. It does not affect automatic
// detection rules.
func (s *Session) JumpToZone(input string) (State, bool) {
	zone, ok := s.matcher.FindZone(input, s.data)
	if !ok {
		return s.State(), false
	}
	candidates := matcher.RouteIndexesForZoneID(s.data, zone.ID)
	if len(candidates) == 0 {
		return s.State(), false
	}
	return s.SelectRouteIndex(candidates[0])
}

func (s *Session) State() State {
	state := State{RouteIndex: s.currentRouteIndex, StepIndex: s.currentStepIndex, LastDetectedArea: s.lastDetectedArea}
	if s.currentRouteIndex >= 0 && s.currentRouteIndex < len(s.data.Route) {
		state.Route = &s.data.Route[s.currentRouteIndex]
		state.Zone = s.zoneByID(state.Route.ZoneID)
		state.RouteCandidates = matcher.RouteIndexesForZoneID(s.data, state.Route.ZoneID)
		state.CompletedStepIndexes = copyCompletedSteps(s.completedStepByRoute[state.Route.Order])
		if s.currentStepIndex >= 0 && s.currentStepIndex < len(state.Route.Steps) {
			state.Step = &state.Route.Steps[s.currentStepIndex]
		}
	}
	return state
}

func (s *Session) ToggleCurrentStepDone() (State, bool) {
	route := s.currentRoute()
	if route == nil || s.currentStepIndex < 0 || s.currentStepIndex >= len(route.Steps) {
		return s.State(), false
	}
	if s.completedStepByRoute[route.Order] == nil {
		s.completedStepByRoute[route.Order] = map[int]bool{}
	}
	if s.completedStepByRoute[route.Order][s.currentStepIndex] {
		delete(s.completedStepByRoute[route.Order], s.currentStepIndex)
		if len(s.completedStepByRoute[route.Order]) == 0 {
			delete(s.completedStepByRoute, route.Order)
		}
	} else {
		s.completedStepByRoute[route.Order][s.currentStepIndex] = true
	}
	return s.State(), true
}

func (s *Session) ToggleCurrentStepDoneAdvance() (State, bool) {
	route := s.currentRoute()
	if route == nil || s.currentStepIndex < 0 || s.currentStepIndex >= len(route.Steps) {
		return s.State(), false
	}

	wasDone := s.completedStepByRoute[route.Order] != nil && s.completedStepByRoute[route.Order][s.currentStepIndex]
	state, ok := s.ToggleCurrentStepDone()
	if !ok {
		return state, false
	}
	if wasDone {
		return state, true
	}
	if _, ok := s.NextStep(); ok {
		return s.State(), true
	}
	return state, true
}

func (s *Session) NextStep() (State, bool) {
	route := s.currentRoute()
	if route == nil || s.currentStepIndex+1 >= len(route.Steps) {
		return s.State(), false
	}
	s.currentStepIndex++
	return s.State(), true
}

func (s *Session) PrevStep() (State, bool) {
	if s.currentStepIndex <= 0 {
		return s.State(), false
	}
	s.currentStepIndex--
	return s.State(), true
}

func (s *Session) NextRoute() (State, bool) {
	if s.currentRouteIndex+1 >= len(s.data.Route) {
		return s.State(), false
	}
	return s.SelectRouteIndex(s.currentRouteIndex + 1)
}

func (s *Session) PrevRoute() (State, bool) {
	if s.currentRouteIndex <= 0 {
		return s.State(), false
	}
	return s.SelectRouteIndex(s.currentRouteIndex - 1)
}

func (s *Session) Start() (State, bool) {
	if len(s.data.Route) == 0 {
		return s.State(), false
	}
	return s.SelectRouteIndex(0)
}

func (s *Session) End() (State, bool) {
	if len(s.data.Route) == 0 {
		return s.State(), false
	}
	return s.SelectRouteIndex(len(s.data.Route) - 1)
}

func (s *Session) SelectRouteOrder(order int) (State, bool) {
	idx := s.routeIndexByOrder(order)
	if idx < 0 {
		return s.State(), false
	}
	return s.SelectRouteIndex(idx)
}

func (s *Session) SelectRouteIndex(index int) (State, bool) {
	if index < 0 || index >= len(s.data.Route) {
		return s.State(), false
	}
	s.currentRouteIndex = index
	if len(s.data.Route[index].Steps) > 0 {
		s.currentStepIndex = 0
	} else {
		s.currentStepIndex = -1
	}
	return s.State(), true
}

func (s *Session) currentRoute() *campaign.RouteEntry {
	if s.currentRouteIndex < 0 || s.currentRouteIndex >= len(s.data.Route) {
		return nil
	}
	return &s.data.Route[s.currentRouteIndex]
}

func (s *Session) zoneByID(id string) *campaign.Zone {
	idx, ok := s.zoneIndexByID[strings.ToLower(id)]
	if !ok {
		return nil
	}
	return &s.data.Zones[idx]
}

func (s *Session) routeOrder(index int) int {
	if index < 0 || index >= len(s.data.Route) {
		return 0
	}
	return s.data.Route[index].Order
}

func (s *Session) routeIndexByOrder(order int) int {
	if order == 0 {
		return -1
	}
	for i, entry := range s.data.Route {
		if entry.Order == order {
			return i
		}
	}
	return -1
}

func copyCompletedSteps(src map[int]bool) map[int]bool {
	if len(src) == 0 {
		return nil
	}
	out := make(map[int]bool, len(src))
	for stepIndex, done := range src {
		if done {
			out[stepIndex] = true
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func clampStep(step, steps int) int {
	if steps <= 0 {
		return -1
	}
	if step < 0 {
		return 0
	}
	if step >= steps {
		return steps - 1
	}
	return step
}
