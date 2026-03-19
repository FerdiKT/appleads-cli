package cmd

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

func init() {
	accountCmd.AddCommand(accountMeCmd)
	accountCmd.AddCommand(accountACLsCmd)
}

var accountMeCmd = &cobra.Command{
	Use:   "me",
	Short: "Get current API user details",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, _, err := authedClient(context.Background(), 0, false)
		if err != nil {
			return err
		}
		paths := []string{"/me", "/me/details"}
		var lastErr error
		for _, p := range paths {
			var out any
			err := client.DoJSON(context.Background(), http.MethodGet, p, nil, nil, &out)
			if err == nil {
				return printJSON(out)
			}
			lastErr = err
			time.Sleep(50 * time.Millisecond)
		}
		return fmt.Errorf("me lookup failed for paths [%s]: %w", strings.Join(paths, ", "), lastErr)
	},
}

var accountACLsCmd = &cobra.Command{
	Use:   "acls",
	Short: "Get accessible organizations (ACLs)",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, _, err := authedClient(context.Background(), 0, false)
		if err != nil {
			return err
		}
		resp, err := client.ListUserACLs(context.Background())
		if err != nil {
			return err
		}
		return printJSON(resp)
	},
}
