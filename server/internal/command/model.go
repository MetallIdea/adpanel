package command

type ExecuteRequest struct {
	Action string `json:"action"`
}

type ExecuteResponse struct {
	Output string `json:"output"`
}
