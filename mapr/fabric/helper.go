package fabric

import (
	"bytes"
	"encoding/binary"
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
					Value:   []byte{FwdType_FwdIpv4Unicast}},
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

func createHashedSelectorMember(e *translate.NextHopEntry, smac []byte) v1.ActionProfileMember {
	return v1.ActionProfileMember{
		ActionProfileId: ActionProfile_FabricIngressNextHashedSelector,
		MemberId:        e.Id,
		Action: &v1.Action{
			ActionId: Action_FabricIngressNextRoutingHashed,
			Params: []*v1.Action_Param{
				{
					ParamId: ActionParam_FabricIngressNextRoutingHashed_PortNum,
					Value:   e.Port,
				},
				{
					ParamId: ActionParam_FabricIngressNextRoutingHashed_Dmac,
					Value:   e.MacAddr,
				},
				{
					ParamId: ActionParam_FabricIngressNextRoutingHashed_Smac,
					Value:   smac,
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

func createRouteV4Entry(e *translate.RouteV4Entry) v1.TableEntry {
	return v1.TableEntry{
		TableId: Table_FabricIngressForwardingRoutingV4,
		Match: []*v1.FieldMatch{{
			FieldId: Hdr_FabricIngressForwardingRoutingV4_Ipv4Dst,
			FieldMatchType: &v1.FieldMatch_Lpm{Lpm: &v1.FieldMatch_LPM{
				Value:     e.Ipv4Addr,
				PrefixLen: e.PrefixLen,
			}}}},
		Action: &v1.TableAction{Type: &v1.TableAction_Action{Action: &v1.Action{
			ActionId: Action_FabricIngressForwardingSetNextIdRoutingV4,
			Params: []*v1.Action_Param{{
				ParamId: ActionParam_FabricIngressForwardingSetNextIdRoutingV4_NextId,
				Value:   getNextIdValue(e.NextHopGroupId),
			}},
		}}},
	}
}

func createNextVlanEntry(e *translate.RouteV4Entry, outputTag []byte) v1.TableEntry {
	return v1.TableEntry{
		TableId: Table_FabricIngressNextNextVlan,
		Match: []*v1.FieldMatch{{
			FieldId: Hdr_FabricIngressForwardingRoutingV4_Ipv4Dst,
			FieldMatchType: &v1.FieldMatch_Exact_{Exact: &v1.FieldMatch_Exact{
				Value: getNextIdValue(e.NextHopGroupId),
			}}}},
		Action: &v1.TableAction{Type: &v1.TableAction_Action{Action: &v1.Action{
			ActionId: Action_FabricIngressNextSetVlan,
			Params: []*v1.Action_Param{{
				ParamId: ActionParam_FabricIngressNextSetVlan_VlanId,
				Value:   outputTag,
			}},
		}}},
	}
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

func getTargetEntriesUpstreamByLineId(p fabricProcessor, lineId []byte) []*v1.TableEntry {
	return p.ctx.Target().FilterTableEntries(
		func(entry *v1.TableEntry) bool {
			if entry.TableId == Table_FabricIngressBngIngressTLineMap &&
				entry.GetAction().GetAction().ActionId == Action_FabricIngressBngIngressSetLine {
				for _, v := range entry.GetAction().GetAction().GetParams() {
					if bytes.Equal(v.Value, lineId) {
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
