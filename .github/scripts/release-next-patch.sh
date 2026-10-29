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
# Releases TARGET_SHA as the next patch version: tags it vX.Y.(Z+1) after the
# highest existing vX.Y.Z tag and pushes the tag. The tag push fires the
# Cloud Build release trigger (build, scan, BCID gate, publish of the container
# image) through GitHub's webhook. It does not fire GitHub Actions workflows,
# because pushes made with GITHUB_TOKEN never do, so the binary and docs
# workflows are started explicitly with workflow_dispatch, which is exempt from
# that rule.
#
# Env: GITHUB_REPOSITORY, GITHUB_TOKEN (contents: write, actions: write), TARGET_SHA.
set -euo pipefail
: "${GITHUB_REPOSITORY:?}" "${GITHUB_TOKEN:?}" "${TARGET_SHA:?}"

git fetch --quiet --tags origin

if git tag --points-at "${TARGET_SHA}" | grep -Eq '^v[0-9]+\.[0-9]+\.[0-9]+$'; then
  echo "${TARGET_SHA} already carries a release tag; nothing to do."
  exit 0
fi

LAST=$(git tag -l 'v*' | grep -E '^v[0-9]+\.[0-9]+\.[0-9]+$' | sort -V | tail -n 1 || true)
if [ -z "${LAST}" ]; then
  echo "ERROR: no vX.Y.Z tag found to bump from." >&2
  exit 1
fi
NEXT=$(printf '%s' "${LAST}" | awk -F. -v OFS=. '{$3 = $3 + 1; print}')
echo "Releasing ${TARGET_SHA} as ${NEXT} (previous release ${LAST})."

git -c user.name="github-actions[bot]" \
    -c user.email="41898282+github-actions[bot]@users.noreply.github.com" \
    tag -a "${NEXT}" "${TARGET_SHA}" -m "${NEXT}: automated dependency release"
git push "https://x-access-token:${GITHUB_TOKEN}@github.com/${GITHUB_REPOSITORY}.git" "refs/tags/${NEXT}"
echo "Pushed ${NEXT}. Cloud Build publishes the container image from this tag."

# Give the new ref a moment to be visible to the Actions API.
sleep 15
gh workflow run gorelease-action.yml --ref "${NEXT}" --repo "${GITHUB_REPOSITORY}"
gh workflow run gen-docs.yml --ref "${NEXT}" --repo "${GITHUB_REPOSITORY}"
echo "Started the binary release and docs workflows on ${NEXT}."
