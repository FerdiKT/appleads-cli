package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/ferdikt/appleads-cli/internal/config"
	"github.com/spf13/cobra"
)

func init() {
	authCmd.AddCommand(authProfilesCmd)
	authProfilesCmd.AddCommand(authProfilesListCmd)
	authProfilesCmd.AddCommand(authProfilesUseCmd)
	authProfilesCmd.AddCommand(authProfilesCreateCmd)
	authProfilesCmd.AddCommand(authProfilesDeleteCmd)
	authProfilesCmd.AddCommand(authProfilesRenameCmd)
	authProfilesCmd.AddCommand(authProfilesCloneCmd)
	authProfilesCmd.AddCommand(authProfilesExportCmd)
	authProfilesCmd.AddCommand(authProfilesImportCmd)
}

var authProfilesCmd = &cobra.Command{
	Use:   "profiles",
	Short: "Manage auth profiles",
}

var authProfilesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List configured profiles",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return err
		}
		names := cfg.ProfileNames()
		active := cfg.EffectiveProfileName(opts.Profile)

		if opts.Output == "json" {
			items := make([]map[string]any, 0, len(names))
			for _, name := range names {
				p := cfg.Profiles[name]
				items = append(items, map[string]any{
					"name":      name,
					"active":    name == active,
					"client_id": p.ClientID,
					"team_id":   p.TeamID,
					"org_id":    p.OrgID,
				})
			}
			return printJSON(map[string]any{
				"active_profile": active,
				"profiles":       items,
			})
		}

		if len(names) == 0 {
			fmt.Println("no profiles configured")
			return nil
		}

		w := tableWriter()
		fmt.Fprintln(w, "NAME\tACTIVE\tORG_ID\tCLIENT_ID\tTEAM_ID")
		for _, name := range names {
			p := cfg.Profiles[name]
			isActive := ""
			if name == active {
				isActive = "*"
			}
			fmt.Fprintf(w, "%s\t%s\t%d\t%s\t%s\n", name, isActive, p.OrgID, p.ClientID, p.TeamID)
		}
		return w.Flush()
	},
}

var authProfilesUseCmd = &cobra.Command{
	Use:     "use <profile>",
	Aliases: []string{"switch"},
	Short:   "Switch active profile",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		profileName := args[0]
		cfg, err := loadConfig()
		if err != nil {
			return err
		}
		if _, err := cfg.GetProfile(profileName); err != nil {
			return err
		}
		cfg.ActiveProfile = profileName
		if err := cfg.Save(opts.ConfigPath); err != nil {
			return err
		}
		if opts.Output == "json" {
			return printJSON(map[string]any{
				"active_profile": profileName,
				"saved":          true,
			})
		}
		fmt.Printf("active profile set to %q\n", profileName)
		return nil
	},
}

var authProfilesCreateCmd = &cobra.Command{
	Use:   "create <profile>",
	Short: "Create an empty profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		profileName := args[0]
		cfg, err := loadConfig()
		if err != nil {
			return err
		}
		if _, exists := cfg.Profiles[profileName]; exists {
			return fmt.Errorf("profile %q already exists", profileName)
		}
		cfg.EnsureProfile(profileName)
		if cfg.ActiveProfile == "" {
			if _, ok := cfg.Profiles["default"]; ok {
				cfg.ActiveProfile = "default"
			} else {
				cfg.ActiveProfile = profileName
			}
		}
		if err := cfg.Save(opts.ConfigPath); err != nil {
			return err
		}
		if opts.Output == "json" {
			return printJSON(map[string]any{
				"profile":        profileName,
				"active_profile": cfg.ActiveProfile,
				"created":        true,
			})
		}
		fmt.Printf("profile %q created\n", profileName)
		return nil
	},
}

var authProfilesDeleteFlags struct {
	Force bool
}

var authProfilesExportFlags struct {
	Out string
}

var authProfilesImportFlags struct {
	Name      string
	SetActive bool
}

var authProfilesDeleteCmd = &cobra.Command{
	Use:   "delete <profile>",
	Short: "Delete a profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		profileName := args[0]
		cfg, err := loadConfig()
		if err != nil {
			return err
		}
		if _, exists := cfg.Profiles[profileName]; !exists {
			return fmt.Errorf("profile %q not found", profileName)
		}
		if cfg.ActiveProfile == profileName && !authProfilesDeleteFlags.Force {
			return fmt.Errorf("profile %q is active; switch profile first or use --force", profileName)
		}
		delete(cfg.Profiles, profileName)

		if cfg.ActiveProfile == profileName {
			cfg.ActiveProfile = ""
		}
		if cfg.ActiveProfile == "" {
			names := cfg.ProfileNames()
			if len(names) > 0 {
				cfg.ActiveProfile = names[0]
			}
		}

		if err := cfg.Save(opts.ConfigPath); err != nil {
			return err
		}
		if opts.Output == "json" {
			return printJSON(map[string]any{
				"deleted_profile": profileName,
				"active_profile":  cfg.ActiveProfile,
				"deleted":         true,
			})
		}
		fmt.Printf("profile %q deleted\n", profileName)
		if cfg.ActiveProfile != "" {
			fmt.Printf("active profile: %q\n", cfg.ActiveProfile)
		}
		return nil
	},
}

