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

package fabric
const (
    // Header field IDs
	Hdr_FabricIngressNextMulticast_NextId uint32 = 1
	Hdr_FabricIngressNextHashed_NextId uint32 = 1
	Hdr_FabricEgressEgressNextEgressVlan_VlanId uint32 = 1
	Hdr_FabricEgressEgressNextEgressVlan_EgPort uint32 = 2
	Hdr_FabricIngressForwardingMpls_MplsLabel uint32 = 1
	Hdr_FabricIngressBngIngressTLineMap_CTag uint32 = 2
	Hdr_FabricIngressBngIngressTLineMap_STag uint32 = 1
	Hdr_FabricIngressForwardingRoutingV4_Ipv4Dst uint32 = 1
	Hdr_FabricIngressBngIngressDownstreamTLineSessionMap_LineId uint32 = 1
	Hdr_FabricIngressFilteringFwdClassifier_IpEthType uint32 = 4
	Hdr_FabricIngressFilteringFwdClassifier_IgPort uint32 = 1
	Hdr_FabricIngressFilteringFwdClassifier_EthType uint32 = 3
	Hdr_FabricIngressFilteringFwdClassifier_EthDst uint32 = 2
	Hdr_FabricIngressForwardingBridging_VlanId uint32 = 1
	Hdr_FabricIngressForwardingBridging_EthDst uint32 = 2
	Hdr_FabricIngressAclAcl_Ipv4Src uint32 = 9
	Hdr_FabricIngressAclAcl_Ipv4Dst uint32 = 10
	Hdr_FabricIngressAclAcl_EthSrc uint32 = 6
	Hdr_FabricIngressAclAcl_IcmpCode uint32 = 12
	Hdr_FabricIngressAclAcl_IpProto uint32 = 2
	Hdr_FabricIngressAclAcl_EthType uint32 = 8
	Hdr_FabricIngressAclAcl_L4Sport uint32 = 3
	Hdr_FabricIngressAclAcl_EthDst uint32 = 5
	Hdr_FabricIngressAclAcl_L4Dport uint32 = 4
	Hdr_FabricIngressAclAcl_VlanId uint32 = 7
	Hdr_FabricIngressAclAcl_IgPort uint32 = 1
	Hdr_FabricIngressAclAcl_IcmpType uint32 = 11
	Hdr_FabricIngressBngIngressDownstreamTQosV4_Ipv4Dscp uint32 = 3
	Hdr_FabricIngressBngIngressDownstreamTQosV4_Ipv4Src uint32 = 2
	Hdr_FabricIngressBngIngressDownstreamTQosV4_LineId uint32 = 1
	Hdr_FabricIngressBngIngressDownstreamTQosV4_Ipv4Ecn uint32 = 4
	Hdr_FabricIngressFilteringIngressPortVlan_VlanId uint32 = 3
	Hdr_FabricIngressFilteringIngressPortVlan_VlanIsValid uint32 = 2
	Hdr_FabricIngressFilteringIngressPortVlan_IgPort uint32 = 1
	Hdr_FabricIngressFilteringIngressPortVlan_InnerVlanId uint32 = 4
	Hdr_FabricIngressNextNextVlan_NextId uint32 = 1
	Hdr_FabricIngressBngIngressUpstreamTPppoeCp_PppoeProtocol uint32 = 2
	Hdr_FabricIngressBngIngressUpstreamTPppoeCp_PppoeCode uint32 = 1
	Hdr_FabricIngressBngIngressUpstreamTPppoeTermV4_PppoeSessionId uint32 = 3
	Hdr_FabricIngressBngIngressUpstreamTPppoeTermV4_Ipv4Src uint32 = 2
	Hdr_FabricIngressBngIngressUpstreamTPppoeTermV4_LineId uint32 = 1
    // Table IDs
	Table_FabricIngressForwardingBridging uint32 = 33596749
	Table_FabricIngressBngIngressDownstreamTQosV4 uint32 = 33602462
	Table_FabricIngressBngIngressUpstreamTPppoeTermV4 uint32 = 33595047
	Table_FabricIngressNextHashed uint32 = 33608588
	Table_FabricIngressForwardingMpls uint32 = 33574274
	Table_FabricIngressForwardingRoutingV4 uint32 = 33562650
	Table_FabricIngressFilteringFwdClassifier uint32 = 33596298
	Table_FabricIngressBngIngressUpstreamTPppoeCp uint32 = 33603300
	Table_FabricIngressNextNextVlan uint32 = 33599709
	Table_FabricIngressAclAcl uint32 = 33618978
	Table_FabricIngressNextMulticast uint32 = 33606828
	Table_FabricIngressBngIngressDownstreamTLineSessionMap uint32 = 33594775
	Table_FabricIngressFilteringIngressPortVlan uint32 = 33611649
	Table_FabricEgressEgressNextEgressVlan uint32 = 33599342
	Table_FabricIngressBngIngressTLineMap uint32 = 33592041
    // Indirect Counter IDs
	Counter_FabricIngressBngIngressUpstreamCDropped uint32 = 302043418
	Counter_FabricIngressPortCountersControlIngressPortCounter uint32 = 302002771
	Counter_FabricIngressPortCountersControlEgressPortCounter uint32 = 302011205
	Counter_FabricEgressBngEgressDownstreamCLineTx uint32 = 302046535
	Counter_FabricIngressBngIngressDownstreamCLineRx uint32 = 302004781
	Counter_FabricIngressBngIngressUpstreamCTerminated uint32 = 302022672
	Counter_FabricIngressBngIngressUpstreamCControl uint32 = 302008909
    // Direct Counter IDs
	DirectCounter_FabricEgressEgressNextEgressVlanCounter uint32 = 318827144
	DirectCounter_FabricIngressNextNextVlanCounter uint32 = 318768144
	DirectCounter_FabricIngressNextHashedCounter uint32 = 318800532
	DirectCounter_FabricIngressForwardingBridgingCounter uint32 = 318770289
	DirectCounter_FabricIngressNextMulticastCounter uint32 = 318801752
	DirectCounter_FabricIngressAclAclCounter uint32 = 318801025
	DirectCounter_FabricIngressFilteringIngressPortVlanCounter uint32 = 318815501
	DirectCounter_FabricIngressForwardingMplsCounter uint32 = 318830507
	DirectCounter_FabricIngressFilteringFwdClassifierCounter uint32 = 318827326
    // Action IDs
	Action_FabricIngressForwardingSetNextIdRoutingV4 uint32 = 16777434
	Action_FabricIngressNextMplsRoutingHashed uint32 = 16779255
	Action_FabricEgressBngEgressDownstreamEncapV4 uint32 = 16784000
	Action_FabricIngressBngIngressDownstreamQosBesteff uint32 = 16804676
	Action_FabricIngressBngIngressDownstreamQosPrio uint32 = 16830304
	Action_FabricIngressBngIngressDownstreamSetSession uint32 = 16795395
	Action_FabricIngressBngIngressUpstreamTermDisabled uint32 = 16785853
	Action_FabricIngressAclPuntToCpu uint32 = 16829684
	Action_FabricIngressNextOutputHashed uint32 = 16815357
	Action_FabricIngressBngIngressDownstreamDrop uint32 = 16822844
	Action_FabricIngressNextSetVlan uint32 = 16790685
	Action_FabricIngressForwardingNopRoutingV4 uint32 = 16804187
	Action_FabricIngressFilteringPermitWithInternalVlan uint32 = 16794911
	Action_FabricIngressNextRoutingHashed uint32 = 16791402
	Action_FabricIngressForwardingSetNextIdBridging uint32 = 16811012
	Action_FabricIngressAclSetNextIdAcl uint32 = 16807382
	Action_FabricIngressBngIngressUpstreamTermEnabledV4 uint32 = 16780562
	Action_FabricIngressFilteringDeny uint32 = 16836487
	Action_FabricIngressForwardingPopMplsAndNext uint32 = 16827758
	Action_FabricIngressFilteringPermit uint32 = 16818236
	Action_FabricIngressNextSetMcastGroupId uint32 = 16779917
	Action_FabricIngressAclNopAcl uint32 = 16827694
	Action_FabricIngressFilteringSetForwardingType uint32 = 16840921
	Action_FabricIngressBngIngressSetLine uint32 = 16829385
	Action_FabricIngressNextSetDoubleVlan uint32 = 16803337
	Action_Nop uint32 = 16819938
	Action_FabricIngressAclDrop uint32 = 16820765
	Action_FabricIngressBngIngressUpstreamPuntToCpu uint32 = 16830893
	Action_FabricIngressAclSetCloneSessionId uint32 = 16781601
	Action_FabricEgressEgressNextPopVlan uint32 = 16790030
    // Action Param IDs
	ActionParam_FabricIngressNextSetVlan_VlanId uint32 = 1
	ActionParam_FabricIngressForwardingPopMplsAndNext_NextId uint32 = 1
	ActionParam_FabricIngressBngIngressSetLine_LineId uint32 = 1
	ActionParam_FabricIngressNextSetMcastGroupId_GroupId uint32 = 1
	ActionParam_FabricIngressFilteringSetForwardingType_FwdType uint32 = 1
	ActionParam_FabricIngressBngIngressDownstreamSetSession_PppoeSessionId uint32 = 1
	ActionParam_FabricIngressFilteringPermitWithInternalVlan_VlanId uint32 = 1
	ActionParam_FabricIngressAclSetNextIdAcl_NextId uint32 = 1
	ActionParam_FabricIngressNextOutputHashed_PortNum uint32 = 1
	ActionParam_FabricIngressForwardingSetNextIdBridging_NextId uint32 = 1
	ActionParam_FabricIngressNextSetDoubleVlan_OuterVlanId uint32 = 1
	ActionParam_FabricIngressNextSetDoubleVlan_InnerVlanId uint32 = 2
	ActionParam_FabricIngressForwardingSetNextIdRoutingV4_NextId uint32 = 1
	ActionParam_FabricIngressAclSetCloneSessionId_CloneId uint32 = 1
	ActionParam_FabricIngressNextMplsRoutingHashed_Label uint32 = 4
	ActionParam_FabricIngressNextMplsRoutingHashed_Dmac uint32 = 3
	ActionParam_FabricIngressNextMplsRoutingHashed_PortNum uint32 = 1
	ActionParam_FabricIngressNextMplsRoutingHashed_Smac uint32 = 2
	ActionParam_FabricIngressNextRoutingHashed_Dmac uint32 = 3
	ActionParam_FabricIngressNextRoutingHashed_PortNum uint32 = 1
	ActionParam_FabricIngressNextRoutingHashed_Smac uint32 = 2
    // Action Profile IDs
	ActionProfile_FabricIngressNextHashedSelector uint32 = 285217164
    // Packet Metadata IDs
	PacketMeta_IngressPort uint32 = 1
	PacketMeta_EgressPort uint32 = 1
    // Meter IDs
	Meter_FabricIngressBngIngressDownstreamMBesteff uint32 = 335569952
	Meter_FabricIngressBngIngressDownstreamMPrio uint32 = 335568260
)