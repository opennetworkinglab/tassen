#include <core.p4>
#include <v1model.p4>

typedef bit<2> bng_type_t;
const bng_type_t BNG_TYPE_INVALID = 2w0x0;
const bng_type_t BNG_TYPE_UPSTREAM = 2w0x1;
const bng_type_t BNG_TYPE_DOWNSTREAM = 2w0x2;

typedef bit<3> fwd_type_t;
typedef bit<32> next_id_t;
typedef bit<20> mpls_label_t;
typedef bit<9> port_num_t;
typedef bit<48> mac_addr_t;
typedef bit<16> mcast_group_id_t;
typedef bit<12> vlan_id_t;
typedef bit<32> ipv4_addr_t;
typedef bit<16> l4_port_t;
typedef bit<2> direction_t;
typedef bit<1> pcc_gate_status_t;
typedef bit<32> sdf_rule_id_t;
typedef bit<32> pcc_rule_id_t;
const ipv4_addr_t S1U_SGW_PREFIX = 2348810240;
const bit<16> ETHERTYPE_QINQ = 0x88a8;
const bit<16> ETHERTYPE_QINQ_NON_STD = 0x9100;
const bit<16> ETHERTYPE_VLAN = 0x8100;
const bit<16> ETHERTYPE_MPLS = 0x8847;
const bit<16> ETHERTYPE_MPLS_MULTICAST = 0x8848;
const bit<16> ETHERTYPE_IPV4 = 0x800;
const bit<16> ETHERTYPE_IPV6 = 0x86dd;
const bit<16> ETHERTYPE_ARP = 0x806;
const bit<16> ETHERTYPE_PPPOED = 0x8863;
const bit<16> ETHERTYPE_PPPOES = 0x8864;
const bit<16> PPPOE_PROTOCOL_IP4 = 0x21;
const bit<16> PPPOE_PROTOCOL_IP6 = 0x57;
const bit<16> PPPOE_PROTOCOL_MPLS = 0x281;
const bit<8> PROTO_ICMP = 1;
const bit<8> PROTO_TCP = 6;
const bit<8> PROTO_UDP = 17;
const bit<8> PROTO_ICMPV6 = 58;
const bit<4> IPV4_MIN_IHL = 5;
const fwd_type_t FWD_BRIDGING = 0;
const fwd_type_t FWD_MPLS = 1;
const fwd_type_t FWD_IPV4_UNICAST = 2;
const fwd_type_t FWD_IPV4_MULTICAST = 3;
const fwd_type_t FWD_IPV6_UNICAST = 4;
const fwd_type_t FWD_IPV6_MULTICAST = 5;
const fwd_type_t FWD_UNKNOWN = 7;
const vlan_id_t DEFAULT_VLAN_ID = 12w4094;
const bit<8> DEFAULT_MPLS_TTL = 64;
const bit<8> DEFAULT_IPV4_TTL = 64;
const sdf_rule_id_t DEFAULT_SDF_RULE_ID = 0;
const pcc_rule_id_t DEFAULT_PCC_RULE_ID = 0;
const direction_t SPGW_DIR_UNKNOWN = 2w0;
const direction_t SPGW_DIR_UPLINK = 2w1;
const direction_t SPGW_DIR_DOWNLINK = 2w2;
const pcc_gate_status_t PCC_GATE_OPEN = 1w0;
const pcc_gate_status_t PCC_GATE_CLOSED = 1w1;
const bit<6> INT_DSCP = 0x1;
const bit<8> INT_HEADER_LEN_WORDS = 4;
const bit<16> INT_HEADER_LEN_BYTES = 16;
const bit<8> CPU_MIRROR_SESSION_ID = 250;
const bit<32> REPORT_MIRROR_SESSION_ID = 500;
const bit<4> NPROTO_ETHERNET = 0;
const bit<4> NPROTO_TELEMETRY_DROP_HEADER = 1;
const bit<4> NPROTO_TELEMETRY_SWITCH_LOCAL_HEADER = 2;
const bit<6> HW_ID = 1;
const bit<8> REPORT_FIXED_HEADER_LEN = 12;
const bit<8> DROP_REPORT_HEADER_LEN = 12;
const bit<8> LOCAL_REPORT_HEADER_LEN = 16;
const bit<8> ETH_HEADER_LEN = 14;
const bit<8> IPV4_MIN_HEAD_LEN = 20;
const bit<8> UDP_HEADER_LEN = 8;

