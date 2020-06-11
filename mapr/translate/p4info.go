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
	Hdr_IngressPipeUpstreamLines_Port                uint32 = 1
	Hdr_IngressPipeUpstreamLines_CTag                uint32 = 2
	Hdr_IngressPipeUpstreamLines_STag                uint32 = 3
	Hdr_IngressPipeDownstreamPppoeSessions_LineId    uint32 = 1
	Hdr_IngressPipeUpstreamPppoePunts_PppoeCode      uint32 = 1
	Hdr_IngressPipeUpstreamPppoePunts_PppoeProto     uint32 = 2
	Hdr_IngressPipeDownstreamVids_LineId             uint32 = 1
	Hdr_IngressPipeRoutingRoutesV4_Ipv4Dst           uint32 = 1
	Hdr_IngressPipeMyStations_Port                   uint32 = 1
	Hdr_IngressPipeMyStations_EthDst                 uint32 = 2
	Hdr_IngressPipeUpstreamAttachmentsV4_EthSrc      uint32 = 2
	Hdr_IngressPipeUpstreamAttachmentsV4_Ipv4Src     uint32 = 3
	Hdr_IngressPipeUpstreamAttachmentsV4_LineId      uint32 = 1
	Hdr_IngressPipeUpstreamAttachmentsV4_PppoeSessId uint32 = 4
	Hdr_IngressPipeIfTypes_Port                      uint32 = 1
	// Table IDs
	Table_IngressPipeMyStations              uint32 = 49392761
	Table_IngressPipeUpstreamLines           uint32 = 33956689
	Table_IngressPipeUpstreamAttachmentsV4   uint32 = 44507663
	Table_IngressPipeAclAcls                 uint32 = 43911884
	Table_IngressPipeUpstreamPppoePunts      uint32 = 39053621
	Table_IngressPipeIfTypes                 uint32 = 38498675
	Table_IngressPipeDownstreamLinesV4       uint32 = 44334275
	Table_IngressPipeRoutingRoutesV4         uint32 = 40572658
	Table_IngressPipeDownstreamVids          uint32 = 34456456
	Table_IngressPipeDownstreamPppoeSessions uint32 = 39589935
	// Indirect Counter IDs
	Counter_IngressPipeRoutingRouted     uint32 = 307231920
	Counter_IngressPipeUpstreamSpoofed   uint32 = 314616893
	Counter_IngressPipeUpstreamPunted    uint32 = 310787420
	Counter_IngressPipeUpstreamAll       uint32 = 304792521
	Counter_IngressPipeDownstreamDropped uint32 = 315685570
	Counter_IngressPipeRoutingTtlExpired uint32 = 314096939
	// Direct Counter IDs
	DirectCounter_IfTypes    uint32 = 331661032
	DirectCounter_MyStations uint32 = 333390111
	DirectCounter_Acls       uint32 = 325583051
	// Action IDs
	Action_IngressPipeSetIfType              uint32 = 18538368
	Action_IngressPipeRoutingRouteV4         uint32 = 21408227
	Action_IngressPipeDownstreamSetVids      uint32 = 23385620
	Action_IngressPipeUpstreamPunt           uint32 = 27908888
	Action_IngressPipeSetMyStation           uint32 = 29456969
	Action_IngressPipeDownstreamSetLine      uint32 = 17097684
	Action_IngressPipeAclDrop                uint32 = 29272903
	Action_IngressPipeUpstreamSetLine        uint32 = 17659136
	Action_Nop                               uint32 = 28485346
	Action_IngressPipeDownstreamSetPppoeSess uint32 = 27412451
	Action_IngressPipeUpstreamAllow          uint32 = 23630800
	Action_IngressPipeUpstreamReject         uint32 = 18981580
	Action_IngressPipeDownstreamMiss         uint32 = 27308170
	Action_IngressPipeAclSetPort             uint32 = 21835758
	Action_IngressPipeAclPunt                uint32 = 22515864
	Action_DropNow                           uint32 = 31962786
	// Action Param IDs
	ActionParam_IngressPipeDownstreamSetPppoeSess_PppoeSessId uint32 = 1
	ActionParam_IngressPipeDownstreamSetLine_LineId           uint32 = 1
	ActionParam_IngressPipeSetIfType_IfType                   uint32 = 1
	ActionParam_IngressPipeAclSetPort_Port                    uint32 = 1
	ActionParam_IngressPipeUpstreamSetLine_LineId             uint32 = 1
	ActionParam_IngressPipeDownstreamSetVids_STag             uint32 = 2
	ActionParam_IngressPipeDownstreamSetVids_CTag             uint32 = 1
	ActionParam_IngressPipeRoutingRouteV4_Dmac                uint32 = 2
	ActionParam_IngressPipeRoutingRouteV4_Port                uint32 = 1
	// Action Profile IDs
	ActionProfile_IngressPipeRoutingEcmp uint32 = 293424976
	// Packet Metadata IDs
	PacketMeta_IngressPort uint32 = 1
	PacketMeta_EgressPort  uint32 = 1
)
