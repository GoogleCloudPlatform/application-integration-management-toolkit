#!/usr/bin/env bash
# Copyright 2026 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
# Runs on a Dependabot pull request: waits for the other checks, merges the PR
# once it is safe, then releases the merge commit as the next patch version
# (release-next-patch.sh).
#
# Safe means all of: opened by Dependabot from a dependabot/ branch; still open
# and mergeable; CI ran on its current head; no update in it changes a major
# version (digest updates of the pinned base image count as safe); and every
# check on the PR passed. Anything else is left for a human, with the
# reason in the job log.
#
# Env: GITHUB_REPOSITORY, GITHUB_TOKEN, PR_NUMBER, HEAD_SHA (the PR head this run
# is for), SELF_CHECK (this job's own check name, ignored while waiting),
# WAIT_MINUTES (default 30).
set -euo pipefail
: "${GITHUB_REPOSITORY:?}" "${GITHUB_TOKEN:?}" "${PR_NUMBER:?}" "${HEAD_SHA:?}" "${SELF_CHECK:?}"
WAIT_MINUTES="${WAIT_MINUTES:-30}"
skip() { echo "Not merging #${PR_NUMBER}: $*"; exit 0; }

PR_JSON=$(gh pr view "${PR_NUMBER}" --repo "${GITHUB_REPOSITORY}" \
  --json author,headRefName,state,mergeable,headRefOid,title,body)

read -r AUTHOR BRANCH STATE MERGEABLE HEAD_OID < <(printf '%s' "${PR_JSON}" | python3 -c '
import json, sys
p = json.load(sys.stdin)
print(p["author"]["login"], p["headRefName"], p["state"], p["mergeable"], p["headRefOid"])')
TITLE=$(printf '%s' "${PR_JSON}" | python3 -c 'import json, sys; print(json.load(sys.stdin)["title"])')

case "${AUTHOR}" in app/dependabot|dependabot\[bot\]|dependabot) ;; *) skip "author is ${AUTHOR}, not Dependabot" ;; esac
case "${BRANCH}" in dependabot/*) ;; *) skip "branch ${BRANCH} is not a dependabot/ branch" ;; esac
[ "${STATE}" = "OPEN" ] || skip "state is ${STATE}"
[ "${HEAD_OID}" = "${HEAD_SHA}" ] || skip "CI ran on ${HEAD_SHA} but the PR head is now ${HEAD_OID}"
[ "${MERGEABLE}" = "MERGEABLE" ] || skip "mergeable is ${MERGEABLE}"

# Dependabot writes every update as "from A to B" in the PR title (single
# update) or body (grouped). Allow it when no update changes the major version.
# Digest updates of the pinned base image have no version at all; they are
# allowed on docker branches only.
VERDICT=$(printf '%s' "${PR_JSON}" | BRANCH="${BRANCH}" python3 -c '
import json, os, re, sys
p = json.load(sys.stdin)
text = p["title"] + "\n" + (p.get("body") or "")
pairs = re.findall(r"from `?v?([0-9][\w.+-]*|[0-9a-f]{7,})`? to `?v?([0-9][\w.+-]*|[0-9a-f]{7,})`?", text)
if not pairs:
    print("skip: no \"from A to B\" update found in the PR text"); sys.exit()
def major(v):
    m = re.match(r"(\d+)\.(\d+)(?:\.(\d+))?", v)
    return m.group(1) if m else None
problems = []
for a, b in pairs:
    ma, mb = major(a), major(b)
    if ma is None or mb is None:
        if not os.environ["BRANCH"].startswith("dependabot/docker/"):
            problems.append(f"{a} -> {b} is not a version")
    elif ma != mb:
        problems.append(f"{a} -> {b} changes the major version")
print("ok" if not problems else "skip: " + "; ".join(problems))')
[ "${VERDICT}" = "ok" ] || skip "${VERDICT#skip: }"

# Wait for every other check on the PR. This job is itself one of them, so it
# is skipped by name. Any failure or a wait longer than WAIT_MINUTES leaves the
# PR for a human.
for ((i = 0; i < WAIT_MINUTES * 2; i++)); do
  SUMMARY=$(gh pr checks "${PR_NUMBER}" --repo "${GITHUB_REPOSITORY}" --json name,bucket | SELF_CHECK="${SELF_CHECK}" python3 -c '
import json, os, sys
checks = [c for c in json.load(sys.stdin) if c["name"] != os.environ["SELF_CHECK"]]
bad = [c["name"] + "=" + c["bucket"] for c in checks if c["bucket"] in ("fail", "cancel")]
pending = [c["name"] for c in checks if c["bucket"] == "pending"]
if bad: print("fail " + ", ".join(bad))
elif pending: print("pending " + ", ".join(pending))
elif not checks: print("pending no checks reported yet")
else: print("ok")')
  case "${SUMMARY}" in
    ok) break ;;
    fail*) skip "checks not green: ${SUMMARY#fail }" ;;
    *) echo "Waiting for: ${SUMMARY#pending }"; sleep 30 ;;
  esac
done
[ "${SUMMARY}" = "ok" ] || skip "checks still pending after ${WAIT_MINUTES} minutes: ${SUMMARY#pending }"

# CI may have run on an older head if Dependabot rebased meanwhile.
HEAD_NOW=$(gh pr view "${PR_NUMBER}" --repo "${GITHUB_REPOSITORY}" --json headRefOid --jq .headRefOid)
[ "${HEAD_NOW}" = "${HEAD_SHA}" ] || skip "the PR head moved from ${HEAD_SHA} to ${HEAD_NOW} while waiting"

echo "Merging #${PR_NUMBER}: ${TITLE}"
gh pr merge "${PR_NUMBER}" --repo "${GITHUB_REPOSITORY}" --squash --delete-branch \
  --subject "${TITLE} (#${PR_NUMBER})" \
  --body "Merged automatically: Dependabot update with all checks green. The next patch release follows."

MERGE_SHA=$(gh pr view "${PR_NUMBER}" --repo "${GITHUB_REPOSITORY}" --json mergeCommit --jq '.mergeCommit.oid')
[ -n "${MERGE_SHA}" ] || { echo "ERROR: merged but no merge commit reported" >&2; exit 1; }
echo "Merged as ${MERGE_SHA}."

TARGET_SHA="${MERGE_SHA}" "$(dirname "$0")/release-next-patch.sh"
