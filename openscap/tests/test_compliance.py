#!/usr/bin/env python3

# Copyright 2025 Marcus Furlong <furlongm@gmail.com>
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

import importlib.machinery
import importlib.util
import json
import os
import subprocess
import sys
import tempfile
import textwrap
import unittest
from unittest.mock import mock_open, patch
from xml.etree import ElementTree

# load patchman-openscap as a module (it has no .py extension)
_script = os.path.join(os.path.dirname(os.path.abspath(__file__)),
                       '..', 'patchman-openscap')
_loader = importlib.machinery.SourceFileLoader('patchman_compliance', _script)
_spec = importlib.util.spec_from_loader('patchman_compliance', _loader)
pc = importlib.util.module_from_spec(_spec)
_spec.loader.exec_module(pc)


class TestDetectOS(unittest.TestCase):
    """Test OS detection from /etc/os-release."""

    def _mock_os_release(self, content):
        return patch('builtins.open', mock_open(read_data=content))

    def test_rhel9(self):
        data = 'ID="rhel"\nVERSION_ID="9.3"\n'
        with self._mock_os_release(data):
            os_id, ver_major, ver_full = pc.detect_os()
        self.assertEqual(os_id, 'rhel')
        self.assertEqual(ver_major, '9')
        self.assertEqual(ver_full, '9.3')

    def test_ubuntu2204(self):
        data = 'ID=ubuntu\nVERSION_ID="22.04"\n'
        with self._mock_os_release(data):
            os_id, ver_major, ver_full = pc.detect_os()
        self.assertEqual(os_id, 'ubuntu')
        self.assertEqual(ver_major, '22')
        self.assertEqual(ver_full, '22.04')

    def test_debian12(self):
        data = 'ID=debian\nVERSION_ID="12"\n'
        with self._mock_os_release(data):
            os_id, ver_major, ver_full = pc.detect_os()
        self.assertEqual(os_id, 'debian')
        self.assertEqual(ver_major, '12')
        self.assertEqual(ver_full, '12')

    def test_rocky9(self):
        data = 'ID="rocky"\nVERSION_ID="9.4"\n'
        with self._mock_os_release(data):
            os_id, ver_major, ver_full = pc.detect_os()
        self.assertEqual(os_id, 'rocky')
        self.assertEqual(ver_major, '9')
        self.assertEqual(ver_full, '9.4')

    def test_almalinux9(self):
        data = 'ID="almalinux"\nVERSION_ID="9.4"\n'
        with self._mock_os_release(data):
            os_id, ver_major, ver_full = pc.detect_os()
        self.assertEqual(os_id, 'almalinux')
        self.assertEqual(ver_major, '9')
        self.assertEqual(ver_full, '9.4')

    def test_fedora(self):
        data = 'ID=fedora\nVERSION_ID="41"\n'
        with self._mock_os_release(data):
            os_id, ver_major, ver_full = pc.detect_os()
        self.assertEqual(os_id, 'fedora')
        self.assertEqual(ver_major, '41')
        self.assertEqual(ver_full, '41')

    def test_sles15(self):
        data = 'ID="sles"\nVERSION_ID="15.5"\n'
        with self._mock_os_release(data):
            os_id, ver_major, ver_full = pc.detect_os()
        self.assertEqual(os_id, 'sles')
        self.assertEqual(ver_major, '15')
        self.assertEqual(ver_full, '15.5')

    def test_missing_os_release(self):
        with patch('builtins.open', side_effect=FileNotFoundError):
            os_id, ver_major, ver_full = pc.detect_os()
        self.assertIsNone(os_id)
        self.assertIsNone(ver_major)
        self.assertIsNone(ver_full)


