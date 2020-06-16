package fabric

import (
	"bytes"
	"encoding/binary"
	"fmt"
	v1 "github.com/p4lang/p4runtime/go/p4/v1"
	"mapr/translate"
)

func createUpdateEntry(entry *v1.TableEntry, uType v1.Update_Type) *v1.Update {
	return &v1.Update{
		Type:   uType,
		Entity: &v1.Entity{Entity: &v1.Entity_TableEntry{TableEntry: entry}},
	}
}

func createUpdateEntries(entries []*v1.TableEntry, uType v1.Update_Type) []*v1.Update {
	phyUpdateEntry := make([]*v1.Update, 0)
	for i := range entries {
		phyUpdateEntry = append(phyUpdateEntry, createUpdateEntry(entries[i], uType))
	}
	return phyUpdateEntry
}

func createUpdateActProfMember(member *v1.ActionProfileMember, uType v1.Update_Type) *v1.Update {
	return &v1.Update{
		Type:   uType,
		Entity: &v1.Entity{Entity: &v1.Entity_ActionProfileMember{ActionProfileMember: member}},
	}
}

func createUpdateActProfGroup(group *v1.ActionProfileGroup, uType v1.Update_Type) *v1.Update {
	return &v1.Update{
		Type:   uType,
		Entity: &v1.Entity{Entity: &v1.Entity_ActionProfileGroup{ActionProfileGroup: group}},
	}
}

func getVlanIdValue(vlanId uint16) []byte {
	vlanIdByteSlice := make([]byte, 2)
	binary.BigEndian.PutUint16(vlanIdByteSlice, vlanId)
	return vlanIdByteSlice
}

func getEthTypeValue(ethType uint16) []byte {
	ethTypeByteSlice := make([]byte, 2)
	binary.BigEndian.PutUint16(ethTypeByteSlice, ethType)
	return ethTypeByteSlice
}
func getNextIdValue(nextId uint32) []byte {
	bytes := make([]byte, 4)
	binary.BigEndian.PutUint32(bytes, nextId)
	return bytes
}

func getUInt32FromByteSlice(val []byte) uint32 {
	return binary.BigEndian.Uint32(val)
}

func createEgressVlanPopEntry(port []byte, internalVlan uint16) v1.TableEntry {
	matchVlanId := v1.FieldMatch{
		FieldId: Hdr_FabricEgressEgressNextEgressVlan_VlanId,
		FieldMatchType: &v1.FieldMatch_Exact_{
			Exact: &v1.FieldMatch_Exact{
				Value: getVlanIdValue(internalVlan),
			},
		},
	}
	matchEgressPort := v1.FieldMatch{
		FieldId: Hdr_FabricEgressEgressNextEgressVlan_EgPort,
		FieldMatchType: &v1.FieldMatch_Exact_{
			Exact: &v1.FieldMatch_Exact{
				Value: port,
			},
		},
	}
	actionPop := v1.TableAction{
		Type: &v1.TableAction_Action{Action: &v1.Action{
			ActionId: Action_FabricEgressEgressNextPopVlan,
			Params:   nil,
		}},
	}
	return v1.TableEntry{
		TableId: Table_FabricEgressEgressNextEgressVlan,
		Match:   []*v1.FieldMatch{&matchVlanId, &matchEgressPort},
		Action:  &actionPop,
	}
}

