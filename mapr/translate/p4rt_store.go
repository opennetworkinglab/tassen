/*
 * Copyright 2020-present Open Networking Foundation
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package translate

import (
	"fmt"
	p4v1 "github.com/p4lang/p4runtime/go/p4/v1"
	log "github.com/sirupsen/logrus"
)

// A store of P4Runtime entities with map semantics.
type P4RtStore interface {
	// Updates the store using the content of the given P4Runtime WriteRequest's Update.
	ApplyUpdate(r *p4v1.Update, dryRun bool) error
	// Stores the given table entry.
	PutTableEntry(*p4v1.TableEntry)
	// Returns the table entry associated with the given key, or nil.
	GetTableEntry(*string) *p4v1.TableEntry
	// Removes the given table entry.
	RemoveTableEntry(*p4v1.TableEntry)
	// Returns a slice of table entries that satisfy the predicate f.
	FilterTableEntries(f func(*p4v1.TableEntry) bool) []*p4v1.TableEntry
	// Returns all table entries.
	TableEntries() []*p4v1.TableEntry
	// Returns the number of table entries in the store.
	TableEntryCount() int
	// Stores the given action profile group.
	PutActProfGroup(*p4v1.ActionProfileGroup)
	// Returns the action profile group associated with the given key, or nil.
	GetActProfGroup(*string) *p4v1.ActionProfileGroup
	// Removes the given action profile group.
	RemoveActProfGroup(*p4v1.ActionProfileGroup)
	// Returns a slice of action profile groups that satisfy the predicate f.
	FilterActProfGroups(f func(*p4v1.ActionProfileGroup) bool) []*p4v1.ActionProfileGroup
	// Returns all action profile groups.
	ActProfGroups() []*p4v1.ActionProfileGroup
	// Returns the number of action profile groups in the store.
	ActProfGroupCount() int
	// Stores the given action profile member.
	PutActProfMember(*p4v1.ActionProfileMember)
	// Returns the action profile member associated with the given key, or nil.
	GetActProfMember(*string) *p4v1.ActionProfileMember
	// Removes the given action profile member.
	RemoveActProfMember(*p4v1.ActionProfileMember)
	// Returns a slice of action profile members that satisfy the predicate f.
	FilterActProfMembers(f func(*p4v1.ActionProfileMember) bool) []*p4v1.ActionProfileMember
	// Returns all action profile members.
	ActProfMembers() []*p4v1.ActionProfileMember
	// Returns the number of action profile members in the store.
	ActProfMemberCount() int
}

type p4RtStore struct {
	name           string
	tableEntries   map[string]*p4v1.TableEntry
	actProfGroups  map[string]*p4v1.ActionProfileGroup
	actProfMembers map[string]*p4v1.ActionProfileMember
}

func NewP4RtStore(name string) *p4RtStore {
	return &p4RtStore{
		name:           name,
		tableEntries:   make(map[string]*p4v1.TableEntry),
		actProfGroups:  make(map[string]*p4v1.ActionProfileGroup),
		actProfMembers: make(map[string]*p4v1.ActionProfileMember),
	}
}

func (s *p4RtStore) ApplyUpdate(u *p4v1.Update, dryRun bool) error {
	if dryRun {
		// TODO: implement validation logic
		return nil
	}
	defer s.logStoreSummary()
	switch x := u.Entity.Entity.(type) {
	case *p4v1.Entity_TableEntry:
		if u.Type == p4v1.Update_DELETE {
			s.RemoveTableEntry(x.TableEntry)
		} else {
			s.PutTableEntry(x.TableEntry)
		}
	case *p4v1.Entity_ActionProfileGroup:
		if u.Type == p4v1.Update_DELETE {
			s.RemoveActProfGroup(x.ActionProfileGroup)
		} else {
			s.PutActProfGroup(x.ActionProfileGroup)
		}
	case *p4v1.Entity_ActionProfileMember:
		if u.Type == p4v1.Update_DELETE {
			s.RemoveActProfMember(x.ActionProfileMember)
		} else {
			s.PutActProfMember(x.ActionProfileMember)
		}
	default:
		log.Warnf("P4RtStore(%s): storing %T not implemented, ignoring... [%v]", s.name, x, x)
	}
	return nil
}

func (s *p4RtStore) logStoreSummary() {
	log.Debugf("P4RtStore(%s) summary: TableEntryCount=%d, ActProfGroupCount=%d, ActProfMemberCount=%d",
		s.name, s.TableEntryCount(), s.ActProfGroupCount(), s.ActProfMemberCount())
}

// Returns a string that uniquely identifies a table entry.
func TableEntryKey(tableId uint32, match []*p4v1.FieldMatch, priority int32) string {
	// Fields that determine uniqueness are defined by the P4RT spec.
	// We return a string as that's a comparable and can be used as a map key. Is there a more efficient way of getting
	// a comparable key out of a protobuf TableEntry message?
	return fmt.Sprintf("%v-%v-%v", tableId, match, priority)
}

// Returns a string that uniquely identifies the given table entry.
func KeyFromTableEntry(t *p4v1.TableEntry) string {
	return TableEntryKey(t.TableId, t.Match, t.Priority)
}

func (s *p4RtStore) PutTableEntry(entry *p4v1.TableEntry) {
	s.tableEntries[KeyFromTableEntry(entry)] = entry
}

func (s *p4RtStore) GetTableEntry(key *string) *p4v1.TableEntry {
	return s.tableEntries[*key]
}

func (s *p4RtStore) RemoveTableEntry(entry *p4v1.TableEntry) {
	delete(s.tableEntries, KeyFromTableEntry(entry))
}

func (s *p4RtStore) FilterTableEntries(f func(*p4v1.TableEntry) bool) []*p4v1.TableEntry {
	filtered := make([]*p4v1.TableEntry, 0)
	for _, value := range s.tableEntries {
		if f(value) {
			filtered = append(filtered, value)
		}
	}
	return filtered
}

func (s *p4RtStore) TableEntries() []*p4v1.TableEntry {
	return s.FilterTableEntries(func(*p4v1.TableEntry) bool {
		return true
	})
}

func (s *p4RtStore) TableEntryCount() int {
	return len(s.tableEntries)
}

func ActProfGroupKey(actProfId uint32, groupIp uint32) string {
	return fmt.Sprintf("%d-%d", actProfId, groupIp)
}

// Returns a string that uniquely identifies the given table entry.
func KeyFromActProfGroup(g *p4v1.ActionProfileGroup) string {
	return ActProfGroupKey(g.ActionProfileId, g.GroupId)
}

func (s *p4RtStore) PutActProfGroup(g *p4v1.ActionProfileGroup) {
	s.actProfGroups[KeyFromActProfGroup(g)] = g
}

func (s *p4RtStore) GetActProfGroup(key *string) *p4v1.ActionProfileGroup {
	return s.actProfGroups[*key]
}

func (s *p4RtStore) RemoveActProfGroup(g *p4v1.ActionProfileGroup) {
	delete(s.actProfGroups, KeyFromActProfGroup(g))
}

func (s *p4RtStore) FilterActProfGroups(f func(*p4v1.ActionProfileGroup) bool) []*p4v1.ActionProfileGroup {
	filtered := make([]*p4v1.ActionProfileGroup, 0)
	for _, value := range s.actProfGroups {
		if f(value) {
			filtered = append(filtered, value)
		}
	}
	return filtered
}

func (s *p4RtStore) ActProfGroups() []*p4v1.ActionProfileGroup {
	return s.FilterActProfGroups(func(*p4v1.ActionProfileGroup) bool {
		return true
	})
}

func (s *p4RtStore) ActProfGroupCount() int {
	return len(s.actProfGroups)
}

func ActProfMemberKey(actProfId uint32, memberId uint32) string {
	return fmt.Sprintf("%d-%d", actProfId, memberId)
}

// Returns a string that uniquely identifies the given table entry.
func KeyFromActProfMember(g *p4v1.ActionProfileMember) string {
	return ActProfMemberKey(g.ActionProfileId, g.MemberId)
}

func (s *p4RtStore) PutActProfMember(g *p4v1.ActionProfileMember) {
	s.actProfMembers[KeyFromActProfMember(g)] = g
}

func (s *p4RtStore) GetActProfMember(key *string) *p4v1.ActionProfileMember {
	return s.actProfMembers[*key]
}

func (s *p4RtStore) RemoveActProfMember(g *p4v1.ActionProfileMember) {
	delete(s.actProfMembers, KeyFromActProfMember(g))
}

func (s *p4RtStore) FilterActProfMembers(f func(*p4v1.ActionProfileMember) bool) []*p4v1.ActionProfileMember {
	filtered := make([]*p4v1.ActionProfileMember, 0)
	for _, value := range s.actProfMembers {
		if f(value) {
			filtered = append(filtered, value)
		}
	}
	return filtered
}

func (s *p4RtStore) ActProfMembers() []*p4v1.ActionProfileMember {
	return s.FilterActProfMembers(func(*p4v1.ActionProfileMember) bool {
		return true
	})
}

func (s *p4RtStore) ActProfMemberCount() int {
	return len(s.actProfMembers)
}
