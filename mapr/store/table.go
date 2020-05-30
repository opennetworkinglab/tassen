package store

import (
	"fmt"
	"github.com/p4lang/p4runtime/go/p4/v1"
)

// Returns a string that uniquely identifies a table entry.
func tableEntryKey(k *v1.TableEntry) string {
	// Fields that determine uniqueness are defined by the P4RT spec.
	// We return a string as that's a comparable and can be used as a map key. Is there a more efficient way of getting
	// a comparable key out of a protobuf TableEntry message?
	return fmt.Sprintf("%d-%s-%d", k.TableId, k.Match, k.Priority)
}

func (s store) PutTableEntry(entry *v1.TableEntry) {
	s.tableEntries[tableEntryKey(entry)] = entry
}

func (s store) RemoveTableEntry(entry *v1.TableEntry) {
	delete(s.tableEntries, tableEntryKey(entry))
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
