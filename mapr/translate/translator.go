package translate

import (
	"fmt"
	p4v1 "github.com/p4lang/p4runtime/go/p4/v1"
	log "github.com/sirupsen/logrus"
)

// Produces updates for the target pipeline state by handling changes to the logical one.
//
// The operations of a Translator are supported by a Processor, which represent the target-specific logic.
//
// The main duties of a Translator are:
// - Maintain state (context), for both the logical pipeline (high-level objects) and target one (low-level P4RT
//   entities);
// - Parse P4RT WriteRequest's Updates and detect changes to the logical state;
// - When detecting a change, invoke the Processor to generate the corresponding updates for the target pipeline.
type Translator interface {
	// Given a P4RT Update for the logical pipeline, Translate() returns zero or more updates for the target pipeline.
	// If the returned updates are zero (nil), it means the translation was successful but it doesn't require any
	// changes to the target (e.g., for many-to-one mapping, when we require multiple logical entries to produce a
	// physical one). Calling Translate() does NOT alter the pipeline context.
	Translate(logical *p4v1.Update) (target []*p4v1.Update, err error)
	// Modifies the pipeline context by applying the given logical and target updates.
	ApplyUpdate(logical *p4v1.Update, target []*p4v1.Update) error
}

// A processor of changes in the logical pipeline state. Provides methods that generate updates for the target.
type Processor interface {
	// TODO (carmelo): methods should have proper store event semantics, i.e. pass old value and new one, and let
	//  implementation act on the diff.

	// Returns P4RT updates to apply changes for the given IfTypeEntry
	HandleIfTypeEntry(e *IfTypeEntry, uType p4v1.Update_Type) ([]*p4v1.Update, error)
	// Returns P4RT updates to apply changes for the given MyStationEntry
	HandleMyStationEntry(e *MyStationEntry, uType p4v1.Update_Type) ([]*p4v1.Update, error)
	// Returns P4RT updates to apply changes for the given snapshot of an AttachmentEntry. Since the state of an
	// attachment might depend on multiple tables, the ok flag signals whether the given snapshot of is complete (e.g.,
	// all fields are known), or not.
	HandleAttachmentEntry(e *AttachmentEntry, ok bool) ([]*p4v1.Update, error)
	// TODO docs
	HandleRouteV4NextHopEntry(e *NextHopEntry, uType p4v1.Update_Type) ([]*p4v1.Update, error)
	HandleRouteV4NextHopGroup(e *NextHopGroup, uType p4v1.Update_Type) ([]*p4v1.Update, error)
	HandleRouteV4Entry(e *RouteV4Entry, uType p4v1.Update_Type) ([]*p4v1.Update, error)
}

// Translator context.
type Context interface {
	// The store of logical pipeline objects. Should be treated as read-only.
	// Updates to this store are performed by the Write RPC handler in main.go.
	Logical() *LogicalStore
	// A mirror of the target device's state (P4Runtime). Should be treated as read-only.
	// Updates to this store are performed by the Write RPC handler in main.go.
	Target() P4RtStore
}

type context struct {
	logical LogicalStore
	target  P4RtStore
}

func (p context) Logical() *LogicalStore {
	return &p.logical
}

func (p context) Target() P4RtStore {
	return p.target
}

// A collection of maps holding the logical state.
type LogicalStore struct {
	IfTypes                map[PortKey]*IfTypeEntry
	MyStations             map[PortKey]*MyStationEntry
	UpstreamAttachments    map[LineIdKey]*AttachmentEntry
	DownstreamAttachments  map[LineIdKey]*AttachmentEntry
	UpstreamRoutesV4       map[Ipv4LpmKey]*RouteV4Entry
	UpstreamNextHopGroups  map[uint32]*NextHopGroup
	UpstreamNextHopEntries map[uint32]*NextHopEntry
}

// Creates a new Context.
func NewContext() Context {
	return &context{
		logical: LogicalStore{
			IfTypes:                make(map[PortKey]*IfTypeEntry),
			MyStations:             make(map[PortKey]*MyStationEntry),
			UpstreamAttachments:    make(map[LineIdKey]*AttachmentEntry),
			DownstreamAttachments:  make(map[LineIdKey]*AttachmentEntry),
			UpstreamRoutesV4:       make(map[Ipv4LpmKey]*RouteV4Entry),
			UpstreamNextHopGroups:  make(map[uint32]*NextHopGroup),
			UpstreamNextHopEntries: make(map[uint32]*NextHopEntry),
		},
		target: NewP4RtStore("target"),
	}
}

type translator struct {
	proc Processor
	ctx  Context
}

// Creates a new Translator
func NewTranslator(proc Processor, ctx Context) Translator {
	return &translator{
		proc: proc,
		ctx:  ctx,
	}
}

