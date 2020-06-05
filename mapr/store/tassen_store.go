package store

import (
	"fmt"
	v1 "github.com/p4lang/p4runtime/go/p4/v1"
	log "github.com/sirupsen/logrus"
	"mapr/p4info"
)

type IfTypeEntry struct {
	Port   []byte
	IfType []byte
}

func (i IfTypeEntry) String() string {
	return fmt.Sprintf("Port: %x, IfType: %x", i.Port, i.IfType)
}

type MyStationEntry struct {
	Port   []byte
	EthDst []byte
}

func (m MyStationEntry) String() string {
	return fmt.Sprintf("Port: %x, EthDst: %x", m.Port, m.EthDst)
}

type Direction string

const (
	DirectionUpstream   Direction = "UP"
	DirectionDownstream Direction = "DOWN"
)

type AttachmentEntry struct {
	Direction   Direction
	Port        []byte
	LineId      []byte
	STag        []byte
	CTag        []byte
	MacAddr     []byte
	Ipv4Addr    []byte
	PppoeSessId []byte
}

func (a AttachmentEntry) String() string {
	return fmt.Sprintf("Dir: %s, Port: %s, LineId: %x, STag: %x, CTag: %x, MacAddr: %x, Ipv4Addr: %x, PppoeSessId: %x",
		a.Direction, a.Port, a.LineId, a.STag, a.CTag, a.MacAddr, a.Ipv4Addr, a.PppoeSessId)
}

type PortKey [2]byte
type LineIdKey [4]byte

func toPortKey(b []byte) PortKey {
	return PortKey{b[0], b[1]}
}

func toLineIdKey(b []byte) LineIdKey {
	return LineIdKey{b[0], b[1], b[2], b[3]}
}

type TassenStore interface {
	// Updates the store using the content of the given P4Runtime write request
	Update(r *v1.WriteRequest, dryRun bool) error
	// Returns IfTypeEntry for the given port
	GetIfType(PortKey) *IfTypeEntry
	// Returns MyStationEntry for the given port
	GetMyStation(PortKey) *MyStationEntry
	// Returns upstream AttachmentEntry for the given line ID
	GetUpAttachment(LineIdKey) *AttachmentEntry
	// Returns downstream AttachmentEntry for the given line ID
	GetDownAttachment(LineIdKey) *AttachmentEntry
	// Returns an attachment entry with populated with data from the given table entry.
	EvalAttachment(*v1.TableEntry) (a AttachmentEntry, ok bool, err error)
}

type tassenStore struct {
	ifTypes     map[PortKey]*IfTypeEntry
	myStations  map[PortKey]*MyStationEntry
	upAttachs   map[LineIdKey]*AttachmentEntry
	downAttachs map[LineIdKey]*AttachmentEntry
}

func (s tassenStore) Update(r *v1.WriteRequest, dryRun bool) (err error) {
	for _, u := range r.Updates {
		switch e := u.Entity.Entity.(type) {
		case *v1.Entity_TableEntry:
			switch e.TableEntry.TableId {
			case p4info.Table_IngressPipeIfTypes:
				entry, err := ParseIfTypeEntry(e.TableEntry)
				if err != nil {
					return err
				}
				// TODO: check if key exists
				if !dryRun {
					s.ifTypes[toPortKey(entry.Port)] = &entry
				}
			case p4info.Table_IngressPipeMyStations:
				entry, err := ParseMyStationEntry(e.TableEntry)
				if err != nil {
					return err
				}
				// TODO: check if key exists
				if !dryRun {
					s.myStations[toPortKey(entry.EthDst)] = &entry
				}
			case p4info.Table_IngressPipeUpstreamLines, p4info.Table_IngressPipeUpstreamAttachmentsV4: // TODO: downstream tables
				attach, _, err := s.EvalAttachment(e.TableEntry)
				if err != nil {
					return err
				}
				if !dryRun {
					if attach.Direction == DirectionUpstream {
						s.upAttachs[toLineIdKey(attach.LineId)] = &attach
					} else {
						s.downAttachs[toLineIdKey(attach.LineId)] = &attach
					}
				}
			// TODO: case Table_UpstreamRoutesV4 // device-level
			// TODO: case Table_UpstreamPppoePunts // device-level
			default:
				log.Warnf("Update(): Table ID %v not implemented, ignoring... [%s]",
					e.TableEntry.TableId, e.TableEntry.String())
			}
		default:
			log.Warnf("Update(): Updating %T not implemented, ignoring... [%s]",
				e, r.String())
		}
	}
	return nil
}

func (s tassenStore) GetIfType(key PortKey) *IfTypeEntry {
	return s.ifTypes[key]
}

func (s tassenStore) GetMyStation(key PortKey) *MyStationEntry {
	return s.myStations[key]
}

func (s tassenStore) GetUpAttachment(key LineIdKey) *AttachmentEntry {
	return s.upAttachs[key]
}

func (s tassenStore) GetDownAttachment(key LineIdKey) *AttachmentEntry {
	return s.downAttachs[key]
}

