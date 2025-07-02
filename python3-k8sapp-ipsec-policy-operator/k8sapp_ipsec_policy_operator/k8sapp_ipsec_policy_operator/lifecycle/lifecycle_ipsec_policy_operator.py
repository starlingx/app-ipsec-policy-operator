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
from sysinv.common import exception
from sysinv.common import utils as cutils
from sysinv.helm import lifecycle_base as base
from sysinv.helm import lifecycle_utils
from sysinv.helm.lifecycle_constants import LifecycleConstants

from k8sapp_ipsec_policy_operator.common import constants as app_constants

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
        if hook_info.lifecycle_type == LifecycleConstants.APP_LIFECYCLE_TYPE_RESOURCE:
            if hook_info.operation == constants.APP_APPLY_OP:
                if hook_info.relative_timing == LifecycleConstants.APP_LIFECYCLE_TIMING_PRE:
                    return self.pre_apply(app_op, app, hook_info)
            elif hook_info.operation == constants.APP_REMOVE_OP:
                if hook_info.relative_timing == LifecycleConstants.APP_LIFECYCLE_TIMING_POST:
                    return self.post_remove(app_op, app, hook_info)

        elif hook_info.lifecycle_type == LifecycleConstants.APP_LIFECYCLE_TYPE_SEMANTIC_CHECK:
            if hook_info.operation in [constants.APP_APPLY_OP,
                                       constants.APP_EVALUATE_REAPPLY_OP]:
                return self.pre_apply_semantic_check(app_op, hook_info.operation)

        super(IPsecPolicyOperatorAppLifecycleOperator, self).app_lifecycle_actions(
            context, conductor_obj, app_op, app, hook_info)

    def get_existing_labels(self, dbapi, host):
        LOG.info("Getting existing_labels")
        existing_labels = {}
        label_key = 'ipsec-policy-agent-operator'

        label = None
        try:
            label = dbapi.label_query(host.id, label_key)
        except exception.HostLabelNotFoundByKey:
            pass
        if label:
            existing_labels.update({label_key: label.uuid})
        return existing_labels

    def save_label(self, dbapi, host, label):
        LOG.info(f"Save '{label['key']}' label to host: {host.hostname}")
        existing_labels = self.get_existing_labels(dbapi, host)

        values = {
            'host_id': host.id,
            'label_key': label['key'],
            'label_value': label['value']
        }

        try:
            if existing_labels.get(label['key'], None):
                # Update the value
                label_uuid = existing_labels.get(label['key'])
                new_label = dbapi.label_update(label_uuid, {'label_value': label['value']})
            else:
                new_label = dbapi.label_create(host.uuid, values)
        except exception.HostLabelAlreadyExists:
            LOG.error("Error creating label %s" % label['key'])
        return new_label

    def apply_labels(self, app_op):
        dbapi = app_op._dbapi
        ipsec_policy_operator_keys = app_constants.APP_LABELS
        hosts = dbapi.ihost_get_list()

        for host in hosts:
            # IPsec Policy Operator labels are not applied to storage nodes
            if host.personality == "storage":
                continue

            # Save 'ipsec-policy-agent-operator' label to hosts
            label = {'key': ipsec_policy_operator_keys['ipsec-policy-agent'],
                     'value': 'enabled'}
            self.save_label(dbapi, host, label)

    def remove_labels(self, app_op, host, label):
        LOG.info(f"Removing '{label}' label to host: {host.hostname}")
        # confirming if the label exists before removing it
        dbapi = app_op._dbapi
        lbl_obj = self.get_existing_labels(dbapi, host)
        if label in lbl_obj:
            dbapi.label_destroy(lbl_obj[label])
            app_op._update_kubernetes_labels(host.hostname, {label: None})

    def cleanup_labels(self, app_op):
        hosts = app_op._dbapi.ihost_get_list()
        label = app_constants.APP_LABELS['ipsec-policy-agent']

        for host in hosts:
            # Save 'ipsec-policy-agent-operator' label to hosts
            self.remove_labels(app_op, host, label)

    def pre_apply(self, app_op, app, hook_info):
        LOG.info("Executing pre_apply for IPsec Policy Operator app")

        LOG.info("Creating local registry secrets")
        lifecycle_utils.create_local_registry_secrets(app_op, app, hook_info)

        LOG.info("Applying ipsec-policy-agent-operator labels")
        self.apply_labels(app_op)

    def post_remove(self, app_op, app, hook_info):
        LOG.info("Executing post_remove for IPsec Policy Operator app")

        LOG.info("Removing local registry secrets")
        lifecycle_utils.delete_local_registry_secrets(app_op, app, hook_info)

        LOG.info("Removing ipsec-policy-agent-operator labels")
        self.cleanup_labels(app_op)

    def pre_apply_semantic_check(self, app_op, operation):
        LOG.info("Executing pre_apply_semantic_check for IPsec Policy Operator app")

        # Stop apply the app since the IPsec is for inter host pod-to-pod traffic
        if cutils.is_aio_simplex_system(app_op._dbapi):
            raise exception.LifecycleSemanticCheckException(
                "Cannot apply application: the app is only for multiple nodes system."
            )

        if operation == constants.APP_EVALUATE_REAPPLY_OP:
            LOG.info("Reapplying ipsec-policy-agent-operator labels to all nodes")
            self.apply_labels(app_op)
