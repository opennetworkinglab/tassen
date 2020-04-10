/*
 * Copyright 2020-present Open Networking Foundation
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

#include <core.p4>
#include <v1model.p4>

typedef bit<16> ethertype_t;
typedef bit<9>  port_t;
typedef bit<48> mac_addr_t;
typedef bit<12> vid_t;

const port_t CPU_PORT = 255;

const vid_t DEFAULT_VID = 0;

const ethertype_t ETHERTYPE_QINQ   = 0x88a8;
const ethertype_t ETHERTYPE_QINQ_2 = 0x9100;
const ethertype_t ETHERTYPE_VLAN   = 0x8100;
const ethertype_t ETHERTYPE_IPV4   = 0x0800;
const ethertype_t ETHERTYPE_PPPOED = 0x8863;
const ethertype_t ETHERTYPE_PPPOES = 0x8864;

const bit<8> PROTO_ICMP = 1;
const bit<8> PROTO_TCP = 6;
const bit<8> PROTO_UDP = 17;
const bit<8> PROTO_ICMPV6 = 58;

const bit<16> PPPOE_PROTO_IP4 = 0x21;

action nop() {
    NoAction();
}

@controller_header("packet_in") header cpu_in_t {
    port_t ingress_port;
    bit<7> _pad;
}

@controller_header("packet_out") header cpu_out_t {
    port_t egress_port;
    bit<7> _pad;
}

header ethernet_t {
    mac_addr_t dst_addr;
    mac_addr_t src_addr;
}

header eth_type_t {
    ethertype_t value;
}

header vlan_t {
    ethertype_t pid;
    bit<3>      pcp;
    bit<1>      dei;
    vid_t       vid;
}

header pppoe_t {
    bit<4>  ver;
    bit<4>  type;
    bit<8>  code;
    bit<16> session_id;
    bit<16> length;
    bit<16> proto;
}

header ipv4_t {
    bit<4>   ver;
    bit<4>   ihl;
    bit<6>   dscp;
    bit<2>   ecn;
    bit<16>  len;
    bit<16>  id;
    bit<3>   flags;
    bit<13>  frag_offset;
    bit<8>   ttl;
    bit<8>   proto;
    bit<16>  checksum;
    bit<32>  src_addr;
    bit<32>  dst_addr;
}

header tcp_t {
    bit<16> sport;
    bit<16> dport;
    bit<32> seq_no;
    bit<32> ack_no;
    bit<4>  data_offset;
    bit<3>  res;
    bit<3>  ecn;
    bit<6>  ctrl;
    bit<16> window;
    bit<16> checksum;
    bit<16> urgent_ptr;
}

header udp_t {
    bit<16> sport;
    bit<16> dport;
    bit<16> len;
    bit<16> checksum;
}

header icmp_t {
    bit<8>  icmp_type;
    bit<8>  icmp_code;
    bit<16> checksum;
    bit<16> identifier;
    bit<16> seq_number;
    bit<64> timestamp;
}

struct bng_meta_t {
    bit<2>    type;
    bit<32>   line_id;
    bit<16>   pppoe_session_id;
    bit<32>   ds_meter_result;
    vid_t     s_tag;
    vid_t     c_tag;
}

struct local_metadata_t {
    bit<16>      ip_eth_type;
    vid_t        vlan_vid;
    bit<3>       vlan_pcp;
    bit<1>       vlan_cfi;
    bool         push_double_vlan;
    vid_t        inner_vlan_vid;
    bit<3>       inner_vlan_pcp;
    bit<1>       inner_vlan_cfi;
    bool         skip_forwarding;
    bool         skip_next;
    // fwd_type_t   fwd_type;
    // next_id_t    next_id;
    bool         is_multicast;
    bool         is_controller_packet_out;
    bit<8>       ip_proto;
    bit<16>      l4_sport;
    bit<16>      l4_dport;
    // TODO: move top-level
    bng_meta_t   bng;
}

struct parsed_headers_t {
    ethernet_t  ethernet;
    // TODO: rename to c-tag and s-tag
    vlan_t      vlan;
    vlan_t      inner_vlan;
    eth_type_t  eth_type;
    pppoe_t     pppoe;
    ipv4_t      ipv4;
    tcp_t       tcp;
    udp_t       udp;
    icmp_t      icmp;
    cpu_out_t   cpu_out;
    cpu_in_t    cpu_in;
}

parser ParserImpl (
    packet_in packet,
    out parsed_headers_t hdr,
    inout local_metadata_t local_meta,
    inout standard_metadata_t std_meta) {

    state start {
        transition select(std_meta.ingress_port) {
            CPU_PORT: parse_cpu_out;
            default: parse_ethernet;
        }
    }

    state parse_cpu_out {
        packet.extract(hdr.cpu_out);
        transition parse_ethernet;
    }

    state parse_ethernet {
        packet.extract(hdr.ethernet);
        local_meta.vlan_vid = DEFAULT_VID;
        transition select(packet.lookahead<ethertype_t>()) {
            ETHERTYPE_QINQ: parse_vlan;
            ETHERTYPE_QINQ_2: parse_vlan;
            ETHERTYPE_VLAN: parse_vlan;
            default: parse_eth_type;
        }
    }

    state parse_vlan {
        packet.extract(hdr.vlan);
        local_meta.bng.s_tag = hdr.vlan.vid;
        transition select(packet.lookahead<ethertype_t>()) {
            ETHERTYPE_VLAN: parse_inner_vlan;
            default: parse_eth_type;
        }
    }

    state parse_inner_vlan {
        packet.extract(hdr.inner_vlan);
        local_meta.bng.c_tag = hdr.inner_vlan.vid;
        transition parse_eth_type;
    }

    state parse_eth_type {
        packet.extract(hdr.eth_type);
        transition select(hdr.eth_type.value) {
            ETHERTYPE_IPV4: parse_ipv4;
            ETHERTYPE_PPPOED: parse_pppoe;
            ETHERTYPE_PPPOES: parse_pppoe;
            default: accept;
        }
    }

    state parse_pppoe {
        packet.extract(hdr.pppoe);
        transition select(hdr.pppoe.proto) {
            PPPOE_PROTO_IP4: parse_ipv4;
            default: accept;
        }
    }

    state parse_ipv4 {
        packet.extract(hdr.ipv4);
        local_meta.ip_proto = hdr.ipv4.proto;
        local_meta.ip_eth_type = ETHERTYPE_IPV4;
        transition select(hdr.ipv4.proto) {
            PROTO_TCP: parse_tcp;
            PROTO_UDP: parse_udp;
            PROTO_ICMP: parse_icmp;
            default: accept;
        }
    }

    state parse_tcp {
        packet.extract(hdr.tcp);
        local_meta.l4_sport = hdr.tcp.sport;
        local_meta.l4_dport = hdr.tcp.dport;
        transition accept;
    }

    state parse_udp {
        packet.extract(hdr.udp);
        local_meta.l4_sport = hdr.udp.sport;
        local_meta.l4_dport = hdr.udp.dport;
        transition select(hdr.udp.dport) {
            default: accept;
        }
    }

    state parse_icmp {
        packet.extract(hdr.icmp);
        transition accept;
    }
}

control IngreePipeImpl(
    inout parsed_headers_t hdr,
    inout local_metadata_t local_meta,
    inout standard_metadata_t std_meta) {

    apply {
        nop();
    }
}

control EgressPipeImpl(
    inout parsed_headers_t hdr,
    inout local_metadata_t local_meta,
    inout standard_metadata_t std_meta) {

    apply {
        nop();
    }
}

control ComputeChecksumImpl(
    inout parsed_headers_t hdr,
    inout local_metadata_t meta) {

    apply {
        update_checksum(
            hdr.ipv4.isValid(),
            {
                hdr.ipv4.ver,
                hdr.ipv4.ihl,
                hdr.ipv4.dscp,
                hdr.ipv4.ecn,
                hdr.ipv4.len,
                hdr.ipv4.id,
                hdr.ipv4.flags,
                hdr.ipv4.frag_offset,
                hdr.ipv4.ttl,
                hdr.ipv4.proto,
                hdr.ipv4.src_addr,
                hdr.ipv4.dst_addr
            },
            hdr.ipv4.checksum,
            HashAlgorithm.csum16);
    }
}

control VerifyChecksumImpl(
    inout parsed_headers_t hdr,
    inout local_metadata_t meta) {

    apply {
        verify_checksum(
            hdr.ipv4.isValid(),
            {
                hdr.ipv4.ver,
                hdr.ipv4.ihl,
                hdr.ipv4.dscp,
                hdr.ipv4.ecn,
                hdr.ipv4.len,
                hdr.ipv4.id,
                hdr.ipv4.flags,
                hdr.ipv4.frag_offset,
                hdr.ipv4.ttl,
                hdr.ipv4.proto,
                hdr.ipv4.src_addr,
                hdr.ipv4.dst_addr
            },
            hdr.ipv4.checksum,
            HashAlgorithm.csum16);
    }
}

control DeparserImpl(
    packet_out packet,
    in parsed_headers_t hdr) {

    apply {
        packet.emit(hdr.cpu_in);
        packet.emit(hdr.ethernet);
        packet.emit(hdr.vlan);
        packet.emit(hdr.inner_vlan);
        packet.emit(hdr.eth_type);
        packet.emit(hdr.pppoe);
        packet.emit(hdr.ipv4);
        packet.emit(hdr.tcp);
        packet.emit(hdr.udp);
        packet.emit(hdr.icmp);
    }
}

V1Switch(
    ParserImpl(),
    VerifyChecksumImpl(),
    IngreePipeImpl(),
    EgressPipeImpl(),
    ComputeChecksumImpl(),
    DeparserImpl()) main;
