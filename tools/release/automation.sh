#!/usr/bin/env bash

release_notice() {
	echo "::notice::$*"
}

release_error() {
	echo "::error::$*" >&2
}

release_current_sg_date() {
	TZ=Asia/Singapore date '+%F'
}

release_to_epoch() {
	local timestamp="$1"

	if date -d "${timestamp}" +%s >/dev/null 2>&1; then
		TZ=Asia/Singapore date -d "${timestamp}" +%s
		return 0
	fi

	TZ=Asia/Singapore date -jf '%Y-%m-%d %H:%M:%S' "${timestamp}" +%s
}

release_is_biweekly_date() {
	local anchor_date="$1"
	local current_date="$2"
	local anchor_epoch current_epoch diff_days

	anchor_epoch=$(release_to_epoch "${anchor_date} 00:00:00")
	current_epoch=$(release_to_epoch "${current_date} 00:00:00")

	if (( current_epoch < anchor_epoch )); then
		echo "false"
		return 0
	fi

	diff_days=$(( (current_epoch - anchor_epoch) / 86400 ))
	if (( diff_days % 14 == 0 )); then
		echo "true"
		return 0
	fi

	echo "false"
}

release_latest_stable_tag() {
	git fetch --tags --force >/dev/null 2>&1
	git tag --list 'v*' --sort=-v:refname | grep -E '^v[0-9]+[.][0-9]+[.][0-9]+$' | head -n 1
}

release_next_minor_version() {
	local latest_tag="$1"
	local version major minor patch

	version="${latest_tag#v}"
	IFS='.' read -r major minor patch <<< "${version}"

	if ! [[ "${major}" =~ ^[0-9]+$ && "${minor}" =~ ^[0-9]+$ && "${patch}" =~ ^[0-9]+$ ]]; then
		release_error "Latest tag ${latest_tag} is not numeric semver"
		return 1
	fi

	echo "v${major}.$((minor + 1)).0"
}

release_latest_rc_tag() {
	git fetch --tags --force >/dev/null 2>&1
	git tag --list 'v*-rc*' --sort=-v:refname | grep -E '^v[0-9]+\.[0-9]+\.[0-9]+-rc[0-9]+$' | head -n 1
}

release_base_version_from_rc_tag() {
	local rc_tag="$1"

	if ! [[ "${rc_tag}" =~ ^(v[0-9]+\.[0-9]+\.[0-9]+)-rc[0-9]+$ ]]; then
		release_error "Cannot parse version from RC tag ${rc_tag}"
		return 1
	fi

	echo "${BASH_REMATCH[1]}"
}

release_wait_for_pr() {
	local repo="$1"
	local head_branch="$2"
	local base_branch="$3"
	local timeout_seconds="${4:-3600}"
	local interval_seconds="${5:-15}"
	local deadline pr_number

	deadline=$(( $(date +%s) + timeout_seconds ))

	while (( $(date +%s) < deadline )); do
		pr_number=$(gh pr list \
			--repo "${repo}" \
			--state open \
			--head "${head_branch}" \
			--base "${base_branch}" \
			--json number \
			--jq '.[0].number // empty')

		if [[ -n "${pr_number}" ]]; then
			echo "${pr_number}"
			return 0
		fi

		release_notice "Waiting for PR head=${head_branch} base=${base_branch}"
		sleep "${interval_seconds}"
	done

	release_error "Timed out waiting for PR head=${head_branch} base=${base_branch}"
	return 1
}

release_list_dependabot_prs() {
	local repo="$1"

	gh pr list \
		--repo "${repo}" \
		--author app/dependabot \
		--base main \
		--state open \
		--json number,createdAt,isDraft \
		--jq 'sort_by(.createdAt) | map(select(.isDraft | not)) | .[].number'
}

