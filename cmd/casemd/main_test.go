package main

import (
	"bytes"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestMainVersionOutput(t *testing.T) {
	tempDir := t.TempDir()
	binaryPath := filepath.Join(tempDir, "casemd")

	buildCmd := exec.Command("go", "build", "-ldflags=-X main.version=v1.2.3", "-o", binaryPath, "./cmd/casemd")
	buildCmd.Dir = filepath.Join("..", "..")
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("go build failed: %v\nOutput: %s", err, string(out))
	}

	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "--version flag",
			args: []string{"--version"},
			want: "v1.2.3\n",
		},
		{
			name: "version command",
			args: []string{"version"},
			want: "v1.2.3\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(binaryPath, tt.args...)
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			if err := cmd.Run(); err != nil {
				t.Fatalf("command failed: %v, stderr: %s", err, stderr.String())
			}

			if got := stdout.String(); got != tt.want {
				t.Errorf("stdout = %q, want %q", got, tt.want)
			}
		})
	}
}