action nop() {
    NoAction();
}

@controller_header("packet_in") header cpu_in_t {
    port_num_t ingress_port;
    bit<7>     _pad;
}

@controller_header("packet_out") header cpu_out_t {
    port_num_t egress_port;
    bit<7>     _pad;
}

header ethernet_t {
    mac_addr_t dst_addr;
    mac_addr_t src_addr;
}

header eth_type_t {
    bit<16> value;
}

header vlan_t {
    bit<16>   eth_type;
    bit<3>    pri;
    bit<1>    cfi;
    vlan_id_t vlan_id;
}

header pppoe_t {
    bit<4>  version;
    bit<4>  type_id;
    bit<8>  code;
    bit<16> session_id;
    bit<16> length;
    bit<16> protocol;
}

header ipv4_t {
    bit<4>  version;
    bit<4>  ihl;
    bit<6>  dscp;
    bit<2>  ecn;
    bit<16> total_len;
    bit<16> identification;
    bit<3>  flags;
    bit<13> frag_offset;
    bit<8>  ttl;
    bit<8>  protocol;
    bit<16> hdr_checksum;
    bit<32> src_addr;
    bit<32> dst_addr;
}

header ipv6_t {
    bit<4>   version;
    bit<8>   traffic_class;
    bit<20>  flow_label;
    bit<16>  payload_len;
    bit<8>   next_hdr;
    bit<8>   hop_limit;
    bit<128> src_addr;
    bit<128> dst_addr;
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
    bit<16> sequence_number;
    bit<64> timestamp;
}

struct bng_meta_t {
    bit<2>    type;
    bit<32>   line_id;
    bit<16>   pppoe_session_id;
    bit<32>   ds_meter_result;
    vlan_id_t s_tag;
    vlan_id_t c_tag;
}

struct local_metadata_t {
    bit<16>      ip_eth_type;
    vlan_id_t    vlan_id;
    bit<3>       vlan_pri;
    bit<1>       vlan_cfi;
    bool         push_double_vlan;
    vlan_id_t    inner_vlan_id;
    bit<3>       inner_vlan_pri;
    bit<1>       inner_vlan_cfi;
    bool         skip_forwarding;
    bool         skip_next;
    fwd_type_t   fwd_type;
    next_id_t    next_id;
    bool         is_multicast;
    bool         is_controller_packet_out;
    bit<8>       ip_proto;
    bit<16>      l4_sport;
    bit<16>      l4_dport;
    bng_meta_t   bng;
}

struct parsed_headers_t {
    ethernet_t          ethernet;
    vlan_t          vlan;
    vlan_t          inner_vlan;
    eth_type_t          eth_type;
    pppoe_t             pppoe;
    ipv4_t              ipv4;
    tcp_t               tcp;
    udp_t               udp;
    icmp_t              icmp;
    cpu_out_t packet_out;
    cpu_in_t  packet_in;
}

