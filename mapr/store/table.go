package store

import (
	"fmt"
	"github.com/p4lang/p4runtime/go/p4/v1"
)

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

func (s store) PutTableEntry(entry *v1.TableEntry) {
	s.tableEntries[KeyFromTableEntry(entry)] = entry
}

func (s store) GetTableEntry(key *string) *v1.TableEntry {
	return s.tableEntries[*key]
}

func (s store) RemoveTableEntry(entry *v1.TableEntry) {
	delete(s.tableEntries, KeyFromTableEntry(entry))
}

func (s store) FilterTableEntries(f func(*v1.TableEntry) bool) []*v1.TableEntry {
	filtered := make([]*v1.TableEntry, 0)
	for _, value := range s.tableEntries {
		if f(value) {
			filtered = append(filtered, value)
		}
	}
	return filtered
}

func (s store) TableEntries() []*v1.TableEntry {
	return s.FilterTableEntries(func(*v1.TableEntry) bool {
		return true
	})
}

func (s store) TableEntryCount() int {
	return len(s.tableEntries)
}
