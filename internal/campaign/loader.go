package campaign

import (
	"encoding/json"
	"fmt"
	"os"
)

func Load(path string) (*CampaignData, ValidationReport, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, ValidationReport{}, err
	}

	var data CampaignData
	if err := json.Unmarshal(b, &data); err != nil {
		return nil, ValidationReport{}, fmt.Errorf("decode campaign data: %w", err)
	}

	report, err := Validate(&data)
	if err != nil {
		return nil, report, err
	}
	return &data, report, nil
}