func createIngressPortVlanEntryPermit(port []byte, vlanId []byte, innerVlanId []byte, internalVlan []byte, prio int32) v1.TableEntry {
	matchFields := make([]*v1.FieldMatch, 0)
	matchFields = append(matchFields, &v1.FieldMatch{
		FieldId: Hdr_FabricIngressFilteringIngressPortVlan_IgPort,
		FieldMatchType: &v1.FieldMatch_Exact_{
			Exact: &v1.FieldMatch_Exact{
				Value: port,
			},
		},
	})
	if vlanId != nil {
		matchFields = append(matchFields, &v1.FieldMatch{
			FieldId: Hdr_FabricIngressFilteringIngressPortVlan_VlanIsValid,
			FieldMatchType: &v1.FieldMatch_Exact_{
				Exact: &v1.FieldMatch_Exact{
					Value: []byte{0x01},
				},
			},
		})
		matchFields = append(matchFields, &v1.FieldMatch{
			FieldId: Hdr_FabricIngressFilteringIngressPortVlan_VlanId,
			FieldMatchType: &v1.FieldMatch_Ternary_{
				Ternary: &v1.FieldMatch_Ternary{
					Value: vlanId,
					Mask:  []byte{0x0F, 0xFF},
				},
			},
		})
		if innerVlanId != nil {
			matchFields = append(matchFields, &v1.FieldMatch{
				FieldId: Hdr_FabricIngressFilteringIngressPortVlan_InnerVlanId,
				FieldMatchType: &v1.FieldMatch_Ternary_{
					Ternary: &v1.FieldMatch_Ternary{
						Value: innerVlanId,
						Mask:  []byte{0x0F, 0xFF},
					},
				},
			})
		}
	} else {
		matchFields = append(matchFields, &v1.FieldMatch{
			FieldId: Hdr_FabricIngressFilteringIngressPortVlan_VlanIsValid,
			FieldMatchType: &v1.FieldMatch_Exact_{
				Exact: &v1.FieldMatch_Exact{
					Value: []byte{0x00},
				},
			},
		})
	}
	var actionPop v1.TableAction
	if internalVlan != nil {
		actionPop = v1.TableAction{
			Type: &v1.TableAction_Action{Action: &v1.Action{
				ActionId: Action_FabricIngressFilteringPermitWithInternalVlan,
				Params: []*v1.Action_Param{
					{
						ParamId: ActionParam_FabricIngressFilteringPermitWithInternalVlan_VlanId,
						Value:   internalVlan,
					},
				},
			}},
		}
	} else {
		actionPop = v1.TableAction{
			Type: &v1.TableAction_Action{Action: &v1.Action{
				ActionId: Action_FabricIngressFilteringPermit,
			}},
		}
	}
	return v1.TableEntry{
		TableId:  Table_FabricIngressFilteringIngressPortVlan,
		Match:    matchFields,
		Action:   &actionPop,
		Priority: prio,
	}
}

func createFwdClassifierEntry(port []byte, EthDst []byte, prio int32) v1.TableEntry {
	matchIngressPort := v1.FieldMatch{
		FieldId: Hdr_FabricIngressFilteringFwdClassifier_IgPort,
		FieldMatchType: &v1.FieldMatch_Exact_{
			Exact: &v1.FieldMatch_Exact{
				Value: port,
			},
		},
	}
	matchEthDst := v1.FieldMatch{
		FieldId: Hdr_FabricIngressFilteringFwdClassifier_EthDst,
		FieldMatchType: &v1.FieldMatch_Ternary_{
			Ternary: &v1.FieldMatch_Ternary{
				Value: EthDst,
				Mask:  []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
			},
		},
	}
	matchIpEthType := v1.FieldMatch{
		FieldId: Hdr_FabricIngressFilteringFwdClassifier_IpEthType,
		FieldMatchType: &v1.FieldMatch_Exact_{
			Exact: &v1.FieldMatch_Exact{
				Value: getEthTypeValue(EthTypeIpv4),
			},
		},
	}
	actionPop := v1.TableAction{
		Type: &v1.TableAction_Action{Action: &v1.Action{
			ActionId: Action_FabricIngressFilteringSetForwardingType,
			Params: []*v1.Action_Param{
				{
					ParamId: ActionParam_FabricIngressFilteringSetForwardingType_FwdType,
					Value:   []byte{FwdTypeIpv4Unicast}},
			},
		}},
	}
	return v1.TableEntry{
		TableId:  Table_FabricIngressFilteringFwdClassifier,
		Match:    []*v1.FieldMatch{&matchIngressPort, &matchEthDst, &matchIpEthType},
		Action:   &actionPop,
		Priority: prio,
	}
}

