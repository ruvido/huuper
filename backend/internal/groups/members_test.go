package groups

import (
	"testing"

	"github.com/pocketbase/pocketbase/core"
)

func TestIsAdminUser(t *testing.T) {
	t.Run("admin flag", func(t *testing.T) {
		user := newGroupTestUser(true)
		if !isAdminUser(user) {
			t.Fatalf("expected admin=true user to be admin")
		}
	})

	t.Run("non admin", func(t *testing.T) {
		user := newGroupTestUser(false)
		if isAdminUser(user) {
			t.Fatalf("expected non-admin user to not be admin")
		}
	})
}

func TestShouldIncludeGroupMember(t *testing.T) {
	admin := newGroupTestUser(true)
	normal := newGroupTestUser(false)

	if shouldIncludeGroupMember(admin, false, false) {
		t.Fatalf("expected plain admin to be excluded from members")
	}
	if !shouldIncludeGroupMember(admin, true, false) {
		t.Fatalf("expected admin assistant to be included")
	}
	if !shouldIncludeGroupMember(admin, false, true) {
		t.Fatalf("expected admin guardian to be included")
	}
	if !shouldIncludeGroupMember(normal, false, false) {
		t.Fatalf("expected non-admin member to be included")
	}
}

func newGroupTestUser(admin bool) *core.Record {
	collection := core.NewBaseCollection("users")
	collection.Fields.Add(&core.BoolField{Name: "admin"})

	record := core.NewRecord(collection)
	record.Set("admin", admin)
	return record
}
