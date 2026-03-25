#!/usr/bin/env bash

release_notice() {
	echo "::notice::$*"
}

release_error() {
	echo "::error::$*" >&2
}

release_current_utc_date() {
	TZ=UTC date '+%F'
}

release_to_epoch() {
	local timestamp="$1"

	if date -d "${timestamp}" +%s >/dev/null 2>&1; then
		TZ=UTC date -d "${timestamp}" +%s
		return 0
	fi

	TZ=UTC date -jf '%Y-%m-%d %H:%M:%S' "${timestamp}" +%s
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

release_validate_stable_version() {
	local version="$1"

	if ! [[ "${version}" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
		release_error "Version ${version} is not stable semver (expected vX.Y.Z)"
		return 1
	fi
}

release_latest_stable_tag() {
	git fetch --tags --force >/dev/null 2>&1
	git tag --list 'v*' --sort=-v:refname | grep -E '^v[0-9]+[.][0-9]+[.][0-9]+$' | head -n 1
}

release_next_minor_version() {
	local latest_tag="$1"
	local version major minor patch

	release_validate_stable_version "${latest_tag}" || return 1

	version="${latest_tag#v}"
	IFS='.' read -r major minor patch <<< "${version}"

	echo "v${major}.$((minor + 1)).0"
}

release_release_branch_name() {
	local version="$1"

	release_validate_stable_version "${version}" || return 1
	echo "release_${version}"
}

release_rc_version() {
	local version="$1"

	release_validate_stable_version "${version}" || return 1
	echo "${version}-rc1"
}

release_rc_branch_name() {
	local version="$1"
	local rc_version

	rc_version=$(release_rc_version "${version}") || return 1
	echo "rc_${rc_version}"
}

release_official_branch_name() {
	local version="$1"

	release_validate_stable_version "${version}" || return 1
	echo "official_${version}"
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

release_version_from_release_branch() {
	local branch_name="$1"

	if ! [[ "${branch_name}" =~ ^release_(v[0-9]+\.[0-9]+\.[0-9]+)$ ]]; then
		release_error "Cannot parse version from branch ${branch_name}"
		return 1
	fi

	echo "${BASH_REMATCH[1]}"
}

release_previous_minor_baseline() {
	local version="$1"
	local version_num major minor patch

	release_validate_stable_version "${version}" || return 1

	version_num="${version#v}"
	IFS='.' read -r major minor patch <<< "${version_num}"

	if (( minor == 0 )); then
		if (( major == 0 )); then
			release_error "No previous version available for release notes (${version})"
			return 1
		fi

		echo "v$((major - 1)).0.0"
		return 0
	fi

	echo "v${major}.$((minor - 1)).0"
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

release_pr_review_decision() {
	local repo="$1"
	local pr_number="$2"

	gh pr view \
		"${pr_number}" \
		--repo "${repo}" \
		--json reviewDecision \
		--jq '.reviewDecision // "NONE"'
}

release_is_self_review_error() {
	local message="$1"

	[[ "${message}" == *"approve your own pull request"* ]] || [[ "${message}" == *"approval review cannot be from pull request author"* ]]
}

release_approve_pr_if_needed() {
	local repo="$1"
	local pr_number="$2"
	local review_decision approval_output status attempt

	review_decision=$(release_pr_review_decision "${repo}" "${pr_number}")
	if [[ "${review_decision}" != "REVIEW_REQUIRED" ]]; then
		return 0
	fi

	release_notice "Approving PR #${pr_number} to satisfy branch policy"
	set +e
	approval_output=$(gh pr review \
		"${pr_number}" \
		--repo "${repo}" \
		--approve \
		--body "Approved by release automation after all required checks passed." 2>&1)
	status=$?
	set -e

	if [[ "${status}" -ne 0 ]]; then
		if release_is_self_review_error "${approval_output}"; then
			release_error "PR #${pr_number} requires review, but release automation cannot approve its own pull request"
			return 1
		fi

		release_error "Failed to approve PR #${pr_number}: ${approval_output}"
		return 1
	fi

	for attempt in 1 2 3 4 5; do
		review_decision=$(release_pr_review_decision "${repo}" "${pr_number}")
		if [[ "${review_decision}" != "REVIEW_REQUIRED" ]]; then
			return 0
		fi

		if (( attempt < 5 )); then
			release_notice "Approval submitted for PR #${pr_number}; waiting for GitHub to reflect the review state"
			sleep 5
		fi
	done

	release_error "PR #${pr_number} still requires approval after release automation submitted a review"
	return 1
}

release_wait_for_pr_merge() {
	local repo="$1"
	local pr_number="$2"
	local timeout_seconds="${3:-14400}"
	local interval_seconds="${4:-30}"
	local deadline pr_state

	deadline=$(( $(date +%s) + timeout_seconds ))

	while (( $(date +%s) < deadline )); do
		pr_state=$(gh pr view \
			"${pr_number}" \
			--repo "${repo}" \
			--json state,mergedAt \
			--jq 'if .mergedAt then "MERGED" else .state end')

		case "${pr_state}" in
			MERGED)
				release_notice "PR #${pr_number} merged"
				return 0
				;;
			CLOSED)
				release_error "PR #${pr_number} closed before it was merged"
				return 1
				;;
		esac

		release_notice "Waiting for PR #${pr_number} to merge"
		sleep "${interval_seconds}"
	done

	release_error "Timed out waiting for PR #${pr_number} to merge"
	return 1
}

release_merge_pr_when_ready() {
	local repo="$1"
	local pr_number="$2"
	local timeout_seconds="${3:-14400}"
	local interval_seconds="${4:-30}"
	local merge_state head_sha max_behind_retries behind_count merge_deadline remaining_seconds merge_output status

	max_behind_retries=5
	behind_count=0
	merge_deadline=$(( $(date +%s) + timeout_seconds ))

	while true; do
		remaining_seconds=$(( merge_deadline - $(date +%s) ))
		if (( remaining_seconds <= 0 )); then
			release_error "Timed out waiting to merge PR #${pr_number}"
			return 1
		fi

		release_refresh_pr_branch_if_needed "${repo}" "${pr_number}"
		release_wait_for_green_checks "${repo}" "${pr_number}" "${remaining_seconds}" "${interval_seconds}"

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
				behind_count=$((behind_count + 1))
				if (( behind_count > max_behind_retries )); then
					release_error "PR #${pr_number} became behind its base branch too many times"
					return 1
				fi
				release_notice "PR #${pr_number} became behind its base after checks completed; refreshing again"
				continue
				;;
		esac

		if [[ "${merge_state}" == "BLOCKED" ]] && [[ "$(release_pr_review_decision "${repo}" "${pr_number}")" == "REVIEW_REQUIRED" ]]; then
			release_approve_pr_if_needed "${repo}" "${pr_number}"
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
					behind_count=$((behind_count + 1))
					if (( behind_count > max_behind_retries )); then
						release_error "PR #${pr_number} became behind its base branch too many times"
						return 1
					fi
					release_notice "PR #${pr_number} became behind its base after approval completed; refreshing again"
					continue
					;;
			esac
		fi

		head_sha=$(gh pr view "${pr_number}" --repo "${repo}" --json headRefOid --jq '.headRefOid')
		release_notice "Merging PR #${pr_number}"
		set +e
		merge_output=$(gh pr merge \
			"${pr_number}" \
			--repo "${repo}" \
			--squash \
			--delete-branch \
			--match-head-commit "${head_sha}" 2>&1)
		status=$?
		set -e

		if [[ "${status}" -eq 0 ]]; then
			return 0
		fi

		if [[ "${merge_output}" == *"base branch policy prohibits the merge"* ]] || [[ "${merge_output}" == *'add the `--auto` flag'* ]]; then
			release_notice "Direct merge blocked for PR #${pr_number}; enabling auto-merge"
			set +e
			merge_output=$(gh pr merge \
				"${pr_number}" \
				--repo "${repo}" \
				--auto \
				--squash \
				--delete-branch \
				--match-head-commit "${head_sha}" 2>&1)
			status=$?
			set -e

			if [[ "${status}" -ne 0 ]]; then
				release_error "Failed to enable auto-merge for PR #${pr_number}: ${merge_output}"
				return 1
			fi

			remaining_seconds=$(( merge_deadline - $(date +%s) ))
			if (( remaining_seconds <= 0 )); then
				release_error "Timed out waiting to merge PR #${pr_number}"
				return 1
			fi

			release_wait_for_pr_merge "${repo}" "${pr_number}" "${remaining_seconds}" "${interval_seconds}"
			return 0
		fi

		release_error "Failed to merge PR #${pr_number}: ${merge_output}"
		return 1
	done
}
