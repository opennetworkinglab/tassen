package translate

import (
	"encoding/binary"
	v1 "github.com/p4lang/p4runtime/go/p4/v1"
	log "github.com/sirupsen/logrus"
	"mapr/p4info"
	"mapr/store"
)

// Implementation of ChangeProcessor interface for ONF's fabric.p4.
// TODO: implement

const (
	internalCoreTag	uint16 = 4094
)

type fabricChangeProcessor struct {
	targetStore store.P4RtStore
}

func NewFabricTranslator(srv store.P4RtStore, trg store.P4RtStore, tsn store.TassenStore) Translator {
	return &translator{
		serverStore: srv,
		tassenStore: tsn,
		processor: &fabricChangeProcessor{
			targetStore: trg,
		},
	}
}

func createUpdateEntry(entry v1.TableEntry, uType v1.Update_Type) *v1.Update {
	return &v1.Update{
		Type:   uType,
		Entity: &v1.Entity{Entity: &v1.Entity_TableEntry{TableEntry: &entry}},
	}
}

func (p fabricChangeProcessor) HandleIfTypeEntry(e *store.IfTypeEntry, uType v1.Update_Type) ([]*v1.Update, error) {
	log.Tracef("IfTypeEntry={ %s }", e)
	switch e.IfType[0] {
	case p4info.IfTypeCore:
		phyUpdateEntry := make([]*v1.Update, 0)
		phyUpdateEntry = append(phyUpdateEntry, createUpdateEntry(createIngressPortVlanEntry(e.Port, e.IfType), uType))
		phyUpdateEntry = append(phyUpdateEntry, createUpdateEntry(createEgressVlanPopEntry(e.Port, e.IfType), uType))
		return phyUpdateEntry, nil
	case p4info.IfTypeAccess:
		log.Warnf("fabricChangeProcessor.HandleIfTypeEntry(): not implemented for ACCESS ports")
	default:
		log.Warnf("IfTypeEntry.IfType=%v not implemented", e.IfType)
	}
	log.Warnf("fabricChangeProcessor.HandleIfTypeEntry(): not implemented")
	return nil, nil
}

func createEgressVlanPopEntry(port []byte, ifType []byte) v1.TableEntry {
	matchVlanId := v1.FieldMatch{
		FieldId: p4info.FieldMatch_FabricEgressNextVlanid,
		FieldMatchType: &v1.FieldMatch_Exact_{
			Exact: &v1.FieldMatch_Exact{
				Value: getInternalVlanCoreTag(),
			},
		},
	}
	matchEgressPort := v1.FieldMatch{
		FieldId: p4info.FieldMatch_FabricEgressNextEgport,
		FieldMatchType: &v1.FieldMatch_Exact_{
			Exact: &v1.FieldMatch_Exact{
				Value: port,
			},
		},
	}
	actionPop := v1.TableAction{
		Type: &v1.TableAction_Action{Action: &v1.Action{
			ActionId: p4info.Action_FabricEgressNextPopVlan,
			Params:   nil,
		}},
	}
	return v1.TableEntry{
		TableId:  p4info.Table_FabricEgressNextEgressVlan,
		Match:    []*v1.FieldMatch{&matchVlanId, &matchEgressPort},
		Action:   &actionPop,
		Priority: 1,
	}
}

func createIngressPortVlanEntry(port []byte, ifType []byte) v1.TableEntry {
	matchIngressPort := v1.FieldMatch{
		FieldId: p4info.FieldMatch_FabricIngressFilteringIngressPort,
		FieldMatchType: &v1.FieldMatch_Exact_{
			Exact: &v1.FieldMatch_Exact{
				Value: port,
			},
		},
	}
	matchVlanIsValid := v1.FieldMatch{
		FieldId: p4info.FieldMatch_FabricIngressFilteringVlanIsValid,
		FieldMatchType: &v1.FieldMatch_Exact_{
			Exact: &v1.FieldMatch_Exact{
				Value: []byte{0}, // TODO: what value is considered false???
			},
		},
	}
	matchVlanId := v1.FieldMatch{
		FieldId: p4info.FieldMatch_FabricIngressFilteringVlanVlanId,
		FieldMatchType: &v1.FieldMatch_Ternary_{
			Ternary: &v1.FieldMatch_Ternary{
				// TODO: If I'm not wrong if we don't want to match on a ternary field, we don't need to add the field in the matchfield
				Value:                []byte{0,0},
				Mask:                 []byte{0,0},
			},
		},
	}
	matchInnerVlanId := v1.FieldMatch{
		FieldId: p4info.FieldMatch_FabricIngressFilteringVlanInnerVlan,
		FieldMatchType: &v1.FieldMatch_Ternary_{
			Ternary: &v1.FieldMatch_Ternary{
				// TODO: If I'm not wrong if we don't want to match on a ternary field, we don't need to add the field in the matchfield
				Value:                []byte{0,0},
				Mask:                 []byte{0,0},
			},
		},
	}
	actionPop := v1.TableAction{
		Type: &v1.TableAction_Action{Action: &v1.Action{
			ActionId: p4info.Action_FabricIngressPermitWithInternalVlan,
			Params: []*v1.Action_Param{
				{
					ParamId: 1,
					Value:   getInternalVlanCoreTag()},
			},
		}},
	}
	return v1.TableEntry{
		TableId:  p4info.Table_FabricIngressFilteringIngressPortVlan,
		Match:    []*v1.FieldMatch{&matchIngressPort, &matchVlanIsValid, &matchVlanId, &matchInnerVlanId},
		Action:   &actionPop,
		Priority: 1,
	}
}

func getInternalVlanCoreTag() []byte {
	// TODO: we should make sure this VLAN (internalCoreTag) is not used by any other flows
	coreTag := make([]byte, 2)
	binary.BigEndian.PutUint16(coreTag, internalCoreTag)
	return coreTag
}

func (p fabricChangeProcessor) HandleMyStationEntry(e *store.MyStationEntry, uType v1.Update_Type) ([]*v1.Update, error) {
	log.Tracef("MyStationEntry={ %s }", e)
	log.Warnf("fabricChangeProcessor.HandleMyStationEntry(): not implemented")
	return nil, nil
}

func (p fabricChangeProcessor) HandleAttachmentEntry(a *store.AttachmentEntry, ok bool) ([]*v1.Update, error) {
	log.Tracef("AttachmentEntry={ %s }, complete=%v", a, ok)
	log.Warnf("fabricChangeProcessor.HandleMyStationEntry(): not implemented")
	return nil, nil
}
