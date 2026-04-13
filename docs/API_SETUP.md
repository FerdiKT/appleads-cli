# Apple Ads API Setup Guide

Before using `appleads`, you need an Apple Ads user with enough access to the target account or campaign groups. This guide walks through the current OAuth flow from zero to the first API call.

> **Time required:** ~10 minutes

---

## Prerequisites

- An [Apple Search Ads Advanced](https://searchads.apple.com) account
- `appleads` CLI installed ([installation guide](../README.md#-installation))
- An Apple Ads user that can access the account you want to manage

### Which user should you use?

Use the Apple Ads login that will own API access.

- If you already have the right Apple Ads access, use your existing login.
- If another person will operate the CLI, invite them first in Apple Ads.
- For Campaign Management API setup, Apple's documented flow uses a designated API user, which is usually a separate Apple ID invited through User Management.
- Do not assume the primary/admin Apple Ads login is the same thing as the API user.
- OAuth does **not** replace Apple Ads permissions. The user must still have access to the target account or campaign groups.
- For third-party OAuth authorizations such as RevenueCat, RevenueCat currently documents that the granting user should be an **Account Admin** or **Campaign Group Manager**.

---

## Step 1 — Make sure the Apple Ads user can access the account

If you are not already using a suitable API user:

1. Sign in at [ads.apple.com](https://ads.apple.com) with an admin-capable account
2. Open **Account Settings** → **User Management**
3. Invite a separate Apple ID that will act as the API user for `appleads`
4. Grant the minimum API-capable access needed for that user
5. Accept the invitation on that separate Apple ID and confirm it can see the relevant account or campaign groups

> The exact role names in Apple Ads can vary by account setup and region. The important point is not the label, but that the user can access the entities you expect the CLI to manage.

---

## Step 2 — Generate a local key pair

You have two options: use `appleads` (recommended) or `openssl` manually.

### Option A: Using `appleads` (recommended)

```bash
# Generate a P-256 key pair locally
appleads auth keygen

# Print the public key again later if needed
appleads auth public-key
```

`appleads` stores the private key path in your selected profile. The private key stays on your machine.

### Option B: Using OpenSSL manually

```bash
# Generate a private key
openssl ecparam -genkey -name prime256v1 -noout -out private-key.pem

# Extract the public key
openssl ec -in private-key.pem -pubout -out public-key.pem

# View the public key to copy it
cat public-key.pem
```

> ⚠️ **Never share your private key.** If it's compromised, regenerate both keys and re-upload.

---

## Step 3 — Upload the public key and collect Apple credentials

1. Sign in at [ads.apple.com](https://ads.apple.com) **as the API user Apple ID** you will use with `appleads`
2. Open **Account Settings** → **Client Credentials** (or **API**, depending on the UI version)
3. Paste the entire public key, including the `-----BEGIN PUBLIC KEY-----` and `-----END PUBLIC KEY-----` lines
4. Click **Save**
5. Apple will display three values. Copy all of them:

```
clientId   SEARCHADS.xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
teamId     SEARCHADS.xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
keyId      xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
```

`appleads auth init` can accept this block as a direct paste. You do not need to copy the three values one by one.

> These values are typically shown **once**. If you lose them, the safest recovery path is to upload a new public key and rotate the key pair.

---

## Step 4 — Save the credentials in `appleads`

Now save these credentials in your `appleads` profile:

```bash
appleads auth set \
  --client-id  "SEARCHADS.xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx" \
  --team-id    "SEARCHADS.xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx" \
  --key-id     "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
```

Or use the interactive setup:

```bash
appleads auth init
```

---

## Step 5 — Create a token and select an org

```bash
# Generate an access token
appleads auth token

# List available orgs and save one interactively
appleads auth orgs --select
```

Notes:

- `org_id` is **not** required to mint the token.
- `org_id` **is** required by most campaign, keyword, ad, and report commands.
- If you prefer not to save one globally, you can pass `--org-id` on individual commands.

---

## Step 6 — Verify everything

Run the built-in health check to make sure everything is configured correctly:

```bash
appleads doctor
```

You should see all green checks:

```
✓ Config file resolved
✓ Profile loaded
✓ Auth fields present
✓ Private key readable
✓ Client secret generated
✓ Token valid
✓ Org resolved
✓ API reachable
```

---

## Step 7 — First API call

```bash
# List your campaigns
appleads campaigns list --limit 5

# Generate a report
appleads reports template campaigns --preset last-7d --run
```

You are ready to use the CLI.

---

## Troubleshooting

| Issue | Solution |
|---|---|
| **"org_id is not set"** | Run `appleads auth orgs --select` |
| **401 Unauthorized** | Your token may have expired. Run `appleads auth token` to refresh |
| **403 Forbidden** | The Apple Ads user lacks access to this account, campaign group, or action |
| **"private key not found"** | If you used OpenSSL, make sure you ran `appleads auth init` and pointed to the key file |
| **Doctor shows ✗ on "API reachable"** | Check your internet connection and that the org is correctly selected |

---

## Multi-Account Setup

If you manage multiple Apple Ads accounts (e.g., agency use), create separate profiles:

```bash
# Create named profiles
appleads auth profiles create client-acme
appleads auth profiles create client-beta

# Set up each profile
appleads -p client-acme auth init
appleads -p client-beta auth init

# Switch between them
appleads auth profiles use client-acme
```

See the [main README](../README.md#-multi-account-workflow) for more profile operations.

---

## Security Best Practices

- **Never commit** `.pem` private keys or `config.json` files to version control
- Use separate Apple Ads users or profiles when ownership boundaries matter
- Prefer the least-privileged Apple Ads role that still fits the workflow
- **Rotate keys** periodically — regenerate and re-upload
- Use `appleads auth profiles export` through **secured channels** only

---

## Further Reading

- [Apple Ads API — Implementing OAuth](https://developer.apple.com/documentation/apple_search_ads/implementing_oauth_for_the_apple_search_ads_api) (official Apple documentation)
- [Apple Ads API Reference](https://developer.apple.com/documentation/apple_search_ads)
- [Apple Search Ads Help — User Management](https://searchads.apple.com/help)
