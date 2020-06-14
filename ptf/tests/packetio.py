# Copyright 2019-present Open Networking Foundation
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#   http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

# ------------------------------------------------------------------------------
# CONTROL PLANE PACKET-IN/OUT TESTS
#
# To run all tests in this file:
#     make check TEST=packetio
#
# To run a specific test case:
#     make check TEST=packetio.<TEST CLASS NAME>
#
# For example:
#     make check TEST=packetio.PacketOutTest
# ------------------------------------------------------------------------------

from base_test import *
from ptf.testutils import group
from scapy.layers.ppp import PPPoED, Ether

CPU_CLONE_SESSION_ID = 99

PPPOED_CODE_PADI = 0x09
PPPOED_CODE_PADO = 0x07
PPPOED_CODE_PADR = 0x19
PPPOED_CODE_PADS = 0x65
PPPOED_CODE_PADT = 0xa7

PPPOED_CODES = (
    PPPOED_CODE_PADI,
    PPPOED_CODE_PADO,
    PPPOED_CODE_PADR,
    PPPOED_CODE_PADS,
    PPPOED_CODE_PADT,
)


@group("packetio")
class PacketOutTest(P4RuntimeTest):
    """Tests controller packet-out capability by sending PacketOut messages and
    expecting a corresponding packet on the output port set in the PacketOut
    metadata.
    """

    def runTest(self):
        for pkt_type in ["tcp", "udp", "icmp", "arp"]:
            print_inline("%s ... " % pkt_type)
            pkt = getattr(testutils, "simple_%s_packet" % pkt_type)()
            self.testPacket(pkt)

    def testPacket(self, pkt):
        for outport in [self.port1, self.port2]:
            # Build PacketOut message
            packet_out_msg = self.helper.build_packet_out(
                payload=str(pkt),
                metadata={
                    "egress_port": outport,
                    "_pad": 0
                })

            # Send message and expect packet on the given data plane port.
            self.send_packet_out(packet_out_msg)

            testutils.verify_packet(self, pkt, outport)

        # Make sure packet was forwarded only on the specified ports
        testutils.verify_no_other_packets(self)


@group("packetio")
class AclPacketInTest(P4RuntimeTest):
    """Tests controller packet-in capability by matching on the packet EtherType
    and cloning to the CPU port.
    """

    def runTest(self):
        for pkt_type in ["tcp", "udp", "icmp", "arp"]:
            print_inline("%s ... " % pkt_type)
            pkt = getattr(testutils, "simple_%s_packet" % pkt_type)()
            self.testPacket(pkt)

    @autocleanup
    def testPacket(self, pkt):

        # Insert ACL entry to match on the given eth_type and clone to CPU.
        eth_type = pkt[Ether].type
        self.insert(self.helper.build_table_entry(
            table_name="IngressPipe.acl.acls",
            match_fields={
                # Ternary match.
                "eth_type": (eth_type, 0xffff)
            },
            action_name="IngressPipe.acl.punt",
            priority=DEFAULT_PRIORITY
        ))

        for inport in [self.port1, self.port2, self.port3]:
            # Expected P4Runtime PacketIn message.
            exp_packet_in_msg = self.helper.build_packet_in(
                payload=str(pkt),
                metadata={
                    "ingress_port": inport,
                    "_pad": 0
                })

            # Send packet to given switch ingress port and expect P4Runtime
            # PacketIn message.
            testutils.send_packet(self, inport, str(pkt))
            self.verify_packet_in(exp_packet_in_msg)


# TODO: add test for LCP, IPCP, CHAP/PAP, keep-alive control plane packets
@group("packetio")
class PppoePuntTest(P4RuntimeTest):
    """Tests controller packet-in capability by matching PPPoE Control Plane packets
    """

    def runTest(self):
        packets = {"PADI": Ether(src="00:11:22:33:44:55", dst="FF:FF:FF:FF:FF:FF") /
                           PPPoED(version=1, type=1, code=PPPOED_CODE_PADI) /
                           "dummy pppoed payload",
                   "PADR": Ether(src="00:11:22:33:44:55", dst="AA:BB:CC:DD:EE:FF") /
                           PPPoED(version=1, type=1, code=PPPOED_CODE_PADR) /
                           "dummy pppoed payload",
                   }

        for pkt_type, pkt in packets.items():
            print_inline("%s ... " % pkt_type)
            self.testPacket(pkt)

    @autocleanup
    def testPacket(self, pkt):

        # Set Access Interface
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

        self.insert(self.helper.build_table_entry(
            table_name='IngressPipe.my_stations',
            match_fields={
                'port': self.port1,
                'eth_dst': pkt[Ether].dst,
            },
            action_name='IngressPipe.set_my_station'
        ))

        for pppoe_code in PPPOED_CODES:
            self.insert(self.helper.build_table_entry(
                table_name="IngressPipe.upstream.pppoe_punts",
                match_fields={
                    "pppoe_code": stringify(pppoe_code, 1),
                },
                action_name="IngressPipe.upstream.punt",
                priority=DEFAULT_PRIORITY
            ))

        testutils.send_packet(self, self.port1, str(pkt))
        exp_packet_in_msg = self.helper.build_packet_in(
            payload=str(pkt),
            metadata={
                "ingress_port": self.port1,
                "_pad": 0
            })
        self.verify_packet_in(exp_packet_in_msg)
