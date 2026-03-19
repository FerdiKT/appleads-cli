package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

func confirmJSONPayload(title string, payload any, skip bool) error {
	if skip {
		return nil
	}
	if !stdinIsTTY() {
		return fmt.Errorf("%s requires interactive confirmation, but stdin is not a TTY; re-run with --yes", title)
	}

	fmt.Printf("%s payload:\n", title)
	if err := printJSON(payload); err != nil {
		return err
	}
	fmt.Print("Apply changes? [y/N]: ")

	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		if errors.Is(err, io.EOF) {
			return fmt.Errorf("%s confirmation not provided; re-run with --yes for non-interactive usage", title)
		}
		return fmt.Errorf("read confirmation: %w", err)
	}
	answer := strings.ToLower(strings.TrimSpace(line))
	if answer != "y" && answer != "yes" {
		return fmt.Errorf("operation cancelled")
	}
	return nil
}
