package campaign

// CampaignData is the normalized runtime data format.
type CampaignData struct {
	SchemaVersion int          `json:"schema_version"`
	Game          string       `json:"game"`
	GeneratedAt   string       `json:"generated_at"`
	Sources       []Source     `json:"sources"`
	Zones         []Zone       `json:"zones"`
	Route         []RouteEntry `json:"route"`
}

type Source struct {
	Name    string   `json:"name"`
	Repo    string   `json:"repo"`
	Commit  string   `json:"commit"`
	License string   `json:"license"`
	Files   []string `json:"files"`
}

type Zone struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Aliases []string `json:"aliases"`
	Act     string   `json:"act"`
	Kind    string   `json:"kind"`
	Flags   []string `json:"flags"`

	// Level info — used for display in the TUI header
	LevelRecommendation string   `json:"level_recommendation"`
	LevelMin            *float64 `json:"level_min,omitempty"`
	LevelMax            *float64 `json:"level_max,omitempty"`
}

type RouteEntry struct {
	Order       int         `json:"order"`
	Act         string      `json:"act"`
	ZoneID      string      `json:"zone_id"`
	ZoneName    string      `json:"zone_name"`
	DisplayName string      `json:"display_name"`
	LevelRange  string      `json:"level_range"`
	LevelMin    *float64    `json:"level_min,omitempty"`
	LevelMax    *float64    `json:"level_max,omitempty"`
	Flags       []string    `json:"flags"`
	Steps       []Step      `json:"steps"`
	Notes       []string    `json:"notes"`
	NextZoneIDs []string    `json:"next_zone_ids"`
	NextZones   []string    `json:"next_zones"`
}

type Step struct {
	Order        int      `json:"order"`
	Type         string   `json:"type"`
	Text         string   `json:"text"`
	Optional     bool     `json:"optional"`
	TargetZoneID *string  `json:"target_zone_id"`
	TargetZone   *string  `json:"target_zone"`
	Tags         []string `json:"tags"`
	Rewards      []string `json:"rewards"`
	Source       string   `json:"source"`
}