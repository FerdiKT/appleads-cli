package cmd

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/spf13/cobra"
)

var boListFlags struct {
	OrgID  int64
	Offset int
	Limit  int
	All    bool
}

var boGetFlags struct {
	OrgID int64
}

var boCreateFlags struct {
	OrgID    int64
	Body     string
	BodyFile string
}

var boUpdateFlags struct {
	OrgID    int64
	Body     string
	BodyFile string
}

var boFindFlags struct {
	OrgID    int64
	Body     string
	BodyFile string
}

func init() {
	budgetOrdersCmd.AddCommand(boListCmd)
	budgetOrdersCmd.AddCommand(boGetCmd)
	budgetOrdersCmd.AddCommand(boCreateCmd)
	budgetOrdersCmd.AddCommand(boUpdateCmd)
	budgetOrdersCmd.AddCommand(boFindCmd)

	boListCmd.Flags().Int64Var(&boListFlags.OrgID, "org-id", 0, "Organization ID override")
	boListCmd.Flags().IntVar(&boListFlags.Offset, "offset", 0, "Pagination offset")
	boListCmd.Flags().IntVar(&boListFlags.Limit, "limit", 20, "Pagination limit")
	boListCmd.Flags().BoolVar(&boListFlags.All, "all", false, "Fetch all pages")

	boGetCmd.Flags().Int64Var(&boGetFlags.OrgID, "org-id", 0, "Organization ID override")

	boCreateCmd.Flags().Int64Var(&boCreateFlags.OrgID, "org-id", 0, "Organization ID override")
	boCreateCmd.Flags().StringVar(&boCreateFlags.Body, "body", "", "Inline JSON body")
	boCreateCmd.Flags().StringVar(&boCreateFlags.BodyFile, "body-file", "", "Path to JSON body file")

	boUpdateCmd.Flags().Int64Var(&boUpdateFlags.OrgID, "org-id", 0, "Organization ID override")
	boUpdateCmd.Flags().StringVar(&boUpdateFlags.Body, "body", "", "Inline JSON body")
	boUpdateCmd.Flags().StringVar(&boUpdateFlags.BodyFile, "body-file", "", "Path to JSON body file")

	boFindCmd.Flags().Int64Var(&boFindFlags.OrgID, "org-id", 0, "Organization ID override")
	boFindCmd.Flags().StringVar(&boFindFlags.Body, "body", "", "Inline JSON selector body")
	boFindCmd.Flags().StringVar(&boFindFlags.BodyFile, "body-file", "", "Path to JSON selector body file")
}

var boListCmd = &cobra.Command{
	Use:   "list",
	Short: "List budget orders",
	RunE: func(cmd *cobra.Command, args []string) error {
		if boListFlags.Offset < 0 {
			return fmt.Errorf("--offset must be >= 0")
		}
		if boListFlags.Limit <= 0 {
			return fmt.Errorf("--limit must be > 0")
		}
		client, _, _, err := authedClient(context.Background(), boListFlags.OrgID, true)
		if err != nil {
			return err
		}
		q := url.Values{}
		q.Set("offset", strconv.Itoa(boListFlags.Offset))
		q.Set("limit", strconv.Itoa(boListFlags.Limit))
		return callListEndpointWithFallback(context.Background(), client, []string{
			"/budgetorders",
			"/budget-orders",
		}, q, boListFlags.Offset, boListFlags.Limit, boListFlags.All)
	},
}

var boGetCmd = &cobra.Command{
	Use:   "get <budget-order-id>",
	Short: "Get budget order by ID",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		budgetOrderID, err := parsePositiveInt64("budget-order-id", args[0])
		if err != nil {
			return err
		}
		client, _, _, err := authedClient(context.Background(), boGetFlags.OrgID, true)
		if err != nil {
			return err
		}
		return callAPIAndPrintWithFallback(context.Background(), client, http.MethodGet, []string{
			fmt.Sprintf("/budgetorders/%d", budgetOrderID),
			fmt.Sprintf("/budget-orders/%d", budgetOrderID),
		}, nil, nil)
	},
}

var boCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create budget order with JSON payload",
	Example: "  appleads budget-orders create --body-file ./payloads/budget-order-create.json\n" +
		"  appleads budget-orders create --body '{\"name\":\"BO-2026-Q1\"}'",
	RunE: func(cmd *cobra.Command, args []string) error {
		payload, err := readJSONPayload(boCreateFlags.Body, boCreateFlags.BodyFile, false)
		if err != nil {
			return err
		}
		client, _, _, err := authedClient(context.Background(), boCreateFlags.OrgID, true)
		if err != nil {
			return err
		}
		return callAPIAndPrintWithFallback(context.Background(), client, http.MethodPost, []string{
			"/budgetorders",
			"/budget-orders",
		}, nil, payload)
	},
}

var boUpdateCmd = &cobra.Command{
	Use:   "update <budget-order-id>",
	Short: "Update budget order by ID with JSON payload",
	Args:  cobra.ExactArgs(1),
	Example: "  appleads budget-orders update 123456 --body-file ./payloads/budget-order-update.json\n" +
		"  appleads budget-orders update 123456 --body '{\"status\":\"ACTIVE\"}'",
	RunE: func(cmd *cobra.Command, args []string) error {
		budgetOrderID, err := parsePositiveInt64("budget-order-id", args[0])
		if err != nil {
			return err
		}
		payload, err := readJSONPayload(boUpdateFlags.Body, boUpdateFlags.BodyFile, false)
		if err != nil {
			return err
		}
		client, _, _, err := authedClient(context.Background(), boUpdateFlags.OrgID, true)
		if err != nil {
			return err
		}
		return callAPIAndPrintWithFallback(context.Background(), client, http.MethodPut, []string{
			fmt.Sprintf("/budgetorders/%d", budgetOrderID),
			fmt.Sprintf("/budget-orders/%d", budgetOrderID),
		}, nil, payload)
	},
}

var boFindCmd = &cobra.Command{
	Use:   "find",
	Short: "Find budget orders with selector payload",
	Example: "  appleads budget-orders find --body-file ./payloads/budget-order-find.json\n" +
		"  appleads budget-orders find --body '{\"selector\":{\"pagination\":{\"offset\":0,\"limit\":20}}}'",
	RunE: func(cmd *cobra.Command, args []string) error {
		payload, err := readJSONPayload(boFindFlags.Body, boFindFlags.BodyFile, false)
		if err != nil {
			return err
		}
		client, _, _, err := authedClient(context.Background(), boFindFlags.OrgID, true)
		if err != nil {
			return err
		}
		return callAPIAndPrintWithFallback(context.Background(), client, http.MethodPost, []string{
			"/budgetorders/find",
			"/budget-orders/find",
		}, nil, payload)
	},
}
