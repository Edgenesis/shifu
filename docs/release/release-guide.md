# Shifu Release Manager Guide

This document describes the automated biweekly release process for Shifu. The release cycle consists of a Wednesday RC (Release Candidate) release followed by the official release on the following Monday.

## Release Schedule

| Day | Action | Type |
|-----|--------|------|
| Wednesday | RC Release | Automated on a biweekly cadence |
| Monday | Official Release | Automated on the following Monday |

The first scheduled automated Wednesday run is **March 25, 2026 at 05:00 UTC**, and the first scheduled automated Monday run is **March 30, 2026 at 05:00 UTC**. Both schedules repeat every 14 days. `05:00 UTC` is the same release moment as `13:00 SGT`. The underlying release workflows remain manually runnable for hotfixes and off-cycle releases.

## Release Workflow Overview

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           WEDNESDAY - RC RELEASE                            │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  1. [AUTO] Merge open Dependabot PRs to main sequentially                   │
│         ↓                                                                   │
│  2. [AUTO] Trigger "Prepare Release Changelog" workflow                      │
│         ↓                                                                   │
│  3. [AUTO] Changelog PR created and auto-merged after green checks          │
│         ↓                                                                   │
│  4. [AUTO] "Release RC" workflow triggers                                   │
│         - Creates release_vX.Y.Z branch                                     │
│         - Creates RC PR (rc_vX.Y.Z-rc1 → release_vX.Y.Z)                    │
│         ↓                                                                   │
│  5. [AUTO] RC PR auto-merged after green checks                             │
│         ↓                                                                   │
│  6. [AUTO] "Release RC Tag" workflow triggers                               │
│         - Creates vX.Y.Z-rc1 pre-release tag                                │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────────┐
│                          MONDAY - OFFICIAL RELEASE                          │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  7. [AUTO] Trigger "Release Official" workflow                              │
│         ↓                                                                   │
│  8. [AUTO] Official PR created (official_vX.Y.Z → release_vX.Y.Z)           │
│         ↓                                                                   │
│  9. [AUTO] Official PR auto-merged after green checks                       │
│         ↓                                                                   │
│  10. [AUTO] "Release Official Tag" workflow triggers                        │
│          - Creates vX.Y.Z release tag (marked as "Latest")                  │
│          - Dispatches shifu.dev version bump workflow                       │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Detailed Steps

### Wednesday: RC Release

#### Automated Wednesday sequence

1. The **Biweekly Release Wednesday** workflow checks whether the current UTC date is on the 14-day cadence anchored to March 25, 2026.
2. If the cadence matches, the workflow merges all open Dependabot PRs targeting `main`, oldest first. Each PR must have a clean merge state and fully passing checks before it is merged.
3. After the Dependabot queue is empty, the workflow triggers **Prepare Release Changelog** to generate the next changelog PR.
4. The changelog PR is merged automatically after its checks pass.
5. Merging the changelog PR triggers **Release RC**, which creates the `release_vX.Y.Z` branch and the RC PR.
6. The RC PR is merged automatically after its checks pass.
7. Merging the RC PR triggers **Release RC Tag**, which creates the `vX.Y.Z-rc1` pre-release.

#### Verify RC Pre-release

1. Go to **Releases** page
2. Verify `vX.Y.Z-rc1` appears as a **Pre-release**
3. Verify the release notes are generated correctly

### Monday: Official Release

#### Automated Monday sequence

1. The **Biweekly Release Monday** workflow checks whether the current UTC date is on the 14-day cadence anchored to March 30, 2026.
2. If the cadence matches, it triggers **Release Official**.
3. **Release Official** finds the latest RC pre-release tag, creates the official release PR, and returns the PR details to the orchestrator.
4. The official PR is merged automatically after its checks pass.
5. Merging the official PR triggers **Release Official Tag**, which creates the stable `vX.Y.Z` release and dispatches `Edgenesis/shifu.dev` to open its version-bump PR.

#### Verify Official Release

1. Go to **Releases** page
2. Verify `vX.Y.Z` appears as the **Latest** release
3. Verify the release links to the changelog
4. Check the **Discussions** → **Announcements** for the release announcement
5. Check `Edgenesis/shifu.dev` for a new PR titled **"Bump Shifu references to vX.Y.Z"**

## Workflow Files Reference

