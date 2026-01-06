#
# Copyright (c) 2025 Wind River Systems, Inc.
#
# SPDX-License-Identifier: Apache-2.0
#

# App name
HELM_APP_IPSEC_POLICY_OPERATOR = 'ipsec-policy-operator'

# IPsec Policy Agent Operator
HELM_CHART_IPSEC_POLICY_AGENT = 'ipsec-policy-agent'

# IPsec Policy Manager Operator
HELM_CHART_IPSEC_POLICY_MANAGER = 'ipsec-policy-manager'

# IPsec Policy Operator App shared constants
APP_LABELS = {
    'ipsec-policy-agent': 'ipsec-policy-agent-operator',
}
HELM_NS_IPSEC_POLICY_OPERATOR = 'ipsec-policy-operator'
CHART_GROUP_IPSEC_POLICY_AGENT = 'ipsec-policy-operator-charts'
