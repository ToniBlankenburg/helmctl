package main

import (
	"io"
	"os"
	"strings"
	"testing"
)

func captureOutput(t *testing.T, fn func()) string {
	t.Helper()

	origStdout := os.Stdout
	origStderr := os.Stderr
	defer func() {
		os.Stdout = origStdout
		os.Stderr = origStderr
	}()

	rOut, wOut, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create stdout pipe: %v", err)
	}
	rErr, wErr, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create stderr pipe: %v", err)
	}

	os.Stdout = wOut
	os.Stderr = wErr

	fn()

	_ = wOut.Close()
	_ = wErr.Close()

	outBytes, _ := io.ReadAll(rOut)
	errBytes, _ := io.ReadAll(rErr)
	_ = rOut.Close()
	_ = rErr.Close()

	return string(outBytes) + string(errBytes)
}

func TestMain_HelpFlag_PrintsRootHelp(t *testing.T) {
	origArgs := os.Args
	defer func() { os.Args = origArgs }()

	os.Args = []string{"helmctl", "-h"}

	output := captureOutput(t, main)

	if !strings.Contains(output, "helmctl abstracts selected helm workflows.") {
		t.Fatalf("expected root short description, got: %q", output)
	}
	if !strings.Contains(output, "Usage:") {
		t.Fatalf("expected usage in output, got: %q", output)
	}
}

func TestMain_HelpCommand_PrintsRootHelp(t *testing.T) {
	origArgs := os.Args
	defer func() { os.Args = origArgs }()

	os.Args = []string{"helmctl", "help"}

	output := captureOutput(t, main)

	if !strings.Contains(output, "helmctl abstracts selected helm workflows.") {
		t.Fatalf("expected root help text, got: %q", output)
	}
	if !strings.Contains(output, "Available Commands:") {
		t.Fatalf("expected available commands section, got: %q", output)
	}
}

func TestMain_InstallHelp_PrintsInstallHelp(t *testing.T) {
	origArgs := os.Args
	defer func() { os.Args = origArgs }()

	os.Args = []string{"helmctl", "install", "--help"}

	output := captureOutput(t, main)

	if !strings.Contains(output, "Install a helm chart") {
		t.Fatalf("expected install help text, got: %q", output)
	}
	if !strings.Contains(output, "helmctl install [flags]") {
		t.Fatalf("expected install usage, got: %q", output)
	}
}
