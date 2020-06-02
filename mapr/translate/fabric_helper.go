package translate

import (
	"encoding/binary"
	v1 "github.com/p4lang/p4runtime/go/p4/v1"
	"mapr/p4info"
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
		FieldId: p4info.FieldMatch_EgressVlanVlanId,
		FieldMatchType: &v1.FieldMatch_Exact_{
			Exact: &v1.FieldMatch_Exact{
				Value: getVlanIdValue(internalVlan),
			},
		},
	}
	matchEgressPort := v1.FieldMatch{
		FieldId: p4info.FieldMatch_EgressVlanEgPort,
		FieldMatchType: &v1.FieldMatch_Exact_{
			Exact: &v1.FieldMatch_Exact{
				Value: port,
			},
		},
	}
	actionPop := v1.TableAction{
		Type: &v1.TableAction_Action{Action: &v1.Action{
			ActionId: p4info.Action_EgressVlanPopVlan,
			Params:   nil,
		}},
	}
	return v1.TableEntry{
		TableId: p4info.Table_EgressVlan,
		Match:   []*v1.FieldMatch{&matchVlanId, &matchEgressPort},
		Action:  &actionPop,
	}
}

func createIngressPortVlanEntry(port []byte, internlVlan uint16, prio int32) v1.TableEntry {
	matchIngressPort := v1.FieldMatch{
		FieldId: p4info.FieldMatch_IngressPortVlanIgPort,
		FieldMatchType: &v1.FieldMatch_Exact_{
			Exact: &v1.FieldMatch_Exact{
				Value: port,
			},
		},
	}
	matchVlanIsValid := v1.FieldMatch{
		FieldId: p4info.FieldMatch_IngressPortVlanVlanIsValid,
		FieldMatchType: &v1.FieldMatch_Exact_{
			Exact: &v1.FieldMatch_Exact{
				Value: []byte{0},
			},
		},
	}
	actionPop := v1.TableAction{
		Type: &v1.TableAction_Action{Action: &v1.Action{
			ActionId: p4info.Action_IngressPortVlanPermitWithInternalVlan,
			Params: []*v1.Action_Param{
				{
					ParamId: 1,
					Value:   getVlanIdValue(internlVlan),
				},
			},
		}},
	}
	return v1.TableEntry{
		TableId:  p4info.Table_IngressPortVlan,
		Match:    []*v1.FieldMatch{&matchIngressPort, &matchVlanIsValid},
		Action:   &actionPop,
		Priority: prio,
	}
}

func createFwdClassifierEntry(port []byte, EthDst []byte, prio int32) v1.TableEntry {
	matchIngressPort := v1.FieldMatch{
		FieldId: p4info.FieldMatch_FwdClassifierIgPort,
		FieldMatchType: &v1.FieldMatch_Exact_{
			Exact: &v1.FieldMatch_Exact{
				Value: port,
			},
		},
	}
	matchEthDst := v1.FieldMatch{
		FieldId: p4info.FieldMatch_FwdClassifierEthDst,
		FieldMatchType: &v1.FieldMatch_Ternary_{
			Ternary: &v1.FieldMatch_Ternary{
				Value: EthDst,
				Mask:  []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
			},
		},
	}
	matchIpEthType := v1.FieldMatch{
		FieldId: p4info.FieldMatch_FwdClassifierIpEthType,
		FieldMatchType: &v1.FieldMatch_Exact_{
			Exact: &v1.FieldMatch_Exact{
				Value: getEthTypeValue(p4info.EthTypeIpv4),
			},
		},
	}
	actionPop := v1.TableAction{
		Type: &v1.TableAction_Action{Action: &v1.Action{
			ActionId: p4info.Action_FwdClassifierSetForwardingType,
			Params: []*v1.Action_Param{
				{
					ParamId: 1,
					Value:   []byte{p4info.FwdType_FwdIpv4Unicast}},
			},
		}},
	}
	return v1.TableEntry{
		TableId:  p4info.Table_FwdClassifier,
		Match:    []*v1.FieldMatch{&matchIngressPort, &matchEthDst, &matchIpEthType},
		Action:   &actionPop,
		Priority: prio,
	}
}