release_refresh_pr_branch_if_needed() {
	local repo="$1"
	local pr_number="$2"
	local pr_state

	pr_state=$(gh pr view \
		"${pr_number}" \
		--repo "${repo}" \
		--json isDraft,mergeStateStatus,mergeable \
		--jq 'if .isDraft then "DRAFT" else if .mergeable == "CONFLICTING" then "CONFLICTING" else .mergeStateStatus end end')

	case "${pr_state}" in
		DRAFT)
			release_error "PR #${pr_number} is still a draft"
			return 1
			;;
		DIRTY|CONFLICTING)
			release_error "PR #${pr_number} cannot be merged cleanly (${pr_state})"
			return 1
			;;
		BEHIND)
			release_notice "Updating PR #${pr_number} because it is behind the base branch"
			gh pr update-branch "${pr_number}" --repo "${repo}"
			;;
	esac
}

release_wait_for_green_checks() {
	local repo="$1"
	local pr_number="$2"
	local timeout_seconds="${3:-14400}"
	local interval_seconds="${4:-30}"
	local deadline saw_checks status checks_json

	deadline=$(( $(date +%s) + timeout_seconds ))
	saw_checks="false"

	while (( $(date +%s) < deadline )); do
		set +e
		checks_json=$(gh pr checks \
			"${pr_number}" \
			--repo "${repo}" \
			--json name,bucket,state,link 2>&1)
		status=$?
		set -e

		if [[ "${status}" -ne 0 && "${status}" -ne 8 ]]; then
			release_error "Failed to read checks for PR #${pr_number}: ${checks_json}"
			return 1
		fi

		if ! jq -e 'type == "array"' >/dev/null 2>&1 <<< "${checks_json}"; then
			release_error "Unexpected check payload for PR #${pr_number}: ${checks_json}"
			return 1
		fi

		if jq -e 'length == 0' >/dev/null 2>&1 <<< "${checks_json}"; then
			release_notice "PR #${pr_number} has no checks yet"
			sleep "${interval_seconds}"
			continue
		fi

		saw_checks="true"

		if jq -e 'map(select(.bucket == "fail" or .bucket == "cancel")) | length > 0' >/dev/null 2>&1 <<< "${checks_json}"; then
			release_error "PR #${pr_number} has failing checks"
			jq -r '.[] | select(.bucket == "fail" or .bucket == "cancel") | "- \(.name): \(.state) (\(.link // "no link"))"' <<< "${checks_json}" >&2
			return 1
		fi

		if jq -e 'all(.[]; .bucket == "pass" or .bucket == "skipping")' >/dev/null 2>&1 <<< "${checks_json}"; then
			release_notice "All checks passed for PR #${pr_number}"
			return 0
		fi

		release_notice "Waiting for checks on PR #${pr_number}"
		sleep "${interval_seconds}"
	done

	if [[ "${saw_checks}" != "true" ]]; then
		release_error "Timed out waiting for checks to start on PR #${pr_number}"
		return 1
	fi

	release_error "Timed out waiting for checks to pass on PR #${pr_number}"
	return 1
}

release_merge_pr_when_ready() {
	local repo="$1"
	local pr_number="$2"
	local timeout_seconds="${3:-14400}"
	local interval_seconds="${4:-30}"
	local merge_state head_sha

	while true; do
		release_refresh_pr_branch_if_needed "${repo}" "${pr_number}"
		release_wait_for_green_checks "${repo}" "${pr_number}" "${timeout_seconds}" "${interval_seconds}"

		merge_state=$(gh pr view \
			"${pr_number}" \
			--repo "${repo}" \
			--json isDraft,mergeStateStatus,mergeable \
			--jq 'if .isDraft then "DRAFT" else if .mergeable == "CONFLICTING" then "CONFLICTING" else .mergeStateStatus end end')

		case "${merge_state}" in
			DRAFT)
				release_error "PR #${pr_number} is still a draft"
				return 1
				;;
			DIRTY|CONFLICTING)
				release_error "PR #${pr_number} cannot be merged cleanly (${merge_state})"
				return 1
				;;
			BEHIND)
				release_notice "PR #${pr_number} became behind its base after checks completed; refreshing again"
				continue
				;;
		esac

		head_sha=$(gh pr view "${pr_number}" --repo "${repo}" --json headRefOid --jq '.headRefOid')
		release_notice "Merging PR #${pr_number}"
		gh pr merge \
			"${pr_number}" \
			--repo "${repo}" \
			--squash \
			--delete-branch \
			--match-head-commit "${head_sha}"
		return 0
	done
}
