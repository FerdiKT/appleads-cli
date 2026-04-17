package cmd

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

func maybePrepareOrgMutationContext(cmd *cobra.Command) error {
	if !shouldGuardOrgMutation(cmd) {
		return nil
	}

	_, profile, err := loadProfile()
	if err != nil {
		return err
	}

	overrideOrgID, err := commandInt64Flag(cmd, "org-id")
	if err != nil {
		return err
	}
	orgID, err := resolveOrgID(overrideOrgID, profile.OrgID)
	if err != nil {
		return err
	}
	if opts.ConfirmOrg > 0 && opts.ConfirmOrg != orgID {
		return fmt.Errorf("resolved org_id=%d for profile %q does not match --confirm-org=%d", orgID, opts.Profile, opts.ConfirmOrg)
	}

	_, _ = fmt.Fprintf(os.Stderr, "Active context: profile=%s org_id=%d\n", opts.Profile, orgID)
	return nil
}

func shouldGuardOrgMutation(cmd *cobra.Command) bool {
	if cmd == nil {
		return false
	}
	if strings.HasPrefix(cmd.CommandPath(), "appleads api") {
		return apiCallLooksMutating()
	}
	if cmd.Flags().Lookup("org-id") == nil {
		return false
	}

	path := cmd.CommandPath()
	switch {
	case strings.HasPrefix(path, "appleads auth "):
		return false
	case strings.HasPrefix(path, "appleads doctor"):
		return false
	case strings.HasPrefix(path, "appleads account"):
		return false
	case strings.HasPrefix(path, "appleads orgs"):
		return false
	}

	switch cmd.Name() {
	case "create", "update", "delete", "enable", "pause", "set", "clear", "replace", "add", "remove", "only", "only-iphone", "only-ipad", "only-both":
		return true
	default:
		return false
	}
}

func apiCallLooksMutating() bool {
	method := strings.ToUpper(strings.TrimSpace(apiFlags.Method))
	switch method {
	case http.MethodGet, http.MethodHead:
		return false
	case http.MethodPost:
		path := strings.TrimSpace(apiFlags.Path)
		if strings.HasSuffix(path, "/find") || strings.HasPrefix(path, "/reports") {
			return false
		}
		return true
	case http.MethodPut, http.MethodDelete:
		return true
	default:
		return false
	}
}

func commandInt64Flag(cmd *cobra.Command, name string) (int64, error) {
	if cmd == nil {
		return 0, nil
	}
	flag := cmd.Flags().Lookup(name)
	if flag == nil {
		return 0, nil
	}
	value := strings.TrimSpace(flag.Value.String())
	if value == "" {
		return 0, nil
	}
	n, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse --%s: %w", name, err)
	}
	return n, nil
}
