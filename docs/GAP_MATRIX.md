# Apple Ads UI vs CLI Gap Matrix

Last updated: 2026-03-18

## Legend

- `Covered`: Typed CLI commands are present.
- `Partial`: Some workflows are covered; advanced UI actions still rely on raw API or are account-gated.
- `Missing`: No typed command yet.

## Matrix

| UI Area | Status | CLI Coverage |
|---|---|---|
| Auth / OAuth client setup | Partial | `auth init`, `auth keygen`, `auth public-key`, `auth set`, `auth token`, `auth orgs` |
| Organization selection | Covered | `auth orgs --select`, global `--org-id` override |
| Campaign list/details/create/update/delete | Covered | `campaigns list/get/create/update/delete/find` |
| Campaign quick status actions | Covered | `campaigns enable/pause` |
| Ad group list/details/create/update/delete | Covered | `adgroups list/get/create/update/delete/find` |
| Ad group quick status actions | Covered | `adgroups enable/pause` |
| Ads list/details/create/update/delete | Covered | `ads list/get/create/update/delete/find` |
| Ad quick status actions | Covered | `ads enable/pause` |
| Targeting keywords CRUD/find | Covered | `keywords targeting list/get/create/update/delete/find` |
| Targeting keyword quick status actions | Covered | `keywords targeting enable/pause` |
| Negative keywords (campaign/adgroup) | Covered | `keywords campaign-negative ...`, `keywords adgroup-negative ...` |
| Campaign negative keyword quick status actions | Covered | `keywords campaign-negative enable/pause` |
| Keyword recommendations | Partial | `keywords recommendations list/find` (endpoint availability account-dependent) |
| Targeting dimensions editing | Covered | `targeting show/set/clear/replace`, `targeting country ...`, `targeting device ...` |
| Reports | Partial | `reports campaigns/adgroups/keywords/searchterms/ads/impressionshare`, `reports template ...` preset generator (impression share account-dependent) |
| Apps metadata in ads context | Partial | `apps get/eligibilities/product-pages` (eligibility/product pages can be account/app dependent) |
| Search lookups | Covered | `search apps`, `search geo` |
| Creatives | Partial | `creatives list/get/find/create/update/delete` (fallback paths used; account/endpoint dependent) |
| Budget orders | Partial | `budget-orders list/get/create/update/find` |
| Account identity/ACL | Covered | `account me`, `account acls` |
| Any undocumented endpoint | Covered (raw) | `api --method ... --path ...` |

## Top Missing / Partial UI-Parity Items

1. Creative set management flows beyond list/get/find (depends on account capabilities and endpoint exposure).
2. Full recommendation workflows (accept/reject/apply) if exposed in specific accounts.
3. Extended reporting templates/presets and saved-report UX parity.
4. UI-only convenience around onboarding/billing/account management steps not exposed in Campaign Management API.
5. End-to-end “wizardized” campaign/adgroup/ad creation with guided defaults.

## Current Implementation Strategy

1. Prefer typed commands for high-frequency actions.
2. Keep `appleads api` as escape hatch for newly released or niche endpoints.
3. Add commands incrementally with live validation against the account to avoid dead commands.
