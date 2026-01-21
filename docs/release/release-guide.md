# Shifu Release Manager Guide

This document describes the automated release process for Shifu. The release cycle consists of a Wednesday RC (Release Candidate) release followed by a Monday official release.

## Release Schedule

| Day | Action | Type |
|-----|--------|------|
| Wednesday | RC Release | Semi-automated |
| Monday | Official Release | Semi-automated |

## Release Workflow Overview

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           WEDNESDAY - RC RELEASE                            │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  1. [MANUAL] Trigger "Prepare Release Changelog" workflow                   │
│         ↓                                                                   │
│  2. [AUTO] Changelog PR created (changelog-vX.Y.Z branch → main)            │
│         ↓                                                                   │
│  3. [MANUAL] Review changelog, approve and merge PR                         │
│         ↓                                                                   │
│  4. [AUTO] "Release RC" workflow triggers                                   │
│         - Creates release_vX.Y.Z branch                                     │
│         - Creates RC PR (rc_vX.Y.Z-rc1 → release_vX.Y.Z)                    │
│         ↓                                                                   │
│  5. [MANUAL] Review RC PR, approve and merge                                │
│         ↓                                                                   │
│  6. [AUTO] "Release RC Tag" workflow triggers                               │
│         - Creates vX.Y.Z-rc1 pre-release tag                                │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────────┐
│                          MONDAY - OFFICIAL RELEASE                          │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  7. [MANUAL] Trigger "Release Official" workflow                            │
│         ↓                                                                   │
│  8. [AUTO] Official PR created (official_vX.Y.Z → release_vX.Y.Z)           │
│         ↓                                                                   │
│  9. [MANUAL] Review official PR, approve and merge                          │
│         ↓                                                                   │
│  10. [AUTO] "Release Official Tag" workflow triggers                        │
│          - Creates vX.Y.Z release tag (marked as "Latest")                  │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Detailed Steps

### Wednesday: RC Release

#### Step 1: Generate Changelog

1. Go to **Actions** → **Prepare Release Changelog**
2. Click **Run workflow** → Select `main` branch → Click **Run workflow**
3. Wait for the workflow to complete (~2-3 minutes)

The workflow will:
- Detect the latest stable tag (e.g., `v0.88.0`)
- Calculate the next version (e.g., `v0.89.0`)
- Generate changelog using GitHub's release notes API
- Process changelog with AI enhancement
- Create a PR with the changelog file

Note: This workflow always bumps the MINOR version. For patch/hotfix releases, skip it and follow the Patch/Hotfix section below.

#### Step 2: Review and Merge Changelog PR

1. Navigate to the created PR titled **"chore: add changelog for vX.Y.Z"**
2. Review the changelog in `CHANGELOG/CHANGELOG-vX.Y.Z.md`
3. Verify:
   - All significant changes are captured
   - Categorization is correct (features, fixes, dependencies, etc.)
   - No sensitive information is exposed
4. Approve and merge the PR

#### Step 3: Review and Merge RC PR

After merging the changelog PR, the **Release RC** workflow automatically triggers.

1. Wait for the workflow to create the RC PR (~1-2 minutes)
2. Navigate to the created PR titled **"chore: release vX.Y.Z-rc1"**
3. Verify:
   - PR targets the correct `release_vX.Y.Z` branch
   - Version files are updated correctly
4. Approve and merge the PR

#### Step 4: Verify RC Pre-release

After merging the RC PR, the **Release RC Tag** workflow automatically creates the pre-release.

1. Go to **Releases** page
2. Verify `vX.Y.Z-rc1` appears as a **Pre-release**
3. Verify the release notes are generated correctly

### Monday: Official Release

#### Step 5: Create Official Release PR

1. Go to **Actions** → **Release Official**
2. Click **Run workflow** → Select `main` branch → Click **Run workflow**
3. Wait for the workflow to complete (~1-2 minutes)

The workflow will:
- Find the latest RC pre-release tag
- Create an official branch from the RC tag
- Update version files (remove `-rc1` suffix)
- Create a PR for the official release

#### Step 6: Review and Merge Official PR

1. Navigate to the created PR titled **"chore: release vX.Y.Z"**
2. Verify:
   - PR targets the correct `release_vX.Y.Z` branch
   - Version matches the RC that was tested
3. Approve and merge the PR

#### Step 7: Verify Official Release

After merging the official PR, the **Release Official Tag** workflow automatically creates the release.

1. Go to **Releases** page
2. Verify `vX.Y.Z` appears as the **Latest** release
3. Verify the release links to the changelog
4. Check the **Discussions** → **Announcements** for the release announcement

## Workflow Files Reference

| Workflow | File | Trigger |
|----------|------|---------|
| Prepare Release Changelog | `.github/workflows/release.yml` | Manual |
| Release RC | `.github/workflows/release-rc.yml` | Auto (changelog PR merge) or Manual |
| Release RC Tag | `.github/workflows/release-rc-tag.yml` | Auto (RC PR merge) |
| Release Official | `.github/workflows/release-official.yml` | Manual |
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

### Secrets Required

The changelog generation workflow requires these secrets:
- `AZURE_OPENAI_APIKEY` - Azure OpenAI API key for changelog enhancement
- `AZURE_OPENAI_HOST` - Azure OpenAI endpoint host
- `DEPLOYMENT_NAME` - Azure OpenAI deployment name

## Version Numbering

The release process follows semantic versioning:
- Version format: `vMAJOR.MINOR.PATCH`
- RC format: `vMAJOR.MINOR.PATCH-rc1`
- Each release cycle increments the MINOR version
- PATCH is reset to 0 for regular releases
- For hotfixes, manually specify the version with incremented PATCH and provide a manual changelog
