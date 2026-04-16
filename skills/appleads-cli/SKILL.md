---
name: appleads-cli
description: Use this skill when working with the local `appleads` CLI for Apple Ads profile-based auth, API user setup, org discovery, campaign management, reports, targeting, and agent-safe mutations across campaigns, ad groups, ads, keywords, creatives, and budget orders.
---

# Apple Ads CLI

Use this skill for repository-local Apple Ads CLI work.

## Workflow

1. Resolve the correct auth profile first.
2. Verify token and org readiness before app, campaign, or report operations.
3. Prefer JSON output for agent workflows.
4. Use typed resource commands before falling back to `appleads api`.

## Profile Resolution

- Inspect profiles with `appleads auth profiles list` or `appleads auth show`.
- Select the active profile with `appleads auth profiles use <name>`.
- Override per call with `-p <name>`.
- Use `appleads doctor` if auth, token, org, or endpoint access is uncertain.

If a profile is missing `org_id`, run `appleads auth orgs --select` or pass `--org-id` on the command.

## Auth Guardrail

- Campaign Management API setup usually uses a designated API user, often a separate Apple ID invited via Apple Ads User Management.
- Use `appleads auth init` for guided setup and key rotation.
- `appleads auth init` can parse the Apple credential block as a direct paste.
- Use `appleads auth token` to mint or refresh a token.
- Never print private keys or raw access tokens in normal output.

## Read Pattern

- Start with `appleads doctor` for health checks.
- Use `appleads account me` and `appleads account acls` for access discovery.
- Use `appleads <resource> list|get|find` for focused reads.
- Add `--all` where pagination matters.
- Prefer `--output json` for downstream agent processing.

## Mutation Pattern

- Use typed `create`, `update`, `delete`, `enable`, `pause`, `set`, `clear`, and `replace` commands first.
- Prefer `--body-file` for larger JSON payloads.
- Keep mutations single-profile and single-org.
- Use `--dry-run` before destructive or broad updates.
- Respect interactive confirmation or use `--yes` only when the target is already validated.

## Reports And Escape Hatch

- Use `appleads reports template ... --run` for preset report queries.
- Use `appleads api` only when a typed command is missing.
- When using `appleads api`, confirm whether the endpoint needs org context and pass `--org-id` if needed.
