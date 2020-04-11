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

const int MAX_LINES = 8192;
const int MAX_ATTACH_PER_LINE = 4;
const int MAX_UPSTREAM_ROUTES = 1024;
// For some reason P4 constants cannot be used in annotations.
#define MAX_ECMP_GROUP_SIZE 16
const int MAX_PPPOE_PUNTS = 32;
const int MAX_ACLS = 256;

const bit<12> DEFAULT_VID = 0;

const bit<16> ETHERTYPE_QINQ   = 0x88a8;
const bit<16> ETHERTYPE_QINQ2 = 0x9100;
const bit<16> ETHERTYPE_VLAN   = 0x8100;
const bit<16> ETHERTYPE_IPV4   = 0x0800;
const bit<16> ETHERTYPE_PPPOED = 0x8863;
const bit<16> ETHERTYPE_PPPOES = 0x8864;

const bit<8> IP_PROTO_ICMP = 1;
const bit<8> IP_PROTO_TCP = 6;
const bit<8> IP_PROTO_UDP = 17;
const bit<8> IP_PROTO_ICMPV6 = 58;

const bit<16> PPPOE_PROTO_IP4 = 0x21;

typedef bit<9>  port_t;
const port_t CPU_PORT = 255;
const bit<32> CPU_CLONE_SESSION_ID = 99;

typedef bit<3>  if_type_t;
const if_type_t IF_UNKNOWN = 0;
const if_type_t IF_CORE = 1;
const if_type_t IF_ACCESS = 2;

typedef bit<32> line_id_t;
const line_id_t LINE_UNKNOWN = 0;

action nop() { NoAction(); }

action drop_now(inout standard_metadata_t smeta) {
    // Exit the pipeline now and drop.
    mark_to_drop(smeta);
    exit;
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
    bit<48> dst_addr;
    bit<48> src_addr;
}

header eth_type_t {
    bit<16> value;
}

header vlan_t {
    bit<16> pid;
    bit<3>  pcp;
    bit<1>  dei;
    bit<12> vid;
}

header pppoe_t {
    bit<4>  ver;
    bit<4>  type;
    bit<8>  code;
    bit<16> sess_id;
    bit<16> length;
    bit<16> proto;
}