class TestResolveDatastream(unittest.TestCase):
    """Test OS-to-datastream mapping."""

    def test_rhel9(self):
        path, profile = pc.resolve_datastream('rhel', '9', '9.3', '/content')
        self.assertEqual(path, '/content/ssg-rhel9-ds.xml')
        self.assertIn('cis', profile)

    def test_rocky9_uses_rhel(self):
        path, profile = pc.resolve_datastream('rocky', '9', '9.4', '/content')
        self.assertEqual(path, '/content/ssg-rhel9-ds.xml')

    def test_almalinux9_uses_rhel(self):
        path, profile = pc.resolve_datastream('almalinux', '9', '9.4', '/content')
        self.assertEqual(path, '/content/ssg-rhel9-ds.xml')

    def test_ubuntu2204(self):
        path, profile = pc.resolve_datastream('ubuntu', '22', '22.04', '/content')
        self.assertEqual(path, '/content/ssg-ubuntu2204-ds.xml')

    def test_ubuntu2404(self):
        path, profile = pc.resolve_datastream('ubuntu', '24', '24.04', '/content')
        self.assertEqual(path, '/content/ssg-ubuntu2404-ds.xml')

    def test_debian12(self):
        path, profile = pc.resolve_datastream('debian', '12', '12', '/content')
        self.assertEqual(path, '/content/ssg-debian12-ds.xml')

    def test_fedora(self):
        path, profile = pc.resolve_datastream('fedora', '41', '41', '/content')
        self.assertEqual(path, '/content/ssg-fedora-ds.xml')

    def test_sles15(self):
        path, profile = pc.resolve_datastream('sles', '15', '15.5', '/content')
        self.assertEqual(path, '/content/ssg-sle15-ds.xml')

    def test_opensuse_uses_sle(self):
        path, profile = pc.resolve_datastream('opensuse-leap', '15', '15.6', '/content')
        self.assertEqual(path, '/content/ssg-sle15-ds.xml')

    def test_unsupported_os(self):
        path, profile = pc.resolve_datastream('arch', '', '', '/content')
        self.assertIsNone(path)
        self.assertIsNone(profile)


class TestParseResults(unittest.TestCase):
    """Test XCCDF results XML parsing."""

    SAMPLE_XML = textwrap.dedent(f'''\
        <?xml version="1.0" encoding="UTF-8"?>
        <Benchmark xmlns="{pc.XCCDF_NS}">
          <TestResult id="test1" start-time="2026-01-01T00:00:00" end-time="2026-01-01T00:05:00">
            <profile idref="xccdf_org.ssgproject.content_profile_cis"/>
            <target>server1.example.com</target>
            <score maximum="100">85.5</score>
            <rule-result idref="rule_1" severity="high">
              <result>pass</result>
            </rule-result>
            <rule-result idref="rule_2" severity="medium">
              <result>fail</result>
            </rule-result>
            <rule-result idref="rule_3" severity="low">
              <result>notapplicable</result>
            </rule-result>
            <rule-result idref="rule_4" severity="medium">
              <result>pass</result>
            </rule-result>
          </TestResult>
        </Benchmark>
    ''')

    def test_parse_sample(self):
        with tempfile.NamedTemporaryFile(mode='w', suffix='.xml', delete=False) as f:
            f.write(self.SAMPLE_XML)
            f.flush()
            try:
                result = pc.parse_results(f.name)
            finally:
                os.unlink(f.name)

        self.assertEqual(result['profile'], 'xccdf_org.ssgproject.content_profile_cis')
        self.assertEqual(result['target'], 'server1.example.com')
        self.assertAlmostEqual(result['score'], 85.5)
        self.assertEqual(result['score_maximum'], 100.0)
        self.assertEqual(result['summary']['pass'], 2)
        self.assertEqual(result['summary']['fail'], 1)
        self.assertEqual(result['summary']['notapplicable'], 1)
        self.assertEqual(len(result['rules']), 4)

    def test_parse_rule_details(self):
        with tempfile.NamedTemporaryFile(mode='w', suffix='.xml', delete=False) as f:
            f.write(self.SAMPLE_XML)
            f.flush()
            try:
                result = pc.parse_results(f.name)
            finally:
                os.unlink(f.name)

        rules_by_id = {r['id']: r for r in result['rules']}
        self.assertEqual(rules_by_id['rule_1']['severity'], 'high')
        self.assertEqual(rules_by_id['rule_1']['result'], 'pass')
        self.assertEqual(rules_by_id['rule_2']['severity'], 'medium')
        self.assertEqual(rules_by_id['rule_2']['result'], 'fail')

    def test_parse_no_test_result_raises(self):
        xml = f'<?xml version="1.0"?><Benchmark xmlns="{pc.XCCDF_NS}"></Benchmark>'
        with tempfile.NamedTemporaryFile(mode='w', suffix='.xml', delete=False) as f:
            f.write(xml)
            f.flush()
            try:
                with self.assertRaises(ValueError):
                    pc.parse_results(f.name)
            finally:
                os.unlink(f.name)


