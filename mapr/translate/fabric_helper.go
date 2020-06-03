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
		FieldId: fabric.FieldMatch_EgressVlanVlanId,
		FieldMatchType: &v1.FieldMatch_Exact_{
			Exact: &v1.FieldMatch_Exact{
				Value: getVlanIdValue(internalVlan),
			},
		},
	}
	matchEgressPort := v1.FieldMatch{
		FieldId: fabric.FieldMatch_EgressVlanEgPort,
		FieldMatchType: &v1.FieldMatch_Exact_{
			Exact: &v1.FieldMatch_Exact{
				Value: port,
			},
		},
	}
	actionPop := v1.TableAction{
		Type: &v1.TableAction_Action{Action: &v1.Action{
			ActionId: fabric.Action_EgressVlanPopVlan,
			Params:   nil,
		}},
	}
	return v1.TableEntry{
		TableId: fabric.Table_EgressVlan,
		Match:   []*v1.FieldMatch{&matchVlanId, &matchEgressPort},
		Action:  &actionPop,
	}
}

func createIngressPortVlanEntry(port []byte, internalVlan uint16, prio int32) v1.TableEntry {
	matchIngressPort := v1.FieldMatch{
		FieldId: fabric.FieldMatch_IngressPortVlanIgPort,
		FieldMatchType: &v1.FieldMatch_Exact_{
			Exact: &v1.FieldMatch_Exact{
				Value: port,
			},
		},
	}
	matchVlanIsValid := v1.FieldMatch{
		FieldId: fabric.FieldMatch_IngressPortVlanVlanIsValid,
		FieldMatchType: &v1.FieldMatch_Exact_{
			Exact: &v1.FieldMatch_Exact{
				Value: []byte{0},
			},
		},
	}
	actionPop := v1.TableAction{
		Type: &v1.TableAction_Action{Action: &v1.Action{
			ActionId: fabric.Action_IngressPortVlanPermitWithInternalVlan,
			Params: []*v1.Action_Param{
				{
					ParamId: 1,
					Value:   getVlanIdValue(internalVlan),
				},
			},
		}},
	}
	return v1.TableEntry{
		TableId:  fabric.Table_IngressPortVlan,
		Match:    []*v1.FieldMatch{&matchIngressPort, &matchVlanIsValid},
		Action:   &actionPop,
		Priority: prio,
	}
}

func createFwdClassifierEntry(port []byte, EthDst []byte, prio int32) v1.TableEntry {
	matchIngressPort := v1.FieldMatch{
		FieldId: fabric.FieldMatch_FwdClassifierIgPort,
		FieldMatchType: &v1.FieldMatch_Exact_{
			Exact: &v1.FieldMatch_Exact{
				Value: port,
			},
		},
	}
	matchEthDst := v1.FieldMatch{
		FieldId: fabric.FieldMatch_FwdClassifierEthDst,
		FieldMatchType: &v1.FieldMatch_Ternary_{
			Ternary: &v1.FieldMatch_Ternary{
				Value: EthDst,
				Mask:  []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
			},
		},
	}
	matchIpEthType := v1.FieldMatch{
		FieldId: fabric.FieldMatch_FwdClassifierIpEthType,
		FieldMatchType: &v1.FieldMatch_Exact_{
			Exact: &v1.FieldMatch_Exact{
				Value: getEthTypeValue(fabric.EthTypeIpv4),
			},
		},
	}
	actionPop := v1.TableAction{
		Type: &v1.TableAction_Action{Action: &v1.Action{
			ActionId: fabric.Action_FwdClassifierSetForwardingType,
			Params: []*v1.Action_Param{
				{
					ParamId: 1,
					Value:   []byte{fabric.FwdType_FwdIpv4Unicast}},
			},
		}},
	}
	return v1.TableEntry{
		TableId:  fabric.Table_FwdClassifier,
		Match:    []*v1.FieldMatch{&matchIngressPort, &matchEthDst, &matchIpEthType},
		Action:   &actionPop,
		Priority: prio,
	}
}
