#
# Copyright (c) 2026 Wind River Systems, Inc.
#
# SPDX-License-Identifier: Apache-2.0
#
"""Unit tests for constants module."""

import unittest

from k8sapp_ipsec_policy_operator.common import constants as app_constants


class TestIPsecConstants(unittest.TestCase):
    """Tests for common constants."""

    def test_helm_app_name(self):
        self.assertEqual(app_constants.HELM_APP_IPSEC_POLICY_OPERATOR,
                         'ipsec-policy-operator')

    def test_helm_chart_agent(self):
        self.assertEqual(app_constants.HELM_CHART_IPSEC_POLICY_AGENT,
                         'ipsec-policy-agent')

    def test_helm_chart_manager(self):
        self.assertEqual(app_constants.HELM_CHART_IPSEC_POLICY_MANAGER,
                         'ipsec-policy-manager')

    def test_app_labels_structure(self):
        self.assertIsInstance(app_constants.APP_LABELS, dict)
        self.assertIn('ipsec-policy-agent', app_constants.APP_LABELS)
        self.assertEqual(app_constants.APP_LABELS['ipsec-policy-agent'],
                         'ipsec-policy-agent-operator')

    def test_helm_namespace(self):
        self.assertEqual(app_constants.HELM_NS_IPSEC_POLICY_OPERATOR,
                         'ipsec-policy-operator')

    def test_chart_group(self):
        self.assertEqual(app_constants.CHART_GROUP_IPSEC_POLICY_AGENT,
                         'ipsec-policy-operator-charts')
