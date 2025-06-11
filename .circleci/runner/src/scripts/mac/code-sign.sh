#!/usr/bin/env bash
# Copyright 2022-2025 Salesforce, Inc.
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

set -euo pipefail

cd "${ARTIFACTS_DIR}"

for package in "${PROD_NAME}"*macOS_*.zip; do
    echo "Signing: ${package}"
    unzip "${package}"
    codesign --force --deep --verbose --verify --sign "Developer ID Application: SLACK TECHNOLOGIES L.L.C. (BQR82RBBHL)" --options runtime "${PROD_NAME}"
    codesign -vvv --deep --strict "${PROD_NAME}"
    zip -r "${package}" "${PROD_NAME}"
    rm "${PROD_NAME}"
    xcrun notarytool submit "${package}" -p "HERMES_NOTARY"
done