control Filtering(inout parsed_headers_t hdr, inout local_metadata_t local_meta, inout standardd_meta_t std_meta) {
    direct_counter(CounterType.packets_and_bytes) ingress_port_vlan_counter;
    action deny() {
        local_meta.skip_forwarding = true;
        local_meta.skip_next = true;
        ingress_port_vlan_counter.count();
    }
    action permit() {
        ingress_port_vlan_counter.count();
    }
    action permit_with_internal_vlan(vlan_id_t vlan_id) {
        local_meta.vlan_id = vlan_id;
        permit();
    }
    table ingress_port_vlan {
        key = {
            std_meta.ingress_port: exact @name("ig_port") ;
            hdr.vlan.isValid()        : exact @name("vlan_is_valid") ;
            hdr.vlan.vlan_id          : ternary @name("vlan_id") ;
            hdr.inner_vlan.vlan_id    : ternary @name("inner_vlan_id") ;
        }
        actions = {
            deny();
            permit();
            permit_with_internal_vlan();
        }
        const default_action = deny();
        counters = ingress_port_vlan_counter;
        size = 8192;
    }
    direct_counter(CounterType.packets_and_bytes) fwd_classifier_counter;
    action set_forwarding_type(fwd_type_t fwd_type) {
        local_meta.fwd_type = fwd_type;
        fwd_classifier_counter.count();
    }
    table fwd_classifier {
        key = {
            std_meta.ingress_port: exact @name("ig_port") ;
            hdr.ethernet.dst_addr         : ternary @name("eth_dst") ;
            hdr.eth_type.value            : ternary @name("eth_type") ;
            local_meta.ip_eth_type   : exact @name("ip_eth_type") ;
        }
        actions = {
            set_forwarding_type;
        }
        const default_action = set_forwarding_type(FWD_BRIDGING);
        counters = fwd_classifier_counter;
        size = 1024;
    }
    apply {
        if (hdr.vlan.isValid()) {
            local_meta.vlan_id = hdr.vlan.vlan_id;
            local_meta.vlan_pri = hdr.vlan.pri;
            local_meta.vlan_cfi = hdr.vlan.cfi;
        }
        if (hdr.inner_vlan.isValid()) {
            local_meta.inner_vlan_id = hdr.inner_vlan.vlan_id;
            local_meta.inner_vlan_pri = hdr.inner_vlan.pri;
            local_meta.inner_vlan_cfi = hdr.inner_vlan.cfi;
        }
        ingress_port_vlan.apply();
        fwd_classifier.apply();
    }
}

control Forwarding(inout parsed_headers_t hdr, inout local_metadata_t local_meta, inout standardd_meta_t std_meta) {
    @hidden
    action set_next_id(next_id_t next_id) {
        local_meta.next_id = next_id;
    }
    // direct_counter(CounterType.packets_and_bytes) bridging_counter;
    // action set_next_id_bridging(next_id_t next_id) {
    //     set_next_id(next_id);
    //     bridging_counter.count();
    // }
    // table bridging {
    //     key = {
    //         local_meta.vlan_id: exact @name("vlan_id") ;
    //         hdr.ethernet.dst_addr  : ternary @name("eth_dst") ;
    //     }
    //     actions = {
    //         set_next_id_bridging;
    //         @defaultonly nop;
    //     }
    //     const default_action = nop();
    //     counters = bridging_counter;
    //     size = 1024;
    // }
    // direct_counter(CounterType.packets_and_bytes) mpls_counter;
    // action pop_mpls_and_next(next_id_t next_id) {
    //     local_meta.mpls_label = 0;
    //     set_next_id(next_id);
    //     mpls_counter.count();
    // }
    // table mpls {
    //     key = {
    //         local_meta.mpls_label: exact @name("mpls_label") ;
    //     }
    //     actions = {
    //         pop_mpls_and_next;
    //         @defaultonly nop;
    //     }
    //     const default_action = nop();
    //     counters = mpls_counter;
    //     size = 1024;
    // }
    action set_next_id_routing_v4(next_id_t next_id) {
        set_next_id(next_id);
    }
    action nop_routing_v4() {
    }
    table routing_v4 {
        key = {
            hdr.ipv4.dst_addr: lpm @name("ipv4_dst") ;
        }
        actions = {
            set_next_id_routing_v4;
            nop_routing_v4;
            @defaultonly nop;
        }
        default_action = nop();
        size = 1024;
    }
    apply {
        if (local_meta.fwd_type == FWD_BRIDGING) {
            bridging.apply();
        } else if (local_meta.fwd_type == FWD_MPLS) {
            mpls.apply();
        } else if (local_meta.fwd_type == FWD_IPV4_UNICAST) {
            routing_v4.apply();
        }
    }
}

