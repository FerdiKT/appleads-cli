package cmd

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

const (
	bundledSkillName = "appleads-cli"
	bundledSkillDoc  = `---
name: appleads-cli
description: Use this skill when working with the local ` + "`appleads`" + ` CLI for Apple Ads profile-based auth, API user setup, org discovery, campaign management, reports, targeting, and agent-safe mutations across campaigns, ad groups, ads, keywords, creatives, and budget orders.
---

# Apple Ads CLI

Use this skill for repository-local Apple Ads CLI work.

## Workflow

1. Resolve the correct auth profile first.
2. Verify token and org readiness before app, campaign, or report operations.
3. Prefer JSON output for agent workflows.
4. Use typed resource commands before falling back to ` + "`appleads api`" + `.

## Profile Resolution

- Inspect profiles with ` + "`appleads auth profiles list`" + ` or ` + "`appleads auth show`" + `.
- Select the active profile with ` + "`appleads auth profiles use <name>`" + `.
- Override per call with ` + "`-p <name>`" + `.
- Use ` + "`appleads doctor`" + ` if auth, token, org, or endpoint access is uncertain.

If a profile is missing ` + "`org_id`" + `, run ` + "`appleads auth orgs --select`" + ` or pass ` + "`--org-id`" + ` on the command.

## Auth Guardrail

- Campaign Management API setup usually uses a designated API user, often a separate Apple ID invited via Apple Ads User Management.
- Use ` + "`appleads auth init`" + ` for guided setup and key rotation.
- ` + "`appleads auth init`" + ` can parse the Apple credential block as a direct paste.
- Use ` + "`appleads auth token`" + ` to mint or refresh a token.
- Never print private keys or raw access tokens in normal output.

## Read Pattern

- Start with ` + "`appleads doctor`" + ` for health checks.
- Use ` + "`appleads account me`" + ` and ` + "`appleads account acls`" + ` for access discovery.
- Use ` + "`appleads <resource> list|get|find`" + ` for focused reads.
- Add ` + "`--all`" + ` where pagination matters.
- Prefer ` + "`--output json`" + ` for downstream agent processing.

## Mutation Pattern

- Use typed ` + "`create`" + `, ` + "`update`" + `, ` + "`delete`" + `, ` + "`enable`" + `, ` + "`pause`" + `, ` + "`set`" + `, ` + "`clear`" + `, and ` + "`replace`" + ` commands first.
- Prefer ` + "`--body-file`" + ` for larger JSON payloads.
- Keep mutations single-profile and single-org.
- Use ` + "`--dry-run`" + ` before destructive or broad updates.
- Respect interactive confirmation or use ` + "`--yes`" + ` only when the target is already validated.

## Reports And Escape Hatch

- Use ` + "`appleads reports template ... --run`" + ` for preset report queries.
- Use ` + "`appleads api`" + ` only when a typed command is missing.
- When using ` + "`appleads api`" + `, confirm whether the endpoint needs org context and pass ` + "`--org-id`" + ` if needed.
`
	bundledAgentYAML = `display_name: Apple Ads CLI
short_description: Work with the local Apple Ads agent-first CLI using profile-first, org-aware workflows.
default_prompt: Use the local appleads CLI. Resolve the target auth profile first, verify token and org readiness before data commands, prefer JSON output, and use typed commands before falling back to appleads api.
`
)

func init() {
	rootCmd.AddCommand(newAgentCmd())
}

func newAgentCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "agent",
		Short: "Install or link bundled agent assets such as the Codex skill",
	}

	cmd.AddCommand(
		newAgentInstallSkillCmd(),
		newAgentLinkSkillCmd(),
		newAgentShowSkillPathCmd(),
	)
	return cmd
}

func newAgentInstallSkillCmd() *cobra.Command {
	var codexHome string
	var force bool

	cmd := &cobra.Command{
		Use:   "install-skill",
		Short: "Install the bundled appleads Codex skill into CODEX_HOME",
		RunE: func(cmd *cobra.Command, args []string) error {
			targetDir, err := resolveSkillInstallDir(codexHome)
			if err != nil {
				return err
			}
			if err := ensureReplaceableTarget(targetDir, force); err != nil {
				return err
			}
			if err := writeBundledSkill(targetDir); err != nil {
				return err
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "skill installed: %s\n", targetDir)
			return err
		},
	}

	cmd.Flags().StringVar(&codexHome, "codex-home", "", "Override CODEX_HOME for the destination")
	cmd.Flags().BoolVar(&force, "force", false, "Replace an existing installed skill")
	return cmd
}

