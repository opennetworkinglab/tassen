package translate

import (
	"github.com/p4lang/p4runtime/go/p4/v1"
	log "github.com/sirupsen/logrus"
	"mapr/p4info"
	"mapr/store"
)

// A translator of P4RT write requests.
type Translator interface {
	// Given a write request for the logical pipeline, returns a write request for the physical pipeline that will
	// produce a forwarding state that is equivalent to that of the logical pipeline after applying the given write
	// request. The returned request might be nil, signaling that the translation was successful but that should not
	// produce any change in the physical pipeline.
	Translate(logical *v1.WriteRequest) (physical *v1.WriteRequest, err error)
}

// A processor of changes in the logical pipeline state.
type ChangeProcessor interface {
	// Returns P4RT updates to apply changes for the given IfTypeEntry
	HandleIfTypeEntry(e *store.IfTypeEntry, uType v1.Update_Type) ([]*v1.Update, error)
	// Returns P4RT updates to apply changes for the given MyStationEntry
	HandleMyStationEntry(e *store.MyStationEntry, uType v1.Update_Type) ([]*v1.Update, error)
	// Returns P4RT updates to apply changes for the given AttachmentEntry. Since the state of an attachment is
	// derived by multiple tables, the ok flag signals whether the state of the attachment is complete (e.g., all
	// fields are known), or not.
	HandleAttachmentEntry(a *store.AttachmentEntry, ok bool) ([]*v1.Update, error)
}

type translator struct {
	serverStore store.P4RtStore
	tassenStore store.TassenStore
	processor   ChangeProcessor
}

// Implementation of translation logic that delegates changes to a ChangeProcessor.
//
// P4 tables and other objects can be categorized based on the type of information they hold, which can be device-level,
// or attachment-level. For updates to device-level objects, we parse them and call the corresponding method in the
// change processor. For attachment-level updates, the state of an attachment can be derived by multiple objects (mostly
// tables), as such, a translator might need to  "accumulate" attachment state before it can modify the physical
// pipeline. When receiving an update for an attachment-level table, we first evaluate a snapshot of the attachment that
// includes data from the store and from the given update. The snapshot is passed to the change processor with an
// indication of whether the attachment state is complete (all fields are populated) or not.
//
// Considering an example implementation for a change processor, if an attachment is complete, the processor might
// generate all necessary updates to insert the corresponding entries in the target to enable termination/forwarding. If
// the attachment is not complete, the processor might decide to remove all entries from the target related to that
// attachment.
func (t *translator) Translate(r *v1.WriteRequest) (*v1.WriteRequest, error) {
	result := make([]*v1.Update, 0)
	for _, u := range r.Updates {
		switch e := u.Entity.Entity.(type) {
		case *v1.Entity_TableEntry:
			switch e.TableEntry.TableId {
			case p4info.Table_IngressPipeIfTypes: // device-level
				entry, err := store.ParseIfTypeEntry(e.TableEntry)
				if err != nil {
					return nil, err
				}
				newUpdates, err := t.processor.HandleIfTypeEntry(&entry, u.Type)
				if err != nil {
					return nil, err
				}
				if newUpdates != nil {
					result = append(result, newUpdates...)
				}
			case p4info.Table_IngressPipeMyStations: // device-level
				entry, err := store.ParseMyStationEntry(e.TableEntry)
				if err != nil {
					return nil, err
				}
				newUpdates, err := t.processor.HandleMyStationEntry(&entry, u.Type)
				if err != nil {
					return nil, err
				}
				if newUpdates != nil {
					result = append(result, newUpdates...)
				}
			case p4info.Table_IngressPipeUpstreamLines, p4info.Table_IngressPipeUpstreamAttachmentsV4: // attachment-level for upstream
				a, ok, err := t.tassenStore.EvalAttachment(e.TableEntry)
				if err != nil {
					return nil, err
				}
				newUpdates, err := t.processor.HandleAttachmentEntry(&a, ok)
				if err != nil {
					return nil, err
				}
				if newUpdates != nil {
					result = append(result, newUpdates...)
				}
			// TODO: case Table_UpstreamRoutesV4 // device-level
			// TODO: case Table_UpstreamPppoePunts // device-level
			// TODO: downstream tables
			default:
				log.Warnf("Table ID %v not implemented, ignoring... [%s]",
					e.TableEntry.TableId, e.TableEntry.String())
			}
		default:
			log.Warnf("Translating %T not implemented, ignoring... [%s]",
				e, r.String())
		}
	}
	if len(result) == 0 {
		return nil, nil
	}
	// Return original write request but with modified updates.
	return &v1.WriteRequest{
		DeviceId:   r.DeviceId,
		RoleId:     r.RoleId,
		ElectionId: r.ElectionId,
		Updates:    result,
		Atomicity:  r.Atomicity,
	}, nil
}