control Acl(inout parsed_headers_t hdr, inout local_metadata_t local_meta, inout standardd_meta_t std_meta) {
    direct_counter(CounterType.packets_and_bytes) acl_counter;
    action set_next_id_acl(next_id_t next_id) {
        local_meta.next_id = next_id;
        acl_counter.count();
    }
    action punt_to_cpu() {
        std_meta.egress_spec = 255;
        local_meta.skip_next = true;
        acl_counter.count();
    }
    action set_clone_session_id(bit<32> clone_id) {
        clone3(CloneType.I2E, clone_id, { std_meta.ingress_port });
        acl_counter.count();
    }
    action drop() {
        mark_to_drop(std_meta);
        local_meta.skip_next = true;
        acl_counter.count();
    }
    action nop_acl() {
        acl_counter.count();
    }
    table acl {
        key = {
            std_meta.ingress_port: ternary @name("ig_port") ;
            local_meta.ip_proto      : ternary @name("ip_proto") ;
            local_meta.l4_sport      : ternary @name("l4_sport") ;
            local_meta.l4_dport      : ternary @name("l4_dport") ;
            hdr.ethernet.dst_addr         : ternary @name("eth_dst") ;
            hdr.ethernet.src_addr         : ternary @name("eth_src") ;
            hdr.vlan.vlan_id              : ternary @name("vlan_id") ;
            hdr.eth_type.value            : ternary @name("eth_type") ;
            hdr.ipv4.src_addr             : ternary @name("ipv4_src") ;
            hdr.ipv4.dst_addr             : ternary @name("ipv4_dst") ;
            hdr.icmp.icmp_type            : ternary @name("icmp_type") ;
            hdr.icmp.icmp_code            : ternary @name("icmp_code") ;
        }
        actions = {
            set_next_id_acl;
            punt_to_cpu;
            set_clone_session_id;
            drop;
            nop_acl;
        }
        const default_action = nop_acl();
        size = 1024;
        counters = acl_counter;
    }
    apply {
        acl.apply();
    }
}

