package internal

// ErrorResponse represents the structure of error responses sent to the client
type ErrorResponse struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// Code Exection Request
type CodeExecutionReq struct {
	Code     string `json:"code"`
	Language string `json:"language"`
}

// Code Execution Response
type CodeExecutionResp struct {
	Output     string `json:"output,omitempty"`
	Error      string `json:"error,omitempty"`
	StatusCode int
}

// Healthz Response
type HealthzResp struct {
	TotalPods   int                 `json:"total_pods"`
	StandbyPods map[string][]string `json:"standby_pods"`
	InUsePods   []string            `json:"in_use_pods"`
}
