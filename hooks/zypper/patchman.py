#!/usr/bin/env python
#
# Copyright 2016-2026 Marcus Furlong <furlongm@gmail.com>
#
# This file is part of Patchman.
#
# Patchman is free software: you can redistribute it and/or modify
# it under the terms of the GNU General Public License as published by
# the Free Software Foundation, version 3 only.
#
# Patchman is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU General Public License for more details.
#
# You should have received a copy of the GNU General Public License
# along with Patchman. If not, see <http://www.gnu.org/licenses/>
#
# zypp system plugin for patchman

import logging
import os
import subprocess

from zypp_plugin import Plugin


class PatchmanPlugin(Plugin):

    def PLUGINBEGIN(self, headers, body):  # noqa: N802
        logging.info('PLUGINBEGIN')
        self.ack()

    def PACKAGESETCHANGED(self, headers, body):  # noqa: N802
        logging.info('PACKAGESETCHANGED')
        servicecmd = '/usr/sbin/patchman-client'

        if not os.path.isfile(servicecmd) or not os.access(servicecmd, os.X_OK):
            self.ack()
            return

        print('Sending report to patchman server...')
        try:
            result = subprocess.run(
                [servicecmd, '--post-hook'],
                capture_output=True,
                text=True,
                timeout=60
            )
            if result.returncode != 0:
                if 'not configured' in result.stderr.lower():
                    print('patchman-client: Server not configured.')
                    print('  Edit /etc/patchman/patchman-client.conf and set server= to your patchman server.')
                else:
                    print(f'patchman-client: Failed to send report (exit {result.returncode})')
        except subprocess.TimeoutExpired:
            print('patchman-client: Timeout sending report')
        except Exception as e:
            print(f'patchman-client: Error - {e}')

        self.ack()

    def PLUGINEND(self, headers, body):  # noqa: N802
        logging.info('PLUGINEND')
        self.ack()


plugin = PatchmanPlugin()
plugin.main()
