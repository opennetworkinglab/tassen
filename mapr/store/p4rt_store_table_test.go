package store

import (
	v1 "github.com/p4lang/p4runtime/go/p4/v1"
	"github.com/stretchr/testify/assert"
	"testing"
)

var mockFieldMatch1 = v1.FieldMatch{
	FieldId: 1,
	FieldMatchType: &v1.FieldMatch_Exact_{
		Exact: &v1.FieldMatch_Exact{
			Value: []byte{0x01},
		},
	},
}

var mockFieldMatch2 = v1.FieldMatch{
	FieldId: 2,
	FieldMatchType: &v1.FieldMatch_Exact_{
		Exact: &v1.FieldMatch_Exact{
			Value: []byte{0x02},
		},
	},
}

var mockFieldMatch3 = v1.FieldMatch{
	FieldId: 3,
	FieldMatchType: &v1.FieldMatch_Exact_{
		Exact: &v1.FieldMatch_Exact{
			Value: []byte{0x03},
		},
	},
}

var mockFieldMatch4 = v1.FieldMatch{
	FieldId: 4,
	FieldMatchType: &v1.FieldMatch_Exact_{
		Exact: &v1.FieldMatch_Exact{
			Value: []byte{0x04},
		},
	},
}

var mockTableAction1 = v1.TableAction{
	Type: &v1.TableAction_Action{Action: &v1.Action{
		ActionId: 1,
		Params: []*v1.Action_Param{
			{ParamId: 1, Value: []byte{0x0A}},
			{ParamId: 2, Value: []byte{0x0B}},
		},
	}},
}

var mockTableAction2 = v1.TableAction{
	Type: &v1.TableAction_Action{Action: &v1.Action{
		ActionId: 1,
		Params: []*v1.Action_Param{
			{ParamId: 1, Value: []byte{0x0C}},
			{ParamId: 2, Value: []byte{0x0D}},
		},
	}},
}

var mockTableEntry1 = v1.TableEntry{
	TableId:  1,
	Match:    []*v1.FieldMatch{&mockFieldMatch1, &mockFieldMatch2},
	Action:   &mockTableAction1,
	Priority: 1,
}

var mockTableEntryKey1 = "1-[field_id:1 exact:<value:\"\\001\" >  field_id:2 exact:<value:\"\\002\" > ]-1"

var sameAsMockTableEntry1 = v1.TableEntry{
	TableId:  1,
	Match:    []*v1.FieldMatch{&mockFieldMatch1, &mockFieldMatch2},
	Action:   &mockTableAction1,
	Priority: 1,
}

var mockTableEntry2 = v1.TableEntry{
	TableId:  2,
	Match:    []*v1.FieldMatch{&mockFieldMatch3, &mockFieldMatch4},
	Action:   &mockTableAction2,
	Priority: 1,
}

var mockTableEntryKey2 = "2-[field_id:3 exact:<value:\"\\003\" >  field_id:4 exact:<value:\"\\004\" > ]-1"

func Test_store_FilterTableEntries(t *testing.T) {
	type fields struct {
		tableEntries map[string]*v1.TableEntry
	}
	type args struct {
		f func(*v1.TableEntry) bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []*v1.TableEntry
	}{
		{
			name:   "empty",
			fields: fields{map[string]*v1.TableEntry{ /* empty */ }},
			args: args{f: func(e *v1.TableEntry) bool {
				return e.TableId == 1
			}},
			want: []*v1.TableEntry{ /* empty */ },
		},
		{
			name: "get 1",
			fields: fields{map[string]*v1.TableEntry{
				mockTableEntryKey1: &mockTableEntry1,
				mockTableEntryKey2: &mockTableEntry2,
			}},
			args: args{f: func(e *v1.TableEntry) bool {
				return e.TableId == mockTableEntry1.TableId
			}},
			want: []*v1.TableEntry{
				&mockTableEntry1,
			},
		},
		{
			name: "get 2",
			fields: fields{map[string]*v1.TableEntry{
				mockTableEntryKey1: &mockTableEntry1,
				mockTableEntryKey2: &mockTableEntry2,
			}},
			args: args{f: func(e *v1.TableEntry) bool {
				return e.TableId <= 99
			}},
			want: []*v1.TableEntry{
				&mockTableEntry1,
				&mockTableEntry2,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := p4RtStore{
				tableEntries: tt.fields.tableEntries,
			}
			got := s.FilterTableEntries(tt.args.f)
			assert.ElementsMatch(t, tt.want, got, "FilterTableEntries(): elements don't match the expected ones")
		})
	}
}

