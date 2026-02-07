# Copyright 2013-2026 Marcus Furlong <furlongm@gmail.com>
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

import os
import subprocess

from yum.plugins import TYPE_CORE

requires_api_version = '2.1'
plugin_type = (TYPE_CORE,)


def posttrans_hook(conduit):
    servicecmd = conduit.confString('main', 'servicecmd', '/usr/sbin/patchman-client')

    if not os.path.isfile(servicecmd) or not os.access(servicecmd, os.X_OK):
        return

    conduit.info(2, 'Sending report to patchman server...')
    try:
        result = subprocess.run(
            [servicecmd, '--post-hook'],
            capture_output=True,
            text=True,
            timeout=60
        )
        if result.returncode != 0:
            if 'not configured' in result.stderr.lower():
                conduit.info(2, 'patchman-client: Server not configured.')
                conduit.info(2, '  Edit /etc/patchman/patchman-client.conf and set server= to your patchman server.')
            else:
                conduit.info(2, f'patchman-client: Failed to send report (exit {result.returncode})')
    except subprocess.TimeoutExpired:
        conduit.info(2, 'patchman-client: Timeout sending report')
    except Exception as e:
        conduit.info(2, f'patchman-client: Error - {e}')
