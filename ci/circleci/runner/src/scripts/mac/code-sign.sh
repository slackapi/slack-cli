#!/bin/bash
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

APP_ZIP_PATTERN="${PROD_NAME}*macOS_64-bit.zip"

cd ${ARTIFACTS_DIR} && APP_ZIP=$(find . -type f -iname "${APP_ZIP_PATTERN}");

echo $APP_ZIP

APP_ZIP_NAME=${APP_ZIP:2}

echo "$APP_ZIP_NAME"

unzip $APP_ZIP_NAME

rm $APP_ZIP_NAME

ls -l ${ARTIFACTS_DIR}

codesign --force --deep --verbose --verify --sign "Developer ID Application: SLACK TECHNOLOGIES L.L.C. (BQR82RBBHL)" --options runtime "${PROD_NAME}"

codesign -vvv --deep --strict "${PROD_NAME}"

zip -r "${APP_ZIP_NAME}" "${PROD_NAME}"

xcrun notarytool submit "${APP_ZIP_NAME}" -p "HERMES_NOTARY"



