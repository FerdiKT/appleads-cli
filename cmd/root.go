package cmd

import (
	"fmt"
	"os"

	"github.com/ferdikt/appleads-cli/internal/buildinfo"
	"github.com/ferdikt/appleads-cli/internal/config"
	"github.com/spf13/cobra"
)

type globalOptions struct {
	Profile    string
	Output     string
	JSON       bool
	ConfirmOrg int64
	ConfigPath string
}

var opts globalOptions

var rootCmd = &cobra.Command{
	Use:   "appleads",
	Short: "Apple Ads CLI",
	Long:  "A command line interface for Apple Ads API operations.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		profileChanged := cmd.Flags().Changed("profile")
		if opts.JSON {
			opts.Output = "json"
		}
		switch opts.Output {
		case "table", "json":
			// continue
		default:
			return fmt.Errorf("invalid --output value %q (allowed: table, json)", opts.Output)
		}

		if !profileChanged {
			cfg, err := config.Load(opts.ConfigPath)
			if err == nil {
				opts.Profile = cfg.EffectiveProfileName(opts.Profile)
			}
		}
		if err := maybePrepareOrgMutationContext(cmd); err != nil {
			return err
		}
		return nil
	},
}

func init() {
	rootCmd.Version = buildinfo.Version
	rootCmd.SetVersionTemplate("{{.Version}}\n")

	defaultConfigPath, err := config.DefaultPath()
	if err != nil {
		defaultConfigPath = "appleads-config.json"
	}

	rootCmd.PersistentFlags().StringVarP(&opts.Profile, "profile", "p", "default", "Credential profile name")
	rootCmd.PersistentFlags().StringVarP(&opts.Output, "output", "o", "table", "Output format (table|json)")
	rootCmd.PersistentFlags().BoolVar(&opts.JSON, "json", false, "Alias for --output json")
	rootCmd.PersistentFlags().Int64Var(&opts.ConfirmOrg, "confirm-org", 0, "Abort if the resolved org_id does not match this value")
	rootCmd.PersistentFlags().StringVar(&opts.ConfigPath, "config", defaultConfigPath, "Config file path")
}

func Execute() error {
	rootCmd.SilenceUsage = true
	rootCmd.SilenceErrors = true
	return rootCmd.Execute()
}

func Exitf(format string, args ...any) {
	_, _ = fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