var authProfilesRenameCmd = &cobra.Command{
	Use:   "rename <old> <new>",
	Short: "Rename a profile",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		oldName := strings.TrimSpace(args[0])
		newName := strings.TrimSpace(args[1])
		if oldName == "" || newName == "" {
			return fmt.Errorf("profile names cannot be empty")
		}

		cfg, err := loadConfig()
		if err != nil {
			return err
		}
		p, ok := cfg.Profiles[oldName]
		if !ok || p == nil {
			return fmt.Errorf("profile %q not found", oldName)
		}
		if _, exists := cfg.Profiles[newName]; exists {
			return fmt.Errorf("profile %q already exists", newName)
		}
		delete(cfg.Profiles, oldName)
		cfg.Profiles[newName] = p
		if cfg.ActiveProfile == oldName {
			cfg.ActiveProfile = newName
		}
		if err := cfg.Save(opts.ConfigPath); err != nil {
			return err
		}
		if opts.Output == "json" {
			return printJSON(map[string]any{
				"renamed":        true,
				"old":            oldName,
				"new":            newName,
				"active_profile": cfg.ActiveProfile,
			})
		}
		fmt.Printf("profile %q renamed to %q\n", oldName, newName)
		return nil
	},
}

var authProfilesCloneCmd = &cobra.Command{
	Use:   "clone <source> <target>",
	Short: "Clone profile values into a new profile",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		source := strings.TrimSpace(args[0])
		target := strings.TrimSpace(args[1])
		if source == "" || target == "" {
			return fmt.Errorf("profile names cannot be empty")
		}
		cfg, err := loadConfig()
		if err != nil {
			return err
		}
		src, ok := cfg.Profiles[source]
		if !ok || src == nil {
			return fmt.Errorf("profile %q not found", source)
		}
		if _, exists := cfg.Profiles[target]; exists {
			return fmt.Errorf("profile %q already exists", target)
		}

		cp := *src
		cfg.Profiles[target] = &cp
		if err := cfg.Save(opts.ConfigPath); err != nil {
			return err
		}
		if opts.Output == "json" {
			return printJSON(map[string]any{
				"cloned":         true,
				"source":         source,
				"target":         target,
				"active_profile": cfg.ActiveProfile,
			})
		}
		fmt.Printf("profile %q cloned to %q\n", source, target)
		return nil
	},
}

type exportedProfile struct {
	Profile string          `json:"profile"`
	Data    *config.Profile `json:"data"`
}

var authProfilesExportCmd = &cobra.Command{
	Use:   "export <profile>",
	Short: "Export a profile as JSON",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := strings.TrimSpace(args[0])
		cfg, err := loadConfig()
		if err != nil {
			return err
		}
		p, err := cfg.GetProfile(name)
		if err != nil {
			return err
		}

		payload := exportedProfile{Profile: name, Data: p}
		raw, err := json.MarshalIndent(payload, "", "  ")
		if err != nil {
			return err
		}
		raw = append(raw, '\n')

		if authProfilesExportFlags.Out == "" || authProfilesExportFlags.Out == "-" {
			_, err = os.Stdout.Write(raw)
			return err
		}
		if err := os.WriteFile(authProfilesExportFlags.Out, raw, 0o600); err != nil {
			return err
		}
		if opts.Output == "json" {
			return printJSON(map[string]any{
				"exported": true,
				"profile":  name,
				"out":      authProfilesExportFlags.Out,
			})
		}
		fmt.Printf("profile %q exported to %s\n", name, authProfilesExportFlags.Out)
		return nil
	},
}

var authProfilesImportCmd = &cobra.Command{
	Use:   "import <file>",
	Short: "Import a profile JSON export",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := strings.TrimSpace(args[0])
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		var payload exportedProfile
		if err := json.Unmarshal(raw, &payload); err != nil {
			return fmt.Errorf("parse profile export: %w", err)
		}
		if payload.Data == nil {
			return fmt.Errorf("invalid export: missing data")
		}

		targetName := strings.TrimSpace(authProfilesImportFlags.Name)
		if targetName == "" {
			targetName = strings.TrimSpace(payload.Profile)
		}
		if targetName == "" {
			return fmt.Errorf("target profile name is empty (pass --name)")
		}

		cfg, err := loadConfig()
		if err != nil {
			return err
		}
		if _, exists := cfg.Profiles[targetName]; exists {
			return fmt.Errorf("profile %q already exists", targetName)
		}
		cp := *payload.Data
		cfg.Profiles[targetName] = &cp
		if authProfilesImportFlags.SetActive || cfg.ActiveProfile == "" {
			cfg.ActiveProfile = targetName
		}
		if err := cfg.Save(opts.ConfigPath); err != nil {
			return err
		}
		if opts.Output == "json" {
			return printJSON(map[string]any{
				"imported":       true,
				"profile":        targetName,
				"active_profile": cfg.ActiveProfile,
			})
		}
		fmt.Printf("profile imported as %q\n", targetName)
		return nil
	},
}

func init() {
	authProfilesDeleteCmd.Flags().BoolVar(&authProfilesDeleteFlags.Force, "force", false, "Delete even if profile is active")
	authProfilesExportCmd.Flags().StringVar(&authProfilesExportFlags.Out, "out", "-", "Output file path ('-' for stdout)")
	authProfilesImportCmd.Flags().StringVar(&authProfilesImportFlags.Name, "name", "", "Override imported profile name")
	authProfilesImportCmd.Flags().BoolVar(&authProfilesImportFlags.SetActive, "set-active", false, "Set imported profile as active")
}
