#
# Copyright (c) 2026 Wind River Systems, Inc.
#
# SPDX-License-Identifier: Apache-2.0
#
"""Unit tests for helm chart modules."""

import unittest

from k8sapp_ipsec_policy_operator.common import constants as app_constants
from k8sapp_ipsec_policy_operator.helm.ipsec_policy_agent import IPsecPolicyAgentHelm
from k8sapp_ipsec_policy_operator.helm.ipsec_policy_manager import IPsecPolicyManagerHelm
from k8sapp_ipsec_policy_operator.tests_coverage.conftest import InvalidHelmNamespace


class TestIPsecPolicyAgentHelm(unittest.TestCase):
    """Tests for IPsecPolicyAgentHelm class."""

    def setUp(self):
        self.helm_obj = object.__new__(IPsecPolicyAgentHelm)

    def test_chart_name(self):
        self.assertEqual(IPsecPolicyAgentHelm.CHART,
                         app_constants.HELM_CHART_IPSEC_POLICY_AGENT)

    def test_supported_namespaces_contains_operator_ns(self):
        self.assertIn(app_constants.HELM_NS_IPSEC_POLICY_OPERATOR,
                      IPsecPolicyAgentHelm.SUPPORTED_NAMESPACES)

    def test_supported_app_namespaces(self):
        self.assertIn(app_constants.HELM_APP_IPSEC_POLICY_OPERATOR,
                      IPsecPolicyAgentHelm.SUPPORTED_APP_NAMESPACES)

    def test_get_overrides_valid_namespace(self):
        ns = app_constants.HELM_NS_IPSEC_POLICY_OPERATOR
        result = self.helm_obj.get_overrides(namespace=ns)
        self.assertEqual(result, {})

    def test_get_overrides_no_namespace(self):
        result = self.helm_obj.get_overrides(namespace=None)
        self.assertIsInstance(result, dict)
        self.assertIn(app_constants.HELM_NS_IPSEC_POLICY_OPERATOR, result)

    def test_get_overrides_invalid_namespace(self):
        self.assertRaises(InvalidHelmNamespace,
                          self.helm_obj.get_overrides,
                          namespace='invalid-ns')


class TestIPsecPolicyManagerHelm(unittest.TestCase):
    """Tests for IPsecPolicyManagerHelm class."""

    def setUp(self):
        self.helm_obj = object.__new__(IPsecPolicyManagerHelm)

    def test_chart_name(self):
        self.assertEqual(IPsecPolicyManagerHelm.CHART,
                         app_constants.HELM_CHART_IPSEC_POLICY_MANAGER)

    def test_supported_namespaces_contains_operator_ns(self):
        self.assertIn(app_constants.HELM_NS_IPSEC_POLICY_OPERATOR,
                      IPsecPolicyManagerHelm.SUPPORTED_NAMESPACES)

    def test_supported_app_namespaces(self):
        self.assertIn(app_constants.HELM_APP_IPSEC_POLICY_OPERATOR,
                      IPsecPolicyManagerHelm.SUPPORTED_APP_NAMESPACES)

    def test_get_overrides_valid_namespace(self):
        ns = app_constants.HELM_NS_IPSEC_POLICY_OPERATOR
        result = self.helm_obj.get_overrides(namespace=ns)
        self.assertEqual(result, {})

    def test_get_overrides_no_namespace(self):
        result = self.helm_obj.get_overrides(namespace=None)
        self.assertIsInstance(result, dict)
        self.assertIn(app_constants.HELM_NS_IPSEC_POLICY_OPERATOR, result)

    def test_get_overrides_invalid_namespace(self):
        self.assertRaises(InvalidHelmNamespace,
                          self.helm_obj.get_overrides,
                          namespace='invalid-ns')
