package synthetics

import (
	"path"
)

const (
	listMonitorsLimit = 100
)

// Monitor represents a New Relic Synthetics monitor.
type Monitor struct {
	ID           string            `json:"id,omitempty"`
	Name         string            `json:"name"`
	Type         MonitorType       `json:"type"`
	Frequency    uint              `json:"frequency"`
	URI          string            `json:"uri"`
	Locations    []string          `json:"locations"`
	Status       MonitorStatusType `json:"status"`
	SLAThreshold float64           `json:"slaThreshold"`
	UserID       uint              `json:"userId,omitempty"`
	APIVersion   string            `json:"apiVersion,omitempty"`
	ModifiedAt   *Time             `json:"modifiedAt,omitempty"`
	CreatedAt    *Time             `json:"createdAt,omitempty"`
	Options      MonitorOptions    `json:"options,omitempty"`
}

// MonitorScriptLocation represents a New Relic Synthetics monitor script location.
type MonitorScriptLocation struct {
	Name string `json:"name"`
	HMAC string `json:"hmac"`
}

// MonitorScript represents a New Relic Synthetics monitor script.
type MonitorScript struct {
	Text      string                  `json:"scriptText"`
	Locations []MonitorScriptLocation `json:"scriptLocations"`
}

// MonitorType represents a Synthetics monitor type.
type MonitorType string

// MonitorStatusType represents a Synthetics monitor status type.
type MonitorStatusType string

// MonitorOptions represents the options for a New Relic Synthetics monitor.
type MonitorOptions struct {
	ValidationString       string `json:"validationString,omitempty"`
	VerifySSL              bool   `json:"verifySSL,omitempty"`
	BypassHEADRequest      bool   `json:"bypassHEADRequest,omitempty"`
	TreatRedirectAsFailure bool   `json:"treatRedirectAsFailure,omitempty"`
}

var (
	// MonitorTypes specifies the possible types for a Synthetics monitor.
	MonitorTypes = struct {
		Ping            MonitorType
		Browser         MonitorType
		ScriptedBrowser MonitorType
		APITest         MonitorType
	}{
		Ping:            "SIMPLE",
		Browser:         "BROWSER",
		ScriptedBrowser: "SCRIPT_BROWSER",
		APITest:         "SCRIPT_API",
	}

	// MonitorStatus specifies the possible Synthetics monitor status types.
	MonitorStatus = struct {
		Enabled  MonitorStatusType
		Muted    MonitorStatusType
		Disabled MonitorStatusType
	}{
		Enabled:  "ENABLED",
		Muted:    "MUTED",
		Disabled: "DISABLED",
	}
)

// ListMonitors is used to retrieve New Relic Synthetics monitors.
func (s *Synthetics) ListMonitors(queryParams *ListMonitorsParams) (ListMonitorsResponse, error) {
	resp := ListMonitorsResponse{}

	if queryParams.Limit == 0 {
		queryParams.Limit = listMonitorsLimit
	}

	_, err := s.client.Get(s.config.Region().SyntheticsURL("/v4/monitors"), queryParams, &resp)

	if err != nil {
		return resp, err
	}

	return resp, nil
}

// GetMonitor is used to retrieve a specific New Relic Synthetics monitor.
func (s *Synthetics) GetMonitor(monitorID string) (*Monitor, error) {
	resp := Monitor{}

	_, err := s.client.Get(s.config.Region().SyntheticsURL("/v4/monitors", monitorID), nil, &resp)

	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// CreateMonitor is used to create a New Relic Synthetics monitor.
func (s *Synthetics) CreateMonitor(monitor Monitor) (*Monitor, error) {
	resp, err := s.client.Post(s.config.Region().SyntheticsURL("/v4/monitors"), nil, &monitor, nil)

	if err != nil {
		return nil, err
	}

	l := resp.Header.Get("location")
	monitorID := path.Base(l)

	monitor.ID = monitorID

	return &monitor, nil
}

// UpdateMonitor is used to update a New Relic Synthetics monitor.
func (s *Synthetics) UpdateMonitor(monitor Monitor) (*Monitor, error) {
	_, err := s.client.Put(s.config.Region().SyntheticsURL("/v4/monitors", monitor.ID), nil, &monitor, nil)

	if err != nil {
		return nil, err
	}

	return &monitor, nil
}

// DeleteMonitor is used to delete a New Relic Synthetics monitor.
func (s *Synthetics) DeleteMonitor(monitorID string) error {
	_, err := s.client.Delete(s.config.Region().SyntheticsURL("/v4/monitors", monitorID), nil, nil)

	if err != nil {
		return err
	}

	return nil
}

type ListMonitorsResponse struct {
	Monitors []*Monitor `json:"monitors,omitempty"`
	Count int `json:"count,omitempty"`
}

type ListMonitorsParams struct {
	Limit int `url:"limit,omitempty"`
	Offset int `url:"offset,omitempty"`
}