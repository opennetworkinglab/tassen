package store

import (
	"github.com/p4lang/p4runtime/go/p4/v1"
)

// A store of P4Runtime entities with map semantics
type P4RtStore interface {
	// Stores the given table entry.
	PutTableEntry(*v1.TableEntry)
	// Returns a slice of table entries that satisfy the predicate f.
	FilterTableEntries(f func(*v1.TableEntry) bool) []*v1.TableEntry
	// Returns all table entries.
	TableEntries() []*v1.TableEntry
	// Returns the number of table entries in the store.
	TableEntryCount() int
}

type store struct {
	tableEntries map[string]*v1.TableEntry
}