control Next(inout parsed_headers_t hdr, inout local_metadata_t local_meta, inout standardd_meta_t std_meta) {
    @hidden action output(port_num_t port_num) {
        std_meta.egress_spec = port_num;
    }
    @hidden action rewrite_smac(mac_addr_t smac) {
        hdr.ethernet.src_addr = smac;
    }
    @hidden action rewrite_dmac(mac_addr_t dmac) {
        hdr.ethernet.dst_addr = dmac;
    }
    @hidden action routing(port_num_t port_num, mac_addr_t smac, mac_addr_t dmac) {
        rewrite_smac(smac);
        rewrite_dmac(dmac);
        output(port_num);
    }
    direct_counter(CounterType.packets_and_bytes) next_vlan_counter;
    action set_vlan(vlan_id_t vlan_id) {
        local_meta.vlan_id = vlan_id;
        next_vlan_counter.count();
    }
    action set_double_vlan(vlan_id_t outer_vlan_id, vlan_id_t inner_vlan_id) {
        set_vlan(outer_vlan_id);
        local_meta.push_double_vlan = true;
        local_meta.inner_vlan_id = inner_vlan_id;
        local_meta.bng.s_tag = outer_vlan_id;
        local_meta.bng.c_tag = inner_vlan_id;
    }
    table next_vlan {
        key = {
            local_meta.next_id: exact @name("next_id") ;
        }
        actions = {
            set_vlan;
            set_double_vlan;
            @defaultonly nop;
        }
        const default_action = nop();
        counters = next_vlan_counter;
        size = 1024;
    }

    @max_group_size(16)
    action_selector(HashAlgorithm.crc16, 32w1024, 32w16) hashed_selector;

    direct_counter(CounterType.packets_and_bytes) hashed_counter;

    action output_hashed(port_num_t port_num) {
        output(port_num);
        hashed_counter.count();
    }
    action routing_hashed(port_num_t port_num, mac_addr_t smac, mac_addr_t dmac) {
        routing(port_num, smac, dmac);
        hashed_counter.count();
    }

    table hashed {
        key = {
            local_meta.next_id : exact @name("next_id") ;
            hdr.ipv4.dst_addr       : selector;
            hdr.ipv4.src_addr       : selector;
            local_meta.ip_proto: selector;
            local_meta.l4_sport: selector;
            local_meta.l4_dport: selector;
        }
        actions = {
            output_hashed;
            routing_hashed;
            mpls_routing_hashed;
            @defaultonly nop;
        }
        implementation = hashed_selector;
        counters = hashed_counter;
        const default_action = nop();
        size = 1024;
    }

    direct_counter(CounterType.packets_and_bytes) multicast_counter;
    action set_mcast_group_id(mcast_group_id_t group_id) {
        std_meta.mcast_grp = group_id;
        local_meta.is_multicast = true;
        multicast_counter.count();
    }

    table multicast {
        key = {
            local_meta.next_id: exact @name("next_id") ;
        }
        actions = {
            set_mcast_group_id;
            @defaultonly nop;
        }
        counters = multicast_counter;
        const default_action = nop();
        size = 1024;
    }

    apply {
        hashed.apply();
        multicast.apply();
        next_vlan.apply();
    }
}

control EgressNextControl(inout parsed_headers_t hdr, inout local_metadata_t local_meta, inout standardd_meta_t std_meta) {

    @hidden
    action push_vlan() {
        hdr.vlan.setValid();
        hdr.vlan.cfi = local_meta.vlan_cfi;
        hdr.vlan.pri = local_meta.vlan_pri;
        hdr.vlan.eth_type = ETHERTYPE_VLAN;
        hdr.vlan.vlan_id = local_meta.vlan_id;
    }
    @hidden
    action push_inner_vlan() {
        hdr.inner_vlan.setValid();
        hdr.inner_vlan.cfi = local_meta.inner_vlan_cfi;
        hdr.inner_vlan.pri = local_meta.inner_vlan_pri;
        hdr.inner_vlan.vlan_id = local_meta.inner_vlan_id;
        hdr.inner_vlan.eth_type = ETHERTYPE_VLAN;
        hdr.vlan.eth_type = ETHERTYPE_VLAN;
    }

    direct_counter(CounterType.packets_and_bytes) egress_vlan_counter;

    action pop_vlan() {
        hdr.vlan.setInvalid();
        egress_vlan_counter.count();
    }

    table egress_vlan {
        key = {
            local_meta.vlan_id   : exact @name("vlan_id");
            std_meta.egress_port : exact @name("eg_port");
        }
        actions = {
            pop_vlan;
            @defaultonly nop;
        }
        const default_action = nop();
        counters = egress_vlan_counter;
        size = 1024;
    }

    apply {
        if (local_meta.is_multicast == true &&
                std_meta.ingress_port == std_meta.egress_port) {
            mark_to_drop(std_meta);
        }

        if (local_meta.push_double_vlan == true) {
            push_vlan();
            push_inner_vlan();
        } else {
            hdr.inner_vlan.setInvalid();
            if (!egress_vlan.apply().hit) {
                if (local_meta.vlan_id != DEFAULT_VLAN_ID) {
                    push_vlan();
                }
            }
        }

        if (hdr.ipv4.isValid()) {
            hdr.ipv4.ttl = hdr.ipv4.ttl - 1;
            if (hdr.ipv4.ttl == 0) {
                mark_to_drop(std_meta);
            }
        }
    }
}

