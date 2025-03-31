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


# Lint
bash -n install.sh

# Test that we can install the latest version at the default location.
rm -rf ~/.deno
rm -rf ~/.slack
brew uninstall deno

echo "Make sure deno is uninstalled"
deno --version

# Run install script with skip deno install flag
echo "Run install script with skip deno install flag"
bash ./install.sh -d slack-test
slack-test --version
deno --version

# Clean up
rm -rf ~/.slack
slack-test --version

# Run install script with skip deno install flag and version flag
echo "Run install script with skip deno install flag and version flag"
bash ./install.sh -v 2.2.0 -d slack-test
slack-test --version
deno --version

# Clean up
rm -rf ~/.slack
slack-test --version

echo "Run install script"
bash ./install.sh slack-test
slack-test --version
deno --version