| Workflow | File | Trigger |
|----------|------|---------|
| Biweekly Release Wednesday | `.github/workflows/release-biweekly-wednesday.yml` | Scheduled weekly; runs only on the 14-day Wednesday cadence |
| Prepare Release Changelog | `.github/workflows/release.yml` | Manual or reusable from Wednesday automation |
| Release RC | `.github/workflows/release-rc.yml` | Auto (changelog PR merge) or Manual |
| Release RC Tag | `.github/workflows/release-rc-tag.yml` | Auto (RC PR merge) |
| Biweekly Release Monday | `.github/workflows/release-biweekly-monday.yml` | Scheduled weekly; runs only on the 14-day Monday cadence |
| Release Official | `.github/workflows/release-official.yml` | Manual or reusable from Monday automation |
| Release Official Tag | `.github/workflows/release-official-tag.yml` | Auto (official PR merge) |

## Branch Naming Convention

| Branch | Purpose |
|--------|---------|
| `main` | Development branch |
| `changelog-vX.Y.Z` | Changelog PR branch |
| `release_vX.Y.Z` | Release branch (long-lived) |
| `rc_vX.Y.Z-rc1` | RC PR branch |
| `official_vX.Y.Z` | Official release PR branch |

## Troubleshooting

### Workflow failed: "No stable semver tags found"

The repository has no existing release tags. Create an initial tag manually:
```bash
git tag v0.1.0
git push origin v0.1.0
```

### Workflow failed: "Branch already exists"

A previous release attempt left branches behind. Delete the conflicting temporary branches:
```bash
git push origin --delete changelog-vX.Y.Z
git push origin --delete rc_vX.Y.Z-rc1
# or
git push origin --delete official_vX.Y.Z
```
Keep `release_vX.Y.Z` unless you are intentionally restarting the release and no hotfixes are planned.

### Workflow failed: "Tag already exists"

The release tag already exists. This usually means the release was already completed. Check the Releases page. If you need to re-release, delete the existing tag first:
```bash
git push origin --delete vX.Y.Z
# Also delete the GitHub release via the web UI
```

### Workflow failed: "No RC tags found"

Run the RC workflow first before attempting the official release.

### Workflow failed: Discussion category error

Ensure the repository has **Discussions** enabled with an **"Announcements"** category:
1. Go to repository **Settings** → **General** → **Features**
2. Enable **Discussions**
3. Create an "Announcements" category if it doesn't exist

### Manual workflow trigger (version override)

If you need to release a specific version (e.g., hotfix), use the **Release RC** workflow with version override:
1. Go to **Actions** → **Release RC**
2. Click **Run workflow**
3. Enter version in **Override version** field (e.g., `v0.88.1`)
4. Click **Run workflow**

### Dry-run the biweekly orchestrators

Both orchestrators support a `dry_run` input for validation without creating or merging PRs:
1. Go to **Actions** → **Biweekly Release Wednesday** or **Biweekly Release Monday**
2. Click **Run workflow**
3. Set **dry_run** to `true`
4. Review the workflow log for cadence status and the planned release targets

The automated merge helpers are intentionally patient with CI:
- they wait up to 1 hour for the expected changelog or release PR to appear
- they wait up to 4 hours for PR checks to finish before failing the run

### Patch/Hotfix releases (manual changelog)

Patch releases are manual because the changelog workflow always bumps MINOR. For a patch/hotfix:
1. Create `CHANGELOG/CHANGELOG-vX.Y.Z.md` manually (use the latest changelog format).
2. Open a PR to `main` with the changelog and merge it.
3. Run **Release RC** with the version override set to the patch version.

## Post-Release Cleanup (Optional)

After a successful release cycle, you may clean up temporary branches:

```bash
# Delete changelog branch
git push origin --delete changelog-vX.Y.Z

# Delete RC branch
git push origin --delete rc_vX.Y.Z-rc1

# Delete official branch
git push origin --delete official_vX.Y.Z
```

Note: The `release_vX.Y.Z` branch should be kept for potential hotfixes.

## Prerequisites

### Repository Settings

1. **Discussions** enabled with "Announcements" category
2. **Actions** enabled with write permissions for `GITHUB_TOKEN`
3. **Branch protection** rules should allow the workflows to create branches and PRs
4. `shifu-release-bot` should be added as a PR bypass actor for both `main` and `release*`

### Secrets Required

The changelog generation workflow requires these secrets:
- `AZURE_OPENAI_APIKEY` - Azure OpenAI API key for changelog enhancement
- `AZURE_OPENAI_HOST` - Azure OpenAI endpoint host
- `DEPLOYMENT_NAME` - Azure OpenAI deployment name
- `SHIFU_DEV_DISPATCH_TOKEN` - token with permission to dispatch the version-bump workflow in `Edgenesis/shifu.dev`

## Version Numbering

The release process follows semantic versioning:
- Version format: `vMAJOR.MINOR.PATCH`
- RC format: `vMAJOR.MINOR.PATCH-rc1`
- Each release cycle increments the MINOR version
- PATCH is reset to 0 for regular releases
- For hotfixes, manually specify the version with incremented PATCH and provide a manual changelog
