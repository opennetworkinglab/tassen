package store

import (
	v1 "github.com/p4lang/p4runtime/go/p4/v1"
	"testing"
)

var emptyTableEntries = map[string]*v1.TableEntry{}

var mockUpdateInsertTableEntry1 = v1.Update{
	Type:   v1.Update_INSERT,
	Entity: &v1.Entity{Entity: &v1.Entity_TableEntry{TableEntry: &mockTableEntry1}},
}

var mockUpdateInsertTableEntry2 = v1.Update{
	Type:   v1.Update_INSERT,
	Entity: &v1.Entity{Entity: &v1.Entity_TableEntry{TableEntry: &mockTableEntry2}},
}

func Test_store_PutAll(t *testing.T) {
	type fields struct {
		tableEntries map[string]*v1.TableEntry
	}
	type args struct {
		req *v1.WriteRequest
	}
	tests := []struct {
		name                string
		fields              fields
		args                args
		wantTableEntryCount int
	}{
		{
			name:   "insert 1 table entry",
			fields: fields{emptyTableEntries},
			args: args{&v1.WriteRequest{
				Updates: []*v1.Update{
					&mockUpdateInsertTableEntry1,
				},
			}},
			wantTableEntryCount: 1,
		},
		{
			name:   "insert 2 table entries",
			fields: fields{emptyTableEntries},
			args: args{&v1.WriteRequest{
				Updates: []*v1.Update{
					&mockUpdateInsertTableEntry1,
					&mockUpdateInsertTableEntry2,
				},
			}},
			wantTableEntryCount: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &store{
				tableEntries: tt.fields.tableEntries,
			}
			s.PutAll(tt.args.req)
			if gotTableEntryCount := s.TableEntryCount(); gotTableEntryCount != tt.wantTableEntryCount {
				t.Errorf("TableEntryCount() = %v, want %v", gotTableEntryCount, tt.wantTableEntryCount)
			}
		})
	}
}
