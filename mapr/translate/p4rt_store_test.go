package translate

import (
	p4v1 "github.com/p4lang/p4runtime/go/p4/v1"
	"testing"
)

var emptyTableEntries = map[string]*p4v1.TableEntry{}

var mockUpdateInsertTableEntry1 = p4v1.Update{
	Type:   p4v1.Update_INSERT,
	Entity: &p4v1.Entity{Entity: &p4v1.Entity_TableEntry{TableEntry: &mockTableEntry1}},
}

func Test_store_Update(t *testing.T) {
	type fields struct {
		tableEntries map[string]*p4v1.TableEntry
	}
	type args struct {
		update *p4v1.Update
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
				update: &mockUpdateInsertTableEntry1,
				dryRun: false,
			},
			wantTableEntryCount: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &p4RtStore{
				tableEntries: tt.fields.tableEntries,
			}
			s.Update(tt.args.update, tt.args.dryRun)
			if gotTableEntryCount := s.TableEntryCount(); gotTableEntryCount != tt.wantTableEntryCount {
				t.Errorf("TableEntryCount() = %v, want %v", gotTableEntryCount, tt.wantTableEntryCount)
			}
		})
	}
}