func createPppoePuntEntry(pppoeCode []byte, pppoeProto []byte, prio int32) v1.TableEntry {
	match := []*v1.FieldMatch{{
		FieldId: Hdr_FabricIngressBngIngressUpstreamTPppoeCp_PppoeCode,
		FieldMatchType: &v1.FieldMatch_Exact_{
			Exact: &v1.FieldMatch_Exact{
				Value: pppoeCode,
			}}},
	}
	if pppoeProto != nil {
		match = append(match, &v1.FieldMatch{
			FieldId: Hdr_FabricIngressBngIngressUpstreamTPppoeCp_PppoeProtocol,
			FieldMatchType: &v1.FieldMatch_Ternary_{
				Ternary: &v1.FieldMatch_Ternary{
					Value: pppoeProto,
					Mask:  []byte{0xFF, 0xFF},
				}}})
	}
	return v1.TableEntry{
		TableId: Table_FabricIngressBngIngressUpstreamTPppoeCp,
		Match:   match,
		Action: &v1.TableAction{
			Type: &v1.TableAction_Action{Action: &v1.Action{
				ActionId: Action_FabricIngressBngIngressUpstreamPuntToCpu,
			}}},
		Priority: prio,
	}
}

func createHashedSelectorMember(memberId uint32, port []byte, dMac []byte, sMac []byte) v1.ActionProfileMember {
	return v1.ActionProfileMember{
		ActionProfileId: ActionProfile_FabricIngressNextHashedSelector,
		MemberId:        memberId,
		Action: &v1.Action{
			ActionId: Action_FabricIngressNextRoutingHashed,
			Params: []*v1.Action_Param{
				{
					ParamId: ActionParam_FabricIngressNextRoutingHashed_PortNum,
					Value:   port,
				},
				{
					ParamId: ActionParam_FabricIngressNextRoutingHashed_Dmac,
					Value:   dMac,
				},
				{
					ParamId: ActionParam_FabricIngressNextRoutingHashed_Smac,
					Value:   sMac,
				},
			},
		},
	}
}

func createNextHashedEntry(nextId uint32) v1.TableEntry {
	return v1.TableEntry{
		TableId: Table_FabricIngressNextHashed,
		Match: []*v1.FieldMatch{{
			FieldId: Hdr_FabricIngressNextHashed_NextId,
			FieldMatchType: &v1.FieldMatch_Exact_{Exact: &v1.FieldMatch_Exact{
				Value: getNextIdValue(nextId),
			}}}},
		Action: &v1.TableAction{Type: &v1.TableAction_ActionProfileGroupId{
			ActionProfileGroupId: nextId}},
	}
}

func createRouteV4Entry(nextId uint32, ipv4Addr []byte, prefixLen int32) v1.TableEntry {
	return v1.TableEntry{
		TableId: Table_FabricIngressForwardingRoutingV4,
		Match: []*v1.FieldMatch{{
			FieldId: Hdr_FabricIngressForwardingRoutingV4_Ipv4Dst,
			FieldMatchType: &v1.FieldMatch_Lpm{Lpm: &v1.FieldMatch_LPM{
				Value:     ipv4Addr,
				PrefixLen: prefixLen,
			}}}},
		Action: &v1.TableAction{Type: &v1.TableAction_Action{Action: &v1.Action{
			ActionId: Action_FabricIngressForwardingSetNextIdRoutingV4,
			Params: []*v1.Action_Param{{
				ParamId: ActionParam_FabricIngressForwardingSetNextIdRoutingV4_NextId,
				Value:   getNextIdValue(nextId),
			}},
		}}},
	}
}

