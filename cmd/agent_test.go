package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func executeCommand(t *testing.T, cmd *cobra.Command, args []string) (string, string, error) {
	t.Helper()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs(args)

	err := cmd.Execute()
	return stdout.String(), stderr.String(), err
}

func TestAgentInstallSkillWritesBundledFiles(t *testing.T) {
	codexHome := t.TempDir()
	cmd := newAgentCmd()

	stdout, stderr, err := executeCommand(t, cmd, []string{
		"install-skill",
		"--codex-home", codexHome,
	})
	if err != nil {
		t.Fatalf("Execute: %v stderr=%s", err, stderr)
	}
	if !strings.Contains(stdout, "skill installed:") {
		t.Fatalf("stdout = %q, want install message", stdout)
	}

	skillDir := filepath.Join(codexHome, "skills", bundledSkillName)
	skillDoc, err := os.ReadFile(filepath.Join(skillDir, "SKILL.md"))
	if err != nil {
		t.Fatalf("ReadFile SKILL.md: %v", err)
	}
	if !strings.Contains(string(skillDoc), "Use this skill for repository-local Apple Ads CLI work.") {
		t.Fatalf("unexpected SKILL.md content: %s", string(skillDoc))
	}

	agentYAML, err := os.ReadFile(filepath.Join(skillDir, "agents", "openai.yaml"))
	if err != nil {
		t.Fatalf("ReadFile openai.yaml: %v", err)
	}
	if !strings.Contains(string(agentYAML), "display_name: Apple Ads CLI") {
		t.Fatalf("unexpected openai.yaml content: %s", string(agentYAML))
	}
}

func TestAgentLinkSkillCreatesSymlink(t *testing.T) {
	sourceRoot := t.TempDir()
	sourceDir := filepath.Join(sourceRoot, bundledSkillName)
	if err := os.MkdirAll(filepath.Join(sourceDir, "agents"), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sourceDir, "SKILL.md"), []byte("test skill"), 0o644); err != nil {
		t.Fatalf("WriteFile SKILL.md: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sourceDir, "agents", "openai.yaml"), []byte("display_name: Test"), 0o644); err != nil {
		t.Fatalf("WriteFile openai.yaml: %v", err)
	}

	codexHome := t.TempDir()
	cmd := newAgentCmd()
	_, stderr, err := executeCommand(t, cmd, []string{
		"link-skill",
		"--codex-home", codexHome,
		"--source", sourceDir,
	})
	if err != nil {
		t.Fatalf("Execute: %v stderr=%s", err, stderr)
	}

	targetDir := filepath.Join(codexHome, "skills", bundledSkillName)
	info, err := os.Lstat(targetDir)
	if err != nil {
		t.Fatalf("Lstat: %v", err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("target is not a symlink: mode=%v", info.Mode())
	}
	resolved, err := os.Readlink(targetDir)
	if err != nil {
		t.Fatalf("Readlink: %v", err)
	}
	if resolved != sourceDir {
		t.Fatalf("symlink target = %q, want %q", resolved, sourceDir)
	}
}

func TestAgentShowSkillPath(t *testing.T) {
	codexHome := t.TempDir()
	cmd := newAgentCmd()
	stdout, stderr, err := executeCommand(t, cmd, []string{
		"show-skill-path",
		"--codex-home", codexHome,
	})
	if err != nil {
		t.Fatalf("Execute: %v stderr=%s", err, stderr)
	}

	want := filepath.Join(codexHome, "skills", bundledSkillName)
	if strings.TrimSpace(stdout) != want {
		t.Fatalf("stdout = %q, want %q", stdout, want)
	}
}
