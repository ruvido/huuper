package requests

import (
	"errors"
	"reflect"
	"testing"

	"github.com/pocketbase/pocketbase/core"
)

func newGroupRecord(id string) *core.Record {
	collection := core.NewBaseCollection("groups")
	collection.Fields.Add(&core.TextField{Name: "id"})
	record := core.NewRecord(collection)
	record.Id = id
	return record
}

func TestGeneratePromoteInvitesAllSucceed(t *testing.T) {
	targetIDs := []string{"g1", "g2"}
	attempted := []string{}

	created := generatePromoteInvites(
		"user1",
		targetIDs,
		func(id string) (*core.Record, error) { return newGroupRecord(id), nil },
		func(*core.Record) bool { return true },
		func(group *core.Record) error {
			attempted = append(attempted, group.Id)
			return nil
		},
	)

	if !reflect.DeepEqual(created, targetIDs) {
		t.Fatalf("expected all groups tracked as created, got %v", created)
	}
	if !reflect.DeepEqual(attempted, targetIDs) {
		t.Fatalf("expected invite attempted for all groups, got %v", attempted)
	}
}

func TestGeneratePromoteInvitesSkipsGroupsWithoutChatID(t *testing.T) {
	attempted := []string{}

	created := generatePromoteInvites(
		"user1",
		[]string{"g1", "g2"},
		func(id string) (*core.Record, error) { return newGroupRecord(id), nil },
		func(group *core.Record) bool { return group.Id == "g2" },
		func(group *core.Record) error {
			attempted = append(attempted, group.Id)
			return nil
		},
	)

	if !reflect.DeepEqual(created, []string{"g2"}) {
		t.Fatalf("expected only g2 tracked, got %v", created)
	}
	if !reflect.DeepEqual(attempted, []string{"g2"}) {
		t.Fatalf("expected invite attempted only for g2, got %v", attempted)
	}
}

func TestGeneratePromoteInvitesNonBlockingOnInviteFailure(t *testing.T) {
	attempted := []string{}

	created := generatePromoteInvites(
		"user1",
		[]string{"g1", "g2", "g3"},
		func(id string) (*core.Record, error) { return newGroupRecord(id), nil },
		func(*core.Record) bool { return true },
		func(group *core.Record) error {
			attempted = append(attempted, group.Id)
			if group.Id == "g2" {
				return errors.New("Bad Request: chat not found")
			}
			return nil
		},
	)

	if !reflect.DeepEqual(created, []string{"g1", "g3"}) {
		t.Fatalf("expected only successful groups tracked, got %v", created)
	}
	if !reflect.DeepEqual(attempted, []string{"g1", "g2", "g3"}) {
		t.Fatalf("expected invite attempted for every group, got %v", attempted)
	}
}

func TestGeneratePromoteInvitesNonBlockingOnLoadFailure(t *testing.T) {
	attempted := []string{}

	created := generatePromoteInvites(
		"user1",
		[]string{"missing", "g2"},
		func(id string) (*core.Record, error) {
			if id == "missing" {
				return nil, errors.New("not found")
			}
			return newGroupRecord(id), nil
		},
		func(*core.Record) bool { return true },
		func(group *core.Record) error {
			attempted = append(attempted, group.Id)
			return nil
		},
	)

	if !reflect.DeepEqual(created, []string{"g2"}) {
		t.Fatalf("expected only g2 tracked, got %v", created)
	}
	if !reflect.DeepEqual(attempted, []string{"g2"}) {
		t.Fatalf("expected invite attempted only for g2, got %v", attempted)
	}
}

func TestGeneratePromoteInvitesAllFailuresStillReturnsEmptySlice(t *testing.T) {
	created := generatePromoteInvites(
		"user1",
		[]string{"g1", "g2"},
		func(id string) (*core.Record, error) { return newGroupRecord(id), nil },
		func(*core.Record) bool { return true },
		func(*core.Record) error { return errors.New("chat not found") },
	)

	if len(created) != 0 {
		t.Fatalf("expected no groups tracked when all invites fail, got %v", created)
	}
}
