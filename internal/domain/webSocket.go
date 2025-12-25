package domain

// ... AnalysisStatusMessage 保持不变 ...
type AnalysisStatusMessage struct {
	Type      string `json:"type"`
	TaskID    string `json:"taskId"`
	Status    string `json:"status"`
	Message   string `json:"message,omitempty"`
	ResultURL string `json:"resultUrl,omitempty"`
	Error     string `json:"error,omitempty"`
	Timestamp int64  `json:"timestamp"`
}
