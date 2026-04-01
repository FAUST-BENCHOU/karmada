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

# Compare docs/command-flags (extract) vs docs/binary-command-flags (binary --help):
# strip extract-only Deprecated flags sections, normalize whitespace, diff semantics.
#
# Usage:
#   hack/compare-command-flags-dirs.sh [REF_DIR] [OTHER_DIR]
# Defaults: docs/command-flags docs/binary-command-flags

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
REF_DIR="${1:-${SCRIPT_ROOT}/docs/command-flags}"
OTHER_DIR="${2:-${SCRIPT_ROOT}/docs/binary-command-flags}"

exec python3 "${SCRIPT_ROOT}/hack/tools/compare-command-flags-dirs.py" "${REF_DIR}" "${OTHER_DIR}"
