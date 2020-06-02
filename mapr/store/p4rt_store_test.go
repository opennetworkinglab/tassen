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

func Test_store_Update(t *testing.T) {
	type fields struct {
		tableEntries map[string]*v1.TableEntry
	}
	type args struct {
		req    *v1.WriteRequest
		dryRun bool
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
			args: args{
				req: &v1.WriteRequest{
					Updates: []*v1.Update{
						&mockUpdateInsertTableEntry1,
					},
				},
				dryRun: false,
			},
			wantTableEntryCount: 1,
		},
		{
			name:   "insert 2 table entries",
			fields: fields{emptyTableEntries},
			args: args{
				req: &v1.WriteRequest{
					Updates: []*v1.Update{
						&mockUpdateInsertTableEntry1,
						&mockUpdateInsertTableEntry2,
					},
				},
				dryRun: false,
			},
			wantTableEntryCount: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &p4RtStore{
				tableEntries: tt.fields.tableEntries,
			}
			s.Update(tt.args.req, tt.args.dryRun)
			if gotTableEntryCount := s.TableEntryCount(); gotTableEntryCount != tt.wantTableEntryCount {
				t.Errorf("TableEntryCount() = %v, want %v", gotTableEntryCount, tt.wantTableEntryCount)
			}
		})
	}
}