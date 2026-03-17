package config

import (
	"fmt"
	"os"
	"strings"
)

type Runtime struct {
	APIURL    string
	APIToken  string
	OrgID     string
	AppID     string
	InstallID string
}

func LoadFromEnv() (*Runtime, error) {
	apiURL := strings.TrimSpace(os.Getenv("NUON_API_URL"))
	if apiURL == "" {
		apiURL = "https://api.nuon.co"
	}

	apiToken := strings.TrimSpace(os.Getenv("NUON_API_TOKEN"))
	if apiToken == "" {
		return nil, fmt.Errorf("NUON_API_TOKEN is required")
	}

	orgID := strings.TrimSpace(os.Getenv("NUON_ORG_ID"))
	if orgID == "" {
		return nil, fmt.Errorf("NUON_ORG_ID is required: select an org with `nuon orgs select`")
	}

	return &Runtime{
		APIURL:    apiURL,
		APIToken:  apiToken,
		OrgID:     orgID,
		AppID:     strings.TrimSpace(os.Getenv("NUON_APP_ID")),
		InstallID: strings.TrimSpace(os.Getenv("NUON_INSTALL_ID")),
	}, nil
}
