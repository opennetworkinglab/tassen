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

CPU_CLONE_SESSION_ID = 99


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
        for port in [self.port1, self.port2]:
            # Build PacketOut message
            packet_out_msg = self.helper.build_packet_out(
                payload=str(pkt),
                metadata={
                    "egress_port": port,
                    "_pad": 0
                })

            # Send the P4RT message and expect the packet on the given port.
            self.send_packet_out(packet_out_msg)

            testutils.verify_packet(self, pkt, port)

        # Make sure packet came out only on the specified ports
        testutils.verify_no_other_packets(self)


@group("packetio")
class PacketInTest(P4RuntimeTest):
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

        # Insert ACL entry to match on the given eth_type and punt to CPU.
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

        for inport in (self.port1, self.port2, self.port3):
            # Expected P4Runtime PacketIn message.
            exp_packet_in_msg = self.helper.build_packet_in(
                payload=str(pkt),
                metadata={
                    "ingress_port": inport,
                    "_pad": 0
                })

            # Send packet to given switch ingress port and expect P4RT
            # PacketIn message.
            testutils.send_packet(self, inport, str(pkt))
            self.verify_packet_in(exp_packet_in_msg)
