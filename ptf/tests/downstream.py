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
# DOWNSTREAM TESTS
#
# To run all tests in this file:
#     make check TEST=downstream
#
# To run a specific test case:
#     make check TEST=downstream.<TEST CLASS NAME>
#
# For example:
#     make check TEST=downstream.PacketOutTest
# ------------------------------------------------------------------------------

from base_test import *
from ptf.testutils import group


@group('downstream')
class PppoeIp4UnicastTest(P4RuntimeTest):
    """Tests downstream PPPoE aggregation and routing of IPv4 unicast packets.
    """

    def runTest(self):
        # Test with different type of packets.
        for pkt_type in ['tcp', 'udp', 'icmp']:
            print_inline('%s ... ' % pkt_type)
            pkt = getattr(testutils, 'simple_%s_packet' % pkt_type)()
            self.testPacket(pkt)

    @autocleanup
    def testPacket(self, pkt):
        next_hop_mac = HOST1_MAC
        c_tag = 10
        s_tag = 20
        line_id = 100
        pppoe_sess_id = 90

        self.insert(self.helper.build_table_entry(
            table_name='IngressPipe.if_types',
            match_fields={
                'port': self.port1
            },
            action_name='IngressPipe.set_if_type',
            action_params={
                'if_type': IF_CORE,
            }
        ))

        self.insert(self.helper.build_table_entry(
            table_name='IngressPipe.if_types',
            match_fields={
                'port': self.port2
            },
            action_name='IngressPipe.set_if_type',
            action_params={
                'if_type': IF_ACCESS,
            }
        ))

        # Consider the given pkt's eth dst addr
        # as the bng mac.
        self.insert(self.helper.build_table_entry(
            table_name='IngressPipe.my_stations',
            match_fields={
                'port': self.port1,
                'eth_dst': pkt[Ether].dst,
            },
            action_name='IngressPipe.set_my_station'
        ))
        self.insert(self.helper.build_table_entry(
            table_name='IngressPipe.my_stations',
            match_fields={
                'port': self.port2,
                'eth_dst': pkt[Ether].dst,
            },
            action_name='IngressPipe.set_my_station'
        ))

        self.insert(self.helper.build_table_entry(
            table_name='IngressPipe.downstream.lines_v4',
            match_fields={
                'ipv4_dst': pkt[IP].dst
            },
            action_name='IngressPipe.downstream.set_line',
            action_params={'line_id': line_id}
        ))

        self.insert(self.helper.build_table_entry(
            table_name='IngressPipe.downstream.attachments_v4',
            match_fields={
                'line_id': line_id
            },
            action_name='IngressPipe.downstream.set_pppoe_attachment_v4',
            action_params={
                'port': self.port2,
                'dmac': next_hop_mac,
                's_tag': s_tag,
                'c_tag': c_tag,
                'pppoe_sess_id': pppoe_sess_id,
            }
        ))

        # Transform the given input packet as it would be transmitted out of an ONU.
        pppoe_pkt = pkt_add_pppoe(
            pkt, code=PPPOE_CODE_SESSION_STAGE, session_id=pppoe_sess_id)
        pppoe_pkt = pkt_add_vlan(pppoe_pkt, vid=c_tag)
        pppoe_pkt = pkt_add_vlan(pppoe_pkt, vid=s_tag)

        # Expected pkt should have vlan tags, PPPoE header, routed MAC addresses
        # and decremented TTL.
        exp_pkt = pkt.copy()
        exp_pkt = pkt_add_pppoe(
            exp_pkt, code=PPPOE_CODE_SESSION_STAGE, session_id=pppoe_sess_id)
        exp_pkt = pkt_add_vlan(exp_pkt, vid=c_tag)
        exp_pkt = pkt_add_vlan(exp_pkt, vid=s_tag)
        pkt_route(exp_pkt, next_hop_mac)
        pkt_decrement_ttl(exp_pkt)

        testutils.send_packet(self, self.port1, str(pkt))
        testutils.verify_packet(self, exp_pkt, self.port2)
