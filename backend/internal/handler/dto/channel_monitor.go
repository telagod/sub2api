package dto

// ChannelMonitorExtraModelStatus captures the most recent health-check result
// for a model on a given channel. Shared by both admin and user list responses.
type ChannelMonitorExtraModelStatus struct {
	Model     string `json:"model"`
	Status    string `json:"status"`
	LatencyMs *int   `json:"latency_ms"`
}
