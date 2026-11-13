#
# Copyright (c) 2026 Wind River Systems, Inc.
#
# SPDX-License-Identifier: Apache-2.0
#
from k8sapp_ipsec_policy_operator.common import constants as app_constants
from oslo_log import log
from sysinv.common import exception
from sysinv.common import kubernetes
from sysinv.db import api as dbapi

import yaml


LOG = log.getLogger(__name__)


def set_component_label_override(overrides, app, chart):
    """ Set user overrides for core affinity

    Analyzes and returns proper overrides for ipsec-policy-manager pods.
    Supported overrides are app.starlingx.io/component:
    [platform,application]

    """

    dbapi_instance = dbapi.get_instance()

    db_app = dbapi_instance.kube_app_get(
        app_constants.HELM_APP_IPSEC_POLICY_OPERATOR)

    # Current chart overrides
    current_chart_overrides = get_current_pod_labels(app)

    # User chart overrides
    new_chart_overrides = get_helm_overrides(
        dbapi_instance,
        db_app,
        chart,
        app_constants.HELM_NS_IPSEC_POLICY_OPERATOR,
        'user_overrides')

    labels = app_constants.SUPPORTED_COMPONENT_LABELS
    namespace = app_constants.HELM_NS_IPSEC_POLICY_OPERATOR

    for pod_name in app_constants.DEFAULT_LABEL_OVERRIDES:
        overrides[namespace].setdefault(pod_name, {})
        overrides[namespace][pod_name][app_constants.LABEL_PARAMETER] = (
            new_chart_overrides
            if new_chart_overrides and new_chart_overrides in labels
            else current_chart_overrides[pod_name])


def get_current_pod_labels(app):
    current_pod_labels = {}
    try:
        kube_client = kubernetes.KubeOperator()
        for pod_name in app_constants.DEFAULT_LABEL_OVERRIDES:
            pod_desc = kube_client.kube_get_pods_by_selector(
                app_constants.HELM_NS_IPSEC_POLICY_OPERATOR,
                "app=%s" % app, "")
            if not pod_desc:
                continue
            for label, value in pod_desc[0].metadata.labels.items():
                item = {label: value}
                if item in app_constants.SUPPORTED_COMPONENT_LABELS:
                    current_pod_labels[pod_name] = item
    except exception.HelmOverrideNotFound:
        LOG.debug("Can not read pods's labels. Default value will be "
                    "used instead for label overrides.")
    finally:
        if not current_pod_labels:
            current_pod_labels = app_constants.DEFAULT_LABEL_OVERRIDES
    return current_pod_labels


def get_helm_overrides(dbapi_instance, app, chart, namespace, type_of_overrides):
    helm_overrides = {}
    try:
        helm_overrides = dbapi_instance.helm_override_get(
            app_id=app.id,
            name=chart,
            namespace=namespace,
        )[type_of_overrides]

        if type(helm_overrides) is str:
            helm_overrides = yaml.safe_load(helm_overrides)
    except exception.HelmOverrideNotFound:
        LOG.debug("Overrides for this chart not found, nothing to be "
                    "done.")
    return helm_overrides