control PacketIoIngress(inout parsed_headers_t hdr, inout local_metadata_t local_meta, inout standardd_meta_t std_meta) {
    apply {
        if (hdr.packet_out.isValid()) {
            std_meta.egress_spec = hdr.packet_out.egress_port;
            hdr.packet_out.setInvalid();
            local_meta.is_controller_packet_out = true;
            exit;
        }
    }
}

control PacketIoEgress(inout parsed_headers_t hdr, inout local_metadata_t local_meta, inout standardd_meta_t std_meta) {
    apply {
        if (local_meta.is_controller_packet_out == true) {
            exit;
        }
        if (std_meta.egress_port == 255) {
            hdr.packet_in.setValid();
            hdr.packet_in.ingress_port = std_meta.ingress_port;
            exit;
        }
    }
}

control FabricComputeChecksum(inout parsed_headers_t hdr, inout local_metadata_t meta) {
    apply {
        update_checksum(hdr.ipv4.isValid(), { hdr.ipv4.version, hdr.ipv4.ihl, hdr.ipv4.dscp, hdr.ipv4.ecn, hdr.ipv4.total_len, hdr.ipv4.identification, hdr.ipv4.flags, hdr.ipv4.frag_offset, hdr.ipv4.ttl, hdr.ipv4.protocol, hdr.ipv4.src_addr, hdr.ipv4.dst_addr }, hdr.ipv4.hdr_checksum, HashAlgorithm.csum16);
    }
}

control FabricVerifyChecksum(inout parsed_headers_t hdr, inout local_metadata_t meta) {
    apply {
        verify_checksum(hdr.ipv4.isValid(), { hdr.ipv4.version, hdr.ipv4.ihl, hdr.ipv4.dscp, hdr.ipv4.ecn, hdr.ipv4.total_len, hdr.ipv4.identification, hdr.ipv4.flags, hdr.ipv4.frag_offset, hdr.ipv4.ttl, hdr.ipv4.protocol, hdr.ipv4.src_addr, hdr.ipv4.dst_addr }, hdr.ipv4.hdr_checksum, HashAlgorithm.csum16);
    }
}