func (t *translator) logLogicalSummary() {
	log.Debugf("Context summary: ifTypes=%d, myStations=%d, upAttachs=%d, downAttachs=%d, "+
		"upRoutesV4=%d, upNextHopGroups=%d, upNextHopEntries=%d",
		len(t.ctx.Logical().IfTypes), len(t.ctx.Logical().MyStations), len(t.ctx.Logical().UpstreamAttachments),
		len(t.ctx.Logical().DownstreamAttachments), len(t.ctx.Logical().UpstreamRoutesV4),
		len(t.ctx.Logical().UpstreamNextHopGroups), len(t.ctx.Logical().UpstreamNextHopEntries))
}

func (t *translator) Translate(u *p4v1.Update) ([]*p4v1.Update, error) {
	target, err := t.translateOrStore(u, true)
	if err != nil {
		return nil, err
	}
	// Validate updates to target pipeline by performing a dry run.
	for _, u := range target {
		if err := t.ctx.Target().ApplyUpdate(u, true); err != nil {
			return nil, err
		}
	}
	return target, nil
}

func (t translator) ApplyUpdate(logical *p4v1.Update, target []*p4v1.Update) error {
	defer t.logLogicalSummary()
	for _, u := range target {
		if err := t.ctx.Target().ApplyUpdate(u, false); err != nil {
			panic("ApplyUpdate(): error when applying update to target store (BUG?)")
		}
	}
	_, err := t.translateOrStore(logical, false)
	return err
}

func (t translator) translateOrStore(u *p4v1.Update, translate bool) ([]*p4v1.Update, error) {
	switch e := u.Entity.Entity.(type) {
	case *p4v1.Entity_TableEntry:
		switch e.TableEntry.TableId {
		case Table_IngressPipeIfTypes:
			x, err := parseIfTypeEntry(e.TableEntry)
			if err != nil {
				return nil, err
			}
			if translate {
				// TODO: implement validation
				return t.proc.HandleIfTypeEntry(&x, u.Type)
			} else {
				key := ToPortKey(x.Port)
				if u.Type == p4v1.Update_DELETE {
					delete(t.ctx.Logical().IfTypes, key)
				} else {
					t.ctx.Logical().IfTypes[ToPortKey(x.Port)] = &x
				}
				return nil, nil
			}
		case Table_IngressPipeMyStations:
			x, err := parseMyStationEntry(e.TableEntry)
			if err != nil {
				return nil, err
			}
			if translate {
				// TODO: implement validation
				return t.proc.HandleMyStationEntry(&x, u.Type)
			} else {
				key := ToPortKey(x.Port)
				if u.Type == p4v1.Update_DELETE {
					delete(t.ctx.Logical().MyStations, key)
				} else {
					t.ctx.Logical().MyStations[ToPortKey(x.Port)] = &x
				}
				return nil, nil
			}
		case Table_IngressPipeUpstreamLines, Table_IngressPipeUpstreamAttachmentsV4:
			x, ok, err := t.evalAttachment(e.TableEntry)
			if err != nil {
				return nil, err
			}
			if translate {
				// TODO: implement validation
				return t.proc.HandleAttachmentEntry(&x, ok)
			} else {
				key := ToLineIdKey(x.LineId)
				if u.Type == p4v1.Update_DELETE {
					if x.Direction == DirectionUpstream {
						delete(t.ctx.Logical().UpstreamAttachments, key)
					} else {
						delete(t.ctx.Logical().DownstreamAttachments, key)
					}
				} else {
					if x.Direction == DirectionUpstream {
						t.ctx.Logical().UpstreamAttachments[key] = &x
					} else {
						t.ctx.Logical().DownstreamAttachments[key] = &x
					}
				}
				return nil, nil
			}
		case Table_IngressPipeUpstreamRoutesV4:
			x, err := parseUpstreamRouteV4Entry(e.TableEntry)
			if err != nil {
				return nil, err
			}
			if translate {
				// TODO: implement validation
				return t.proc.HandleRouteV4Entry(&x, u.Type)
			} else {
				key := ToIpv4LpmKey(x.Ipv4Addr, x.PrefixLen)
				if u.Type == p4v1.Update_DELETE {
					delete(t.ctx.Logical().UpstreamRoutesV4, key)
				} else {
					t.ctx.Logical().UpstreamRoutesV4[key] = &x
				}
				return nil, nil
			}
		// TODO: case Table_UpstreamPppoePunts // device-level
		// TODO: downstream tables
		default:
			return nil, fmt.Errorf("invalid table ID %v", e.TableEntry.TableId)
		}
	case *p4v1.Entity_ActionProfileGroup:
		switch e.ActionProfileGroup.ActionProfileId {
		case ActionProfile_IngressPipeUpstreamEcmp:
			x, err := parseUpstreamRoutesV4ActProfGroup(e.ActionProfileGroup)
			if err != nil {
				return nil, err
			}
			if translate {
				// TODO: implement validation
				return t.proc.HandleRouteV4NextHopGroup(&x, u.Type)
			} else {
				key := x.GroupId
				if u.Type == p4v1.Update_DELETE {
					delete(t.ctx.Logical().UpstreamNextHopGroups, key)
				} else {
					t.ctx.Logical().UpstreamNextHopGroups[key] = &x
				}
				return nil, nil
			}
		default:
			return nil, fmt.Errorf("invalid action profile ID %v", e.ActionProfileGroup.ActionProfileId)
		}
	case *p4v1.Entity_ActionProfileMember:
		switch e.ActionProfileMember.ActionProfileId {
		case ActionProfile_IngressPipeUpstreamEcmp:
			x, err := parseUpstreamRoutesV4ActProfMember(e.ActionProfileMember)
			if err != nil {
				return nil, err
			}
			if translate {
				// TODO: implement validation
				return t.proc.HandleRouteV4NextHopEntry(&x, u.Type)
			} else {
				key := x.Id
				if u.Type == p4v1.Update_DELETE {
					delete(t.ctx.Logical().UpstreamNextHopEntries, key)
				} else {
					t.ctx.Logical().UpstreamNextHopEntries[key] = &x
				}
				return nil, nil
			}
		default:
			return nil, fmt.Errorf("invalid action profile ID %v", e.ActionProfileMember.ActionProfileId)
		}
	default:
		return nil, fmt.Errorf("processing of %T not implemented", e)
	}
	// Should never be here.
}

