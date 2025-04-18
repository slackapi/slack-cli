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

import abc, configparser
import logging, os, pwd
import subprocess

class CircleRunner(abc.ABC):
  def __init__(self, logger=None):
    if not logger:
      self.logger = logging.getLogger(__name__)
    else:
      self.logger = logger

  def load_config(self, config_file, title=None):
    '''Load necessary secrets from the config file to environment for running the scripts.

    Keyword arguments:
    config_file -- full path of the config file
      the config file is expected to be in ini format (https://en.wikipedia.org/wiki/INI_file)
    title -- title of the config section within the config file (default: DEFAULT)
    '''
    self.security_validation(config_file)
    config = configparser.ConfigParser()
    # preserve case when parsing the config
    config.optionxform=str
    config.read(config_file)
    if not title:
      title = 'DEFAULT'
    for key in config[title]:
      os.environ[key]=config[title][key]

  def run_shell_script(self, script_file, args=None, env=None, cwd=None):
    '''Running a shell script, raise exception if the return code is not 0

    Keyword arguments:
    script_file -- path of the script file to execute
    args -- list of the parameters to pass to the script
    env -- additional environment vars (dictionary)
    cwd -- running directory
    '''
    self.security_validation(script_file)
    cmd = [script_file]
    if isinstance(args, list):
      cmd = cmd + args
    running_env = os.environ.copy()
    self.logger.debug(f"additional running env: {env}")
    if env:
      for key in env:
        running_env[key.upper()] = env[key]
    try:
      result = subprocess.run(cmd, check=True, capture_output=True, env=running_env, text=True, cwd=cwd)
    except Exception as e:
      self.logger.debug(e.stdout)
      self.logger.debug(e.stderr)
      raise e
    self.logger.debug(result.stdout)
    # only log the stderr, don't raise exception
    if result.stderr:
      self.logger.error(f"error executing {script_file}: {result.stderr}")
    # only raise exception if the returncode is 0
    if result.returncode != 0:
      raise Exception(f"command failed: {cmd} with return code {result.returncode}")

  def _check_file_permission(self, file_path, expected_permission='00'):
    '''check if a file has expected permission

    Keyword arguments:
    file_path -- path of the file to check permission of
    expected_permission -- target permission in CHMOD Permissions format,
      default 00 (not accessible by group or others)

    Return:
      False if the file has permission other than expected_permission
      True otherwise
    '''
    file_stat = os.stat(file_path)
    file_perm = oct(file_stat.st_mode)
    self.logger.debug(f"file {file_path} with perm {file_perm}")
    if not file_perm.endswith(expected_permission):
      self.logger.error(f"file permission error: expected to end with {expected_permission}: {file_perm}")
      return False
    return True

  def _check_file_owner(self, file_path, owner):
    '''check if a file has expected owner

    Keyword arguments:
    file_path -- path of the file to check permission of
    owner -- target file owner username

    Return:
      False if the file has owner other than expected owner
      True otherwise
    '''
    file_owner = pwd.getpwuid(os.stat(file_path).st_uid).pw_name
    self.logger.debug(f"file {file_path} with owner {file_owner}")
    if (file_owner != owner):
      self.logger.error(f"script file is not owned by root: {file_owner}")
      return False
    return True

  def security_validation(self, file_path):
    """validate file security for runner execution, besides file existence,
    the file needs to pass the following validation:
    1. it needs to be owned by root
    2. it needs to be only readable (and optionally executable)
      by owner(root) (mode 500 or 400)
    """
    self.logger.debug(f"validating script file {file_path}")
    valid = True
    if not file_path or not os.path.isfile(file_path):
      self.logger.error(f"script file doesn't exist {file_path}")
      valid = False
    else:
      valid = self._check_file_owner(file_path=file_path, owner='root') and \
        self._check_file_permission(file_path=file_path, expected_permission='00')
    if not valid:
      raise Exception(f"permission/ownership error: validation failed on {file_path}")

  def teardown(self):
    pass
