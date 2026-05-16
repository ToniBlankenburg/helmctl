package cmd

import (
	"bytes"
	"testing"
)

func TestRootCmd(t *testing.T) {
	cmd := rootCmd
	if cmd.Use != "helmctl" {
		t.Errorf("expected Use to be 'helmctl', got %q", cmd.Use)
	}
	if cmd.Short == "" {
		t.Error("expected Short description to be set")
	}
	if cmd.Long == "" {
		t.Error("expected Long description to be set")
	}
}

func TestRootCmdHelp(t *testing.T) {
	cmd := rootCmd
	output := new(bytes.Buffer)
	cmd.SetOut(output)
	cmd.SetArgs([]string{"--help"})
	
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	
	helpText := output.String()
	if helpText == "" {
		t.Error("expected help output, got empty string")
	}
	if bytes.Contains(output.Bytes(), []byte("DEBUG: help requested")) {
		t.Log("help hook executed as expected")
	}
}

func TestRootCmdAddCommand(t *testing.T) {
	cmd := rootCmd
	if len(cmd.Commands()) == 0 {
		t.Error("expected rootCmd to have commands, got none")
	}
}