// evalAttachment() evaluates a snapshot of the attachment that includes information from the given table entry, as well
// as the context. ok is true if the snapshot is complete (all fields are known), false otherwise.
func (t translator) evalAttachment(e *p4v1.TableEntry) (a AttachmentEntry, ok bool, err error) {
	switch e.TableId {
	case Table_IngressPipeUpstreamLines:
		err = parseUpstreamLineEntry(e, &a)
	case Table_IngressPipeUpstreamAttachmentsV4:
		err = parseUpstreamAttachmentV4Entry(e, &a)
	default:
		err = fmt.Errorf("table ID %d is not attachment-level", e.TableId)
	}
	if err != nil {
		return
	}
	if a.LineId == nil {
		panic("missing line ID in parsed table entry")
	}
	var stored *AttachmentEntry
	if a.Direction == DirectionUpstream {
		stored = t.ctx.Logical().UpstreamAttachments[ToLineIdKey(a.LineId)]
	} else if a.Direction == DirectionDownstream {
		stored = t.ctx.Logical().DownstreamAttachments[ToLineIdKey(a.LineId)]
	} else {
		panic("direction unknown")
	}
	if stored == nil {
		return
	}
	// Is there a less verbose way of evaluating the attachment?
	if a.Port == nil {
		a.Port = stored.Port
	}
	if a.STag == nil {
		a.STag = stored.STag
	}
	if a.CTag == nil {
		a.CTag = stored.CTag
	}
	if a.MacAddr == nil {
		a.MacAddr = stored.MacAddr
	}
	if a.Ipv4Addr == nil {
		a.Ipv4Addr = stored.Ipv4Addr
	}
	if a.PppoeSessId == nil {
		a.PppoeSessId = stored.PppoeSessId
	}
	ok = a.Port != nil && a.STag != nil && a.CTag != nil && a.MacAddr != nil && a.Ipv4Addr != nil && a.PppoeSessId != nil
	return
}

func parseIfTypeEntry(t *p4v1.TableEntry) (IfTypeEntry, error) {
	entry := IfTypeEntry{}
	// Parse match
	for _, m := range t.Match {
		switch m.FieldId {
		case Hdr_IngressPipeIfTypes_Port:
			entry.Port = m.GetExact().Value
		default:
			return IfTypeEntry{}, fmt.Errorf("invalid %T ID %d", m, m.FieldId)
		}
	}
	// Parse action
	act := t.GetAction().GetAction()
	if act == nil || act.ActionId != Action_IngressPipeSetIfType {
		return IfTypeEntry{}, fmt.Errorf("invalid Action %s", t.GetAction().String())
	}
	for _, p := range act.Params {
		switch p.ParamId {
		case ActionParam_IngressPipeSetIfType_IfType:
			entry.IfType = p.Value
		default:
			return IfTypeEntry{}, fmt.Errorf("invalid %T ID %d", p, p.ParamId)
		}
	}
	return entry, nil
}

