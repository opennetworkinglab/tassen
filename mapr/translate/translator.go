package translate

import (
	"fmt"
	p4v1 "github.com/p4lang/p4runtime/go/p4/v1"
)

// A translator of P4RT updates.
type Translator interface {
	// Given an update for the logical pipeline, returns zero or more updates for the physical pipeline that will
	// produce a forwarding behavior that is equivalent to that of the logical pipeline after applying the given update.
	// If the returned updates are zero (nil), that means the translation was successful but it doesn't require any
	// changes to the physical pipeline (e.g., for many-to-one mapping, when we require multiple logical entries to
	// produce a physical one).
	Translate(logical *p4v1.Update) (physical []*p4v1.Update, err error)
}

// A processor of changes in the logical pipeline state.
type ChangeProcessor interface {
	// Returns P4RT updates to apply changes for the given IfTypeEntry
	HandleIfTypeEntry(e *IfTypeEntry, uType p4v1.Update_Type) ([]*p4v1.Update, error)
	// Returns P4RT updates to apply changes for the given MyStationEntry
	HandleMyStationEntry(e *MyStationEntry, uType p4v1.Update_Type) ([]*p4v1.Update, error)
	// Returns P4RT updates to apply changes for the given AttachmentEntry. Since the state of an attachment is
	// derived by multiple tables, the ok flag signals whether the state of the attachment is complete (e.g., all
	// fields are known), or not.
	HandleAttachmentEntry(a *AttachmentEntry, ok bool) ([]*p4v1.Update, error)
}

type translator struct {
	serverStore P4RtStore
	tassenStore TassenStore
	processor   ChangeProcessor
}

func NewTranslator(srv P4RtStore, tsn TassenStore, prc ChangeProcessor) Translator {
	return &translator{
		serverStore: srv,
		tassenStore: tsn,
		processor:   prc,
	}
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
func (t *translator) Translate(u *p4v1.Update) ([]*p4v1.Update, error) {
	result := make([]*p4v1.Update, 0)
	switch e := u.Entity.Entity.(type) {
	case *p4v1.Entity_TableEntry:
		switch e.TableEntry.TableId {
		case Table_IngressPipeIfTypes: // device-level
			entry, err := ParseIfTypeEntry(e.TableEntry)
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
		case Table_IngressPipeMyStations: // device-level
			entry, err := ParseMyStationEntry(e.TableEntry)
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
		case Table_IngressPipeUpstreamLines, Table_IngressPipeUpstreamAttachmentsV4: // attachment-level for upstream
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
			return nil, fmt.Errorf("table ID %v not implemented", e.TableEntry.TableId)
		}
	default:
		return nil, fmt.Errorf("translating %T not implemented", e)
	}
	return result, nil
}
