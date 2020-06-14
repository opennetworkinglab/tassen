# Copyright 2019-present Open Networking Foundation
#
# Licensed under the Apache License, Version 2.0 (the 'License');
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#   http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an 'AS IS' BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

# ------------------------------------------------------------------------------
# UPSTREAM TESTS
#
# To run all tests in this file:
#     make check TEST=accounting
#
# To run a specific test case:
#     make check TEST=accounting.<TEST CLASS NAME>
#
# For example:
#     make check TEST=accounting.UpstreamPppoeIp4UnicastTest
# ------------------------------------------------------------------------------

from base_test import *
from ptf.testutils import group
import upstream

INGRESS = "Ingress"
EGRESS = "Egress"
UPSTREAM = 'upstream'
DOWNSTREAM = 'downstream'

BYTES = 'bytes'
PKTS = 'packets'

COUNTER_NAME_TEMPLATE = "%sPipe.accounting.%s"


def ctr_name(gress, direction):
    return COUNTER_NAME_TEMPLATE % (gress, direction)


def ctr_key(gress, dir, idx, typ):
    return ctr_name(gress, dir) + "[%d].%s" % (idx, typ)


class AccountingTest(P4RuntimeTest):

    def read_counters(self, idx):
        result = dict()
        for gress in (INGRESS, EGRESS):
            for dir in (UPSTREAM, DOWNSTREAM):
                pkts, bytez = self.read_counter(ctr_name(gress, dir), idx)
                result[ctr_key(gress, dir, idx, PKTS)] = pkts
                result[ctr_key(gress, dir, idx, BYTES)] = bytez
        return result

    def assert_counter_increase(self, pre_counters, post_counters, idx,
                                pkt_direction, ig_bytes, eg_bytes, pkt_count=1):
        """
        Asserts that counters have increased by the given values.
        :param pre_counters: counter values before packet(s)
        :param post_counters: counter values after packet(s)
        :param idx: index of the counters to compare (accounting_id)
        :param pkt_direction: direction of packet(s)
        :param ig_bytes: expected byte count increase at ingress
        :param eg_bytes: expected byte count increase at egress
        :param pkt_count: expected pkt count increase
        """
        for gress in (INGRESS, EGRESS):
            for dir in (UPSTREAM, DOWNSTREAM):
                for typ in (PKTS, BYTES):
                    key = ctr_key(gress, dir, idx, typ)
                    if dir == pkt_direction:
                        # Increase should be non-zero
                        increase_bytes = ig_bytes if dir == UPSTREAM else eg_bytes
                        increase = increase_bytes if typ == BYTES else pkt_count
                    else:
                        increase = 0
                    expected = pre_counters[key] + increase
                    actual = post_counters[key]

                    self.assertEqual(
                        expected, actual,
                        "Invalid count for %s, expected %s but got %s"
                        % (key, expected, actual))


@group('accounting')
class UpstreamPppoeIp4UnicastTest(AccountingTest, upstream.PppoeIp4UnicastTest):
    """Tests counters for PPPoE IPv4 upstream traffic. Uses
    upstream.PppoeIp4UnicastTest as the base class for packet testing, but
    verifies that counters get incremented as expected.
    """

    def runTest(self):
        # Test with different type of packets.
        for pkt_type in ['tcp', 'udp', 'icmp']:
            print_inline('%s ... ' % pkt_type)
            pkt = getattr(testutils, 'simple_%s_packet' % pkt_type)()
            # TODO: add cos and account_id rules
            self.testPacketAndCounters(pkt, 0)

    def testPacketAndCounters(self, pkt, accounting_id):
        # Pkt's byte size at ingress, encapsulated with VLAN s-tag (4 bytes),
        # VLAN c-tag (4 bytes), and PPPoE (8 bytes). NOTE: self.testPacket()
        # performs encapsulation before sending pkt to switch.
        ig_bytes = len(pkt) + 4 + 4 + 8

        # Byte size at egress (decapsulated)
        # FIXME: switch to PSA to count post-encap/decap bytes.
        # Since we use v1model, byte counters in the egress pipe are
        # incremented with the same pkt size seen at ingress. See note in
        # bng.p4's EgressPipe. With PSA, eg_bytes should be the size of
        # the decapped pkt, i.e., len(pkt).
        eg_bytes = ig_bytes

        pre = self.read_counters(accounting_id)
        self.testPacket(pkt)
        post = self.read_counters(accounting_id)

        self.assert_counter_increase(
            pre_counters=pre, post_counters=post, idx=accounting_id,
            pkt_direction=UPSTREAM,
            ig_bytes=ig_bytes, eg_bytes=eg_bytes, pkt_count=1
        )