func createNextVlanEntry(nextId uint32, vlanId []byte, innerVlanid []byte) v1.TableEntry {
	var action *v1.TableAction
	if innerVlanid == nil {
		action = &v1.TableAction{Type: &v1.TableAction_Action{Action: &v1.Action{
			ActionId: Action_FabricIngressNextSetVlan,
			Params: []*v1.Action_Param{{
				ParamId: ActionParam_FabricIngressNextSetVlan_VlanId,
				Value:   vlanId,
			}},
		}}}
	} else {
		action = &v1.TableAction{Type: &v1.TableAction_Action{Action: &v1.Action{
			ActionId: Action_FabricIngressNextSetDoubleVlan,
			Params: []*v1.Action_Param{{
				ParamId: ActionParam_FabricIngressNextSetDoubleVlan_OuterVlanId,
				Value:   vlanId,
			}, {
				ParamId: ActionParam_FabricIngressNextSetDoubleVlan_InnerVlanId,
				Value:   innerVlanid,
			}},
		}}}
	}

	return v1.TableEntry{
		TableId: Table_FabricIngressNextNextVlan,
		Match: []*v1.FieldMatch{{
			FieldId: Hdr_FabricIngressForwardingRoutingV4_Ipv4Dst,
			FieldMatchType: &v1.FieldMatch_Exact_{Exact: &v1.FieldMatch_Exact{
				Value: getNextIdValue(nextId),
			}}}},
		Action: action}
}

func createLineMapEntry(sTag []byte, cTag []byte, lineId []byte) v1.TableEntry {
	matches := []*v1.FieldMatch{{
		FieldId: Hdr_FabricIngressBngIngressTLineMap_STag,
		FieldMatchType: &v1.FieldMatch_Exact_{
			Exact: &v1.FieldMatch_Exact{
				Value: sTag,
			}}}, {
		FieldId: Hdr_FabricIngressBngIngressTLineMap_CTag,
		FieldMatchType: &v1.FieldMatch_Exact_{
			Exact: &v1.FieldMatch_Exact{
				Value: cTag,
			}}},
	}
	return v1.TableEntry{
		TableId: Table_FabricIngressBngIngressTLineMap,
		Match:   matches,
		Action: &v1.TableAction{Type: &v1.TableAction_Action{Action: &v1.Action{
			ActionId: Action_FabricIngressBngIngressSetLine,
			Params: []*v1.Action_Param{{
				ParamId: ActionParam_FabricIngressBngIngressSetLine_LineId,
				Value:   lineId,
			}}}}},
	}
}

func createPppoeTermV4(lineId []byte, ipv4Addr []byte, pppoeSessId []byte) v1.TableEntry {
	matches := []*v1.FieldMatch{{
		FieldId: Hdr_FabricIngressBngIngressUpstreamTPppoeTermV4_LineId,
		FieldMatchType: &v1.FieldMatch_Exact_{
			Exact: &v1.FieldMatch_Exact{
				Value: lineId,
			}}}, {
		FieldId: Hdr_FabricIngressBngIngressUpstreamTPppoeTermV4_Ipv4Src,
		FieldMatchType: &v1.FieldMatch_Exact_{
			Exact: &v1.FieldMatch_Exact{
				Value: ipv4Addr,
			}}}, {
		FieldId: Hdr_FabricIngressBngIngressUpstreamTPppoeTermV4_PppoeSessionId,
		FieldMatchType: &v1.FieldMatch_Exact_{
			Exact: &v1.FieldMatch_Exact{
				Value: pppoeSessId,
			}}},
	}
	return v1.TableEntry{
		TableId: Table_FabricIngressBngIngressUpstreamTPppoeTermV4,
		Match:   matches,
		Action: &v1.TableAction{Type: &v1.TableAction_Action{Action: &v1.Action{
			ActionId: Action_FabricIngressBngIngressUpstreamTermEnabledV4,
		}}},
	}
}

