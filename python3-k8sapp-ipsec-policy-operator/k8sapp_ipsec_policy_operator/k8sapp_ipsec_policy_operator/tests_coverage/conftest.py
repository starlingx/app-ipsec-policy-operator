#
# Copyright (c) 2026 Wind River Systems, Inc.
#
# SPDX-License-Identifier: Apache-2.0
#
"""Pytest conftest - mock sysinv and oslo before test collection."""

import sys
import types
from unittest import mock


# --- Exception classes ---
class HostLabelNotFoundByKey(Exception):
    def __init__(self, *args, **kwargs):
        super().__init__("label not found")


class LifecycleSemanticCheckException(Exception):
    def __init__(self, msg=''):
        super().__init__(msg)


class InvalidHelmNamespace(Exception):
    def __init__(self, *args, **kwargs):
        super().__init__("invalid namespace")


class HostLabelAlreadyExists(Exception):
    pass


# --- Base classes ---
class BaseHelm(object):
    SUPPORTED_NAMESPACES = ['kube-system']

    def __init__(self, *args, **kwargs):
        pass


class AppLifecycleOperator(object):
    def __init__(self, *args, **kwargs):
        pass

    def app_lifecycle_actions(self, context, conductor_obj, app_op, app,
                              hook_info):
        pass


# --- Lifecycle constants ---
class LifecycleConstants(object):
    APP_LIFECYCLE_TYPE_OPERATION = 'operation'
    APP_LIFECYCLE_TYPE_RESOURCE = 'resource'
    APP_LIFECYCLE_TYPE_SEMANTIC_CHECK = 'semantic-check'
    APP_LIFECYCLE_TIMING_PRE = 'pre'
    APP_LIFECYCLE_TIMING_POST = 'post'
    APP_LIFECYCLE_MODE_AUTO = 'auto'


# Only mock sysinv/oslo if real sysinv is not installed.
# This allows stestr (which has real sysinv deps) to coexist
# with pytest-based coverage tests that use these mocks.
if 'sysinv.common' not in sys.modules:
    mod_sysinv = types.ModuleType('sysinv')
    mod_sysinv_common = types.ModuleType('sysinv.common')
    mod_sysinv_helm = types.ModuleType('sysinv.helm')

    # sysinv.common.exception
    mod_exception = types.ModuleType('sysinv.common.exception')
    mod_exception.HostLabelNotFoundByKey = HostLabelNotFoundByKey
    mod_exception.LifecycleSemanticCheckException = \
        LifecycleSemanticCheckException
    mod_exception.InvalidHelmNamespace = InvalidHelmNamespace
    mod_exception.HostLabelAlreadyExists = HostLabelAlreadyExists

    # sysinv.common.constants
    mock_sys_constants = types.ModuleType('sysinv.common.constants')
    mock_sys_constants.APP_REMOVE_OP = 'remove'
    mock_sys_constants.APP_APPLY_OP = 'apply'
    mock_sys_constants.APP_EVALUATE_REAPPLY_OP = 'evaluate-reapply'
    mock_sys_constants.APP_INACTIVE_STATE = 'inactive'

    # sysinv.common.kubernetes
    mock_kubernetes = types.ModuleType('sysinv.common.kubernetes')
    mock_kubernetes.KUBERNETES_ADMIN_CONF = '/etc/kubernetes/admin.conf'

    # sysinv.common.utils
    mock_cutils = mock.MagicMock()
    mock_cutils.__name__ = 'sysinv.common.utils'
    mock_cutils.is_aio_simplex_system = mock.MagicMock(return_value=False)
    mock_cutils.trycmd = mock.MagicMock(return_value=('', ''))

    # sysinv.helm.base
    mod_helm_base = types.ModuleType('sysinv.helm.base')
    mod_helm_base.BaseHelm = BaseHelm

    # sysinv.helm.lifecycle_base
    mod_lifecycle_base = types.ModuleType('sysinv.helm.lifecycle_base')
    mod_lifecycle_base.AppLifecycleOperator = AppLifecycleOperator

    # sysinv.helm.lifecycle_utils
    mock_lifecycle_utils = mock.MagicMock()
    mock_lifecycle_utils.__name__ = 'sysinv.helm.lifecycle_utils'

    # sysinv.helm.lifecycle_constants
    mod_lifecycle_constants = types.ModuleType(
        'sysinv.helm.lifecycle_constants')
    mod_lifecycle_constants.LifecycleConstants = LifecycleConstants

    # oslo_log
    mod_oslo_log = types.ModuleType('oslo_log')
    mod_oslo_log.log = mock.MagicMock()
    mod_oslo_log.log.getLogger = mock.MagicMock(
        return_value=mock.MagicMock())
    mod_oslo_log_log = types.ModuleType('oslo_log.log')
    mod_oslo_log_log.getLogger = mock.MagicMock(
        return_value=mock.MagicMock())

    # Wire up module hierarchy
    mod_sysinv.common = mod_sysinv_common
    mod_sysinv.helm = mod_sysinv_helm
    mod_sysinv_common.exception = mod_exception
    mod_sysinv_common.constants = mock_sys_constants
    mod_sysinv_common.kubernetes = mock_kubernetes
    mod_sysinv_common.utils = mock_cutils
    mod_sysinv_helm.base = mod_helm_base
    mod_sysinv_helm.lifecycle_base = mod_lifecycle_base
    mod_sysinv_helm.lifecycle_utils = mock_lifecycle_utils
    mod_sysinv_helm.lifecycle_constants = mod_lifecycle_constants

    # Register all in sys.modules
    sys.modules['sysinv'] = mod_sysinv
    sys.modules['sysinv.common'] = mod_sysinv_common
    sys.modules['sysinv.common.exception'] = mod_exception
    sys.modules['sysinv.common.constants'] = mock_sys_constants
    sys.modules['sysinv.common.kubernetes'] = mock_kubernetes
    sys.modules['sysinv.common.utils'] = mock_cutils
    sys.modules['sysinv.helm'] = mod_sysinv_helm
    sys.modules['sysinv.helm.base'] = mod_helm_base
    sys.modules['sysinv.helm.lifecycle_base'] = mod_lifecycle_base
    sys.modules['sysinv.helm.lifecycle_utils'] = mock_lifecycle_utils
    sys.modules['sysinv.helm.lifecycle_constants'] = mod_lifecycle_constants
    sys.modules['oslo_log'] = mod_oslo_log
    sys.modules['oslo_log.log'] = mod_oslo_log_log

    mod_oslo_log.logging = mod_oslo_log_log
else:
    # When real sysinv is available, reference the real constants
    from sysinv.common import constants as _real_constants
    from sysinv.common import utils as _real_utils
    mock_sys_constants = _real_constants
    mock_cutils = _real_utils
    mock_lifecycle_utils = mock.MagicMock()
