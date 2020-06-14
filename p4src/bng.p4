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
// FIXME: we agreed to get rid of line IDs in favor of attachment IDs, as such
//  allowing multiple attachments per line is no longer meaningful and should be
//  removed.
const int MAX_ATTACH_PER_LINE = 4;
const int MAX_UPSTREAM_ROUTES = 1024;
// For some reason P4 constants cannot be used in annotations.
#define MAX_ECMP_GROUP_SIZE 16
const int MAX_PPPOE_PUNTS = 32;
const int MAX_ACLS = 256;
const int MAX_COS = 16; // Classes of Service
const int MAX_ACCOUNTING_IDS = MAX_LINES * MAX_COS;

const bit<12> DEFAULT_VID = 0;

const bit<16> ETHERTYPE_QINQ   = 0x88a8;
const bit<16> ETHERTYPE_QINQ2  = 0x9100;
const bit<16> ETHERTYPE_VLAN   = 0x8100;
const bit<16> ETHERTYPE_IPV4   = 0x0800;
const bit<16> ETHERTYPE_PPPOED = 0x8863;
const bit<16> ETHERTYPE_PPPOES = 0x8864;

const bit<8> IP_PROTO_ICMP   = 1;
const bit<8> IP_PROTO_TCP    = 6;
const bit<8> IP_PROTO_UDP    = 17;
const bit<8> IP_PROTO_ICMPV6 = 58;

const bit<16> PPPOE_PROTO_IP4 = 0x21;

typedef bit<9>  port_t;
const port_t CPU_PORT = 255;
const bit<32> CPU_CLONE_SESSION_ID = 99;

typedef bit<3>  if_type_t;
const if_type_t IF_UNKNOWN = 0;
const if_type_t IF_CORE    = 1;
const if_type_t IF_ACCESS  = 2;

typedef bit<3> direction_t;
const direction_t DIR_UNKNOWN    = 0;
const direction_t DIR_UPSTREAM   = 1;
const direction_t DIR_DOWNSTREAM = 2;

typedef bit<32> line_id_t;
const line_id_t LINE_UNKNOWN = 0;

typedef bit<32> cos_id_t;
const line_id_t COS_UNKNOWN = 0;

typedef bit<32> accounting_id_t;
const line_id_t ACCOUNTING_UNKNOWN = 0;

// Global actions common to many controls.
action nop() { /* no-op */ }

action drop_now(inout standard_metadata_t smeta) {
    // Exit the pipeline now and drop.
    mark_to_drop(smeta);
    exit;
}

// Controller packet-in/out headers.
@controller_header("packet_in")
header cpu_in_t {
    port_t ingress_port;
    bit<7> _pad;
}

@controller_header("packet_out")
header cpu_out_t {
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
    bit<16> tpid;
    bit<3>  pcp;
    bit<1>  dei;
    bit<12> vid;
}

