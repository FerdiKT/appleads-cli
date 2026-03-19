# Apple Search Ads API Setup Guide

Before using `appleads`, you need an Apple Search Ads account with API access. This guide walks you through the entire setup — from creating an API user to making your first API call.

> **Time required:** ~10 minutes

---

## Prerequisites

- An [Apple Search Ads Advanced](https://searchads.apple.com) account
- Account administrator access (to invite API users)
- `appleads` CLI installed ([installation guide](../README.md#-installation))

---

## Step 1 — Invite an API User

If you're the account admin **and** the API user, you still need to assign yourself an API role.

1. Sign in at [ads.apple.com](https://ads.apple.com) → **Sign In** → **Advanced**
2. Click the **user menu** (top-right corner) and select the account you want to manage
3. Go to **Account Settings** → **User Management**
4. Click **Invite Users**
5. Fill in the details:

   | Field | Value |
   |---|---|
   | First name / Last name | Your name (or service account name) |
   | Apple ID | The Apple ID email for the API user |
   | Role | **API Account Manager** (read + write) |

   > **💡 Tip:** If you only need read access (reports, listing campaigns), choose **API Account Read Only** instead.

6. Click **Send Invite**
7. The invited user receives an email with a secure code → they must sign in and activate the account

### API roles reference

| Role | Capabilities |
|---|---|
| **API Account Manager** | Full read/write access to all campaigns, ad groups, keywords, reports |
| **API Account Read Only** | Read-only access — can view and report but not modify |

---

## Step 2 — Generate Keys

You have two options: use `appleads` (recommended) or `openssl` manually.

### Option A: Using `appleads` (recommended)

```bash
# Generate an EC P-256 key pair
appleads auth keygen

# Print the public key (PEM format)
appleads auth public-key
```

`appleads` stores the private key securely in your profile configuration.

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

## Step 3 — Upload Public Key to Apple Ads

1. Sign in at [ads.apple.com](https://ads.apple.com) **as the API user** (the one you invited in Step 1)
2. Go to **Account Settings** → **API**
3. Paste the entire public key into the **Public Key** field, including the `-----BEGIN PUBLIC KEY-----` and `-----END PUBLIC KEY-----` lines
4. Click **Save**
5. Apple will display three credentials — **save them all:**

```
clientId   SEARCHADS.xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
teamId     SEARCHADS.xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
keyId      xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
```

> These credentials are shown **only once.** If you lose them, you'll need to re-upload a new public key.

---

## Step 4 — Configure `appleads`

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

## Step 5 — Generate Token & Select Org

```bash
# Generate an access token (valid for 1 hour)
appleads auth token

# List available orgs and select one interactively
appleads auth orgs --select
```

---

## Step 6 — Verify Everything

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

## Step 7 — First API Call

```bash
# List your campaigns
appleads campaigns list --limit 5

# Generate a report
appleads reports template campaigns --preset last-7d --run
```

🎉 **You're all set!**

---

## Troubleshooting

| Issue | Solution |
|---|---|
| **"org_id is not set"** | Run `appleads auth orgs --select` |
| **401 Unauthorized** | Your token may have expired. Run `appleads auth token` to refresh |
| **403 Forbidden** | Check that your API user role has the right permissions |
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
- Use **separate API users** for different team members
- Prefer **API Account Read Only** role when write access isn't needed
- **Rotate keys** periodically — regenerate and re-upload
- Use `appleads auth profiles export` through **secured channels** only

---

## Further Reading

- [Apple Ads API — Implementing OAuth](https://developer.apple.com/documentation/apple_search_ads/implementing_oauth_for_the_apple_search_ads_api) (official Apple documentation)
- [Apple Ads API Reference](https://developer.apple.com/documentation/apple_search_ads)
- [Apple Search Ads Help — User Management](https://searchads.apple.com/help)
