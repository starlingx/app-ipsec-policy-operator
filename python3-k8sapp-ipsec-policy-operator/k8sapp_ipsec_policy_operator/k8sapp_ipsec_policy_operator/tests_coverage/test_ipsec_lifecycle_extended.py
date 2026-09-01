#
# Copyright (c) 2026 Wind River Systems, Inc.
#
# SPDX-License-Identifier: Apache-2.0
#
"""Extended unit tests for lifecycle operations."""

import unittest
from unittest import mock

from k8sapp_ipsec_policy_operator.tests_coverage.conftest import LifecycleSemanticCheckException
from k8sapp_ipsec_policy_operator.tests_coverage.conftest import mock_cutils
from k8sapp_ipsec_policy_operator.tests_coverage.conftest import mock_lifecycle_utils
from k8sapp_ipsec_policy_operator.tests_coverage.conftest import mock_sys_constants
from k8sapp_ipsec_policy_operator.lifecycle import lifecycle_ipsec_policy_operator as lifecycle_mod

IPsecPolicyOperatorAppLifecycleOperator = lifecycle_mod.IPsecPolicyOperatorAppLifecycleOperator


class TestLifecycleRemoveLabels(unittest.TestCase):
    """Tests for remove_labels method."""

    def setUp(self):
        self.operator = object.__new__(
            IPsecPolicyOperatorAppLifecycleOperator)

    def test_remove_labels_existing(self):
        mock_app_op = mock.MagicMock()
        mock_host = mock.MagicMock()
        mock_host.hostname = 'controller-0'
        label_key = 'ipsec-policy-agent-operator'

        with mock.patch.object(self.operator, 'get_existing_labels',
                               return_value={label_key: 'uuid-123'}):
            self.operator.remove_labels(mock_app_op, mock_host, label_key)
            mock_app_op._dbapi.label_destroy.assert_called_once_with('uuid-123')
            mock_app_op._update_kubernetes_labels.assert_called_once_with(
                'controller-0', {label_key: None})

    def test_remove_labels_not_existing(self):
        mock_app_op = mock.MagicMock()
        mock_host = mock.MagicMock()
        mock_host.hostname = 'controller-0'

        with mock.patch.object(self.operator, 'get_existing_labels',
                               return_value={}):
            self.operator.remove_labels(mock_app_op, mock_host,
                                        'ipsec-policy-agent-operator')
            mock_app_op._dbapi.label_destroy.assert_not_called()


class TestLifecycleCleanupLabels(unittest.TestCase):
    """Tests for cleanup_labels method."""

    def setUp(self):
        self.operator = object.__new__(
            IPsecPolicyOperatorAppLifecycleOperator)

    def test_cleanup_labels_all_hosts(self):
        mock_app_op = mock.MagicMock()
        mock_app_op._dbapi.ihost_get_list.return_value = [
            mock.MagicMock(), mock.MagicMock()]

        with mock.patch.object(self.operator, 'remove_labels') as mock_rm:
            self.operator.cleanup_labels(mock_app_op)
            self.assertEqual(mock_rm.call_count, 2)


class TestLifecyclePreApply(unittest.TestCase):
    """Tests for pre_apply method."""

    def setUp(self):
        self.operator = object.__new__(
            IPsecPolicyOperatorAppLifecycleOperator)

    def test_pre_apply(self):
        mock_app_op = mock.MagicMock()
        mock_app = mock.MagicMock()
        mock_hook_info = mock.MagicMock()
        mock_lifecycle_utils.create_local_registry_secrets.reset_mock()

        with mock.patch.object(self.operator, 'apply_labels') as mock_apply:
            self.operator.pre_apply(mock_app_op, mock_app, mock_hook_info)
            mock_lifecycle_utils.create_local_registry_secrets.assert_called_once()
            mock_apply.assert_called_once_with(mock_app_op)


class TestLifecyclePreRemove(unittest.TestCase):
    """Tests for pre_remove method."""

    def setUp(self):
        self.operator = object.__new__(
            IPsecPolicyOperatorAppLifecycleOperator)

    def test_pre_remove_scale_success_no_configmaps(self):
        mock_cutils.trycmd.side_effect = [('scaled', ''), ('', '')]
        self.operator.pre_remove()
        self.assertEqual(mock_cutils.trycmd.call_count, 2)
        mock_cutils.trycmd.reset_mock()
        mock_cutils.trycmd.side_effect = None

    def test_pre_remove_scale_warning_with_configmaps(self):
        mock_cutils.trycmd.side_effect = [
            ('', 'scale error'), ('cm1\ncm2', ''),
            ('deleted', ''), ('deleted', '')]
        self.operator.pre_remove()
        self.assertEqual(mock_cutils.trycmd.call_count, 4)
        mock_cutils.trycmd.reset_mock()
        mock_cutils.trycmd.side_effect = None

    def test_pre_remove_get_configmaps_error(self):
        mock_cutils.trycmd.side_effect = [('scaled', ''), ('', 'get error')]
        self.assertRaises(LifecycleSemanticCheckException,
                          self.operator.pre_remove)
        mock_cutils.trycmd.reset_mock()
        mock_cutils.trycmd.side_effect = None

    def test_pre_remove_delete_configmap_error(self):
        mock_cutils.trycmd.side_effect = [
            ('scaled', ''), ('cm1', ''), ('', 'delete error')]
        self.assertRaises(LifecycleSemanticCheckException,
                          self.operator.pre_remove)
        mock_cutils.trycmd.reset_mock()
        mock_cutils.trycmd.side_effect = None


class TestLifecyclePostRemove(unittest.TestCase):
    """Tests for post_remove method."""

    def setUp(self):
        self.operator = object.__new__(
            IPsecPolicyOperatorAppLifecycleOperator)

    def test_post_remove(self):
        mock_app_op = mock.MagicMock()
        mock_app = mock.MagicMock()
        mock_hook_info = mock.MagicMock()
        mock_lifecycle_utils.delete_local_registry_secrets.reset_mock()

        with mock.patch.object(self.operator, 'cleanup_labels') as mock_cl:
            self.operator.post_remove(mock_app_op, mock_app, mock_hook_info)
            mock_lifecycle_utils.delete_local_registry_secrets.assert_called_once()
            mock_cl.assert_called_once_with(mock_app_op)


class TestLifecycleSemanticCheck(unittest.TestCase):
    """Tests for pre_apply_semantic_check method."""

    def setUp(self):
        self.operator = object.__new__(
            IPsecPolicyOperatorAppLifecycleOperator)

    def test_semantic_check_simplex_raises(self):
        mock_cutils.is_aio_simplex_system.return_value = True
        mock_app_op = mock.MagicMock()
        self.assertRaises(LifecycleSemanticCheckException,
                          self.operator.pre_apply_semantic_check,
                          mock_app_op, mock_sys_constants.APP_APPLY_OP)
        mock_cutils.is_aio_simplex_system.return_value = False

    def test_semantic_check_apply_non_simplex(self):
        mock_cutils.is_aio_simplex_system.return_value = False
        mock_app_op = mock.MagicMock()
        self.operator.pre_apply_semantic_check(
            mock_app_op, mock_sys_constants.APP_APPLY_OP)

    def test_semantic_check_reapply_applies_labels(self):
        mock_cutils.is_aio_simplex_system.return_value = False
        mock_app_op = mock.MagicMock()
        with mock.patch.object(self.operator, 'apply_labels') as mock_apply:
            self.operator.pre_apply_semantic_check(
                mock_app_op, mock_sys_constants.APP_EVALUATE_REAPPLY_OP)
            mock_apply.assert_called_once_with(mock_app_op)
