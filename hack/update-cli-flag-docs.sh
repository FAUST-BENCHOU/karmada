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

# Regenerates docs/componentdocs and docs/karmadactldocs from the in-tree cobra
# commands using hack/tools/genflagdocs.
#
# Usage:
#   hack/update-cli-flag-docs.sh

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
COMPONENT_DOCS="${SCRIPT_ROOT}/docs/componentdocs"
KARMADACTL_DOCS="${SCRIPT_ROOT}/docs/karmadactldocs"

cd "${SCRIPT_ROOT}"

gen() {
  go run ./hack/tools/genflagdocs "$@"
}

mkdir -p "${COMPONENT_DOCS}" "${KARMADACTL_DOCS}"

gen "${COMPONENT_DOCS}" karmada-controller-manager
gen "${COMPONENT_DOCS}" karmada-scheduler
gen "${COMPONENT_DOCS}" karmada-agent
gen "${COMPONENT_DOCS}" karmada-aggregated-apiserver
gen "${COMPONENT_DOCS}" karmada-descheduler
gen "${COMPONENT_DOCS}" karmada-search
gen "${COMPONENT_DOCS}" karmada-scheduler-estimator
gen "${COMPONENT_DOCS}" karmada-webhook
gen "${COMPONENT_DOCS}" karmada-metrics-adapter

gen "${KARMADACTL_DOCS}" karmadactl
