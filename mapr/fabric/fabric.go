package fabric

import (
	v1 "github.com/p4lang/p4runtime/go/p4/v1"
	log "github.com/sirupsen/logrus"
	"mapr/translate"
)

// Implementation of ChangeProcessor interface for ONF's fabric.p4.
// TODO: implement
const (
	defaultInternalTag     uint16 = 4094
	defaultPrio            int32  = 1
	FwdType_FwdIpv4Unicast byte   = 0x02
	EthTypeIpv4            uint16 = 0x0800
)

type fabricChangeProcessor struct {
	targetStore translate.P4RtStore
}

func NewFabricTranslator(srv translate.P4RtStore, trg translate.P4RtStore, tsn translate.TassenStore) translate.Translator {
	prc := &fabricChangeProcessor{
		targetStore: trg,
	}
	return translate.NewTranslator(srv, tsn, prc)
}

func (p fabricChangeProcessor) HandleIfTypeEntry(e *translate.IfTypeEntry, uType v1.Update_Type) ([]*v1.Update, error) {
	log.Tracef("IfTypeEntry={ %s }", e)
	// TODO: check parameter of IfTypeEntry and return error
	phyTableEntries := make([]v1.TableEntry, 0)
	switch e.IfType[0] {
	case translate.IfTypeCore:
		phyTableEntries = append(phyTableEntries,
			createIngressPortVlanEntryPermit(e.Port, nil, nil, getVlanIdValue(defaultInternalTag), defaultPrio))
		phyTableEntries = append(phyTableEntries, createEgressVlanPopEntry(e.Port, defaultInternalTag))
	case translate.IfTypeAccess:
		log.Warnf("fabricChangeProcessor.HandleIfTypeEntry(): not implemented for ACCESS ports")
	default:
		log.Warnf("IfTypeEntry.IfType=%v not implemented", e.IfType)
	}

	phyUpdateEntry := make([]*v1.Update, 0)
	for i := range phyTableEntries {
		phyUpdateEntry = append(phyUpdateEntry, createUpdateEntry(&phyTableEntries[i], uType))
	}
	return phyUpdateEntry, nil
}

func (p fabricChangeProcessor) HandleMyStationEntry(e *translate.MyStationEntry, uType v1.Update_Type) ([]*v1.Update, error) {
	log.Tracef("MyStationEntry={ %s }", e)
	// TODO: check parameter of mystation entry and return error
	phyTableEntry := createFwdClassifierEntry(e.Port, e.EthDst, defaultPrio)
	return []*v1.Update{createUpdateEntry(&phyTableEntry, uType)}, nil
}

func (p fabricChangeProcessor) HandleAttachmentEntry(a *translate.AttachmentEntry, ok bool) ([]*v1.Update, error) {
	log.Tracef("AttachmentEntry={ %s }, complete=%v", a, ok)
	log.Warnf("fabricChangeProcessor.HandleMyStationEntry(): not implemented")
	return nil, nil
}
