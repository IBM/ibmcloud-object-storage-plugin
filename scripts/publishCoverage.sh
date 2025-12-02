#!/bin/bash
#/******************************************************************************
# Copyright 2021 IBM Corp.
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
# *****************************************************************************/
set -e

echo "Publishing coverage results..."

if [ -z "$GHE_TOKEN" ]; then
  echo "GHE_TOKEN not set → skipping coverage publish (normal in public repo)"
  exit 0
fi

REPO_SLUG="${GITHUB_REPOSITORY}"
BRANCH="${GITHUB_REF_NAME}"
COMMIT="${GITHUB_SHA}"
BUILD_NUMBER="${GITHUB_RUN_ID}"
PR_NUMBER="${{ github.event.pull_request.number }}"

TEMP_DIR=$(mktemp -d)
git clone -b gh-pages https://$GHE_TOKEN@github.com/$REPO_SLUG.git "$TEMP_DIR" || git clone https://$GHE_TOKEN@github.com/$REPO_SLUG.git "$TEMP_DIR" && cd "$TEMP_DIR" && git checkout -b gh-pages || true
cd "$TEMP_DIR"
git config user.name "github-actions"
git config user.email "github-actions@github.com"

mkdir -p "coverage/$BRANCH"
mkdir -p "coverage/$COMMIT"

cp "$GITHUB_WORKSPACE/cover.html" "coverage/$BRANCH/cover.html"
cp "$GITHUB_WORKSPACE/cover.html" "coverage/$COMMIT/cover.html"

OLD_COVERAGE=0
NEW_COVERAGE=$(cat coverage/$BRANCH/cover.html | grep "%)" | sed 's/[][()><%]/ /g' | awk '{ print $4 }' | awk '{s+=$1}END{print s/NR}')

GREEN_THRESHOLD=85
YELLOW_THRESHOLD=50
BADGE_COLOR="red"
if (( $(echo "$NEW_COVERAGE >= $GREEN_THRESHOLD" | bc -l) )); then
  BADGE_COLOR="brightgreen"
elif (( $(echo "$NEW_COVERAGE >= $YELLOW_THRESHOLD" | bc -l) )); then
  BADGE_COLOR="yellow"
fi

curl -s https://img.shields.io/badge/coverage-$NEW_COVERAGE%25-$BADGE_COLOR.svg > "coverage/$BRANCH/badge.svg"

git add .
git commit -m "Coverage update for $COMMIT (run $BUILD_NUMBER)"
git push origin gh-pages

echo "Coverage badge updated: https://$REPO_SLUG.github.io/coverage/$BRANCH/badge.svg"

if [ "$GITHUB_EVENT_NAME" = "pull_request" ] && [ -n "$PR_NUMBER" ]; then
  MESSAGE="Coverage: **$NEW_COVERAGE%**"
  curl -X POST \
    -H "Authorization: token $GHE_TOKEN" \
    -H "Content-Type: application/json" \
    "https://api.github.com/repos/$REPO_SLUG/issues/$PR_NUMBER/comments" \
    -d "{\"body\": \"$MESSAGE\"}"
fi