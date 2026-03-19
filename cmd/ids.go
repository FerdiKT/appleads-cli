package cmd

import (
	"fmt"
	"strconv"
)

func parsePositiveInt64(name, value string) (int64, error) {
	n, err := strconv.ParseInt(value, 10, 64)
	if err != nil || n <= 0 {
		return 0, fmt.Errorf("invalid %s %q (must be a positive integer)", name, value)
	}
	return n, nil
}
