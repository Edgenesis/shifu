# v0.106.1

## Release recovery

This patch replaces withdrawn v0.106.0. The v0.106.0 and v0.106.0-rc1 tags remain unchanged for traceability; their deployment image versions were inconsistent and release validation failed.

## Fixes

- Regenerate release-managed deployment, controller, telemetry, and demo image references consistently with the release version.
- Make retagging transactional so a dependency-download or controller-generation failure cannot leave a partially prepared release tree.
- Fail closed on release preparation and pull-request check errors, including failures inside Bash command substitutions.
- Add image-version consistency and release-control regression checks.
- Centralize release publication in the Hermes release manager after tag-specific CI and image-build verification; legacy merged-PR workflows validate candidates without publishing competing releases.

## Verification

Publication of this patch is gated on successful pull-request checks and the exact tag's required Azure build/test/image jobs and GitHub lint checks. Documentation is updated only after release validation.
