#
# Copyright (c) 2026 Wind River Systems, Inc.
#
# SPDX-License-Identifier: Apache-2.0
#
"""Unit tests for lifecycle_ipsec_policy_operator module."""

import unittest
from unittest import mock

from k8sapp_ipsec_policy_operator.tests_coverage.conftest import HostLabelNotFoundByKey
from k8sapp_ipsec_policy_operator.lifecycle import lifecycle_ipsec_policy_operator as lifecycle_mod

IPsecPolicyOperatorAppLifecycleOperator = lifecycle_mod.IPsecPolicyOperatorAppLifecycleOperator


class TestLifecycleGetExistingLabels(unittest.TestCase):
    """Tests for get_existing_labels method."""

    def setUp(self):
        self.operator = object.__new__(
            IPsecPolicyOperatorAppLifecycleOperator)

    def test_get_existing_labels_found(self):
        mock_dbapi = mock.MagicMock()
        mock_label = mock.MagicMock()
        mock_label.uuid = 'test-uuid-123'
        mock_dbapi.label_query.return_value = mock_label
        mock_host = mock.MagicMock()
        mock_host.id = 1

        result = self.operator.get_existing_labels(mock_dbapi, mock_host)
        self.assertEqual(result,
                         {'ipsec-policy-agent-operator': 'test-uuid-123'})

    def test_get_existing_labels_not_found(self):
        mock_dbapi = mock.MagicMock()
        mock_dbapi.label_query.side_effect = HostLabelNotFoundByKey()
        mock_host = mock.MagicMock()
        mock_host.id = 1

        result = self.operator.get_existing_labels(mock_dbapi, mock_host)
        self.assertEqual(result, {})


class TestLifecycleSaveLabel(unittest.TestCase):
    """Tests for save_label method."""

    def setUp(self):
        self.operator = object.__new__(
            IPsecPolicyOperatorAppLifecycleOperator)

    def test_save_label_create_new(self):
        mock_dbapi = mock.MagicMock()
        mock_dbapi.label_query.side_effect = HostLabelNotFoundByKey()
        mock_new_label = mock.MagicMock()
        mock_dbapi.label_create.return_value = mock_new_label
        mock_host = mock.MagicMock()
        mock_host.id = 1
        mock_host.uuid = 'host-uuid'
        mock_host.hostname = 'controller-0'

        label = {'key': 'ipsec-policy-agent-operator', 'value': 'enabled'}
        result = self.operator.save_label(mock_dbapi, mock_host, label)
        self.assertEqual(result, mock_new_label)
        mock_dbapi.label_create.assert_called_once()

    def test_save_label_update_existing(self):
        mock_dbapi = mock.MagicMock()
        mock_label = mock.MagicMock()
        mock_label.uuid = 'existing-uuid'
        mock_dbapi.label_query.return_value = mock_label
        mock_updated = mock.MagicMock()
        mock_dbapi.label_update.return_value = mock_updated
        mock_host = mock.MagicMock()
        mock_host.id = 1
        mock_host.uuid = 'host-uuid'
        mock_host.hostname = 'controller-0'

        label = {'key': 'ipsec-policy-agent-operator', 'value': 'enabled'}
        result = self.operator.save_label(mock_dbapi, mock_host, label)
        self.assertEqual(result, mock_updated)
        mock_dbapi.label_update.assert_called_once_with(
            'existing-uuid', {'label_value': 'enabled'})


class TestLifecycleApplyLabels(unittest.TestCase):
    """Tests for apply_labels method."""

    def setUp(self):
        self.operator = object.__new__(
            IPsecPolicyOperatorAppLifecycleOperator)

    def test_apply_labels_skips_storage(self):
        mock_app_op = mock.MagicMock()
        mock_host = mock.MagicMock()
        mock_host.personality = 'storage'
        mock_app_op._dbapi.ihost_get_list.return_value = [mock_host]

        with mock.patch.object(self.operator, 'save_label') as mock_save:
            self.operator.apply_labels(mock_app_op)
            mock_save.assert_not_called()

    def test_apply_labels_applies_to_controller(self):
        mock_app_op = mock.MagicMock()
        mock_host = mock.MagicMock()
        mock_host.personality = 'controller'
        mock_app_op._dbapi.ihost_get_list.return_value = [mock_host]

        with mock.patch.object(self.operator, 'save_label') as mock_save:
            self.operator.apply_labels(mock_app_op)
            mock_save.assert_called_once()

    def test_apply_labels_applies_to_worker(self):
        mock_app_op = mock.MagicMock()
        mock_host = mock.MagicMock()
        mock_host.personality = 'worker'
        mock_app_op._dbapi.ihost_get_list.return_value = [mock_host]

        with mock.patch.object(self.operator, 'save_label') as mock_save:
            self.operator.apply_labels(mock_app_op)
            mock_save.assert_called_once()
