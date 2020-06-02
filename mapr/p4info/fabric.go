package p4info

//noinspection GoSnakeCaseUsage
const(
	// FabricEgress.egress_next.egress_vlan
	Table_FabricEgressNextEgressVlan  uint32 =		33599342
	Action_FabricEgressNextPopVlan    uint32 =		16790030
	FieldMatch_FabricEgressNextVlanid uint32 =		1
	FieldMatch_FabricEgressNextEgport uint32 =		2

	// Filtering.ingress_port_vlan
	Table_FabricIngressFilteringIngressPortVlan    uint32 =		33611649
	Action_FabricIngressPermitWithInternalVlan     uint32 =		16794911
	FieldMatch_FabricIngressFilteringIngressPort   uint32 =		1
	FieldMatch_FabricIngressFilteringVlanIsValid   uint32 =		2
	FieldMatch_FabricIngressFilteringVlanVlanId    uint32 =		3
	FieldMatch_FabricIngressFilteringVlanInnerVlan uint32 =		4
)
