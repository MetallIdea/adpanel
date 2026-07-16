package command

import (
	"fmt"
	"os/exec"
)

func Execute(action string) (string, error) {
	var cmd *exec.Cmd

	switch action {

	case "pwd":
		cmd = exec.Command("pwd")

	case "ls":
		cmd = exec.Command("ls")

	case "whoami":
		cmd = exec.Command("whoami")

	case "date":
		cmd = exec.Command("date")

	case "hostname":
		cmd = exec.Command("hostname")

	case "uptime":
		cmd = exec.Command("uptime")

	default:
		return "", fmt.Errorf("unknown action")
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	return string(output), nil
}
