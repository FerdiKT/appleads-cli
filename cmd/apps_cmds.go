package cmd

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/spf13/cobra"
)

var appsGetFlags struct {
	OrgID int64
}

var appsEligFlags struct {
	OrgID int64
}

var appsProductPagesFlags struct {
	OrgID int64
}

func init() {
	appsCmd.AddCommand(appsGetCmd)
	appsCmd.AddCommand(appsEligCmd)
	appsCmd.AddCommand(appsProductPagesCmd)

	appsGetCmd.Flags().Int64Var(&appsGetFlags.OrgID, "org-id", 0, "Organization ID override")
	appsEligCmd.Flags().Int64Var(&appsEligFlags.OrgID, "org-id", 0, "Organization ID override")
	appsProductPagesCmd.Flags().Int64Var(&appsProductPagesFlags.OrgID, "org-id", 0, "Organization ID override")
}

var appsGetCmd = &cobra.Command{
	Use:   "get <adam-id>",
	Short: "Get app details by adamId",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		adamID, err := parsePositiveInt64("adam-id", args[0])
		if err != nil {
			return err
		}
		client, _, _, err := authedClient(context.Background(), appsGetFlags.OrgID, true)
		if err != nil {
			return err
		}
		return callAPIAndPrint(context.Background(), client, http.MethodGet, fmt.Sprintf("/apps/%d", adamID), nil, nil)
	},
}

var appsEligCmd = &cobra.Command{
	Use:   "eligibilities <adam-id>",
	Short: "Get app eligibilities",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		adamID, err := parsePositiveInt64("adam-id", args[0])
		if err != nil {
			return err
		}
		client, _, _, err := authedClient(context.Background(), appsEligFlags.OrgID, true)
		if err != nil {
			return err
		}
		return callAPIAndPrint(context.Background(), client, http.MethodGet, fmt.Sprintf("/apps/%d/eligibilities", adamID), nil, nil)
	},
}

var appsProductPagesCmd = &cobra.Command{
	Use:   "product-pages <adam-id>",
	Short: "Get app product pages (default/custom)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		adamID, err := parsePositiveInt64("adam-id", args[0])
		if err != nil {
			return err
		}
		client, _, _, err := authedClient(context.Background(), appsProductPagesFlags.OrgID, true)
		if err != nil {
			return err
		}

		paths := []string{
			fmt.Sprintf("/apps/%d/product-pages", adamID),
			fmt.Sprintf("/apps/%d/productpages", adamID),
		}

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
		return fmt.Errorf("product pages lookup failed for app %d: %w", adamID, lastErr)
	},
}