func createAclEntry(e *translate.AclEntry) (v1.TableEntry, error) {
	matches := make([]*v1.FieldMatch, 0)
	for _, m := range e.Match {
		switch m.FieldId {
		case translate.Hdr_IngressPipeAclAcls_Port:
			matches = append(matches, createMatchAcl(m.GetTernary().Value, m.GetTernary().Mask, Hdr_FabricIngressAclAcl_IgPort))
		//TODO: case translate.Hdr_IngressPipeAclAcls_IfType:
		case translate.Hdr_IngressPipeAclAcls_EthSrc:
			matches = append(matches, createMatchAcl(m.GetTernary().Value, m.GetTernary().Mask, Hdr_FabricIngressAclAcl_EthSrc))
		case translate.Hdr_IngressPipeAclAcls_EthDst:
			matches = append(matches, createMatchAcl(m.GetTernary().Value, m.GetTernary().Mask, Hdr_FabricIngressAclAcl_EthDst))
		case translate.Hdr_IngressPipeAclAcls_EthType:
			matches = append(matches, createMatchAcl(m.GetTernary().Value, m.GetTernary().Mask, Hdr_FabricIngressAclAcl_EthType))
		case translate.Hdr_IngressPipeAclAcls_Ipv4Src:
			matches = append(matches, createMatchAcl(m.GetTernary().Value, m.GetTernary().Mask, Hdr_FabricIngressAclAcl_Ipv4Src))
		case translate.Hdr_IngressPipeAclAcls_Ipv4Dst:
			matches = append(matches, createMatchAcl(m.GetTernary().Value, m.GetTernary().Mask, Hdr_FabricIngressAclAcl_Ipv4Dst))
		case translate.Hdr_IngressPipeAclAcls_Ipv4Proto:
			matches = append(matches, createMatchAcl(m.GetTernary().Value, m.GetTernary().Mask, Hdr_FabricIngressAclAcl_IpProto))
		case translate.Hdr_IngressPipeAclAcls_L4Sport:
			matches = append(matches, createMatchAcl(m.GetTernary().Value, m.GetTernary().Mask, Hdr_FabricIngressAclAcl_L4Sport))
		case translate.Hdr_IngressPipeAclAcls_L4Dport:
			matches = append(matches, createMatchAcl(m.GetTernary().Value, m.GetTernary().Mask, Hdr_FabricIngressAclAcl_L4Dport))
		default:
			return v1.TableEntry{}, fmt.Errorf("unsupported ACL match for fabric.p4: %s", m)
		}

	}
	var action v1.TableAction
	switch e.Action.GetAction().ActionId {
	case translate.Action_IngressPipeAclPunt:
		action = v1.TableAction{Type: &v1.TableAction_Action{Action: &v1.Action{
			ActionId: Action_FabricIngressAclPuntToCpu,
		}}}
	case translate.Action_IngressPipeAclDrop:
		action = v1.TableAction{Type: &v1.TableAction_Action{Action: &v1.Action{
			ActionId: Action_FabricIngressAclDrop,
		}}}
	// TODO: case translate.Action_IngressPipeAclSetPort: this case requires to use the indirect forwarding on fabric (next_id + next.simple/next.hashed)
	default:
		return v1.TableEntry{}, fmt.Errorf("unrecognized acl action: %s", e.Action.GetAction())
	}

	return v1.TableEntry{
		TableId:  Table_FabricIngressAclAcl,
		Match:    matches,
		Action:   &action,
		Priority: e.Priority,
	}, nil
}

func createMatchAcl(value []byte, mask []byte, fieldId uint32) *v1.FieldMatch {
	return &v1.FieldMatch{
		FieldId: fieldId,
		FieldMatchType: &v1.FieldMatch_Ternary_{
			Ternary: &v1.FieldMatch_Ternary{
				Value: value,
				Mask:  mask,
			}}}
}

func createLineSessionMap(lineId []byte, pppoeSessId []byte) v1.TableEntry {
	return v1.TableEntry{
		TableId: Table_FabricIngressBngIngressDownstreamTLineSessionMap,
		Match: []*v1.FieldMatch{{
			FieldId: Hdr_FabricIngressBngIngressDownstreamTLineSessionMap_LineId,
			FieldMatchType: &v1.FieldMatch_Exact_{
				Exact: &v1.FieldMatch_Exact{
					Value: lineId,
				}}}},
		Action: &v1.TableAction{Type: &v1.TableAction_Action{Action: &v1.Action{
			ActionId: Action_FabricIngressBngIngressDownstreamSetSession,
			Params: []*v1.Action_Param{{
				ParamId: ActionParam_FabricIngressBngIngressDownstreamSetSession_PppoeSessionId,
				Value:   pppoeSessId,
			}}}}},
	}
}

