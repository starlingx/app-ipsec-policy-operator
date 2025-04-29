#
# Copyright (c) 2025 Wind River Systems, Inc.
#
# SPDX-License-Identifier: Apache-2.0
#

import mock

from stestr.tests import base
from sysinv.common import constants
from sysinv.db import api as db_api
from sysinv.helm import lifecycle_hook
from sysinv.tests.db import base as db_base
from sysinv.tests.db import utils as dbutils


class IPsecPolicyOperatorTestCase(db_base.DbTestCase):
    def setUp(self):
        super(IPsecPolicyOperatorTestCase, self).setUp()

        self.dbapi = db_api.get_instance()
        self.app_op = mock.MagicMock()
        self.new_app = mock.MagicMock()

        self.new_app.id = 2
        self.new_app.name = 'ipsec-policy-operator'

        self.old_db_app = dbutils.create_test_app(
            id=1,
            name='ipsec-policy-operator',
            app_version='1.0-0')
        # creating an inactive app does not work, need to update the status later
        self.dbapi.kube_app_update(self.old_db_app.id, {'status': constants.APP_INACTIVE_STATE})

        self.new_db_app = dbutils.create_test_app(
            id=self.new_app.id,
            status=constants.APP_APPLY_IN_PROGRESS,
            name='ipsec-policy-operator',
            app_version='1.1-1')

        self.hook_info = lifecycle_hook.LifecycleHookInfo()
        self.hook_info.init(
            constants.APP_LIFECYCLE_MODE_AUTO,
            constants.APP_LIFECYCLE_TYPE_RESOURCE,
            constants.APP_LIFECYCLE_TIMING_PRE,
            constants.APP_APPLY_OP)

    def tearDown(self):
        super(IPsecPolicyOperatorTestCase, self).tearDown()


class IPsecTestCaseDummy(base.TestCase):
    # without a test zuul will fail
    def test_dummy(self):
        pass
