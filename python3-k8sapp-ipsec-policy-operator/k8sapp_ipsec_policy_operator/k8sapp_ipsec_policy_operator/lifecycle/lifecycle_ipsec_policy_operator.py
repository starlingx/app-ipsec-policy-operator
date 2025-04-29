#
# Copyright (c) 2025 Wind River Systems, Inc.
#
# SPDX-License-Identifier: Apache-2.0
#
# All Rights Reserved.
#

"""System inventory App lifecycle operator."""

from oslo_log import log as logging
from sysinv.common import constants
from sysinv.helm import lifecycle_base as base
from sysinv.helm import lifecycle_utils

LOG = logging.getLogger(__name__)


class IPsecPolicyOperatorAppLifecycleOperator(base.AppLifecycleOperator):
    def app_lifecycle_actions(self, context, conductor_obj, app_op, app, hook_info):
        """Perform lifecycle actions for an operation

        :param context: request context, can be None
        :param conductor_obj: conductor object, can be None
        :param app_op: AppOperator object
        :param app: AppOperator.Application object
        :param hook_info: LifecycleHookInfo object

        """
        if hook_info.lifecycle_type == constants.APP_LIFECYCLE_TYPE_RESOURCE:
            if hook_info.operation == constants.APP_APPLY_OP:
                if hook_info.relative_timing == constants.APP_LIFECYCLE_TIMING_PRE:
                    return self.pre_apply(app_op, app, hook_info)

        if hook_info.lifecycle_type == constants.APP_LIFECYCLE_TYPE_OPERATION:
            if hook_info.operation == constants.APP_BACKUP:
                if hook_info.relative_timing == constants.APP_LIFECYCLE_TIMING_PRE:
                    return self.pre_backup(app_op, app)

        if hook_info.lifecycle_type == constants.APP_LIFECYCLE_TYPE_OPERATION:
            if hook_info.operation == constants.APP_BACKUP:
                if hook_info.relative_timing == constants.APP_LIFECYCLE_TIMING_POST:
                    return self.post_backup(app_op, app)

        if hook_info.lifecycle_type == constants.APP_LIFECYCLE_TYPE_OPERATION:
            if hook_info.operation == constants.APP_RESTORE:
                if hook_info.relative_timing == constants.APP_LIFECYCLE_TIMING_POST:
                    return self.post_restore(app_op, app)

        super(IPsecPolicyOperatorAppLifecycleOperator, self).app_lifecycle_actions(
            context, conductor_obj, app_op, app, hook_info)

    def pre_apply(self, app_op, app, hook_info):
        LOG.info("Executing pre_apply for IPsec Policy Operator app")

        # Create local registry secret
        lifecycle_utils.create_local_registry_secrets(app_op, app, hook_info)

    def pre_backup(self, app_op, app):
        LOG.info("Executing pre_backup for IPsec Policy Operator app")
        LOG.info("{} app: pre_backup".format(app.name))

    def post_backup(self, app_op, app):
        LOG.info("Executing post_backup for IPsec Policy Operator app")
        LOG.info("{} app: post_backp".format(app.name))

    def post_restore(self, app_op, app):
        LOG.info("Executing post_restore for IPsec Policy Operator app")
        LOG.info("{} app: pre_restore".format(app.name))
