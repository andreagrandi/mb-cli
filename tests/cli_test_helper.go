package tests

import (
	"bytes"
	"os"
	"os/exec"
	"testing"
)

func runMBCLI(t *testing.T, env map[string]string, args ...string) (string, string, error) {
	t.Helper()

	cmdArgs := append([]string{"run", "./cmd/mb-cli"}, args...)
	cmd := exec.Command("go", cmdArgs...)
	cmd.Dir = ".."

	cmd.Env = os.Environ()
	for key, value := range env {
		cmd.Env = append(cmd.Env, key+"="+value)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}
