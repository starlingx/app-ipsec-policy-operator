#
# Copyright (c) 2026 Wind River Systems, Inc.
#
# SPDX-License-Identifier: Apache-2.0
#
"""Unit tests for helm chart modules."""

import unittest
from unittest import mock

from k8sapp_ipsec_policy_operator.common import constants as app_constants
from k8sapp_ipsec_policy_operator.helm.ipsec_policy_agent import IPsecPolicyAgentHelm
from k8sapp_ipsec_policy_operator.helm.ipsec_policy_manager import IPsecPolicyManagerHelm
from k8sapp_ipsec_policy_operator.tests_coverage.conftest import HelmOverrideNotFound
from k8sapp_ipsec_policy_operator.tests_coverage.conftest import InvalidHelmNamespace

LABEL_KEY = app_constants.HELM_CHART_COMPONENT_LABEL


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
        self.assertEqual(result, {
            'controllerManager': {
                'podLabels': {LABEL_KEY: 'platform'},
            },
        })

    def test_get_overrides_no_namespace(self):
        result = self.helm_obj.get_overrides(namespace=None)
        self.assertIsInstance(result, dict)
        self.assertIn(app_constants.HELM_NS_IPSEC_POLICY_OPERATOR, result)

    def test_get_overrides_invalid_namespace(self):
        self.assertRaises(InvalidHelmNamespace,
                          self.helm_obj.get_overrides,
                          namespace='invalid-ns')

    @mock.patch('k8sapp_ipsec_policy_operator.helm.ipsec_policy_agent.dbapi')
    def test_get_overrides_uses_valid_user_override(self, mock_dbapi):
        mock_dbapi.get_instance.return_value.kube_app_get.return_value = \
            mock.MagicMock(id=1)
        mock_dbapi.get_instance.return_value.helm_override_get.return_value = {
            'user_overrides': {LABEL_KEY: 'application'},
        }
        result = self.helm_obj.get_overrides(
            namespace=app_constants.HELM_NS_IPSEC_POLICY_OPERATOR)
        self.assertEqual(
            result['controllerManager']['podLabels'],
            {LABEL_KEY: 'application'})

    @mock.patch('k8sapp_ipsec_policy_operator.helm.ipsec_policy_agent.dbapi')
    def test_get_overrides_ignores_invalid_user_override(self, mock_dbapi):
        mock_dbapi.get_instance.return_value.kube_app_get.return_value = \
            mock.MagicMock(id=1)
        mock_dbapi.get_instance.return_value.helm_override_get.return_value = {
            'user_overrides': {LABEL_KEY: 'not-a-real-component'},
        }
        result = self.helm_obj.get_overrides(
            namespace=app_constants.HELM_NS_IPSEC_POLICY_OPERATOR)
        # Falls back to the default/current label when the override value
        # isn't one of the supported component labels.
        self.assertEqual(
            result['controllerManager']['podLabels'],
            {LABEL_KEY: 'platform'})

    @mock.patch('k8sapp_ipsec_policy_operator.helm.ipsec_policy_agent.'
                'kubernetes.KubeOperator')
    @mock.patch('k8sapp_ipsec_policy_operator.helm.ipsec_policy_agent.dbapi')
    def test_get_overrides_falls_back_to_live_pod_label(
            self, mock_dbapi, mock_kube_operator):
        mock_dbapi.get_instance.return_value.kube_app_get.return_value = \
            mock.MagicMock(id=1)
        mock_dbapi.get_instance.return_value.helm_override_get.side_effect = \
            HelmOverrideNotFound()

        pod = mock.MagicMock()
        pod.metadata.labels = {LABEL_KEY: 'application'}
        mock_kube_operator.return_value.kube_get_pods_by_selector.return_value = \
            [pod]

        result = self.helm_obj.get_overrides(
            namespace=app_constants.HELM_NS_IPSEC_POLICY_OPERATOR)
        self.assertEqual(
            result['controllerManager']['podLabels'],
            {LABEL_KEY: 'application'})


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
        self.assertEqual(result, {
            'controllerManager': {
                'podLabels': {LABEL_KEY: 'platform'},
            },
        })

    def test_get_overrides_no_namespace(self):
        result = self.helm_obj.get_overrides(namespace=None)
        self.assertIsInstance(result, dict)
        self.assertIn(app_constants.HELM_NS_IPSEC_POLICY_OPERATOR, result)

    def test_get_overrides_invalid_namespace(self):
        self.assertRaises(InvalidHelmNamespace,
                          self.helm_obj.get_overrides,
                          namespace='invalid-ns')

    @mock.patch('k8sapp_ipsec_policy_operator.helm.ipsec_policy_manager.dbapi')
    def test_get_overrides_uses_valid_user_override(self, mock_dbapi):
        mock_dbapi.get_instance.return_value.kube_app_get.return_value = \
            mock.MagicMock(id=1)
        mock_dbapi.get_instance.return_value.helm_override_get.return_value = {
            'user_overrides': {LABEL_KEY: 'application'},
        }
        result = self.helm_obj.get_overrides(
            namespace=app_constants.HELM_NS_IPSEC_POLICY_OPERATOR)
        self.assertEqual(
            result['controllerManager']['podLabels'],
            {LABEL_KEY: 'application'})

    @mock.patch('k8sapp_ipsec_policy_operator.helm.ipsec_policy_manager.dbapi')
    def test_get_overrides_ignores_invalid_user_override(self, mock_dbapi):
        mock_dbapi.get_instance.return_value.kube_app_get.return_value = \
            mock.MagicMock(id=1)
        mock_dbapi.get_instance.return_value.helm_override_get.return_value = {
            'user_overrides': {LABEL_KEY: 'not-a-real-component'},
        }
        result = self.helm_obj.get_overrides(
            namespace=app_constants.HELM_NS_IPSEC_POLICY_OPERATOR)
        self.assertEqual(
            result['controllerManager']['podLabels'],
            {LABEL_KEY: 'platform'})

    @mock.patch('k8sapp_ipsec_policy_operator.helm.ipsec_policy_manager.'
                'kubernetes.KubeOperator')
    @mock.patch('k8sapp_ipsec_policy_operator.helm.ipsec_policy_manager.dbapi')
    def test_get_overrides_falls_back_to_live_pod_label(
            self, mock_dbapi, mock_kube_operator):
        mock_dbapi.get_instance.return_value.kube_app_get.return_value = \
            mock.MagicMock(id=1)
        mock_dbapi.get_instance.return_value.helm_override_get.side_effect = \
            HelmOverrideNotFound()

        pod = mock.MagicMock()
        pod.metadata.labels = {LABEL_KEY: 'application'}
        mock_kube_operator.return_value.kube_get_pods_by_selector.return_value = \
            [pod]

        result = self.helm_obj.get_overrides(
            namespace=app_constants.HELM_NS_IPSEC_POLICY_OPERATOR)
        self.assertEqual(
            result['controllerManager']['podLabels'],
            {LABEL_KEY: 'application'})