func (s tassenStore) EvalAttachment(t *v1.TableEntry) (a AttachmentEntry, ok bool, err error) {
	switch t.TableId {
	case p4info.Table_IngressPipeUpstreamLines:
		err = ParseUpstreamLineEntry(t, &a)
	case p4info.Table_IngressPipeUpstreamAttachmentsV4:
		err = ParseUpstreamAttachmentV4Entry(t, &a)
	default:
		log.Warnf("EvalAttachment(): table ID %d not implemented", t.TableId)
	}
	if err != nil {
		return
	}
	if a.LineId == nil {
		panic("missing line ID in parsed table entry")
	}
	var storedAttach *AttachmentEntry
	if a.Direction == DirectionUpstream {
		storedAttach = s.GetUpAttachment(toLineIdKey(a.LineId))
	} else if a.Direction == DirectionDownstream {
		storedAttach = s.GetDownAttachment(toLineIdKey(a.LineId))
	} else {
		panic("direction unknown")
	}
	if storedAttach == nil {
		return
	}
	// Is there a less verbose way of evaluating the attachment?
	if a.Port == nil {
		a.Port = storedAttach.Port
	}
	if a.STag == nil {
		a.STag = storedAttach.STag
	}
	if a.CTag == nil {
		a.CTag = storedAttach.CTag
	}
	if a.MacAddr == nil {
		a.MacAddr = storedAttach.MacAddr
	}
	if a.Ipv4Addr == nil {
		a.Ipv4Addr = storedAttach.Ipv4Addr
	}
	if a.PppoeSessId == nil {
		a.PppoeSessId = storedAttach.PppoeSessId
	}
	ok = a.Port != nil && a.STag != nil && a.CTag != nil && a.MacAddr != nil && a.Ipv4Addr != nil && a.PppoeSessId != nil
	return
}

func NewTassenStore() *tassenStore {
	return &tassenStore{
		ifTypes:    make(map[PortKey]*IfTypeEntry),
		myStations: make(map[PortKey]*MyStationEntry),
		upAttachs:  make(map[LineIdKey]*AttachmentEntry),
	}
}

func ParseIfTypeEntry(t *v1.TableEntry) (IfTypeEntry, error) {
	entry := IfTypeEntry{}
	// Parse match
	for _, m := range t.Match {
		switch m.FieldId {
		case p4info.Hdr_IngressPipeIfTypes_Port:
			entry.Port = m.GetExact().Value
		default:
			return IfTypeEntry{}, fmt.Errorf("invalid %T ID %d", m, m.FieldId)
		}
	}
	// Parse action
	act := t.GetAction().GetAction()
	if act == nil || act.ActionId != p4info.Action_IngressPipeSetIfType {
		return IfTypeEntry{}, fmt.Errorf("invalid Action %s", t.GetAction().String())
	}
	for _, p := range act.Params {
		switch p.ParamId {
		case p4info.ActionParam_IngressPipeSetIfType_IfType:
			entry.IfType = p.Value
		default:
			return IfTypeEntry{}, fmt.Errorf("invalid %T ID %d", p, p.ParamId)
		}
	}
	return entry, nil
}

func ParseMyStationEntry(t *v1.TableEntry) (MyStationEntry, error) {
	entry := MyStationEntry{}
	// Parse match
	for _, m := range t.Match {
		switch m.FieldId {
		case p4info.Hdr_IngressPipeMyStations_Port:
			entry.Port = m.GetExact().Value
		case p4info.Hdr_IngressPipeMyStations_EthDst:
			entry.EthDst = m.GetExact().Value
		default:
			return MyStationEntry{}, fmt.Errorf("invalid %T ID %d", m, m.FieldId)
		}
	}
	// Parse action
	act := t.GetAction().GetAction()
	if act == nil || act.ActionId != p4info.Action_IngressPipeSetMyStation {
		return MyStationEntry{}, fmt.Errorf("invalid Action %s", t.GetAction().String())
	}
	return entry, nil
}

func ParseUpstreamLineEntry(t *v1.TableEntry, a *AttachmentEntry) error {
	// Parse match
	a.Direction = DirectionUpstream
	for _, m := range t.Match {
		switch m.FieldId {
		case p4info.Hdr_IngressPipeUpstreamLines_CTag:
			a.CTag = m.GetExact().Value
		case p4info.Hdr_IngressPipeUpstreamLines_STag:
			a.STag = m.GetExact().Value
		case p4info.Hdr_IngressPipeUpstreamLines_Port:
			a.Port = m.GetExact().Value
		default:
			return fmt.Errorf("invalid %T ID %d", m, m.FieldId)
		}
	}
	// Parse action
	act := t.GetAction().GetAction()
	if act == nil || act.ActionId != p4info.Action_IngressPipeUpstreamSetLine {
		return fmt.Errorf("invalid Action %s", t.GetAction().String())
	}
	for _, p := range act.Params {
		switch p.ParamId {
		case p4info.ActionParam_IngressPipeUpstreamSetLine_LineId:
			a.LineId = p.Value
		default:
			return fmt.Errorf("invalid %T ID %d", p, p.ParamId)
		}
	}
	return nil
}

func ParseUpstreamAttachmentV4Entry(t *v1.TableEntry, a *AttachmentEntry) error {
	// Parse match
	a.Direction = DirectionUpstream
	for _, m := range t.Match {
		switch m.FieldId {
		case p4info.Hdr_IngressPipeUpstreamAttachmentsV4_LineId:
			a.LineId = m.GetExact().Value
		case p4info.Hdr_IngressPipeUpstreamAttachmentsV4_EthSrc:
			a.MacAddr = m.GetExact().Value
		case p4info.Hdr_IngressPipeUpstreamAttachmentsV4_Ipv4Src:
			a.Ipv4Addr = m.GetExact().Value
		case p4info.Hdr_IngressPipeUpstreamAttachmentsV4_PppoeSessId:
			a.PppoeSessId = m.GetExact().Value
		default:
			return fmt.Errorf("invalid %T ID %d", m, m.FieldId)
		}
	}
	// Parse action
	act := t.GetAction().GetAction()
	if act == nil || act.ActionId != p4info.Action_Nop {
		return fmt.Errorf("invalid Action %s", t.GetAction().String())
	}
	return nil
}
