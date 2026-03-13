package cpwd

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunSuccess(t *testing.T) {
	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	t.Cleanup(func() {
		if chdirErr := os.Chdir(originalWD); chdirErr != nil {
			t.Fatalf("restore cwd: %v", chdirErr)
		}
	})

	tempDir := t.TempDir()
	realDir := filepath.Join(tempDir, "real")
	linkDir := filepath.Join(tempDir, "link")

	if err := os.Mkdir(realDir, 0o755); err != nil {
		t.Fatalf("mkdir real: %v", err)
	}
	if err := os.Symlink(realDir, linkDir); err != nil {
		t.Fatalf("symlink: %v", err)
	}
	if err := os.Chdir(linkDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	originalCopy := copyToClipboard
	t.Cleanup(func() {
		copyToClipboard = originalCopy
	})
	var copied string
	copyToClipboard = func(value string) error {
		copied = value
		return nil
	}

	exitCode := Run(stdout, stderr)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected no stderr, got %q", stderr.String())
	}

	expectedPath, err := filepath.EvalSymlinks(realDir)
	if err != nil {
		t.Fatalf("resolve expected path: %v", err)
	}
	expectedPath = filepath.Clean(expectedPath)
	if copied != expectedPath {
		t.Fatalf("expected copied path %q, got %q", expectedPath, copied)
	}

	expectedOutput := "Copied to clipboard\n" + expectedPath + "\n"
	if stdout.String() != expectedOutput {
		t.Fatalf("expected stdout %q, got %q", expectedOutput, stdout.String())
	}
}

func TestRunClipboardFailure(t *testing.T) {
	originalResolve := resolvePhysicalPath
	originalCopy := copyToClipboard
	t.Cleanup(func() {
		resolvePhysicalPath = originalResolve
		copyToClipboard = originalCopy
	})

	resolvePhysicalPath = func() (string, error) {
		return "/tmp/project", nil
	}
	copyToClipboard = func(string) error {
		return errors.New("pbcopy unavailable")
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	exitCode := Run(stdout, stderr)
	if exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitCode)
	}
	if stdout.String() != "/tmp/project\n" {
		t.Fatalf("unexpected stdout: %q", stdout.String())
	}
	if !strings.Contains(stderr.String(), "failed to copy to clipboard") {
		t.Fatalf("unexpected stderr: %q", stderr.String())
	}
}

func TestRunResolveFailure(t *testing.T) {
	originalResolve := resolvePhysicalPath
	t.Cleanup(func() {
		resolvePhysicalPath = originalResolve
	})

	resolvePhysicalPath = func() (string, error) {
		return "", errors.New("no working directory")
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	exitCode := Run(stdout, stderr)
	if exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitCode)
	}
	if stdout.Len() != 0 {
		t.Fatalf("expected no stdout, got %q", stdout.String())
	}
	if !strings.Contains(stderr.String(), "no working directory") {
		t.Fatalf("unexpected stderr: %q", stderr.String())
	}
}

func TestCopyToClipboardUsesPbcopy(t *testing.T) {
	originalCommandFactory := commandFactory
	t.Cleanup(func() {
		commandFactory = originalCommandFactory
	})

	scriptPath := filepath.Join(t.TempDir(), "capture-stdin.sh")
	outputPath := filepath.Join(t.TempDir(), "stdin.txt")
	script := "#!/bin/sh\ncat > \"$1\"\n"
	if err := os.WriteFile(scriptPath, []byte(script), 0o755); err != nil {
		t.Fatalf("write script: %v", err)
	}

	var name string
	var args []string
	commandFactory = func(cmdName string, cmdArgs ...string) *exec.Cmd {
		name = cmdName
		args = append([]string(nil), cmdArgs...)
		return exec.Command(scriptPath, outputPath)
	}

	if err := CopyToClipboard("/tmp/project"); err != nil {
		t.Fatalf("copy to clipboard: %v", err)
	}
	if name != "pbcopy" {
		t.Fatalf("unexpected command name: %q", name)
	}
	if len(args) != 0 {
		t.Fatalf("unexpected command args: %q", strings.Join(args, " "))
	}
	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read stdin capture: %v", err)
	}
	if string(data) != "/tmp/project" {
		t.Fatalf("unexpected stdin: %q", string(data))
	}
}