header ipv4_t {
    bit<4>  ver;
    bit<4>  ihl;
    bit<6>  dscp;
    bit<2>  ecn;
    bit<16> len;
    bit<16> id;
    bit<3>  flags;
    bit<13> frag_offset;
    bit<8>  ttl;
    bit<8>  proto;
    bit<16> checksum;
    bit<32> src_addr;
    bit<32> dst_addr;
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

struct local_metadata_t {
    if_type_t  if_type;
    bit<48>    my_mac;
    bit<8>     ip_proto;
    bit<16>    l4_sport;
    bit<16>    l4_dport;
    line_id_t  line_id;
    bit<12>    s_tag;
    bit<12>    c_tag;
    bit<16>    pppoe_sess_id;
}

struct parsed_headers_t {
    ethernet_t  ethernet;
    vlan_t      vlan;
    vlan_t      vlan2;
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
    inout local_metadata_t lmeta,
    inout standard_metadata_t smeta) {

    state start {
        transition select(smeta.ingress_port) {
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
        transition select(packet.lookahead<bit<16>>()) {
            ETHERTYPE_QINQ: parse_vlan;
            ETHERTYPE_QINQ2: parse_vlan;
            ETHERTYPE_VLAN: parse_vlan;
            default: parse_eth_type;
        }
    }

    state parse_vlan {
        packet.extract(hdr.vlan);
        lmeta.s_tag = hdr.vlan.vid;
        transition select(packet.lookahead<bit<16>>()) {
            ETHERTYPE_VLAN: parse_vlan2;
            default: parse_eth_type;
        }
    }

    state parse_vlan2 {
        packet.extract(hdr.vlan2);
        lmeta.c_tag = hdr.vlan2.vid;
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
        lmeta.ip_proto = hdr.ipv4.proto;
        transition select(hdr.ipv4.proto) {
            IP_PROTO_TCP: parse_tcp;
            IP_PROTO_UDP: parse_udp;
            IP_PROTO_ICMP: parse_icmp;
            default: accept;
        }
    }

    state parse_tcp {
        packet.extract(hdr.tcp);
        lmeta.l4_sport = hdr.tcp.sport;
        lmeta.l4_dport = hdr.tcp.dport;
        transition accept;
    }

    state parse_udp {
        packet.extract(hdr.udp);
        lmeta.l4_sport = hdr.udp.sport;
        lmeta.l4_dport = hdr.udp.dport;
        transition select(hdr.udp.dport) {
            default: accept;
        }
    }

    state parse_icmp {
        packet.extract(hdr.icmp);
        transition accept;
    }
}

control TtlCheck(
    inout parsed_headers_t hdr,
    inout local_metadata_t lmeta,
    inout standard_metadata_t smeta) {

    counter(MAX_LINES, CounterType.packets_and_bytes) expired;

    apply {
        if (hdr.ipv4.isValid()) {
            if (hdr.ipv4.ttl > 1) {
                hdr.ipv4.ttl = hdr.ipv4.ttl - 1;
            } else {
                expired.count(lmeta.line_id);
                drop_now(smeta);
            }
        }
    }
}

control IngressUpstream(
    inout parsed_headers_t hdr,
    inout local_metadata_t lmeta,
    inout standard_metadata_t smeta) {
    
    counter(MAX_LINES, CounterType.packets_and_bytes) all;
    counter(MAX_LINES, CounterType.packets_and_bytes) punted;
    counter(MAX_LINES, CounterType.packets_and_bytes) spoofed;
    counter(MAX_LINES, CounterType.packets_and_bytes) routed;
    
    action set_line(bit<32> line_id) {
        lmeta.line_id = line_id;
    }

    table lines {
        key = {
            lmeta.c_tag: exact @name("c_tag") ;
            lmeta.s_tag: exact @name("s_tag") ;
        }
        actions = {
            set_line;
        }
        size = MAX_LINES;
        const default_action = set_line(LINE_UNKNOWN);
    }

    action punt() {
        smeta.egress_spec = CPU_PORT;
        punted.count(lmeta.line_id);
        exit;
    }

    table pppoe_punts {
        key = {
            hdr.pppoe.code  : exact @name("pppoe_code");
            hdr.pppoe.proto : ternary @name("pppoe_proto");
        }
        actions = {
            punt;
            @defaultonly nop;
        }
        size = MAX_PPPOE_PUNTS;
        const default_action = nop;
    }

    action reject() {
        spoofed.count(lmeta.line_id);
        drop_now(smeta);
    }

    table attachments_v4 {
        key = {
            lmeta.line_id         : exact @name("line_id");
            hdr.ethernet.src_addr : exact @name("eth_src");
            hdr.ipv4.src_addr     : exact @name("ipv4_src");
            hdr.pppoe.sess_id     : ternary @name("pppoe_sess_id");
        }
        actions = {
            nop;
            @defaultonly reject;
        }
        size = MAX_ATTACH_PER_LINE * MAX_LINES;
        const default_action = reject;
    }

    @hidden
    action decap(bit<16> eth_type) {
        hdr.eth_type.value = eth_type;
        hdr.vlan.setInvalid();
        hdr.vlan2.setInvalid();
        hdr.pppoe.setInvalid();
    }

    @hidden
    action route(port_t port, bit<48> dmac) {
        smeta.egress_spec = port;
        hdr.ethernet.src_addr = lmeta.my_mac;
        hdr.ethernet.dst_addr = dmac;
        routed.count(lmeta.line_id);
    }

    action route_v4(port_t port, bit<48> dmac) {
        decap(ETHERTYPE_IPV4);
        route(port, dmac);
    }

    table routes_v4 {
        key = {
            hdr.ipv4.dst_addr : lpm @name("ipv4_dst");
            hdr.ipv4.dst_addr : selector;
            hdr.ipv4.src_addr : selector;
            lmeta.ip_proto    : selector;
            lmeta.l4_sport    : selector;
            lmeta.l4_dport    : selector;
        }
        actions = {
            route_v4;
            @defaultonly nop;
        }
        default_action = nop();
        @name("ecmp_up")
        @max_group_size(MAX_ECMP_GROUP_SIZE)
        implementation = action_selector(HashAlgorithm.crc16, 32w1024, 32w16);
        size = MAX_UPSTREAM_ROUTES;
    }

    TtlCheck() ttl;

    apply { 
        lines.apply();
        all.count(lmeta.line_id);
        pppoe_punts.apply();
        if (lmeta.line_id == LINE_UNKNOWN) {
            drop_now(smeta);
        }
        // Line is known and pkt was not punted.
        // Verify attachment info, if valid (not spoofed), decap and route.
        // If no route then no-op, we might punt or do something else in ACL.
        if (hdr.ipv4.isValid()) {
            attachments_v4.apply();
            routes_v4.apply();
        }
        ttl.apply(hdr, lmeta, smeta);
    }
}

control IngressDownstream(
    inout parsed_headers_t hdr,
    inout local_metadata_t lmeta,
    inout standard_metadata_t smeta) {
    
    counter(MAX_LINES, CounterType.packets_and_bytes) dropped;
    counter(MAX_LINES, CounterType.packets_and_bytes) routed;

    // In downstream, we expect all tables to be matched.
    // Any table miss will cause the packet to be dropped
    // immediately.
    action miss() {
        dropped.count(lmeta.line_id);
        drop_now(smeta);
    }

    action set_line(bit<32> line_id) {
        lmeta.line_id = line_id;
    }

    table lines_v4 {
        key = {
            hdr.ipv4.dst_addr: exact @name("ipv4_dst");
        }
        actions = {
            set_line;
            @defaultonly miss;
        }
        size = MAX_LINES;
        const default_action = miss;
    }

    action set_vids(bit<12> c_tag, bit<12> s_tag) {
        lmeta.c_tag = c_tag;
        lmeta.s_tag = s_tag;
    }

    table vids {
        key = {
            lmeta.line_id: exact @name("ipv4_dst");
        }
        actions = {
            set_vids;
            @defaultonly miss;
        }
        size = MAX_LINES;
        const default_action = miss;
    }

    action set_pppoe_sess(bit<16> pppoe_sess_id) {
        lmeta.pppoe_sess_id = pppoe_sess_id;
    }

    table pppoe_sessions {
        key = {
            lmeta.line_id: exact @name("line_id");
        }
        actions = {
            set_pppoe_sess;
            @defaultonly miss;
        }
        size = MAX_LINES;
        const default_action = miss;
    }

    action route_v4(port_t port, bit<48> dmac) {
        smeta.egress_spec  = port;
        hdr.ethernet.src_addr = lmeta.my_mac;
        hdr.ethernet.dst_addr = dmac;
        routed.count(lmeta.line_id);
    }

    table routes_v4 {
        key = {
            lmeta.line_id     : lpm @name("line_id");
            hdr.ipv4.dst_addr : selector;
            hdr.ipv4.src_addr : selector;
            lmeta.ip_proto    : selector;
            lmeta.l4_sport    : selector;
            lmeta.l4_dport    : selector;
        }
        actions = {
            route_v4;
            @defaultonly miss;
        }
        default_action = miss;
        @name("ecmp_down")
        @max_group_size(MAX_ECMP_GROUP_SIZE)
        implementation = action_selector(HashAlgorithm.crc16, 32w1024, 32w16);
        size = MAX_LINES;
    }

    TtlCheck() ttl;

    apply {
        lmeta.line_id = LINE_UNKNOWN;
        lines_v4.apply();
        vids.apply();
        pppoe_sessions.apply();
        routes_v4.apply();
        ttl.apply(hdr, lmeta, smeta);
    }
}

control Acl(
    inout parsed_headers_t hdr,
    inout local_metadata_t lmeta,
    inout standard_metadata_t smeta) {

    action set_port(port_t port) {
        smeta.egress_spec = port;
    }

    action punt() {
        set_port(CPU_PORT);
    }

    // FIXME: what's the right way of cloning in v1model?
    // action clone_to_cpu() {
    //     clone3(CloneType.I2E, CPU_CLONE_SESSION_ID, { smeta.ingress_port });
    // }

    action drop() {
        mark_to_drop(smeta);
    }

    table acls {
        key = {
            smeta.ingress_port    : exact @name("port");
            lmeta.if_type         : exact @name("if_type");
            hdr.ethernet.src_addr : exact @name("eth_src");
            hdr.ethernet.dst_addr : exact @name("eth_dst");
            hdr.ipv4.src_addr     : exact @name("ipv4_src");
            hdr.ipv4.dst_addr     : exact @name("ipv4_dst");
            hdr.ipv4.proto        : exact @name("ipv4_proto");
            lmeta.l4_sport        : exact @name("l4_sport");
            lmeta.l4_dport        : exact @name("l4_dport");
        }
        actions = {
            set_port;
            punt;
            // clone_to_cpu;
            drop;
            nop;
        }
        const default_action = nop;
        @name("acls")
        counters = direct_counter(CounterType.packets_and_bytes);
        size = MAX_ACLS;
    }

    apply {
        acls.apply();
    }
}

control IngressPipe(
    inout parsed_headers_t hdr,
    inout local_metadata_t lmeta,
    inout standard_metadata_t smeta) {

    action set_if_type(if_type_t if_type) {
        lmeta.if_type = if_type;
    }

    table if_types {
        key = {
            smeta.ingress_port : exact @name("port");
        }
        actions = {
            set_if_type();
        }
        const default_action = set_if_type(IF_UNKNOWN);
        @name("if_types")
        counters = direct_counter(CounterType.packets_and_bytes);
        size = 1024;
    }

    action set_my_station() {
        lmeta.my_mac = hdr.ethernet.dst_addr;
    }

    table my_stations {
        key = {
            smeta.ingress_port    : exact @name("port");
            hdr.ethernet.dst_addr : exact @name("eth_dst");
        }
        actions = { nop; }
        const default_action = nop;
        @name("my_stations")
        counters = direct_counter(CounterType.packets_and_bytes);
        size = 2048;
    }

    IngressUpstream() upstream;
    IngressDownstream() downstream;
    Acl() acl;

    apply {
        // Controller packet-out.
        if (hdr.cpu_out.isValid()) {
            smeta.egress_spec = hdr.cpu_out.egress_port;
            hdr.cpu_out.setInvalid();
            exit;
        }

        if_types.apply();

        if (my_stations.apply().hit) {
            if (lmeta.if_type == IF_ACCESS) {
                upstream.apply(hdr, lmeta, smeta);
            } else if (lmeta.if_type == IF_CORE) {
                downstream.apply(hdr, lmeta, smeta);
            }
        }

        acl.apply(hdr, lmeta, smeta);
    }
}

control EgressPipe(
    inout parsed_headers_t hdr,
    inout local_metadata_t local_meta,
    inout standard_metadata_t smeta) {

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
        packet.emit(hdr.vlan2);
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
    IngressPipe(),
    EgressPipe(),
    ComputeChecksumImpl(),
    DeparserImpl()) main;
