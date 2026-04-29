#
# Copyright (c) 2026 Wind River Systems, Inc.
#
# SPDX-License-Identifier: Apache-2.0
#
"""Unit tests for app_lifecycle_actions dispatch logic."""

import unittest
from unittest import mock

from k8sapp_ipsec_policy_operator.tests_coverage.conftest import LifecycleConstants
from k8sapp_ipsec_policy_operator.tests_coverage.conftest import mock_sys_constants
from k8sapp_ipsec_policy_operator.lifecycle import lifecycle_ipsec_policy_operator as lifecycle_mod

IPsecPolicyOperatorAppLifecycleOperator = lifecycle_mod.IPsecPolicyOperatorAppLifecycleOperator


class TestLifecycleActions(unittest.TestCase):
    """Tests for app_lifecycle_actions method."""

    def setUp(self):
        self.operator = object.__new__(
            IPsecPolicyOperatorAppLifecycleOperator)

    def test_post_remove_operation_dispatch(self):
        hook_info = mock.MagicMock()
        hook_info.lifecycle_type = LifecycleConstants.APP_LIFECYCLE_TYPE_OPERATION
        hook_info.operation = mock_sys_constants.APP_REMOVE_OP
        hook_info.relative_timing = LifecycleConstants.APP_LIFECYCLE_TIMING_POST

        with mock.patch.object(self.operator, 'post_remove',
                               return_value=None) as mock_pr:
            self.operator.app_lifecycle_actions(
                None, None, mock.MagicMock(), mock.MagicMock(), hook_info)
            mock_pr.assert_called_once()

    def test_pre_apply_resource_dispatch(self):
        hook_info = mock.MagicMock()
        hook_info.lifecycle_type = LifecycleConstants.APP_LIFECYCLE_TYPE_RESOURCE
        hook_info.operation = mock_sys_constants.APP_APPLY_OP
        hook_info.relative_timing = LifecycleConstants.APP_LIFECYCLE_TIMING_PRE

        with mock.patch.object(self.operator, 'pre_apply',
                               return_value=None) as mock_pa:
            self.operator.app_lifecycle_actions(
                None, None, mock.MagicMock(), mock.MagicMock(), hook_info)
            mock_pa.assert_called_once()

    def test_semantic_check_apply_dispatch(self):
        hook_info = mock.MagicMock()
        hook_info.lifecycle_type = LifecycleConstants.APP_LIFECYCLE_TYPE_SEMANTIC_CHECK
        hook_info.operation = mock_sys_constants.APP_APPLY_OP

        with mock.patch.object(self.operator, 'pre_apply_semantic_check',
                               return_value=None) as mock_sc:
            self.operator.app_lifecycle_actions(
                None, None, mock.MagicMock(), mock.MagicMock(), hook_info)
            mock_sc.assert_called_once()

    def test_semantic_check_remove_dispatch(self):
        hook_info = mock.MagicMock()
        hook_info.lifecycle_type = LifecycleConstants.APP_LIFECYCLE_TYPE_SEMANTIC_CHECK
        hook_info.operation = mock_sys_constants.APP_REMOVE_OP
        hook_info.relative_timing = LifecycleConstants.APP_LIFECYCLE_TIMING_PRE

        with mock.patch.object(self.operator, 'pre_remove',
                               return_value=None) as mock_prm:
            self.operator.app_lifecycle_actions(
                None, None, mock.MagicMock(), mock.MagicMock(), hook_info)
            mock_prm.assert_called_once()

    def test_fallback_to_super(self):
        hook_info = mock.MagicMock()
        hook_info.lifecycle_type = 'unknown_type'
        hook_info.operation = 'unknown_op'

        # Should not raise - falls through to super
        self.operator.app_lifecycle_actions(
            None, None, mock.MagicMock(), mock.MagicMock(), hook_info)

    def test_operation_type_non_remove(self):
        hook_info = mock.MagicMock()
        hook_info.lifecycle_type = LifecycleConstants.APP_LIFECYCLE_TYPE_OPERATION
        hook_info.operation = mock_sys_constants.APP_APPLY_OP
        hook_info.relative_timing = LifecycleConstants.APP_LIFECYCLE_TIMING_PRE

        # Should not raise - falls through to super
        self.operator.app_lifecycle_actions(
            None, None, mock.MagicMock(), mock.MagicMock(), hook_info)

    def test_semantic_check_evaluate_reapply_dispatch(self):
        hook_info = mock.MagicMock()
        hook_info.lifecycle_type = LifecycleConstants.APP_LIFECYCLE_TYPE_SEMANTIC_CHECK
        hook_info.operation = mock_sys_constants.APP_EVALUATE_REAPPLY_OP

        with mock.patch.object(self.operator, 'pre_apply_semantic_check',
                               return_value=None) as mock_sc:
            self.operator.app_lifecycle_actions(
                None, None, mock.MagicMock(), mock.MagicMock(), hook_info)
            mock_sc.assert_called_once()
