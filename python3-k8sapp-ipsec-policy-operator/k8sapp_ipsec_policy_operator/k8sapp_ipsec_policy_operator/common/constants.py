#
# Copyright (c) 2026 Wind River Systems, Inc.
#
# SPDX-License-Identifier: Apache-2.0
#

from sysinv.helm import base

# App name
HELM_APP_IPSEC_POLICY_OPERATOR = 'ipsec-policy-operator'

# IPsec Policy Agent Operator
HELM_CHART_IPSEC_POLICY_AGENT = 'ipsec-policy-agent'
APP_LABEL_IPSEC_POLICY_AGENT = 'ipsec-policy-agent'

# IPsec Policy Manager Operator
HELM_CHART_IPSEC_POLICY_MANAGER = 'ipsec-policy-manager'
APP_LABEL_IPSEC_POLICY_MANAGER = 'ipsec-policy-manager'

# IPsec Policy Operator App shared constants
APP_LABELS = {
    'ipsec-policy-agent': 'ipsec-policy-agent-operator',
}
HELM_NS_IPSEC_POLICY_OPERATOR = 'ipsec-policy-operator'
CHART_GROUP_IPSEC_POLICY_AGENT = 'ipsec-policy-operator-charts'

# app.starlingx.io/component pod label support
HELM_CHART_COMPONENT_LABEL = 'app.starlingx.io/component'


class PodNames(object):
    # NOTE: Matches the top-level values.yaml key ('controllerManager')
    # that templates render into pod labels via
    # `.Values.controllerManager.podLabels`.
    CONTROLLER_MANAGER = 'controllerManager'


SUPPORTED_NAMESPACES = base.BaseHelm.SUPPORTED_NAMESPACES + \
    [HELM_NS_IPSEC_POLICY_OPERATOR]

COMPONENT_CORE_PLATFORM = 'platform'

COMPONENT_CORE_APPLICATION = 'application'

SUPPORTED_COMPONENT_LABELS = [
    {HELM_CHART_COMPONENT_LABEL: COMPONENT_CORE_PLATFORM},
    {HELM_CHART_COMPONENT_LABEL: COMPONENT_CORE_APPLICATION},
]

DEFAULT_LABEL_OVERRIDES = {
    PodNames.CONTROLLER_MANAGER: {
        HELM_CHART_COMPONENT_LABEL: COMPONENT_CORE_PLATFORM,
    },
}

LABEL_PARAMETER = 'podLabels'