class TestBuildJsonOutput(unittest.TestCase):
    """Test JSON output structure."""

    def test_output_structure(self):
        scan_results = {
            'profile': 'cis_l1',
            'target': 'host1',
            'score': 90.0,
            'score_maximum': 100.0,
            'start_time': '2026-01-01T00:00:00',
            'end_time': '2026-01-01T00:05:00',
            'summary': {v: 0 for v in pc.RESULT_VALUES},
            'rules': [],
        }
        scan_results['summary']['pass'] = 10
        scan_results['summary']['fail'] = 2

        output = pc.build_json_output(scan_results, 'rhel', '9.3', '/content/ssg-rhel9-ds.xml')

        self.assertEqual(output['version'], pc.VERSION)
        self.assertEqual(output['os_id'], 'rhel')
        self.assertEqual(output['os_version'], '9.3')
        self.assertEqual(output['datastream'], 'ssg-rhel9-ds.xml')
        self.assertEqual(output['score'], 90.0)
        self.assertEqual(output['summary']['pass'], 10)
        self.assertEqual(output['summary']['fail'], 2)
        self.assertIn('scan_date', output)
        self.assertIn('hostname', output)

    def test_output_is_json_serializable(self):
        scan_results = {
            'profile': 'cis',
            'target': 'host1',
            'score': 85.0,
            'score_maximum': 100.0,
            'start_time': '',
            'end_time': '',
            'summary': {v: 0 for v in pc.RESULT_VALUES},
            'rules': [{'id': 'r1', 'severity': 'low', 'result': 'pass'}],
        }
        output = pc.build_json_output(scan_results, 'debian', '12', '/ds.xml')
        serialized = json.dumps(output)
        self.assertIsInstance(json.loads(serialized), dict)


class TestLoadConfig(unittest.TestCase):
    """Test config file loading."""

    def test_defaults_when_missing(self):
        config = pc.load_config('/nonexistent/path')
        self.assertEqual(config['profile'], '')
        self.assertEqual(config['content_dir'], pc.DEFAULT_CONTENT_DIR)
        self.assertEqual(config['results_dir'], pc.DEFAULT_RESULTS_DIR)

    def test_load_custom_config(self):
        conf = textwrap.dedent('''\
            [compliance]
            profile = xccdf_org.ssgproject.content_profile_cis_level2_server
            content_dir = /opt/scap/content
            results_dir = /opt/scap/results
        ''')
        with tempfile.NamedTemporaryFile(mode='w', suffix='.conf', delete=False) as f:
            f.write(conf)
            f.flush()
            try:
                config = pc.load_config(f.name)
            finally:
                os.unlink(f.name)

        self.assertEqual(config['profile'],
                         'xccdf_org.ssgproject.content_profile_cis_level2_server')
        self.assertEqual(config['content_dir'], '/opt/scap/content')
        self.assertEqual(config['results_dir'], '/opt/scap/results')

    def test_partial_config(self):
        conf = '[compliance]\nprofile = custom_profile\n'
        with tempfile.NamedTemporaryFile(mode='w', suffix='.conf', delete=False) as f:
            f.write(conf)
            f.flush()
            try:
                config = pc.load_config(f.name)
            finally:
                os.unlink(f.name)

        self.assertEqual(config['profile'], 'custom_profile')
        self.assertEqual(config['content_dir'], pc.DEFAULT_CONTENT_DIR)


class TestRunScan(unittest.TestCase):
    """Test oscap invocation."""

    @patch('subprocess.run')
    def test_success(self, mock_run):
        mock_run.return_value = type('Result', (), {
            'returncode': 2, 'stdout': '', 'stderr': '',
        })()
        success, retcode = pc.run_scan('/ds.xml', 'cis', '/tmp/results.xml')
        self.assertTrue(success)
        self.assertEqual(retcode, 2)
        mock_run.assert_called_once()

    @patch('subprocess.run')
    def test_all_pass(self, mock_run):
        mock_run.return_value = type('Result', (), {
            'returncode': 0, 'stdout': '', 'stderr': '',
        })()
        success, retcode = pc.run_scan('/ds.xml', 'cis', '/tmp/results.xml')
        self.assertTrue(success)
        self.assertEqual(retcode, 0)

    @patch('subprocess.run')
    def test_oscap_error(self, mock_run):
        mock_run.return_value = type('Result', (), {
            'returncode': 1, 'stdout': '', 'stderr': 'some error',
        })()
        success, retcode = pc.run_scan('/ds.xml', 'cis', '/tmp/results.xml')
        self.assertFalse(success)
        self.assertEqual(retcode, 1)

    @patch('subprocess.run', side_effect=FileNotFoundError)
    def test_oscap_not_found(self, mock_run):
        success, retcode = pc.run_scan('/ds.xml', 'cis', '/tmp/results.xml')
        self.assertFalse(success)
        self.assertEqual(retcode, 2)

    @patch('subprocess.run', side_effect=subprocess.TimeoutExpired('oscap', 3600))
    def test_timeout(self, mock_run):
        success, retcode = pc.run_scan('/ds.xml', 'cis', '/tmp/results.xml')
        self.assertFalse(success)
        self.assertEqual(retcode, 1)


if __name__ == '__main__':
    unittest.main()
