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
OUTPUT_DIR="${REPO_ROOT}/docs/binary-command-flags"

COMPONENTS=(
	karmada-controller-manager
	karmada-scheduler
	karmada-search
	karmada-webhook
	karmada-aggregated-apiserver
	karmada-descheduler
	karmada-metrics-adapter
	karmada-scheduler-estimator
	karmadactl
)

# Subcommand paths after karmadactl (same coverage as extract-flags / command-flags headers).
# Includes completion, help, version (same idea as karmada-webhook Available Commands).
KARMADACTL_PATHS_FILE="${REPO_ROOT}/hack/karmadactl-binary-help-paths.txt"

echo "Building components..."
cd "${REPO_ROOT}"
for comp in "${COMPONENTS[@]}"; do
	echo "  Building ${comp}..."
	"${REPO_ROOT}/hack/build.sh" "${comp}"
done

echo "Generating binary --help output..."
mkdir -p "${OUTPUT_DIR}"
PLATFORM=$(go env GOHOSTOS)/$(go env GOHOSTARCH)
BIN_DIR="${REPO_ROOT}/_output/bin/${PLATFORM}"

for comp in "${COMPONENTS[@]}"; do
	bin="${BIN_DIR}/${comp}"
	if [[ ! -f "${bin}" ]]; then
		echo "  Skipping ${comp}: binary not found at ${bin}"
		continue
	fi
	if [[ "${comp}" == "karmadactl" ]]; then
		echo "  Writing karmadactl.txt (root + paths from ${KARMADACTL_PATHS_FILE})..."
		out="${OUTPUT_DIR}/karmadactl.txt"
		LANG=C "${bin}" --help > "${out}" 2>&1
		while IFS= read -r sub || [[ -n "${sub}" ]]; do
			[[ -z "${sub}" || "${sub}" =~ ^[[:space:]]*# ]] && continue
			read -r -a sub_args <<< "${sub}"
			{
				echo ""
				echo ""
				echo "=== karmadactl ${sub} ==="
				echo ""
				LANG=C "${bin}" "${sub_args[@]}" --help
			} >> "${out}" 2>&1 || true
		done < "${KARMADACTL_PATHS_FILE}"
		continue
	fi
	echo "  Writing ${comp}.txt..."
	LANG=C "${bin}" --help > "${OUTPUT_DIR}/${comp}.txt" 2>&1
done

echo "Binary command flags generated: ${OUTPUT_DIR}/"
