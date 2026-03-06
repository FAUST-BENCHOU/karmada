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

SCRIPT_ROOT=$(dirname "${BASH_SOURCE[0]}")/..

DIFFROOT="${SCRIPT_ROOT}/doc/command-line-flags.txt"
TMP_DIFFROOT="${SCRIPT_ROOT}/_tmp/command-line-flags.txt"
_tmp="${SCRIPT_ROOT}/_tmp"

cleanup() {
  rm -rf "${_tmp}"
}
trap "cleanup" EXIT SIGINT

cleanup

mkdir -p "${SCRIPT_ROOT}/_tmp"

# Generate flags documentation to temporary location
echo "Generating command-line flags documentation..."
go build -o "${SCRIPT_ROOT}/_tmp/extract-flags" "${SCRIPT_ROOT}/hack/tools/extract-flags/main.go"
"${SCRIPT_ROOT}/_tmp/extract-flags" > "${TMP_DIFFROOT}"

echo "diffing ${DIFFROOT} against freshly generated flags documentation"
ret=0
diff -Naupr "${DIFFROOT}" "${TMP_DIFFROOT}" || ret=$?
if [[ $ret -eq 0 ]]
then
  echo "${DIFFROOT} is up to date."
else
  echo "${DIFFROOT} is out of date. Please run hack/generate-command-line-flags.sh"
  echo ""
  echo "Diff:"
  diff -Naupr "${DIFFROOT}" "${TMP_DIFFROOT}" || true
  exit 1
fi