parser FabricParser(packet_in packet, out parsed_headers_t hdr, inout local_metadata_t local_meta, inout standardd_meta_t std_meta) {

    state start {
        transition select(std_meta.ingress_port) {
            255: parse_packet_out;
            default: parse_ethernet;
        }
    }
    state parse_packet_out {
        packet.extract(hdr.packet_out);
        transition parse_ethernet;
    }
    state parse_ethernet {
        packet.extract(hdr.ethernet);
        local_meta.vlan_id = DEFAULT_VLAN_ID;
        transition select(packet.lookahead<bit<16>>()) {
            ETHERTYPE_QINQ: parse_vlan;
            ETHERTYPE_QINQ_NON_STD: parse_vlan;
            ETHERTYPE_VLAN: parse_vlan;
            default: parse_eth_type;
        }
    }
    state parse_vlan {
        packet.extract(hdr.vlan);
        local_meta.bng.s_tag = hdr.vlan.vlan_id;
        transition select(packet.lookahead<bit<16>>()) {
            ETHERTYPE_VLAN: parse_inner_vlan;
            default: parse_eth_type;
        }
    }
    state parse_inner_vlan {
        packet.extract(hdr.inner_vlan);
        local_meta.bng.c_tag = hdr.inner_vlan.vlan_id;
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
        transition select(hdr.pppoe.protocol) {
            PPPOE_PROTOCOL_MPLS: parse_mpls;
            PPPOE_PROTOCOL_IP4: parse_ipv4;
            default: accept;
        }
    }
    state parse_ipv4 {
        packet.extract(hdr.ipv4);
        local_meta.ip_proto = hdr.ipv4.protocol;
        local_meta.ip_eth_type = ETHERTYPE_IPV4;
        transition select(hdr.ipv4.protocol) {
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

control FabricDeparser(packet_out packet, in parsed_headers_t hdr) {
    apply {
        packet.emit(hdr.packet_in);
        packet.emit(hdr.ethernet);
        packet.emit(hdr.vlan);
        packet.emit(hdr.inner_vlan);
        packet.emit(hdr.eth_type);
        packet.emit(hdr.pppoe);
        packet.emit(hdr.mpls);
        packet.emit(hdr.ipv4);
        packet.emit(hdr.tcp);
        packet.emit(hdr.udp);
        packet.emit(hdr.icmp);
    }
}

control bng_ingress_upstream(inout parsed_headers_t hdr, inout local_metadata_t fmeta, inout standardd_meta_t smeta) {
    counter(8192, CounterType.bytes) c_terminated;
    counter(8192, CounterType.bytes) c_dropped;
    counter(8192, CounterType.packets) c_control;
    action punt_to_cpu() {
        smeta.egress_spec = 255;
        smeta.mcast_grp = 0;
        c_control.count(fmeta.bng.line_id);
    }
    table t_pppoe_cp {
        key = {
            hdr.pppoe.code    : exact @name("pppoe_code") ;
            hdr.pppoe.protocol: ternary @name("pppoe_protocol") ;
        }
        actions = {
            punt_to_cpu;
            @defaultonly nop;
        }
        size = 16;
        const default_action = nop;
    }
    @hidden action term_enabled(bit<16> eth_type) {
        hdr.eth_type.value = eth_type;
        hdr.pppoe.setInvalid();
        c_terminated.count(fmeta.bng.line_id);
    }
    action term_disabled() {
        fmeta.bng.type = BNG_TYPE_INVALID;
        mark_to_drop(smeta);
    }
    action term_enabled_v4() {
        term_enabled(ETHERTYPE_IPV4);
    }
    table t_pppoe_term_v4 {
        key = {
            fmeta.bng.line_id   : exact @name("line_id") ;
            hdr.ipv4.src_addr   : exact @name("ipv4_src") ;
            hdr.pppoe.session_id: exact @name("pppoe_session_id") ;
        }
        actions = {
            term_enabled_v4;
            @defaultonly term_disabled;
        }
        size = 4 * 8192;
        const default_action = term_disabled;
    }
    apply {
        if (t_pppoe_cp.apply().hit) {
            return;
        }
        if (hdr.ipv4.isValid()) {
            switch (t_pppoe_term_v4.apply().action_run) {
                term_disabled: {
                    c_dropped.count(fmeta.bng.line_id);
                }
            }
        }
    }
}

control bng_ingress_downstream(inout parsed_headers_t hdr, inout local_metadata_t fmeta, inout standardd_meta_t smeta) {
    counter(8192, CounterType.bytes) c_line_rx;
    meter(8192, MeterType.bytes) m_besteff;
    meter(8192, MeterType.bytes) m_prio;

    action set_session(bit<16> pppoe_session_id) {
        fmeta.bng.type = BNG_TYPE_DOWNSTREAM;
        fmeta.bng.pppoe_session_id = pppoe_session_id;
        c_line_rx.count(fmeta.bng.line_id);
    }

    action drop() {
        fmeta.bng.type = BNG_TYPE_DOWNSTREAM;
        c_line_rx.count(fmeta.bng.line_id);
        mark_to_drop(smeta);
    }

    table t_line_session_map {
        key = {
            fmeta.bng.line_id: exact @name("line_id") ;
        }
        actions = {
            @defaultonly nop;
            set_session;
            drop;
        }
        size = 8192;
        const default_action = nop;
    }

    // table t_qos_v4 {
    //     key = {
    //         fmeta.bng.line_id: ternary @name("line_id") ;
    //         hdr.ipv4.src_addr: lpm @name("ipv4_src") ;
    //         hdr.ipv4.dscp    : ternary @name("ipv4_dscp") ;
    //         hdr.ipv4.ecn     : ternary @name("ipv4_ecn") ;
    //     }
    //     actions = {
    //         qos_prio;
    //         qos_besteff;
    //     }
    //     size = 256;
    //     const default_action = qos_besteff;
    // }

    apply {
        t_line_session_map.apply();
    }
}

control bng_egress_downstream(inout parsed_headers_t hdr, inout local_metadata_t fmeta, inout standardd_meta_t smeta) {
    counter(8192, CounterType.bytes) c_line_tx;

    @hidden
    action encap() {
        hdr.eth_type.value = ETHERTYPE_PPPOES;
        hdr.pppoe.setValid();
        hdr.pppoe.version = 4w1;
        hdr.pppoe.type_id = 4w1;
        hdr.pppoe.code = 8w0;
        hdr.pppoe.session_id = fmeta.bng.pppoe_session_id;
        c_line_tx.count(fmeta.bng.line_id);
    }

    action encap_v4() {
        encap();
        hdr.pppoe.length = hdr.ipv4.total_len + 16w2;
        hdr.pppoe.protocol = PPPOE_PROTOCOL_IP4;
    }

    apply {
        if (hdr.ipv4.isValid()) {
            encap_v4();
        }
    }
}

control bng_ingress(inout parsed_headers_t hdr, inout local_metadata_t fmeta, inout standardd_meta_t smeta) {
    bng_ingress_upstream() upstream;
    bng_ingress_downstream() downstream;

    action set_line(bit<32> line_id) {
        fmeta.bng.line_id = line_id;
    }

    table t_line_map {
        key = {
            fmeta.bng.s_tag: exact @name("s_tag") ;
            fmeta.bng.c_tag: exact @name("c_tag") ;
        }
        actions = {
            set_line;
        }
        size = 8192;
        const default_action = set_line(0);
    }

    apply {
        t_line_map.apply();
        if (hdr.pppoe.isValid()) {
            fmeta.bng.type = BNG_TYPE_UPSTREAM;
            upstream.apply(hdr, fmeta, smeta);
        } else {
            downstream.apply(hdr, fmeta, smeta);
        }
    }
}

control bng_egress(inout parsed_headers_t hdr, inout local_metadata_t fmeta, inout standardd_meta_t smeta) {
    bng_egress_downstream() downstream;
    apply {
        if (fmeta.bng.type == BNG_TYPE_DOWNSTREAM) {
            downstream.apply(hdr, fmeta, smeta);
        }
    }
}

control FabricIngress(inout parsed_headers_t hdr, inout local_metadata_t local_meta, inout standardd_meta_t std_meta) {
    PacketIoIngress() pkt_io_ingress;
    Filtering() filtering;
    Forwarding() forwarding;
    Acl() acl;
    Next() next;
    apply {
        pkt_io_ingress.apply(hdr, local_meta, std_meta);
        filtering.apply(hdr, local_meta, std_meta);
        if (local_meta.skip_forwarding == false) {
            forwarding.apply(hdr, local_meta, std_meta);
        }
        acl.apply(hdr, local_meta, std_meta);
        if (local_meta.skip_next == false) {
            next.apply(hdr, local_meta, std_meta);
        }
        bng_ingress.apply(hdr, local_meta, std_meta);
    }
}

control FabricEgress(inout parsed_headers_t hdr, inout local_metadata_t local_meta, inout standardd_meta_t std_meta) {
    PacketIoEgress() pkt_io_egress;
    EgressNextControl() egress_next;
    apply {
        pkt_io_egress.apply(hdr, local_meta, std_meta);
        egress_next.apply(hdr, local_meta, std_meta);
        bng_egress.apply(hdr, local_meta, std_meta);
    }
}

V1Switch(FabricParser(), FabricVerifyChecksum(), FabricIngress(), FabricEgress(), FabricComputeChecksum(), FabricDeparser()) main;

