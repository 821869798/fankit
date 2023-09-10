//go:build darwin

package fanopen

import (
	"os/exec"
)

func open(input string) *exec.Cmd {
	return exec.Command("fanopen", input)
}

func openWith(input string, appName string) *exec.Cmd {
	return exec.Command("fanopen", "-a", appName, input)
}
