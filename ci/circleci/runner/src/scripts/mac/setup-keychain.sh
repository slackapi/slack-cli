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

if [ -f ${KEYCHAIN_FILE}-db ]; then
  security delete-keychain ${KEYCHAIN_FILE}
fi
security create-keychain -p $KEYCHAIN_PASSWORD ${KEYCHAIN_FILE}
security unlock-keychain -p $KEYCHAIN_PASSWORD $KEYCHAIN_FILE
security list-keychains -d user -s "${KEYCHAIN_FILE}" $(security list-keychains -d user | sed s/\"//g)
security set-keychain-settings "${KEYCHAIN_FILE}"
security import ${CERT_P12_FILE} -k ${KEYCHAIN_FILE} -P ${CERT_P12_PASSWORD} -T /usr/bin/pkgbuild -T /usr/bin/productbuild -T /usr/bin/codesign
# Programmatically derive the identity
CERT_IDENTITY=$(security find-identity -v -p codesigning "${KEYCHAIN_FILE}" | head -1 | grep '"' | sed -e 's/[^"]*"//' -e 's/".*//')
# Handy to have UUID (just in case)
CERT_UUID=$(security find-identity -v -p codesigning "${KEYCHAIN_FILE}" | head -1 | grep '"' | awk '{print $2}')
security set-key-partition-list -S apple-tool:,apple: -s -k $KEYCHAIN_PASSWORD -D "$CERT_IDENTITY" -t private ${KEYCHAIN_FILE}

xcrun notarytool store-credentials "HERMES_NOTARY" --apple-id "${APPSTORE_APPLEID}" --password "${APPSTORE_PWD}" --team-id "BQR82RBBHL" --keychain "${KEYCHAIN_FILE}-db"

# make it read only by owner
chmod 400 ${KEYCHAIN_FILE}-db
