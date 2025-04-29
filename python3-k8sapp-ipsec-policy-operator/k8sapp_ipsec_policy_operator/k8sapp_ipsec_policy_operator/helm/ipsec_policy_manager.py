#
# Copyright (c) 2025 Wind River Systems, Inc.
#
# SPDX-License-Identifier: Apache-2.0
#
from k8sapp_ipsec_policy_operator.common import constants as app_constants
from oslo_log import log
from sysinv.common import exception
from sysinv.helm import base

LOG = log.getLogger(__name__)


class IPsecPolicyManagerHelm(base.BaseHelm):
    """Class to encapsulate helm operations for IPsec Policy Manager"""

    CHART = app_constants.HELM_CHART_IPSEC_POLICY_MANAGER

    SUPPORTED_NAMESPACES = base.BaseHelm.SUPPORTED_NAMESPACES + \
        [app_constants.HELM_NS_IPSEC_POLICY_OPERATOR]

    SUPPORTED_APP_NAMESPACES = {
        app_constants.HELM_APP_IPSEC_POLICY_OPERATOR: SUPPORTED_NAMESPACES
    }

    def get_overrides(self, namespace=None):
        LOG.info("Generating system_overrides for %s chart." % self.CHART)

        overrides = {
            app_constants.HELM_NS_IPSEC_POLICY_OPERATOR: {}
        }

        if namespace in self.SUPPORTED_NAMESPACES:
            return overrides[namespace]
        elif namespace:
            raise exception.InvalidHelmNamespace(chart=self.CHART,
                                                 namespace=namespace)
        else:
            return overrides