func newAgentLinkSkillCmd() *cobra.Command {
	var codexHome string
	var source string
	var force bool

	cmd := &cobra.Command{
		Use:   "link-skill",
		Short: "Symlink the local appleads skill into CODEX_HOME",
		RunE: func(cmd *cobra.Command, args []string) error {
			targetDir, err := resolveSkillInstallDir(codexHome)
			if err != nil {
				return err
			}
			if source == "" {
				source = filepath.Join(".", "skills", bundledSkillName)
			}
			source, err = filepath.Abs(source)
			if err != nil {
				return fmt.Errorf("resolve source path: %w", err)
			}
			if err := validateSkillSource(source); err != nil {
				return err
			}
			if err := ensureReplaceableTarget(targetDir, force); err != nil {
				return err
			}
			if err := os.MkdirAll(filepath.Dir(targetDir), 0o755); err != nil {
				return fmt.Errorf("create destination parent: %w", err)
			}
			if err := os.Symlink(source, targetDir); err != nil {
				return fmt.Errorf("create symlink: %w", err)
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "skill linked: %s -> %s\n", targetDir, source)
			return err
		},
	}

	cmd.Flags().StringVar(&codexHome, "codex-home", "", "Override CODEX_HOME for the destination")
	cmd.Flags().StringVar(&source, "source", "", "Source skill directory to symlink; defaults to ./skills/appleads-cli")
	cmd.Flags().BoolVar(&force, "force", false, "Replace an existing installed skill or symlink")
	return cmd
}

func newAgentShowSkillPathCmd() *cobra.Command {
	var codexHome string

	cmd := &cobra.Command{
		Use:   "show-skill-path",
		Short: "Print the destination path for the appleads Codex skill",
		RunE: func(cmd *cobra.Command, args []string) error {
			targetDir, err := resolveSkillInstallDir(codexHome)
			if err != nil {
				return err
			}
			_, err = fmt.Fprintln(cmd.OutOrStdout(), targetDir)
			return err
		},
	}

	cmd.Flags().StringVar(&codexHome, "codex-home", "", "Override CODEX_HOME for the destination")
	return cmd
}

func resolveSkillInstallDir(codexHome string) (string, error) {
	home := codexHome
	if home == "" {
		home = os.Getenv("CODEX_HOME")
	}
	if home == "" {
		userHome, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("resolve user home: %w", err)
		}
		home = filepath.Join(userHome, ".codex")
	}
	return filepath.Join(home, "skills", bundledSkillName), nil
}

func ensureReplaceableTarget(target string, force bool) error {
	info, err := os.Lstat(target)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("inspect target: %w", err)
	}
	if !force {
		return fmt.Errorf("target already exists: %s (use --force to replace)", target)
	}
	if info.Mode()&os.ModeSymlink != 0 || info.IsDir() {
		return os.RemoveAll(target)
	}
	return os.Remove(target)
}

func writeBundledSkill(target string) error {
	if err := os.MkdirAll(filepath.Join(target, "agents"), 0o755); err != nil {
		return fmt.Errorf("create skill directory: %w", err)
	}
	if err := os.WriteFile(filepath.Join(target, "SKILL.md"), []byte(bundledSkillDoc), 0o644); err != nil {
		return fmt.Errorf("write SKILL.md: %w", err)
	}
	if err := os.WriteFile(filepath.Join(target, "agents", "openai.yaml"), []byte(bundledAgentYAML), 0o644); err != nil {
		return fmt.Errorf("write agents/openai.yaml: %w", err)
	}
	return nil
}

func validateSkillSource(source string) error {
	info, err := os.Stat(source)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("skill source does not exist: %s", source)
		}
		return fmt.Errorf("inspect source: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("skill source is not a directory: %s", source)
	}
	if _, err := os.Stat(filepath.Join(source, "SKILL.md")); err != nil {
		return fmt.Errorf("skill source is missing SKILL.md: %s", source)
	}
	return nil
}