func Test_store_PutTableEntry(t *testing.T) {
	type fields struct {
		tableEntries map[string]*v1.TableEntry
	}
	type args struct {
		entry *v1.TableEntry
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		wantCount int
	}{
		{"empty", fields{map[string]*v1.TableEntry{
			// empty
		}}, args{&mockTableEntry1}, 1},
		{"existing key", fields{map[string]*v1.TableEntry{
			mockTableEntryKey1: &mockTableEntry1,
		}}, args{&sameAsMockTableEntry1}, 1},
		{"new key", fields{map[string]*v1.TableEntry{
			mockTableEntryKey1: &mockTableEntry1,
		}}, args{&mockTableEntry2}, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := p4RtStore{
				tableEntries: tt.fields.tableEntries,
			}
			s.PutTableEntry(tt.args.entry)
			assert.Equal(t, tt.wantCount, s.TableEntryCount(), "TableEntryCount() should return expected count")
		})
	}
}

func Test_store_GetTableEntry(t *testing.T) {
	type fields struct {
		tableEntries map[string]*v1.TableEntry
	}
	type args struct {
		key *string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *v1.TableEntry
	}{
		{"empty", fields{map[string]*v1.TableEntry{
			// empty
		}}, args{&mockTableEntryKey1}, nil},
		{"non existing key", fields{map[string]*v1.TableEntry{
			mockTableEntryKey1: &mockTableEntry1,
		}}, args{&mockTableEntryKey2}, nil},
		{"existing key", fields{map[string]*v1.TableEntry{
			mockTableEntryKey1: &mockTableEntry1,
			mockTableEntryKey2: &mockTableEntry2,
		}}, args{&mockTableEntryKey2}, &mockTableEntry2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := p4RtStore{
				tableEntries: tt.fields.tableEntries,
			}
			assert.Equal(t, tt.want, s.GetTableEntry(tt.args.key), "GetTableEntry() should return expected value")
		})
	}
}

func Test_store_RemoveTableEntry(t *testing.T) {
	type fields struct {
		tableEntries map[string]*v1.TableEntry
	}
	type args struct {
		entry *v1.TableEntry
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		wantCount int
	}{
		{"empty", fields{map[string]*v1.TableEntry{
			// empty
		}}, args{&mockTableEntry1}, 0},
		{"existing key", fields{map[string]*v1.TableEntry{
			mockTableEntryKey1: &mockTableEntry1,
		}}, args{&sameAsMockTableEntry1}, 0},
		{"non existing key", fields{map[string]*v1.TableEntry{
			mockTableEntryKey1: &mockTableEntry1,
		}}, args{&mockTableEntry2}, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := p4RtStore{
				tableEntries: tt.fields.tableEntries,
			}
			s.RemoveTableEntry(tt.args.entry)
			assert.Equal(t, tt.wantCount, s.TableEntryCount(), "TableEntryCount() should return expected count")
		})
	}
}

func Test_store_TableEntries(t *testing.T) {
	type fields struct {
		tableEntries map[string]*v1.TableEntry
	}
	tests := []struct {
		name   string
		fields fields
		want   []*v1.TableEntry
	}{
		{name: "empty", fields: fields{map[string]*v1.TableEntry{
				// empty
		}}, want: []*v1.TableEntry{
			// empty
		}},
		{name: "1 entry", fields: fields{map[string]*v1.TableEntry{
			mockTableEntryKey1: &mockTableEntry1,
		}}, want: []*v1.TableEntry{
			&mockTableEntry1,
		}},
		{name: "2 entries", fields: fields{map[string]*v1.TableEntry{
			mockTableEntryKey1: &mockTableEntry1,
			mockTableEntryKey2: &mockTableEntry2,
		}}, want: []*v1.TableEntry{
			&mockTableEntry1,
			&mockTableEntry2,
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := p4RtStore{
				tableEntries: tt.fields.tableEntries,
			}
			got := s.TableEntries()
			assert.ElementsMatch(t, tt.want, got, "TableEntries(): elements don't match the expected ones")
		})
	}
}

func Test_tableEntryKey(t *testing.T) {
	type args struct {
		k *v1.TableEntry
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"1", args{&mockTableEntry1}, mockTableEntryKey1},
		{"sameAs1", args{&sameAsMockTableEntry1}, mockTableEntryKey1},
		{"2", args{&mockTableEntry2}, mockTableEntryKey2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, KeyFromTableEntry(tt.args.k), "KeyFromTableEntry() should return expected key")
		})
	}
}

func Test_store_TableEntryCount(t *testing.T) {
	type fields struct {
		tableEntries map[string]*v1.TableEntry
	}
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		{"1 entry", fields{map[string]*v1.TableEntry{
			mockTableEntryKey1: &mockTableEntry1,
		}}, 1},
		{"2 entries", fields{map[string]*v1.TableEntry{
			mockTableEntryKey1: &mockTableEntry1,
			mockTableEntryKey2: &mockTableEntry2,
		}}, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := p4RtStore{
				tableEntries: tt.fields.tableEntries,
			}
			assert.Equal(t, tt.want, s.TableEntryCount(), "TableEntryCount() should return expected count")
		})
	}
}
