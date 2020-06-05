package fabric

//noinspection GoSnakeCaseUsage
const (
	// FabricEgress.egress_next.egress_vlan
	Table_EgressVlan            uint32 = 33599342
	Action_EgressVlanPopVlan    uint32 = 16790030
	FieldMatch_EgressVlanVlanId uint32 = 1
	FieldMatch_EgressVlanEgPort uint32 = 2

	// Filtering.ingress_port_vlan
	Table_IngressPortVlan                        uint32 = 33611649
	Action_IngressPortVlanPermitWithInternalVlan uint32 = 16794911
	FieldMatch_IngressPortVlanIgPort             uint32 = 1
	FieldMatch_IngressPortVlanVlanIsValid        uint32 = 2
	FieldMatch_IngressPortVlanVlanId             uint32 = 3
	FieldMatch_IngresPortVlanInnerVlan           uint32 = 4

	// Filtering.fwd_classifier
	Table_FwdClassifier                   uint32 = 33596298
	Action_FwdClassifierSetForwardingType uint32 = 16840921
	FieldMatch_FwdClassifierIgPort        uint32 = 1
	FieldMatch_FwdClassifierEthDst        uint32 = 2
	FieldMatch_FwdClassifierEthType       uint32 = 3
	FieldMatch_FwdClassifierIpEthType     uint32 = 4

	FwdType_FwdIpv4Unicast byte = 0x02

	EthTypeIpv4 uint16 = 0x0800
)
