/*
 * Copyright 2020-present Open Networking Foundation
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package fabric

import (
	"fmt"
	v1 "github.com/p4lang/p4runtime/go/p4/v1"
	log "github.com/sirupsen/logrus"
	"mapr/translate"
)

// Implementation of a Processor interface for ONF's fabric.p4.

const (
	defaultInternalTag uint16 = 4094
	defaultPrio        int32  = 1
	FwdTypeIpv4Unicast byte   = 0x02
	EthTypeIpv4        uint16 = 0x0800
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

func (p fabricProcessor) HandleAttachmentEntry(a *translate.AttachmentEntry, ok bool) (targetUpdateEntries []*v1.Update, err error) {
	log.Tracef("AttachmentEntry={ %s }, complete=%v", a, ok)
	if ok {
		// The attachment is complete, generate the attachment-specific table entries
		targetTableEntries := make([]*v1.TableEntry, 0)
		switch a.Direction {
		case translate.DirectionUpstream:
			// Ingress Port Vlan for double tagged access port
			ingressPortVlanEntry := createIngressPortVlanEntryPermit(a.Port, a.STag, a.CTag, nil, defaultPrio)
			// t_line_map
			lineMapEntry := createLineMapEntry(a.STag, a.CTag, a.LineId)
			// t_pppoe_term_v4
			pppoeTermV4Entry := createPppoeTermV4(a.LineId, a.Ipv4Addr, a.PppoeSessId)
			targetTableEntries = append(targetTableEntries, &lineMapEntry, &ingressPortVlanEntry, &pppoeTermV4Entry)
			targetUpdateEntries = insertOrModifyTableEntries(p, targetTableEntries)
		case translate.DirectionDownstream:
			// Need to retrieve the switchMac from the MyStation entry
			x := p.ctx.Logical().MyStations[translate.ToPortKey(a.Port)]
			if x == nil {
				targetUpdateEntries = nil
				err = fmt.Errorf("missing MyStation entry for port %x, cannot derive source MAC", a.Port)
				return
			}
			// forwarding.routing_v4 entry
			routeV4Entry := createRouteV4Entry(getUInt32FromByteSlice(a.LineId), a.Ipv4Addr, 32)
			// next.routing_hashed entry
			nextHashedEntry := createNextHashedEntry(getUInt32FromByteSlice(a.LineId))
			// hashedSelector member
			// FIXME (daniele): Can member ID clash with other member ID? Currently we are using Line ID as Member ID
			hashedSelectorMember := createHashedSelectorMember(getUInt32FromByteSlice(a.LineId), a.Port, a.MacAddr, x.EthDst)
			memberKey := translate.KeyFromActProfMember(&hashedSelectorMember)
			updateTypeMember := v1.Update_INSERT
			if targetSelectorMember := p.ctx.Target().GetActProfMember(&memberKey); targetSelectorMember != nil {
				updateTypeMember = v1.Update_MODIFY
			}
			// hashedSelector group
			actionProfileGroup := v1.ActionProfileGroup{
				ActionProfileId: ActionProfile_FabricIngressNextHashedSelector,
				GroupId:         getUInt32FromByteSlice(a.LineId),
				Members: []*v1.ActionProfileGroup_Member{{
					MemberId: getUInt32FromByteSlice(a.LineId),
					Weight:   1,
				}},
				MaxSize: 1,
			}
			groupKey := translate.KeyFromActProfGroup(&actionProfileGroup)
			updateTypeGroup := v1.Update_INSERT
			if targetGroup := p.ctx.Target().GetActProfGroup(&groupKey); targetGroup != nil {
				updateTypeGroup = v1.Update_MODIFY
			}

			// next.next_vlan to push double vlan tag
			pushDoubleVlan := createNextVlanEntry(getUInt32FromByteSlice(a.LineId), a.STag, a.CTag)
			// t_line_map
			lineMapEntry := createLineMapEntry(a.STag, a.CTag, a.LineId)
			// t_line_sessionMap
			lineSessionMap := createLineSessionMap(a.LineId, a.PppoeSessId)
			targetTableEntries = append(targetTableEntries, &lineMapEntry, &lineSessionMap, &routeV4Entry, &nextHashedEntry, &pushDoubleVlan)

			// Make sure to have member, group and then next.routing_hashed entry
			targetUpdateEntries = append(targetUpdateEntries, createUpdateActProfMember(&hashedSelectorMember, updateTypeMember))
			targetUpdateEntries = append(targetUpdateEntries, createUpdateActProfGroup(&actionProfileGroup, updateTypeGroup))
			targetUpdateEntries = append(targetUpdateEntries, insertOrModifyTableEntries(p, targetTableEntries)...)
		}
	} else {
		// Query target store to understand which entries to remove
		delEntries := make([]*v1.TableEntry, 0)
		switch a.Direction {
		case translate.DirectionUpstream:
			if a.LineId != nil {
				delEntries = append(delEntries, getTargetEntriesUpstreamByLineId(p, a.LineId)...)
				if p.ctx.Logical().DownstreamAttachments[translate.ToLineIdKey(a.LineId)] != nil {
					log.Trace("Leaving TLineMap entry on the target because Downstream Attachment is still present")
					removeFirstTLineMap(delEntries)
				}
			}
			// Specific case for the ingress port VLAN entry
			if a.STag != nil && a.CTag != nil && a.Port != nil {
				// FIXME: if the first Logical rule removed is the upstream.attachments_v4 we'll never reach this point when removing rules
				// Create a "fake" rule just to get the key from the translate.KeyFromTableEntry helper method
				tempRule := createIngressPortVlanEntryPermit(a.Port, a.STag, a.CTag, nil, defaultPrio)
				key := translate.KeyFromTableEntry(&tempRule)
				// Otherwise it will append nil
				if remEntry := p.ctx.Target().GetTableEntry(&key); remEntry != nil {
					delEntries = append(delEntries, p.ctx.Target().GetTableEntry(&key))
				}
			}
			targetUpdateEntries = append(targetUpdateEntries, createUpdateEntries(delEntries, v1.Update_DELETE)...)
		case translate.DirectionDownstream:
			downEntries := make([]*v1.Update, 0)
			if a.LineId != nil {
				delEntries = append(delEntries, getTargetEntriesDownstreamByLineId(p, a.LineId)...)
				if p.ctx.Logical().UpstreamAttachments[translate.ToLineIdKey(a.LineId)] != nil {
					log.Trace("Leaving TLineMap entry on the target because Upstream Attachment is still present")
					removeFirstTLineMap(delEntries)
				}
				// Retrieve also group and member used for routing
				// This works since we used line ID as group ID and member ID.
				groupMemberKey := translate.ActProfGroupKey(ActionProfile_FabricIngressNextHashedSelector, getUInt32FromByteSlice(a.LineId))
				// Make sure the group is removed before the member
				if group := p.ctx.Target().GetActProfGroup(&groupMemberKey); group != nil {
					downEntries = append(downEntries, createUpdateActProfGroup(group, v1.Update_DELETE))
				}
				// In downstream we have a single member
				if member := p.ctx.Target().GetActProfMember(&groupMemberKey); member != nil {
					downEntries = append(downEntries, createUpdateActProfMember(member, v1.Update_DELETE))
				}
			}
			targetUpdateEntries = append(targetUpdateEntries, createUpdateEntries(delEntries, v1.Update_DELETE)...)
			targetUpdateEntries = append(targetUpdateEntries, downEntries...)
		}
	}
	return
}

func removeFirstTLineMap(updateEntries []*v1.TableEntry) {
	for i, v := range updateEntries {
		if v.TableId == Table_FabricIngressBngIngressTLineMap {
			updateEntries = append(updateEntries[:i], updateEntries[i+1:]...)
			return
		}
	}
}

func (p fabricProcessor) HandleRouteV4NextHopEntry(e *translate.NextHopEntry, uType v1.Update_Type) ([]*v1.Update, error) {
	log.Tracef("NextHopEntry={ %s }", e)
	x := p.ctx.Logical().MyStations[translate.ToPortKey(e.Port)]
	if x == nil {
		return nil, fmt.Errorf("missing MyStation entry for port %x, cannot derive source MAC", e.Port)
	}
	m := createHashedSelectorMember(e.Id, e.Port, e.MacAddr, x.EthDst)
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
	r := createRouteV4Entry(e.NextHopGroupId, e.Ipv4Addr, e.PrefixLen)
	switch e.Direction {
	case translate.DirectionUpstream:
		v := createNextVlanEntry(e.NextHopGroupId, getVlanIdValue(defaultInternalTag), nil)
		return []*v1.Update{createUpdateEntry(&r, uType), createUpdateEntry(&v, uType)}, nil
	default:
		return nil, fmt.Errorf("undefined route direction")
	}
}

func (p fabricProcessor) HandleAclEntry(e *translate.AclEntry, uType v1.Update_Type) ([]*v1.Update, error) {
	log.Tracef("AclEntry={ %s }", e)
	t, err := createAclEntry(e)
	if err != nil {
		return nil, err
	}
	return []*v1.Update{createUpdateEntry(&t, uType)}, nil
}

func (p fabricProcessor) HandlePpppoePunts(e *translate.PppoePuntedEntry, uType v1.Update_Type) ([]*v1.Update, error) {
	log.Tracef("PppoePuntEntry={ %s }", e)
	t := createPppoePuntEntry(e.PppoeCode, e.PppoeProto, defaultPrio)
	return []*v1.Update{createUpdateEntry(&t, uType)}, nil
}
