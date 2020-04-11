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
# CONTROL PLANE PACKET-IN/OUT TESTS
#
# To run all tests in this file:
#     make check TEST=upstream
#
# To run a specific test case:
#     make check TEST=upstream.<TEST CLASS NAME>
#
# For example:
#     make check TEST=upstream.PacketOutTest
# ------------------------------------------------------------------------------

from base_test import *
from ptf.testutils import group


@group('upstream')
class Ip4UnicastTest(P4RuntimeTest):
    """Tests upstream termination and routing of IPv4 unicast packets.
    """

    def runTest(self):
        # Test with different type of packets.
        for pkt_type in ['tcp', 'udp', 'icmp']:
            print_inline('%s ... ' % pkt_type)
            pkt = getattr(testutils, 'simple_%s_packet' % pkt_type)()
            self.testPacket(pkt)

    @autocleanup
    def testPacket(self, pkt):
        next_hop_mac = CORE_MAC
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
            table_name='IngressPipe.upstream.lines',
            match_fields={
                'c_tag': c_tag,
                's_tag': s_tag
            },
            action_name='IngressPipe.upstream.set_line',
            action_params={'line_id': line_id}
        ))

        self.insert(self.helper.build_table_entry(
            table_name='IngressPipe.upstream.attachments_v4',
            match_fields={
                'line_id': line_id,
                'eth_src': pkt[Ether].src,
                'ipv4_src': pkt[IP].src,
                'pppoe_sess_id': pppoe_sess_id
            },
            action_name='nop'
        ))

        self.insert(self.helper.build_act_prof_group(
            act_prof_name="IngressPipe.upstream.ecmp",
            group_id=line_id,
            actions=[
                ('IngressPipe.upstream.route_v4',
                    {'dmac': next_hop_mac, 'port': self.port2}),
            ]
        ))

        # Insert routing entry
        self.insert(self.helper.build_table_entry(
            table_name='IngressPipe.upstream.routes_v4',
            match_fields={
                # LPM match (value, prefix)
                'ipv4_dst': (pkt[IP].dst, 32)
            },
            group_id=line_id
        ))

        # Transform the given input packet as it would be transmitted out of an ONU.
        pppoe_pkt = pkt_add_pppoe(
            pkt, code=PPPOE_CODE_SESSION_STAGE, session_id=pppoe_sess_id)
        pppoe_pkt = pkt_add_vlan(pppoe_pkt, vid=c_tag)
        pppoe_pkt = pkt_add_vlan(pppoe_pkt, vid=s_tag)

        # Expected pkt should have routed MAC addresses and decremented TTL.
        exp_pkt = pkt.copy()
        pkt_route(exp_pkt, next_hop_mac)
        pkt_decrement_ttl(exp_pkt)

        testutils.send_packet(self, self.port1, str(pppoe_pkt))
        testutils.verify_packet(self, exp_pkt, self.port2)
