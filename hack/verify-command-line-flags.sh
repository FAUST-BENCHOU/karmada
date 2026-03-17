#!/usr/bin/env bash
# Copyright 2024 The Karmada Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.


set -o errexit
set -o nounset
set -o pipefail

REPO_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
DIFFROOT="${REPO_ROOT}/docs/command-flags"
TMP_DIFFROOT="${REPO_ROOT}/_tmp/docs/command-flags"
_tmp="${REPO_ROOT}/_tmp"

cleanup() {
  rm -rf "${_tmp}"
}
trap "cleanup" EXIT SIGINT

cleanup

cd "${REPO_ROOT}"
mkdir -p "${TMP_DIFFROOT}"

# Generate flags documentation to temporary location
# Use LANG=C to ensure English output (kubectl i18n uses LANG/LC_MESSAGES for translations)
echo "Generating command-line flags documentation..."
go build -o "${REPO_ROOT}/_tmp/extract-flags" "${REPO_ROOT}/hack/tools/extract-flags/main.go"
LANG=C "${REPO_ROOT}/_tmp/extract-flags" "${TMP_DIFFROOT}"

ret=0
for file in "${DIFFROOT}"/*.txt; do
  if [[ ! -f "${file}" ]]; then
    continue
  fi
  filename=$(basename "${file}")
  tmpfile="${TMP_DIFFROOT}/${filename}"
  if [[ ! -f "${tmpfile}" ]]; then
    echo "Missing generated file: ${tmpfile}"
    ret=1
    continue
  fi
  echo "Diffing ${file} against freshly generated flags documentation"
  if ! diff -Naupr "${file}" "${tmpfile}"; then
    ret=1
  fi
done

# Check for new components: files in tmp but not in docs
for file in "${TMP_DIFFROOT}"/*.txt; do
  if [[ ! -f "${file}" ]]; then
    continue
  fi
  filename=$(basename "${file}")
  docfile="${DIFFROOT}/${filename}"
  if [[ ! -f "${docfile}" ]]; then
    echo "New component flags file not in docs: ${filename}"
    echo "Please run hack/update-command-line-flags.sh"
    ret=1
  fi
done

if [[ $ret -eq 0 ]]; then
  echo "Command-line flags documentation is up to date."
else
  echo "Command-line flags documentation is out of date. Please run hack/update-command-line-flags.sh"
  exit 1
fi
