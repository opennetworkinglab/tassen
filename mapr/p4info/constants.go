package p4info

// TODO: use script to auto-generate constants from p4info file

//noinspection GoSnakeCaseUsage
const (
	Action_Nop uint32 = 28485346

	Table_IfTypes           uint32 = 38498675
	FieldMatch_IfTypes_Port uint32 = 1
	Action_SetIfType        uint32 = 18538368
	Param_SetIfType_IfType  uint32 = 1

	Table_MyStations               uint32 = 49392761
	FieldMatch_MyStations_Port     uint32 = 1
	FieldMatch_MyStations_EthDst   uint32 = 2
	Action_MyStations_SetMyStation uint32 = 29456969

	Table_UpstreamLines           uint32 = 33956689
	FieldMatch_UpstreamLines_Ctag uint32 = 1
	FieldMatch_UpstreamLines_Stag uint32 = 2
	Action_UpstreamSetLine        uint32 = 17659136
	Param_UpstreamSetLine_LineId  uint32 = 1

	Table_UpstreamAttachmentsV4                  uint32 = 44507663
	FieldMatch_UpstreamAttachmentsV4_LineId      uint32 = 1
	FieldMatch_UpstreamAttachmentsV4_EthSrc      uint32 = 2
	FieldMatch_UpstreamAttachmentsV4_Ipv4Src     uint32 = 3
	FieldMatch_UpstreamAttachmentsV4_PppoeSessId uint32 = 4
)

// TODO: use P4 enums and generate values from p4info

const (
	IfTypeUnknown byte = 0x00
	IfTypeCore    byte = 0x01
	IfTypeAccess  byte = 0x02
)
