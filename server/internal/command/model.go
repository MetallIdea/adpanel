package command

type ExecuteRequest struct {
	Action  string   `json:"action"`
	Command string   `json:"command"`
	Args    []string `json:"args"`
}

type ExecuteResponse struct {
	Output   string `json:"output"`
	ExitCode *int   `json:"exit_code,omitempty"`
}
