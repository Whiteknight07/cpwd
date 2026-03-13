package cpwd

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var (
	resolvePhysicalPath = ResolvePhysicalPath
	copyToClipboard     = CopyToClipboard
	commandFactory      = exec.Command
)

func Run(stdout, stderr io.Writer) int {
	path, err := resolvePhysicalPath()
	if err != nil {
		fmt.Fprintf(stderr, "cpwd: %v\n", err)
		return 1
	}

	if err := copyToClipboard(path); err != nil {
		fmt.Fprintln(stdout, path)
		fmt.Fprintf(stderr, "cpwd: failed to copy to clipboard: %v\n", err)
		return 1
	}

	fmt.Fprintln(stdout, "Copied to clipboard")
	fmt.Fprintln(stdout, path)
	return 0
}

func ResolvePhysicalPath() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get working directory: %w", err)
	}

	physical, err := filepath.EvalSymlinks(wd)
	if err != nil {
		return "", fmt.Errorf("resolve physical path: %w", err)
	}

	return filepath.Clean(physical), nil
}

func CopyToClipboard(value string) error {
	cmd := commandFactory("pbcopy")
	cmd.Stdin = strings.NewReader(value)

	if output, err := cmd.CombinedOutput(); err != nil {
		if len(output) == 0 {
			return err
		}

		return fmt.Errorf("%w: %s", err, string(output))
	}

	return nil
}
