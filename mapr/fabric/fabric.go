package fabric

import (
	"fmt"
	v1 "github.com/p4lang/p4runtime/go/p4/v1"
	log "github.com/sirupsen/logrus"
	"mapr/translate"
)

// Implementation of a Processor interface for ONF's fabric.p4.

const (
	defaultInternalTag     uint16 = 4094
	defaultPrio            int32  = 1
	FwdType_FwdIpv4Unicast byte   = 0x02
	EthTypeIpv4            uint16 = 0x0800
)

type fabricProcessor struct {
	ctx translate.Context
}

func NewFabricProcessor(ctx translate.Context) translate.Processor {
	return &fabricProcessor{
		ctx: ctx,
	}
}

func (p fabricProcessor) HandleIfTypeEntry(e *translate.IfTypeEntry, uType v1.Update_Type) ([]*v1.Update, error) {
	log.Tracef("IfTypeEntry={ %s }", e)
	// TODO: check parameter of IfTypeEntry and return error
	switch e.IfType[0] {
	case translate.IfTypeCore:
		ingressPortVlanEntry := createIngressPortVlanEntryPermit(e.Port, nil, nil, getVlanIdValue(defaultInternalTag), defaultPrio)
		egressPopVlanEntry := createEgressVlanPopEntry(e.Port, defaultInternalTag)
		return []*v1.Update{createUpdateEntry(&ingressPortVlanEntry, uType), createUpdateEntry(&egressPopVlanEntry, uType)}, nil
	case translate.IfTypeAccess:
		log.Warnf("fabricProcessor.HandleIfTypeEntry(): not implemented for ACCESS ports")
	default:
		log.Warnf("IfTypeEntry.IfType=%v not implemented", e.IfType)
	}
	return nil, nil
}

func (p fabricProcessor) HandleMyStationEntry(e *translate.MyStationEntry, uType v1.Update_Type) ([]*v1.Update, error) {
	log.Tracef("MyStationEntry={ %s }", e)
	// TODO: check parameter of mystation entry and return error
	phyTableEntry := createFwdClassifierEntry(e.Port, e.EthDst, defaultPrio)
	return []*v1.Update{createUpdateEntry(&phyTableEntry, uType)}, nil
}

func (p fabricProcessor) HandleAttachmentEntry(a *translate.AttachmentEntry, ok bool) ([]*v1.Update, error) {
	log.Tracef("AttachmentEntry={ %s }, complete=%v", a, ok)
	if ok {
		// The attachment is complete, generate the rules
		switch a.Direction {
		case translate.DirectionUpstream:
			// Ingress Port Vlan for double tagged access port
			ingressPortVlanEntry := createIngressPortVlanEntryPermit(a.Port, a.STag, a.CTag, nil, defaultPrio)
			// BNG specific rules
			// - t_line_map
			lineMapEntry := createLineMapEntry(a.STag, a.CTag, a.LineId)
			// - t_pppoe_term_v4
			pppoeTermV4Entry := createPppoeTermV4(a.LineId, a.Ipv4Addr, a.PppoeSessId)

			targetUpdateEntries := make([]*v1.Update, 0)
			// Query target store to understand if insert or modify
			for _, v := range []*v1.TableEntry{&ingressPortVlanEntry, &lineMapEntry, &pppoeTermV4Entry} {
				key := translate.KeyFromTableEntry(v)
				targetTableEntry := p.ctx.Target().GetTableEntry(&key)
				updateType := v1.Update_INSERT
				if targetTableEntry != nil {
					// TODO: we could filter only the modified entries instead of always pushing a MODIFY to the target
					updateType = v1.Update_MODIFY
				}
				targetUpdateEntries = append(targetUpdateEntries, createUpdateEntry(v, updateType))
			}
			return targetUpdateEntries, nil
		case translate.DirectionDownstream:
			log.Tracef("fabricProcessor.HandleAttachmentEntry(): Downstream direction not implemented")
		}
	} else {
		switch a.Direction {
		case translate.DirectionUpstream:
			// Query target store to understand which entries to remove
			toBeRemovedEntries := make([]*v1.TableEntry, 0)
			if a.LineId != nil {
				toBeRemovedEntries = append(toBeRemovedEntries, getTargetEntriesUpstreamByLineId(p, a.LineId)...)
			}
			if a.STag != nil && a.CTag != nil && a.Port != nil {
				tempRule := createIngressPortVlanEntryPermit(a.Port, a.STag, a.CTag, nil, defaultPrio)
				key := translate.KeyFromTableEntry(&tempRule)
				remEntry := p.ctx.Target().GetTableEntry(&key)
				// Otherwise it will append nil
				if remEntry != nil {
					toBeRemovedEntries = append(toBeRemovedEntries, p.ctx.Target().GetTableEntry(&key))
				}
			}
			return createUpdateEntries(toBeRemovedEntries, v1.Update_DELETE), nil
		case translate.DirectionDownstream:
			log.Tracef("fabricProcessor.HandleAttachmentEntry(): Downstream direction not implemented")
		}
	}
	return nil, nil
}

func (p fabricProcessor) HandleRouteV4NextHopEntry(e *translate.NextHopEntry, uType v1.Update_Type) ([]*v1.Update, error) {
	log.Tracef("NextHopEntry={ %s }", e)
	x := p.ctx.Logical().MyStations[translate.ToPortKey(e.Port)]
	if x == nil {
		return nil, fmt.Errorf("missing MyStation entry for port %x, cannot derive source MAC", e.Port)
	}
	m := createHashedSelectorMember(e, x.EthDst)
	return []*v1.Update{createUpdateActProfMember(&m, uType)}, nil
}

func (p fabricProcessor) HandleRouteV4NextHopGroup(g *translate.NextHopGroup, uType v1.Update_Type) ([]*v1.Update, error) {
	log.Tracef("NextHopGroup={ %s }", g)
	// Generating the target group is easy if we use the same IDs for the members and group.
	group := v1.ActionProfileGroup{
		ActionProfileId: ActionProfile_FabricIngressNextHashedSelector,
		GroupId:         g.GroupId,
		Members:         g.Members,
		MaxSize:         g.MaxSize,
	}
	groupUpdate := createUpdateActProfGroup(&group, uType)
	nextEntry := createNextHashedEntry(g.GroupId)
	nextUpdate := createUpdateEntry(&nextEntry, uType)
	if uType == v1.Update_DELETE {
		// Table entry must be removed before group.
		return []*v1.Update{nextUpdate, groupUpdate}, nil
	} else {
		return []*v1.Update{groupUpdate, nextUpdate}, nil
	}
}

func (p fabricProcessor) HandleRouteV4Entry(e *translate.RouteV4Entry, uType v1.Update_Type) ([]*v1.Update, error) {
	log.Tracef("RouteV4Entry={ %s }", e)
	r := createRouteV4Entry(e)
	v := createNextVlanEntry(e, getVlanIdValue(defaultInternalTag))
	return []*v1.Update{createUpdateEntry(&r, uType), createUpdateEntry(&v, uType)}, nil
}
