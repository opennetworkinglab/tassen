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

package translate

//noinspection GoSnakeCaseUsage
const (
	// Header field IDs
	Hdr_IngressPipeDownstreamLinesV4_Ipv4Dst         uint32 = 1
	Hdr_IngressPipeAclAcls_L4Dport                   uint32 = 10
	Hdr_IngressPipeAclAcls_L4Sport                   uint32 = 9
	Hdr_IngressPipeAclAcls_Ipv4Dst                   uint32 = 7
	Hdr_IngressPipeAclAcls_IfType                    uint32 = 2
	Hdr_IngressPipeAclAcls_Ipv4Src                   uint32 = 6
	Hdr_IngressPipeAclAcls_Port                      uint32 = 1
	Hdr_IngressPipeAclAcls_Ipv4Proto                 uint32 = 8
	Hdr_IngressPipeAclAcls_EthType                   uint32 = 5
	Hdr_IngressPipeAclAcls_EthSrc                    uint32 = 3
	Hdr_IngressPipeAclAcls_EthDst                    uint32 = 4
	Hdr_IngressPipeUpstreamCosServicesV4_L4Sport     uint32 = 4
	Hdr_IngressPipeUpstreamCosServicesV4_L4Dport     uint32 = 5
	Hdr_IngressPipeUpstreamCosServicesV4_Ipv4Src     uint32 = 1
	Hdr_IngressPipeUpstreamCosServicesV4_Ipv4Dst     uint32 = 2
	Hdr_IngressPipeUpstreamCosServicesV4_Ipv4Proto   uint32 = 3
	Hdr_IngressPipeDownstreamCosServicesV4_L4Sport   uint32 = 4
	Hdr_IngressPipeDownstreamCosServicesV4_L4Dport   uint32 = 5
	Hdr_IngressPipeDownstreamCosServicesV4_Ipv4Src   uint32 = 1
	Hdr_IngressPipeDownstreamCosServicesV4_Ipv4Dst   uint32 = 2
	Hdr_IngressPipeDownstreamCosServicesV4_Ipv4Proto uint32 = 3
	Hdr_IngressPipeUpstreamLines_Port                uint32 = 1
	Hdr_IngressPipeUpstreamLines_CTag                uint32 = 2
	Hdr_IngressPipeUpstreamLines_STag                uint32 = 3
	Hdr_IngressPipeUpstreamPppoePunts_PppoeCode      uint32 = 1
	Hdr_IngressPipeUpstreamPppoePunts_PppoeProto     uint32 = 2
	Hdr_IngressPipeDownstreamAttachmentsV4_LineId    uint32 = 1
	Hdr_IngressPipeAccountingIds_CosId               uint32 = 2
	Hdr_IngressPipeAccountingIds_LineId              uint32 = 1
	Hdr_IngressPipeMyStations_Port                   uint32 = 1
	Hdr_IngressPipeMyStations_EthDst                 uint32 = 2
	Hdr_IngressPipeUpstreamAttachmentsV4_EthSrc      uint32 = 2
	Hdr_IngressPipeUpstreamAttachmentsV4_Ipv4Src     uint32 = 3
	Hdr_IngressPipeUpstreamAttachmentsV4_LineId      uint32 = 1
	Hdr_IngressPipeUpstreamAttachmentsV4_PppoeSessId uint32 = 4
	Hdr_IngressPipeUpstreamRoutesV4_Ipv4Dst          uint32 = 1
	Hdr_IngressPipeIfTypes_Port                      uint32 = 1
	// Table IDs
	Table_IngressPipeMyStations              uint32 = 49392761
	Table_IngressPipeUpstreamCosServicesV4   uint32 = 43325894
	Table_IngressPipeDownstreamCosServicesV4 uint32 = 44470911
	Table_IngressPipeUpstreamRoutesV4        uint32 = 40314915
	Table_IngressPipeUpstreamLines           uint32 = 33956689
	Table_IngressPipeDownstreamAttachmentsV4 uint32 = 50044735
	Table_IngressPipeUpstreamAttachmentsV4   uint32 = 44507663
	Table_IngressPipeAclAcls                 uint32 = 43911884
	Table_IngressPipeUpstreamPppoePunts      uint32 = 39053621
	Table_IngressPipeIfTypes                 uint32 = 38498675
	Table_IngressPipeDownstreamLinesV4       uint32 = 44334275
	Table_IngressPipeAccountingIds           uint32 = 42617777
	// Indirect Counter IDs
	Counter_IngressPipeUpstreamTtlExpired   uint32 = 307117173
	Counter_IngressPipeDownstreamRouted     uint32 = 315756728
	Counter_IngressPipeUpstreamRouted       uint32 = 310793502
	Counter_IngressPipeUpstreamSpoofed      uint32 = 314616893
	Counter_IngressPipeUpstreamPunted       uint32 = 310787420
	Counter_IngressPipeAccountingUpstream   uint32 = 305793916
	Counter_IngressPipeUpstreamAll          uint32 = 304792521
	Counter_IngressPipeAccountingDownstream uint32 = 302621207
	Counter_IngressPipeDownstreamDropped    uint32 = 315685570
	Counter_EgressPipeAccountingDownstream  uint32 = 311241884
	Counter_EgressPipeAccountingUpstream    uint32 = 309841876
	Counter_IngressPipeDownstreamTtlExpired uint32 = 305954198
	// Direct Counter IDs
	DirectCounter_IfTypes    uint32 = 331661032
	DirectCounter_MyStations uint32 = 333390111
	DirectCounter_Acls       uint32 = 325583051
	// Action IDs
	Action_IngressPipeSetIfType                      uint32 = 18538368
	Action_IngressPipeSetMyStation                   uint32 = 29456969
	Action_IngressPipeDownstreamCosSetCosId          uint32 = 24318838
	Action_IngressPipeUpstreamPunt                   uint32 = 27908888
	Action_IngressPipeAclSetPort                     uint32 = 21835758
	Action_IngressPipeSetAccountingId                uint32 = 31884918
	Action_IngressPipeDownstreamSetLine              uint32 = 17097684
	Action_IngressPipeAclDrop                        uint32 = 29272903
	Action_IngressPipeUpstreamSetLine                uint32 = 17659136
	Action_Nop                                       uint32 = 28485346
	Action_IngressPipeUpstreamRouteV4                uint32 = 31033793
	Action_WillDrop                                  uint32 = 28026654
	Action_IngressPipeUpstreamReject                 uint32 = 18981580
	Action_IngressPipeDownstreamMiss                 uint32 = 27308170
	Action_IngressPipeDownstreamSetPppoeAttachmentV4 uint32 = 32146370
	Action_IngressPipeAclPunt                        uint32 = 22515864
	Action_IngressPipeUpstreamCosSetCosId            uint32 = 26801906
	// Action Param IDs
	ActionParam_IngressPipeDownstreamCosSetCosId_CosId                uint32 = 1
	ActionParam_IngressPipeDownstreamSetLine_LineId                   uint32 = 1
	ActionParam_IngressPipeDownstreamSetPppoeAttachmentV4_Dmac        uint32 = 2
	ActionParam_IngressPipeDownstreamSetPppoeAttachmentV4_Port        uint32 = 1
	ActionParam_IngressPipeDownstreamSetPppoeAttachmentV4_CTag        uint32 = 4
	ActionParam_IngressPipeDownstreamSetPppoeAttachmentV4_STag        uint32 = 3
	ActionParam_IngressPipeDownstreamSetPppoeAttachmentV4_PppoeSessId uint32 = 5
	ActionParam_IngressPipeSetIfType_IfType                           uint32 = 1
	ActionParam_IngressPipeUpstreamCosSetCosId_CosId                  uint32 = 1
	ActionParam_IngressPipeAclSetPort_Port                            uint32 = 1
	ActionParam_IngressPipeUpstreamSetLine_LineId                     uint32 = 1
	ActionParam_IngressPipeSetAccountingId_AccountingId               uint32 = 1
	ActionParam_IngressPipeUpstreamRouteV4_Dmac                       uint32 = 2
	ActionParam_IngressPipeUpstreamRouteV4_Port                       uint32 = 1
	// Action Profile IDs
	ActionProfile_IngressPipeUpstreamEcmp uint32 = 286372544
	// Packet Metadata IDs
	PacketMeta_IngressPort uint32 = 1
	PacketMeta_EgressPort  uint32 = 1
)
