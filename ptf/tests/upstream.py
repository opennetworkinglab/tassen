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


class BngTest(P4RuntimeTest):
    next_up_ecmp_id = 0

    def new_upstream_ecmp_id(self):
        self.next_up_ecmp_id = self.next_up_ecmp_id + 1
        return self.next_up_ecmp_id

    def set_if_type(self, port, if_type):
        self.insert(self.helper.build_table_entry(
            table_name='IngressPipe.if_types',
            match_fields={
                'port': port
            },
            action_name='IngressPipe.set_if_type',
            action_params={
                'if_type': if_type,
            }
        ))

    def set_my_station(self, port, mac_addr):
        self.insert(self.helper.build_table_entry(
            table_name='IngressPipe.my_stations',
            match_fields={
                'port': port,
                'eth_dst': mac_addr,
            },
            action_name='IngressPipe.set_my_station'
        ))

    def set_upstream_line(self, c_tag, s_tag, line_id):
        self.insert(self.helper.build_table_entry(
            table_name='IngressPipe.upstream.lines',
            match_fields={
                'c_tag': c_tag,
                's_tag': s_tag
            },
            action_name='IngressPipe.upstream.set_line',
            action_params={'line_id': line_id}
        ))

    def set_upstream_attachment_v4(self, line_id, eth_src, ipv4_src,
                                   pppoe_sess_id):
        self.insert(self.helper.build_table_entry(
            table_name='IngressPipe.upstream.attachments_v4',
            match_fields={
                'line_id': line_id,
                'eth_src': eth_src,
                'ipv4_src': ipv4_src,
                'pppoe_sess_id': pppoe_sess_id
            },
            action_name='nop'
        ))

    def set_upstream_route_v4(self, ipv4_addr, prefix_len,
                              nexthop_mac_to_ports):
        assert len(ipv4_addr)
        assert prefix_len > 0
        assert len(nexthop_mac_to_ports)

        actions = []
        for mac, port in nexthop_mac_to_ports.items():
            # Each member in the group is a tuple: (action_name, action_params)
            actions.append(
                ('IngressPipe.upstream.route_v4', {'dmac': mac, 'port': port}))

        group_id = self.new_upstream_ecmp_id()

        self.insert(self.helper.build_act_prof_group(
            act_prof_name="IngressPipe.upstream.ecmp",
            group_id=group_id,
            actions=actions
        ))

        self.insert(self.helper.build_table_entry(
            table_name='IngressPipe.upstream.routes_v4',
            match_fields={
                # LPM match (value, prefix)
                'ipv4_dst': (ipv4_addr, prefix_len)
            },
            group_id=group_id
        ))


@group('upstream')
class PppoeIp4UnicastTest(BngTest):
    """Tests upstream PPPoE termination and routing of IPv4 unicast packets.
    """

    def runTest(self):
        # Test with different type of packets.
        for pkt_type in ['tcp', 'udp', 'icmp']:
            print_inline('%s ... ' % pkt_type)
            pkt = getattr(testutils, 'simple_%s_packet' % pkt_type)()
            self.testPacket(pkt)

    @autocleanup
    def testPacket(self, pkt):
        nexthop_mac = CORE_MAC
        c_tag = 10
        s_tag = 20
        line_id = 100
        pppoe_sess_id = 90

        self.set_if_type(self.port1, IF_ACCESS)
        self.set_if_type(self.port2, IF_CORE)
        self.set_my_station(self.port1, pkt[Ether].dst)

        self.set_upstream_line(
            c_tag=c_tag,
            s_tag=s_tag,
            line_id=line_id)
        self.set_upstream_attachment_v4(
            line_id=line_id,
            eth_src=pkt[Ether].src,
            ipv4_src=pkt[IP].src,
            pppoe_sess_id=pppoe_sess_id)
        self.set_upstream_route_v4(
            ipv4_addr=pkt[IP].dst,
            prefix_len=32,
            nexthop_mac_to_ports={nexthop_mac: self.port2})

        # Transform the given input pkt as it would be transmitted out of an ONU.
        pppoe_pkt = pkt_add_pppoe(
            pkt, code=PPPOE_CODE_SESSION_STAGE, session_id=pppoe_sess_id)
        pppoe_pkt = pkt_add_vlan(pppoe_pkt, vid=c_tag)
        pppoe_pkt = pkt_add_vlan(pppoe_pkt, vid=s_tag)

        # The expected pkt should have routed MAC addresses and decremented TTL.
        exp_pkt = pkt.copy()
        pkt_route(exp_pkt, nexthop_mac)
        pkt_decrement_ttl(exp_pkt)

        testutils.send_packet(self, self.port1, str(pppoe_pkt))
        testutils.verify_packet(self, exp_pkt, self.port2)
