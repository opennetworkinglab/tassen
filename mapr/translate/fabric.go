package translate

import (
	v1 "github.com/p4lang/p4runtime/go/p4/v1"
	log "github.com/sirupsen/logrus"
	"mapr/store"
)

// Implementation of ChangeProcessor interface for ONF's fabric.p4.
// TODO: implement
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

func (p fabricChangeProcessor) HandleIfTypeEntry(e *store.IfTypeEntry, uType v1.Update_Type) ([]*v1.Update, error) {
	log.Tracef("IfTypeEntry={ %s }", e)
	log.Warnf("fabricChangeProcessor.HandleIfTypeEntry(): not implemented")
	return nil, nil
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
