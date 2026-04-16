# Apple Ads CLI Agent Guide

This repository contains an agent-first CLI for Apple Ads.

## Working Model

- Prefer JSON output.
- Resolve the correct auth profile before any API call.
- Make sure the profile has a valid token and org context before campaign/report commands.
- Use typed resource commands first; use `appleads api` as the escape hatch.
- Treat private keys, access tokens, and Apple credential values as secrets.

## Profile Workflow

1. Inspect profiles with `appleads auth profiles list` or `appleads auth show`.
2. Use `appleads auth profiles use <name>` or `-p <name>` to lock the target profile.
3. Run `appleads doctor` when auth, org, or API access is unclear.
4. Use `appleads auth orgs --select` when `org_id` is missing or needs to change.

## Auth Workflow

- Apple Ads Campaign Management API typically uses a designated API user, often a separate Apple ID invited through User Management.
- Prefer `appleads auth init` for first-time setup or key rotation.
- Reuse `appleads auth public-key` if the public key must be reprinted for Apple Ads.
- Use `appleads auth token` to refresh access tokens.
- `org_id` is not required for token creation, but most data commands need it.

## Read Before Write

- Start with `appleads doctor` or the relevant `list|get|find` command.
- For account/org discovery: `appleads account me`, `appleads account acls`, `appleads auth orgs --output json`.
- For broad reads: `appleads <resource> list --all`.
- Prefer `--output json` for agent workflows.

## Mutation Rules

- Prefer typed commands for CRUD and quick actions.
- Use `--body-file` or JSON body flags for complex create, update, find, and report payloads.
- Keep mutations scoped to one profile and one org.
- Respect `--dry-run`, interactive confirmation, and `--yes` flows.
- Do not assume the profile `org_id` is correct when a command can accept `--org-id`.

## Reporting And Raw API

- Use `appleads reports template ... --run` for common report presets.
- Use `appleads api` only when no typed command exists yet.
- When using `appleads api`, verify whether the route needs org context before calling it.
