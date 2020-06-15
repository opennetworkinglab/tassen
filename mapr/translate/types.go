package translate

import (
	"fmt"
	p4v1 "github.com/p4lang/p4runtime/go/p4/v1"
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
	return fmt.Sprintf("Dir: %s, Port: %x, LineId: %x, STag: %x, CTag: %x, MacAddr: %x, Ipv4Addr: %x, PppoeSessId: %x",
		a.Direction, a.Port, a.LineId, a.STag, a.CTag, a.MacAddr, a.Ipv4Addr, a.PppoeSessId)
}

// Abstraction of an action profile member for ECMP-capable routing tables
type NextHopEntry struct {
	Id      uint32
	Port    []byte
	MacAddr []byte
}

func (n NextHopEntry) String() string {
	return fmt.Sprintf("Id: %d, Port: %x, MacAddr: %x", n.Id, n.Port, n.MacAddr)
}

// A group of NextHopEntry, i.e. an ECMP group
// No need for higher level abstractions, P4RT ActionProfileGroup works just fine.
type NextHopGroup p4v1.ActionProfileGroup

func (n NextHopGroup) String() string {
	return fmt.Sprintf("Id: %d, Members: %s", n.GroupId, n.Members)
}

// An IPv4 routing entry.
type RouteV4Entry struct {
	Direction      Direction
	Ipv4Addr       []byte
	PrefixLen      int32
	NextHopGroupId uint32
}

func (r RouteV4Entry) String() string {
	return fmt.Sprintf("Dir: %s, Ipv4Addr: %x, PrefixLen: %d, NextHopGroupId: %d",
		r.Direction, r.Ipv4Addr, r.PrefixLen, r.NextHopGroupId)
}

type PortKey [2]byte

func ToPortKey(b []byte) PortKey {
	return PortKey{b[0], b[1]}
}

type LineIdKey [4]byte

func ToLineIdKey(b []byte) LineIdKey {
	return LineIdKey{b[0], b[1], b[2], b[3]}
}

type Ipv4LpmKey string

func ToIpv4LpmKey(addr []byte, prefixLen int32) Ipv4LpmKey {
	return Ipv4LpmKey(fmt.Sprintf("%x/%d", addr, prefixLen))
}

// An entry in the ACL table, for now it's just a TableEntry
type AclEntry p4v1.TableEntry

func (a AclEntry) String() string {
	return fmt.Sprintf("%v", (p4v1.TableEntry)(a))
}

type AclKey string

func ToAclKey(t *AclEntry) AclKey {
	return AclKey(KeyFromTableEntry((*p4v1.TableEntry)(t)))
}

type PppoePuntedEntry struct {
	PppoeCode  []byte
	PppoeProto []byte
}

func (c PppoePuntedEntry) String() string {
	return fmt.Sprintf("PPPoECode: %x, PPPoEProto: %x", c.PppoeCode, c.PppoeProto)
}

type CtrlPuntedKey string

func ToCtrlPuntedKey(pppoeCode []byte, pppoeProto []byte) CtrlPuntedKey {
	return CtrlPuntedKey(fmt.Sprintf("%x/%x", pppoeCode, pppoeProto))
}
