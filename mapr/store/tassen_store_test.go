package store

import (
	v1 "github.com/p4lang/p4runtime/go/p4/v1"
	"github.com/stretchr/testify/assert"
	"mapr/p4info"
	"testing"
)

var mockPort1 = []byte{0x00, 0x01}
var mockPort2 = []byte{0x00, 0x02}

var mockTableEntryIfTypesPort1Core = v1.TableEntry{
	TableId: p4info.Table_IfTypes,
	Match: []*v1.FieldMatch{
		{
			FieldId: p4info.FieldMatch_IfTypes_Port,
			FieldMatchType: &v1.FieldMatch_Exact_{
				Exact: &v1.FieldMatch_Exact{
					Value: mockPort1,
				},
			},
		},
	},
	Action: &v1.TableAction{
		Type: &v1.TableAction_Action{
			Action: &v1.Action{
				ActionId: p4info.Action_SetIfType,
				Params: []*v1.Action_Param{
					{
						ParamId: p4info.Param_SetIfType_IfType,
						Value:   []byte{p4info.IfTypeCore}},
				},
			}},
	},
}

var mockTableEntryIfTypesPort2Access = v1.TableEntry{
	TableId: p4info.Table_IfTypes,
	Match: []*v1.FieldMatch{
		{
			FieldId: p4info.FieldMatch_IfTypes_Port,
			FieldMatchType: &v1.FieldMatch_Exact_{
				Exact: &v1.FieldMatch_Exact{
					Value: mockPort2,
				},
			},
		},
	},
	Action: &v1.TableAction{
		Type: &v1.TableAction_Action{
			Action: &v1.Action{
				ActionId: p4info.Action_SetIfType,
				Params: []*v1.Action_Param{
					{
						ParamId: p4info.Param_SetIfType_IfType,
						Value:   []byte{p4info.IfTypeAccess}},
				},
			}},
	},
}

var mockTableEntryIfTypesInvalidFieldMatch = v1.TableEntry{
	TableId: p4info.Table_IfTypes,
	Match: []*v1.FieldMatch{
		{
			FieldId: p4info.FieldMatch_IfTypes_Port - 1,
			FieldMatchType: &v1.FieldMatch_Exact_{
				Exact: &v1.FieldMatch_Exact{
					Value: mockPort2,
				},
			},
		},
	},
	Action: &v1.TableAction{
		Type: &v1.TableAction_Action{
			Action: &v1.Action{
				ActionId: p4info.Action_SetIfType,
				Params: []*v1.Action_Param{
					{
						ParamId: p4info.Param_SetIfType_IfType,
						Value:   []byte{p4info.IfTypeAccess}},
				},
			}},
	},
}

var mockTableEntryIfTypesInvalidActionId = v1.TableEntry{
	TableId: p4info.Table_IfTypes,
	Match: []*v1.FieldMatch{
		{
			FieldId: p4info.FieldMatch_IfTypes_Port,
			FieldMatchType: &v1.FieldMatch_Exact_{
				Exact: &v1.FieldMatch_Exact{
					Value: mockPort2,
				},
			},
		},
	},
	Action: &v1.TableAction{
		Type: &v1.TableAction_Action{
			Action: &v1.Action{
				ActionId: p4info.Action_SetIfType - 1,
				Params: []*v1.Action_Param{
					{
						ParamId: p4info.Param_SetIfType_IfType,
						Value:   []byte{p4info.IfTypeAccess}},
				},
			}},
	},
}

var mockTableEntryIfTypesInvalidActionParamId = v1.TableEntry{
	TableId: p4info.Table_IfTypes,
	Match: []*v1.FieldMatch{
		{
			FieldId: p4info.FieldMatch_IfTypes_Port,
			FieldMatchType: &v1.FieldMatch_Exact_{
				Exact: &v1.FieldMatch_Exact{
					Value: mockPort2,
				},
			},
		},
	},
	Action: &v1.TableAction{
		Type: &v1.TableAction_Action{
			Action: &v1.Action{
				ActionId: p4info.Action_SetIfType,
				Params: []*v1.Action_Param{
					{
						ParamId: p4info.Param_SetIfType_IfType - 1,
						Value:   []byte{p4info.IfTypeAccess}},
				},
			}},
	},
}

func Test_ParseIfTypeEntry(t *testing.T) {
	type args struct {
		t *v1.TableEntry
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
			want:    IfTypeEntry{Port: mockPort1, IfType: []byte{p4info.IfTypeCore}},
			wantErr: false,
		},
		{
			name:    "port 2 is access",
			args:    args{&mockTableEntryIfTypesPort2Access},
			want:    IfTypeEntry{Port: mockPort2, IfType: []byte{p4info.IfTypeAccess}},
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
