package store

import (
	"github.com/p4lang/p4runtime/go/p4/v1"
	"log"
)

// A store of P4Runtime entities with map semantics
type Store interface {
	// Stores all entities in the given write request.
	PutAll(request *v1.WriteRequest)
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

func NewStore() *store {
	return &store{
		tableEntries: make(map[string]*v1.TableEntry),
	}
}

type store struct {
	tableEntries map[string]*v1.TableEntry
}

func (s *store) PutAll(req *v1.WriteRequest) {
	for _, u := range req.Updates {
		switch x := u.Entity.Entity.(type) {
		case *v1.Entity_TableEntry:
			if u.Type == v1.Update_DELETE {
				s.RemoveTableEntry(x.TableEntry)
			} else {
				s.PutTableEntry(x.TableEntry)
			}
		default:
			log.Printf("WARN! Storing %T not implemented, ignoring... [%s]", x, req.String())
		}
	}
}
