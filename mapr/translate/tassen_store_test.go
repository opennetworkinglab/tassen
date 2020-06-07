package translate

import (
	p4v1 "github.com/p4lang/p4runtime/go/p4/v1"
	"github.com/stretchr/testify/assert"
	"testing"
)

var mockPort1 = []byte{0x00, 0x01}
var mockPort2 = []byte{0x00, 0x02}

var mockTableEntryIfTypesPort1Core = p4v1.TableEntry{
	TableId: Table_IngressPipeIfTypes,
	Match: []*p4v1.FieldMatch{
		{
			FieldId: Hdr_IngressPipeIfTypes_Port,
			FieldMatchType: &p4v1.FieldMatch_Exact_{
				Exact: &p4v1.FieldMatch_Exact{
					Value: mockPort1,
				},
			},
		},
	},
	Action: &p4v1.TableAction{
		Type: &p4v1.TableAction_Action{
			Action: &p4v1.Action{
				ActionId: Action_IngressPipeSetIfType,
				Params: []*p4v1.Action_Param{
					{
						ParamId: ActionParam_IngressPipeSetIfType_IfType,
						Value:   []byte{IfTypeCore}},
				},
			}},
	},
}

var mockTableEntryIfTypesPort2Access = p4v1.TableEntry{
	TableId: Table_IngressPipeIfTypes,
	Match: []*p4v1.FieldMatch{
		{
			FieldId: Hdr_IngressPipeIfTypes_Port,
			FieldMatchType: &p4v1.FieldMatch_Exact_{
				Exact: &p4v1.FieldMatch_Exact{
					Value: mockPort2,
				},
			},
		},
	},
	Action: &p4v1.TableAction{
		Type: &p4v1.TableAction_Action{
			Action: &p4v1.Action{
				ActionId: Action_IngressPipeSetIfType,
				Params: []*p4v1.Action_Param{
					{
						ParamId: ActionParam_IngressPipeSetIfType_IfType,
						Value:   []byte{IfTypeAccess}},
				},
			}},
	},
}

var mockTableEntryIfTypesInvalidFieldMatch = p4v1.TableEntry{
	TableId: Table_IngressPipeIfTypes,
	Match: []*p4v1.FieldMatch{
		{
			FieldId: Hdr_IngressPipeIfTypes_Port - 1,
			FieldMatchType: &p4v1.FieldMatch_Exact_{
				Exact: &p4v1.FieldMatch_Exact{
					Value: mockPort2,
				},
			},
		},
	},
	Action: &p4v1.TableAction{
		Type: &p4v1.TableAction_Action{
			Action: &p4v1.Action{
				ActionId: Action_IngressPipeSetIfType,
				Params: []*p4v1.Action_Param{
					{
						ParamId: ActionParam_IngressPipeSetIfType_IfType,
						Value:   []byte{IfTypeAccess}},
				},
			}},
	},
}

var mockTableEntryIfTypesInvalidActionId = p4v1.TableEntry{
	TableId: Table_IngressPipeIfTypes,
	Match: []*p4v1.FieldMatch{
		{
			FieldId: Hdr_IngressPipeIfTypes_Port,
			FieldMatchType: &p4v1.FieldMatch_Exact_{
				Exact: &p4v1.FieldMatch_Exact{
					Value: mockPort2,
				},
			},
		},
	},
	Action: &p4v1.TableAction{
		Type: &p4v1.TableAction_Action{
			Action: &p4v1.Action{
				ActionId: Action_IngressPipeSetIfType - 1,
				Params: []*p4v1.Action_Param{
					{
						ParamId: ActionParam_IngressPipeSetIfType_IfType,
						Value:   []byte{IfTypeAccess}},
				},
			}},
	},
}

var mockTableEntryIfTypesInvalidActionParamId = p4v1.TableEntry{
	TableId: Table_IngressPipeIfTypes,
	Match: []*p4v1.FieldMatch{
		{
			FieldId: Hdr_IngressPipeIfTypes_Port,
			FieldMatchType: &p4v1.FieldMatch_Exact_{
				Exact: &p4v1.FieldMatch_Exact{
					Value: mockPort2,
				},
			},
		},
	},
	Action: &p4v1.TableAction{
		Type: &p4v1.TableAction_Action{
			Action: &p4v1.Action{
				ActionId: Action_IngressPipeSetIfType,
				Params: []*p4v1.Action_Param{
					{
						ParamId: ActionParam_IngressPipeSetIfType_IfType - 1,
						Value:   []byte{IfTypeAccess}},
				},
			}},
	},
}

func Test_ParseIfTypeEntry(t *testing.T) {
	type args struct {
		t *p4v1.TableEntry
	}
	tests := []struct {
		name    string
		args    args
		want    IfTypeEntry
		wantErr bool
	}{
		{
			name:    "port 1 is core",
			args:    args{&mockTableEntryIfTypesPort1Core},
			want:    IfTypeEntry{Port: mockPort1, IfType: []byte{IfTypeCore}},
			wantErr: false,
		},
		{
			name:    "port 2 is access",
			args:    args{&mockTableEntryIfTypesPort2Access},
			want:    IfTypeEntry{Port: mockPort2, IfType: []byte{IfTypeAccess}},
			wantErr: false,
		},
		{
			name:    "invalid field match id",
			args:    args{&mockTableEntryIfTypesInvalidFieldMatch},
			want:    IfTypeEntry{},
			wantErr: true,
		},
		{
			name:    "invalid action id",
			args:    args{&mockTableEntryIfTypesInvalidActionId},
			want:    IfTypeEntry{},
			wantErr: true,
		},
		{
			name:    "invalid action param id",
			args:    args{&mockTableEntryIfTypesInvalidActionParamId},
			want:    IfTypeEntry{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseIfTypeEntry(tt.args.t)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseIfTypeEntry() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got, "parseIfTypeEntry(): should return expected value")
		})
	}
}