header pppoe_t {
    bit<4>  ver;
    bit<4>  type;
    bit<8>  code;
    bit<16> sess_id;
    bit<16> len;
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

// TODO: Add IPv6 support

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
    if_type_t        if_type;
    direction_t      direction;
    bit<48>          my_mac;
    bit<8>           ip_proto;
    // Used to normalize UDP/TCP ports.
    bit<16>          l4_sport;
    bit<16>          l4_dport;
    // Attachment attributes.
    line_id_t        line_id;
    cos_id_t         cos_id;
    accounting_id_t  accounting_id;
    bit<12>          s_tag;
    bit<12>          c_tag;
    bit<16>          pppoe_sess_id;
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

control Accounting(
    inout parsed_headers_t hdr,
    inout local_metadata_t lmeta,
    inout standard_metadata_t smeta) {

    // +1 is for ACCOUNTING_UNKNOWN (0)
    counter(MAX_ACCOUNTING_IDS+1, CounterType.packets_and_bytes) upstream;
    counter(MAX_ACCOUNTING_IDS+1, CounterType.packets_and_bytes) downstream;

    apply {
        if (lmeta.direction == DIR_UPSTREAM) {
            upstream.count(lmeta.accounting_id);
        } else if (lmeta.direction == DIR_DOWNSTREAM) {
            downstream.count(lmeta.accounting_id);
        }
    }
}

control CoS(
    inout parsed_headers_t hdr,
    inout local_metadata_t lmeta,
    inout standard_metadata_t smeta) {

    action set_cos_id(cos_id_t cos_id) {
        lmeta.cos_id = cos_id;
    }

    table services_v4 {
        key = {
            hdr.ipv4.src_addr     : ternary @name("ipv4_src");
            hdr.ipv4.dst_addr     : ternary @name("ipv4_dst");
            hdr.ipv4.proto        : ternary @name("ipv4_proto");
            lmeta.l4_sport        : range   @name("l4_sport");
            lmeta.l4_dport        : range   @name("l4_dport");
        }
        actions = {
            set_cos_id;
        }
        const default_action = set_cos_id(COS_UNKNOWN);
        // This is fine if we assume we classify traffic in the same way for all
        // attachments. If we expect to be supporting special classification
        // rules only for some attachments, then size should be something
        // different.
        size = MAX_COS;
    }

    apply {
        if (hdr.ipv4.isValid()) {
            services_v4.apply();
        } else {
            set_cos_id(COS_UNKNOWN);
        }
    }
}

control IngressUpstream(
    inout parsed_headers_t hdr,
    inout local_metadata_t lmeta,
    inout standard_metadata_t smeta) {

    // FIXME: do we still need all these counters now that we have proper
    //  accounting controls? Most of these counters seems useful for debugging
    //  only, we could use preprocessor flags to enable/disable them.
    counter(MAX_LINES, CounterType.packets_and_bytes) all;
    counter(MAX_LINES, CounterType.packets_and_bytes) punted;
    counter(MAX_LINES, CounterType.packets_and_bytes) spoofed;
    counter(MAX_LINES, CounterType.packets_and_bytes) routed;
    
    action set_line(bit<32> line_id) {
        lmeta.line_id = line_id;
    }

    table lines {
        key = {
            smeta.ingress_port  : exact @name("port");
            lmeta.c_tag         : exact @name("c_tag");
            lmeta.s_tag         : exact @name("s_tag");
            // If one needs more accounting granularirty, they could add more
            // fields here.
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

    // NOTE: we expect the control plane to populate this at runtime.
    // Should we use static entries instead? The PPPoE packets we want to punt
    // should always be the same.
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

    // Provides anti-spoofing.
    // NOTE: consider merging this table with lines to support arbitrary
    // aggregation of traffic into the same line.
    table attachments_v4 {
        key = {
            lmeta.line_id         : exact @name("line_id");
            hdr.ethernet.src_addr : exact @name("eth_src");
            hdr.ipv4.src_addr     : exact @name("ipv4_src");
            hdr.pppoe.sess_id     : exact @name("pppoe_sess_id");
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
            // The following header fields are used to computed the ECMP hash.
            // They're NOT part of the match key provided by the controller.
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
        // BUG: action profiles don't get a fully qualified name in P4Info.
        @name("IngressPipe.upstream.ecmp")
        @max_group_size(MAX_ECMP_GROUP_SIZE)
        implementation = action_selector(HashAlgorithm.crc16, 32w1024, 32w16);
        size = MAX_UPSTREAM_ROUTES;
    }

    CoS() cos;
    TtlCheck() ttl;

    apply { 
        lines.apply();
        all.count(lmeta.line_id);
        pppoe_punts.apply();
        if (lmeta.line_id == LINE_UNKNOWN) {
            drop_now(smeta);
        }
        cos.apply(hdr, lmeta, smeta);
        // Line is known and pkt was not punted. Verify attachment info, if
        // valid (not spoofed), decap and route. If no route then no-op, we
        // might punt or do something else in ACL.
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

    @hidden
    action encap(bit<16> pppoe_proto) {
        // Outer VLAN (s_tag)
        hdr.vlan.setValid();
        hdr.vlan.tpid = ETHERTYPE_VLAN;
        hdr.vlan.pcp  = 3w0;
        hdr.vlan.dei  = 1w0;
        hdr.vlan.vid  = lmeta.s_tag;
        // Inner VLAN (c_tag)
        hdr.vlan2.setValid();
        hdr.vlan2.tpid = ETHERTYPE_VLAN;
        hdr.vlan2.pcp  = 3w0;
        hdr.vlan2.dei  = 1w0;
        hdr.vlan2.vid  = lmeta.c_tag;
        // PPPoE
        hdr.eth_type.value = ETHERTYPE_PPPOES;
        hdr.pppoe.setValid();
        hdr.pppoe.ver     = 4w1;
        hdr.pppoe.type    = 4w1;
        hdr.pppoe.code    = 8w0; // session stage
        hdr.pppoe.sess_id = lmeta.pppoe_sess_id;
        hdr.pppoe.len     = hdr.ipv4.len + 16w2;
        hdr.pppoe.proto   = pppoe_proto;
    }

    @hidden
    action route(port_t port, bit<48> dmac) {
        smeta.egress_spec  = port;
        hdr.ethernet.src_addr = lmeta.my_mac;
        hdr.ethernet.dst_addr = dmac;
        routed.count(lmeta.line_id);
    }

    action set_pppoe_attachment_v4(
        port_t port, bit<48> dmac,
        bit<12> s_tag, bit<12> c_tag,
        bit<16> pppoe_sess_id)
    {
        lmeta.s_tag = s_tag;
        lmeta.c_tag = c_tag;
        lmeta.pppoe_sess_id = pppoe_sess_id;
        encap(PPPOE_PROTO_IP4);
        route(port, dmac);
    }

    table attachments_v4 {
        key = {
            lmeta.line_id: exact @name("line_id");
        }
        actions = {
            set_pppoe_attachment_v4;
            @defaultonly miss;
        }
        size = MAX_LINES;
        const default_action = miss;
    }

    CoS() cos;
    TtlCheck() ttl;

    apply {
        lmeta.line_id = LINE_UNKNOWN;
        lines_v4.apply();
        cos.apply(hdr, lmeta, smeta);
        // Routing is implicit for downstream and
        // performed by attachments_v4.
        attachments_v4.apply();
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
            smeta.ingress_port    : ternary @name("port");
            lmeta.if_type         : ternary @name("if_type");
            hdr.ethernet.src_addr : ternary @name("eth_src");
            hdr.ethernet.dst_addr : ternary @name("eth_dst");
            hdr.eth_type.value    : ternary @name("eth_type");
            hdr.ipv4.src_addr     : ternary @name("ipv4_src");
            hdr.ipv4.dst_addr     : ternary @name("ipv4_dst");
            hdr.ipv4.proto        : ternary @name("ipv4_proto");
            lmeta.l4_sport        : ternary @name("l4_sport");
            lmeta.l4_dport        : ternary @name("l4_dport");
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

    // In some implementations, the same port might be serving traffic from both
    // sides (ACCESS and CORE). In that case the match key could be extended to
    // include headers to differentiate the packet direction (e.g. MPLS labels
    // in DT, what about Dell?)
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

    // The BNG acts as a router, so we expect packets to have Ethernet dest
    // address the router MAC address.
    table my_stations {
        key = {
            smeta.ingress_port    : exact @name("port");
            hdr.ethernet.dst_addr : exact @name("eth_dst");
        }
        actions = {
            set_my_station;
            @defaultonly nop;
        }
        const default_action = nop;
        @name("my_stations")
        counters = direct_counter(CounterType.packets_and_bytes);
        size = 2048;
    }

    action set_accounting_id(accounting_id_t accounting_id) {
        lmeta.accounting_id = accounting_id;
    }

    table accounting_ids {
        key = {
            lmeta.line_id : exact @name("line_id");
            lmeta.cos_id  : exact @name("cos_id");
        }
        actions = {
            set_accounting_id;
        }
        const default_action = set_accounting_id(ACCOUNTING_UNKNOWN);
        size = MAX_ACCOUNTING_IDS;
    }

    IngressUpstream() upstream;
    IngressDownstream() downstream;
    Acl() acl;
    Accounting() accounting;

    apply {
        // Controller packet-out. Set the egress port according to CPU header
        // and skip the rest of the pipeline.
        if (hdr.cpu_out.isValid()) {
            smeta.egress_spec = hdr.cpu_out.egress_port;
            hdr.cpu_out.setInvalid();
            exit;
        }

        if_types.apply();

        lmeta.direction = DIR_UNKNOWN;

        if (my_stations.apply().hit) {
            if (lmeta.if_type == IF_ACCESS) {
                lmeta.direction = DIR_UPSTREAM;
                upstream.apply(hdr, lmeta, smeta);
            } else if (lmeta.if_type == IF_CORE) {
                lmeta.direction = DIR_DOWNSTREAM;
                downstream.apply(hdr, lmeta, smeta);
            }
        }

        // FIXME: If we want to do ACL after routing, then we need to make sure
        //  the ACL table sees the original header fields, not those modidified
        //  by routing or previous tables. A simple solution is to make sure all
        //  previous tables modify only metadata, and the actual header rewrite
        //  happen after the ACL table.
        acl.apply(hdr, lmeta, smeta);

        // FIXME: stop using exit statement in drop_now() action, otherwise
        //  dropped packets will not make it until here to be counted. This is
        //  pre-qos accounting, so we should count every packet that enters the
        //  switch. Instead of the exit statement, we should use metadata to
        //  signal intention to drop and skip applying tables as needed.

        accounting_ids.apply();
        accounting.apply(hdr, lmeta, smeta);
    }
}

control EgressPipe(
    inout parsed_headers_t hdr,
    inout local_metadata_t lmeta,
    inout standard_metadata_t smeta) {

    Accounting() accounting;

    apply {
        if (smeta.egress_port == CPU_PORT) {
            hdr.cpu_in.setValid();
            hdr.cpu_in.ingress_port = smeta.ingress_port;
            exit;
        }

        // TODO: Consider switching to PSA to count post-encap/decap bytes
        // In v1model, byte counters in the egress pipe are incremented with the
        // same pkt size seen at ingress. For example, even if upstream packets
        // get decapsulated in ingress, egress counters will still account for
        // access headers (VLAN and PPPoE). That might be a problem for
        // operators that do volume-based billing, which need to maintain stats
        // only for data effectivelly sent/received by an attachment, i.e.,
        // without counting bytes for the access headers. While the control
        // plane can approximate such stats (e.g., by knowing the pkt count and
        // for access headers with fixed size), the right solution to this
        // problem would be to use a different architecture other than v1model,
        // such as PSA, where pkts are deparased at the end of ingress, and
        // parsed again at egress. With PSA, if we decapsulate pkts in the
        // ingress pipe, the egress pipe should see shorter size pkts.
        accounting.apply(hdr, lmeta, smeta);
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
