#!/usr/bin/env bash
# Copyright 2026 The Karmada Authors.
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

# Verifies that docs/componentdocs and docs/karmadactldocs match freshly generated
# CLI flag and command reference output (hack/tools/genflagdocs). Run in CI when
# flags or cobra commands change.
#
# Usage:
#   hack/verify-cli-flag-docs.sh

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
COMPONENT_DOCS="${SCRIPT_ROOT}/docs/componentdocs"
KARMADACTL_DOCS="${SCRIPT_ROOT}/docs/karmadactldocs"
_tmp="${SCRIPT_ROOT}/_tmp/cli-flag-docs-verify"
TMP_COMPONENT="${_tmp}/componentdocs"
TMP_KARMADACTL="${_tmp}/karmadactldocs"

cleanup() {
  rm -rf "${_tmp}"
}
trap "cleanup" EXIT SIGINT

cleanup
mkdir -p "${TMP_COMPONENT}" "${TMP_KARMADACTL}" "${COMPONENT_DOCS}" "${KARMADACTL_DOCS}"

cp -a "${COMPONENT_DOCS}/." "${TMP_COMPONENT}/"
cp -a "${KARMADACTL_DOCS}/." "${TMP_KARMADACTL}/"

bash "${SCRIPT_ROOT}/hack/update-cli-flag-docs.sh"

ret=0
echo "diffing ${COMPONENT_DOCS} against freshly generated files"
diff -Naupr "${COMPONENT_DOCS}" "${TMP_COMPONENT}" || ret=1

echo "diffing ${KARMADACTL_DOCS} against freshly generated files"
diff -Naupr "${KARMADACTL_DOCS}" "${TMP_KARMADACTL}" || ret=1

rm -rf "${COMPONENT_DOCS}" "${KARMADACTL_DOCS}"
mkdir -p "${COMPONENT_DOCS}" "${KARMADACTL_DOCS}"
cp -a "${TMP_COMPONENT}/." "${COMPONENT_DOCS}/"
cp -a "${TMP_KARMADACTL}/." "${KARMADACTL_DOCS}/"

if [[ ${ret} -ne 0 ]]; then
  echo "CLI flag docs are out of date. Please run:"
  echo "  hack/update-cli-flag-docs.sh"
  exit 1
fi

echo "CLI flag docs (componentdocs and karmadactldocs) are up to date."