// Get all the table entries for the upstream direction for a given line ID
func getTargetEntriesUpstreamByLineId(p fabricProcessor, lineId []byte) []*v1.TableEntry {
	return p.ctx.Target().FilterTableEntries(
		func(entry *v1.TableEntry) bool {
			if entry.TableId == Table_FabricIngressBngIngressTLineMap &&
				entry.GetAction().GetAction().ActionId == Action_FabricIngressBngIngressSetLine {
				for _, v := range entry.GetAction().GetAction().GetParams() {
					if v.ParamId == ActionParam_FabricIngressBngIngressSetLine_LineId &&
						bytes.Equal(v.Value, lineId) {
						return true
					}
				}
			}
			if entry.TableId == Table_FabricIngressBngIngressUpstreamTPppoeTermV4 {
				for _, m := range entry.GetMatch() {
					if m.FieldId == Hdr_FabricIngressBngIngressUpstreamTPppoeTermV4_LineId &&
						bytes.Equal(m.GetExact().Value, lineId) {
						return true
					}
				}
			}
			return false
		})
}

// Get all the table entries for the downstream direction for a given line ID
func getTargetEntriesDownstreamByLineId(p fabricProcessor, lineId []byte) []*v1.TableEntry {
	return p.ctx.Target().FilterTableEntries(
		func(entry *v1.TableEntry) bool {
			if entry.TableId == Table_FabricIngressBngIngressTLineMap &&
				entry.GetAction().GetAction().ActionId == Action_FabricIngressBngIngressSetLine {
				for _, v := range entry.GetAction().GetAction().GetParams() {
					if v.ParamId == ActionParam_FabricIngressBngIngressSetLine_LineId &&
						bytes.Equal(v.Value, lineId) {
						return true
					}
				}
			}
			if entry.TableId == Table_FabricIngressBngIngressDownstreamTLineSessionMap {
				for _, m := range entry.GetMatch() {
					if m.FieldId == Hdr_FabricIngressBngIngressDownstreamTLineSessionMap_LineId &&
						bytes.Equal(m.GetExact().Value, lineId) {
						return true
					}
				}
			}
			if entry.TableId == Table_FabricIngressForwardingRoutingV4 &&
				entry.GetAction().GetAction().ActionId == Action_FabricIngressForwardingSetNextIdRoutingV4 {
				for _, v := range entry.GetAction().GetAction().GetParams() {
					if v.ParamId == ActionParam_FabricIngressForwardingSetNextIdRoutingV4_NextId &&
						bytes.Equal(v.Value, lineId) {
						return true
					}
				}
			}
			if entry.TableId == Table_FabricIngressNextHashed {
				for _, m := range entry.GetMatch() {
					if m.FieldId == Hdr_FabricIngressNextHashed_NextId &&
						bytes.Equal(m.GetExact().Value, lineId) {
						return true
					}
				}
			}
			if entry.TableId == Table_FabricIngressNextNextVlan {
				for _, m := range entry.GetMatch() {
					if m.FieldId == Hdr_FabricIngressNextNextVlan_NextId &&
						bytes.Equal(m.GetExact().Value, lineId) {
						return true
					}
				}
			}
			return false
		})
}

func insertOrModifyTableEntries(p fabricProcessor, tableEntries []*v1.TableEntry) (updateEntries []*v1.Update) {
	// Query target store to understand if insert or modify
	for _, v := range tableEntries {
		key := translate.KeyFromTableEntry(v)
		targetTableEntry := p.ctx.Target().GetTableEntry(&key)
		updateType := v1.Update_INSERT
		if targetTableEntry != nil {
			// TODO: we could filter only the updated entries instead of always pushing a MODIFY to the target
			updateType = v1.Update_MODIFY
		}
		updateEntries = append(updateEntries, createUpdateEntry(v, updateType))
	}
	return
}
