package store

import (
	"fmt"
	"github.com/p4lang/p4runtime/go/p4/v1"
	log "github.com/sirupsen/logrus"
)

// A store of P4Runtime entities with map semantics.
type P4RtStore interface {
	// Updates the store using the content of the given P4Runtime write request.
	Update(r *v1.WriteRequest, dryRun bool) error
	// Stores the given table entry.
	PutTableEntry(*v1.TableEntry)
	// Returns the table entry associated with the given key, or nil.
	GetTableEntry(*string) *v1.TableEntry
	// Removes the given table entry.
	RemoveTableEntry(*v1.TableEntry)
	// Returns a slice of table entries that satisfy the predicate f.
	FilterTableEntries(f func(*v1.TableEntry) bool) []*v1.TableEntry
	// Returns all table entries.
	TableEntries() []*v1.TableEntry
	// Returns the number of table entries in the store.
	TableEntryCount() int
}

type p4RtStore struct {
	tableEntries map[string]*v1.TableEntry
}

func NewP4RtStore() *p4RtStore {
	return &p4RtStore{
		tableEntries: make(map[string]*v1.TableEntry),
	}
}

func (s *p4RtStore) Update(req *v1.WriteRequest, dryRun bool) error {
	// TODO: implement dry run for validation
	for _, u := range req.Updates {
		switch x := u.Entity.Entity.(type) {
		case *v1.Entity_TableEntry:
			if u.Type == v1.Update_DELETE {
				s.RemoveTableEntry(x.TableEntry)
			} else {
				s.PutTableEntry(x.TableEntry)
			}
		default:
			log.WithField("WriteRequest", req).Warnf("Storing %T not implemented, ignoring...", x)
		}
	}
	return nil
}

// Returns a string that uniquely identifies a table entry.
func TableEntryKey(tableId uint32, match []*v1.FieldMatch, priority int32) string {
	// Fields that determine uniqueness are defined by the P4RT spec.
	// We return a string as that's a comparable and can be used as a map key. Is there a more efficient way of getting
	// a comparable key out of a protobuf TableEntry message?
	return fmt.Sprintf("%v-%v-%v", tableId, match, priority)
}

// Returns a string that uniquely identifies the given table entry.
func KeyFromTableEntry(t *v1.TableEntry) string {
	return TableEntryKey(t.TableId, t.Match, t.Priority)
}

func (s *p4RtStore) PutTableEntry(entry *v1.TableEntry) {
	s.tableEntries[KeyFromTableEntry(entry)] = entry
}

func (s *p4RtStore) GetTableEntry(key *string) *v1.TableEntry {
	return s.tableEntries[*key]
}

func (s *p4RtStore) RemoveTableEntry(entry *v1.TableEntry) {
	delete(s.tableEntries, KeyFromTableEntry(entry))
}

func (s *p4RtStore) FilterTableEntries(f func(*v1.TableEntry) bool) []*v1.TableEntry {
	filtered := make([]*v1.TableEntry, 0)
	for _, value := range s.tableEntries {
		if f(value) {
			filtered = append(filtered, value)
		}
	}
	return filtered
}

func (s *p4RtStore) TableEntries() []*v1.TableEntry {
	return s.FilterTableEntries(func(*v1.TableEntry) bool {
		return true
	})
}

func (s *p4RtStore) TableEntryCount() int {
	return len(s.tableEntries)
}