func parseMyStationEntry(t *p4v1.TableEntry) (MyStationEntry, error) {
	entry := MyStationEntry{}
	// Parse match
	for _, m := range t.Match {
		switch m.FieldId {
		case Hdr_IngressPipeMyStations_Port:
			entry.Port = m.GetExact().Value
		case Hdr_IngressPipeMyStations_EthDst:
			entry.EthDst = m.GetExact().Value
		default:
			return MyStationEntry{}, fmt.Errorf("invalid %T ID %d", m, m.FieldId)
		}
	}
	// Parse action
	act := t.GetAction().GetAction()
	if act == nil || act.ActionId != Action_IngressPipeSetMyStation {
		return MyStationEntry{}, fmt.Errorf("invalid Action %s", t.GetAction())
	}
	return entry, nil
}

func parseUpstreamLineEntry(t *p4v1.TableEntry, a *AttachmentEntry) error {
	// Parse match
	a.Direction = DirectionUpstream
	for _, m := range t.Match {
		switch m.FieldId {
		case Hdr_IngressPipeUpstreamLines_Port:
			a.Port = m.GetExact().Value
		case Hdr_IngressPipeUpstreamLines_STag:
			a.STag = m.GetExact().Value
		case Hdr_IngressPipeUpstreamLines_CTag:
			a.CTag = m.GetExact().Value
		default:
			return fmt.Errorf("invalid %T ID %d", m, m.FieldId)
		}
	}
	// Parse action
	act := t.GetAction().GetAction()
	if act == nil || act.ActionId != Action_IngressPipeUpstreamSetLine {
		return fmt.Errorf("invalid Action %s", t.GetAction())
	}
	for _, p := range act.Params {
		switch p.ParamId {
		case ActionParam_IngressPipeUpstreamSetLine_LineId:
			a.LineId = p.Value
		default:
			return fmt.Errorf("invalid %T ID %d", p, p.ParamId)
		}
	}
	return nil
}

func parseUpstreamAttachmentV4Entry(t *p4v1.TableEntry, a *AttachmentEntry) error {
	// Parse match
	a.Direction = DirectionUpstream
	for _, m := range t.Match {
		switch m.FieldId {
		case Hdr_IngressPipeUpstreamAttachmentsV4_LineId:
			a.LineId = m.GetExact().Value
		case Hdr_IngressPipeUpstreamAttachmentsV4_EthSrc:
			a.MacAddr = m.GetExact().Value
		case Hdr_IngressPipeUpstreamAttachmentsV4_Ipv4Src:
			a.Ipv4Addr = m.GetExact().Value
		case Hdr_IngressPipeUpstreamAttachmentsV4_PppoeSessId:
			a.PppoeSessId = m.GetExact().Value
		default:
			return fmt.Errorf("invalid %T ID %d", m, m.FieldId)
		}
	}
	// Parse action
	act := t.GetAction().GetAction()
	if act == nil || act.ActionId != Action_Nop {
		return fmt.Errorf("invalid Action %s", t.GetAction())
	}
	return nil
}

func parseUpstreamRoutesV4ActProfMember(m *p4v1.ActionProfileMember) (NextHopEntry, error) {
	n := NextHopEntry{
		Id: m.MemberId,
	}
	// Parse action
	act := m.Action
	if act == nil || act.ActionId != Action_IngressPipeUpstreamRouteV4 {
		return n, fmt.Errorf("invalid Action %s", act)
	}
	for _, p := range act.Params {
		switch p.ParamId {
		case ActionParam_IngressPipeUpstreamRouteV4_Dmac:
			n.MacAddr = p.Value
		case ActionParam_IngressPipeUpstreamRouteV4_Port:
			n.Port = p.Value
		default:
			return n, fmt.Errorf("invalid %T ID %d", p, p.ParamId)
		}
	}
	return n, nil
}

func parseUpstreamRoutesV4ActProfGroup(g *p4v1.ActionProfileGroup) (NextHopGroup, error) {
	// No need to parse, simply wrap message in NextHopGroup.
	return NextHopGroup(*g), nil
}

func parseUpstreamRouteV4Entry(t *p4v1.TableEntry) (RouteV4Entry, error) {
	// Parse match
	r := RouteV4Entry{}
	r.Direction = DirectionUpstream
	for _, m := range t.Match {
		switch m.FieldId {
		case Hdr_IngressPipeUpstreamRoutesV4_Ipv4Dst:
			r.Ipv4Addr = m.GetLpm().Value
			r.PrefixLen = m.GetLpm().PrefixLen
		default:
			return r, fmt.Errorf("invalid %T ID %d", m, m.FieldId)
		}
	}
	// Parse action
	gid := t.GetAction().GetActionProfileGroupId()
	if gid == 0 {
		return r, fmt.Errorf("was expecting non-zero ActionProfileGroupId but found %v", t.GetAction())
	}
	r.NextHopGroupId = gid
	return r, nil
}
