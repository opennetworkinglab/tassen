package translate

import (
	"encoding/binary"
	v1 "github.com/p4lang/p4runtime/go/p4/v1"
	"mapr/fabric"
)

func createUpdateEntry(entry *v1.TableEntry, uType v1.Update_Type) *v1.Update {
	return &v1.Update{
		Type:   uType,
		Entity: &v1.Entity{Entity: &v1.Entity_TableEntry{TableEntry: entry}},
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

func createEgressVlanPopEntry(port []byte, internalVlan uint16) v1.TableEntry {
	matchVlanId := v1.FieldMatch{
		FieldId: fabric.Hdr_FabricEgressEgressNextEgressVlan_VlanId,
		FieldMatchType: &v1.FieldMatch_Exact_{
			Exact: &v1.FieldMatch_Exact{
				Value: getVlanIdValue(internalVlan),
			},
		},
	}
	matchEgressPort := v1.FieldMatch{
		FieldId: fabric.Hdr_FabricEgressEgressNextEgressVlan_EgPort,
		FieldMatchType: &v1.FieldMatch_Exact_{
			Exact: &v1.FieldMatch_Exact{
				Value: port,
			},
		},
	}
	actionPop := v1.TableAction{
		Type: &v1.TableAction_Action{Action: &v1.Action{
			ActionId: fabric.Action_FabricEgressEgressNextPopVlan,
			Params:   nil,
		}},
	}
	return v1.TableEntry{
		TableId: fabric.Table_FabricEgressEgressNextEgressVlan,
		Match:   []*v1.FieldMatch{&matchVlanId, &matchEgressPort},
		Action:  &actionPop,
	}
}

func createIngressPortVlanEntryPermit(port []byte, vlanId []byte, innerVlanId []byte, internalVlan []byte, prio int32) v1.TableEntry {
	matchFields := make([]*v1.FieldMatch, 0)
	matchFields = append(matchFields, &v1.FieldMatch{
		FieldId: fabric.Hdr_FabricIngressFilteringIngressPortVlan_IgPort,
		FieldMatchType: &v1.FieldMatch_Exact_{
			Exact: &v1.FieldMatch_Exact{
				Value: port,
			},
		},
	})
	if vlanId != nil {
		matchFields = append(matchFields, &v1.FieldMatch{
			FieldId: fabric.Hdr_FabricIngressFilteringIngressPortVlan_VlanIsValid,
			FieldMatchType: &v1.FieldMatch_Exact_{
				Exact: &v1.FieldMatch_Exact{
					Value: []byte{1},
				},
			},
		})
		matchFields = append(matchFields, &v1.FieldMatch{
			FieldId: fabric.Hdr_FabricIngressFilteringIngressPortVlan_VlanId,
			FieldMatchType: &v1.FieldMatch_Ternary_{
				Ternary: &v1.FieldMatch_Ternary{
					Value: vlanId,
					Mask:  []byte{0xFF, 0xFF},
				},
			},
		})
		if innerVlanId != nil {
			matchFields = append(matchFields, &v1.FieldMatch{
				FieldId: fabric.Hdr_FabricIngressFilteringIngressPortVlan_InnerVlanId,
				FieldMatchType: &v1.FieldMatch_Ternary_{
					Ternary: &v1.FieldMatch_Ternary{
						Value: innerVlanId,
						Mask:  []byte{0xFF, 0xFF},
					},
				},
			})
		}
	} else {
		matchFields = append(matchFields, &v1.FieldMatch{
			FieldId: fabric.Hdr_FabricIngressFilteringIngressPortVlan_VlanIsValid,
			FieldMatchType: &v1.FieldMatch_Exact_{
				Exact: &v1.FieldMatch_Exact{
					Value: []byte{0},
				},
			},
		})
	}
	var actionPop v1.TableAction
	if innerVlanId != nil {
		actionPop = v1.TableAction{
			Type: &v1.TableAction_Action{Action: &v1.Action{
				ActionId: fabric.Action_FabricIngressFilteringPermitWithInternalVlan,
				Params: []*v1.Action_Param{
					{
						ParamId: fabric.ActionParam_FabricIngressFilteringPermitWithInternalVlan_VlanId,
						Value:   internalVlan,
					},
				},
			}},
		}
	} else {
		actionPop = v1.TableAction{
			Type: &v1.TableAction_Action{Action: &v1.Action{
				ActionId: fabric.Action_FabricIngressFilteringPermit,
			}},
		}
	}
	return v1.TableEntry{
		TableId:  fabric.Table_FabricIngressFilteringIngressPortVlan,
		Match:    matchFields,
		Action:   &actionPop,
		Priority: prio,
	}
}

func createFwdClassifierEntry(port []byte, EthDst []byte, prio int32) v1.TableEntry {
	matchIngressPort := v1.FieldMatch{
		FieldId: fabric.Hdr_FabricIngressFilteringFwdClassifier_IgPort,
		FieldMatchType: &v1.FieldMatch_Exact_{
			Exact: &v1.FieldMatch_Exact{
				Value: port,
			},
		},
	}
	matchEthDst := v1.FieldMatch{
		FieldId: fabric.Hdr_FabricIngressFilteringFwdClassifier_EthDst,
		FieldMatchType: &v1.FieldMatch_Ternary_{
			Ternary: &v1.FieldMatch_Ternary{
				Value: EthDst,
				Mask:  []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
			},
		},
	}
	matchIpEthType := v1.FieldMatch{
		FieldId: fabric.Hdr_FabricIngressFilteringFwdClassifier_IpEthType,
		FieldMatchType: &v1.FieldMatch_Exact_{
			Exact: &v1.FieldMatch_Exact{
				Value: getEthTypeValue(EthTypeIpv4),
			},
		},
	}
	actionPop := v1.TableAction{
		Type: &v1.TableAction_Action{Action: &v1.Action{
			ActionId: fabric.Action_FabricIngressFilteringSetForwardingType,
			Params: []*v1.Action_Param{
				{
					ParamId: fabric.ActionParam_FabricIngressFilteringSetForwardingType_FwdType,
					Value:   []byte{FwdType_FwdIpv4Unicast}},
			},
		}},
	}
	return v1.TableEntry{
		TableId:  fabric.Table_FabricIngressFilteringFwdClassifier,
		Match:    []*v1.FieldMatch{&matchIngressPort, &matchEthDst, &matchIpEthType},
		Action:   &actionPop,
		Priority: prio,
	}
}
