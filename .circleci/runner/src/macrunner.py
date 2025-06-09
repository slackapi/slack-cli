#!/usr/bin/python3
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

import logging, json
import os, optparse, time

import boto3, github, glob, hashlib

import circlerunner

# file_name = '/tmp/circleci-workspace/artifacts/my_file.txt'
# f = open(file_name, 'w+')  # open file in write mode
# f.write('python rules')
# f.close()


class MacCircleRunner(circlerunner.CircleRunner):
  def __init__(self):
    logger = logging.getLogger(__name__)
    logging.basicConfig(level=logging.DEBUG, format='%(asctime)s %(levelname)-8s %(message)s', datefmt='%a, %d %b %Y %H:%M:%S', filename='runner.log', filemode='w')
    # dump log to stdout as well so it can be read from circle ci UI
    consoleHandler = logging.StreamHandler()
    logger.addHandler(consoleHandler)
    super().__init__(logger)

    self.secret_file_path = '/var/root/vault/mac-code-signing.ini'
    src_dir = os.path.abspath(os.path.dirname(__file__))
    self.keychain_script = os.path.join(src_dir, 'scripts/mac/setup-keychain.sh')
    self.codesign_script = os.path.join(src_dir, 'scripts/mac/code-sign.sh')
    self.teardown_script = os.path.join(src_dir, 'scripts/mac/teardown.sh')

  def teardown(self):
    super().teardown()
    if self.teardown_script:
      self.run_shell_script(self.teardown_script)

  def sign(self, sign_param):
    '''
    calling the signing script to load the keychain and code sign the built artifact
    '''
    self.logger.debug(f"sign: {sign_param}")
    self.load_config(self.secret_file_path, 'DEFAULT')
    # this should be defined in the config ini file
    # make sure the cert file is secure
    self.security_validation(os.getenv('CERT_P12_FILE'))
    self.run_shell_script(self.keychain_script)
    # make sure the keychain is secure
    self.security_validation('-'.join([os.getenv('KEYCHAIN_FILE'), 'db']))
    self.run_shell_script(script_file=self.codesign_script, env=sign_param, cwd=os.getcwd())

  def upload_file_to_s3(self, source_file_path, s3_path):
    bucket = 'slack-cli-binary-artifacts-prod'
    base_name = os.path.basename(source_file_path)
    self.load_config(self.secret_file_path, 'S3')
    s3 = boto3.client('s3')
    if os.path.isfile(source_file_path):
      s3_target_path = os.path.join(s3_path, base_name)
      response = s3.upload_file(Filename=source_file_path, Bucket=bucket,
        Key=s3_target_path, ExtraArgs={'ACL': 'public-read'})
    else:
      raise Exception(f"upload to s3: invalid file {source_file_path}")

  def do_the_job(self, job_name, job_params):
    if not job_name:
      raise Exception(f"undefined job name to execute")

    if job_name == 'MAC_CODE_SIGN':
      # check the owner and permission of this file itself
      self.security_validation(os.path.abspath(__file__))
      self.sign(sign_param=job_params)
    elif job_name == "S3_UPLOAD":
      artifact_dir = job_params.get("ARTIFACTS_DIR")
      s3_path = job_params.get("S3_TARGET_PATH")
      file_name = job_params.get('FILE_NAME')
      file_path = os.path.join(artifact_dir, file_name)
      # handle wildcard
      files = glob.glob(file_path)
      for local_path in files:
        self.upload_file_to_s3(source_file_path=local_path, s3_path=s3_path)
    else:
      raise Exception(f"unknown job to execute {job_name}")

if __name__ == "__main__":
  parser = optparse.OptionParser()
  options, args = parser.parse_args()
  # keep the json blob so the change of params won't
  # involve modifying the runner entrypoint script on the host
  # (entrypoint will blindly passing the jsonblob over)
  job_params = json.loads(args[0])
  # set the build url for notification
  job_name = job_params.get('JOB_NAME')
  try:
    ms = MacCircleRunner()
    ms.do_the_job(job_name, job_params)
  finally:
    ms.teardown()

