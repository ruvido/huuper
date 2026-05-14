package requests

import "testing"

func TestPromotedRequestSnapshotUsesUniformEventShape(t *testing.T) {
	record := testRequestRecord()
	record.Id = "request123"
	record.Set("group", "group123")
	record.Set("guardian", "guardian123")

	data := map[string]any{
		"assign_group": map[string]any{
			"assigned_at": "2026-04-17T20:32:59Z",
			"assigned_by": "Ruvido",
			"group": map[string]any{
				"id":   "group123",
				"name": "Il Branco del Sud",
			},
		},
		"guardian": map[string]any{
			"name":        "Daniele Croce",
			"assigned_at": "2026-04-19T14:35:58Z",
			"assigned_by": "Daniele Croce",
		},
		"mentoring": map[string]any{
			"done_at": "2026-05-06T09:48:41Z",
			"done_by": "Daniele Croce",
			"notes": []any{
				map[string]any{
					"at":   "2026-05-06T09:48:30Z",
					"by":   "Daniele Croce",
					"text": "Call completate",
				},
			},
		},
		"group_approved_at": "2026-05-06T09:48:52Z",
		"group_approved_by": "Daniele Croce",
		"admin_approved_at": "2026-05-06T13:52:08Z",
		"admin_approved_by": "Ruvido",
	}

	snapshot := promotedRequestSnapshot(nil, record, data)
	if snapshot["id"] != "request123" {
		t.Fatalf("expected request id in snapshot, got %v", snapshot["id"])
	}

	groupAssignment := requireMap(t, snapshot, "group_assignment")
	if groupAssignment["at"] != "2026-04-17T20:32:59Z" || groupAssignment["by"] != "Ruvido" {
		t.Fatalf("unexpected group assignment event: %#v", groupAssignment)
	}
	group := requireMap(t, groupAssignment, "group")
	if group["id"] != "group123" || group["name"] != "Il Branco del Sud" {
		t.Fatalf("unexpected group snapshot: %#v", group)
	}

	guardianAssignment := requireMap(t, snapshot, "guardian_assignment")
	if guardianAssignment["at"] != "2026-04-19T14:35:58Z" || guardianAssignment["by"] != "Daniele Croce" {
		t.Fatalf("unexpected guardian assignment event: %#v", guardianAssignment)
	}
	guardian := requireMap(t, guardianAssignment, "guardian")
	if guardian["id"] != "guardian123" || guardian["name"] != "Daniele Croce" {
		t.Fatalf("unexpected guardian snapshot: %#v", guardian)
	}

	mentoring := requireMap(t, snapshot, "mentoring")
	if mentoring["at"] != "2026-05-06T09:48:41Z" || mentoring["by"] != "Daniele Croce" {
		t.Fatalf("unexpected mentoring event: %#v", mentoring)
	}
	if _, ok := mentoring["notes"]; !ok {
		t.Fatalf("expected mentoring notes in snapshot")
	}

	groupApproval := requireMap(t, snapshot, "group_approval")
	if groupApproval["at"] != "2026-05-06T09:48:52Z" || groupApproval["by"] != "Daniele Croce" {
		t.Fatalf("unexpected group approval event: %#v", groupApproval)
	}
	adminApproval := requireMap(t, snapshot, "admin_approval")
	if adminApproval["at"] != "2026-05-06T13:52:08Z" || adminApproval["by"] != "Ruvido" {
		t.Fatalf("unexpected admin approval event: %#v", adminApproval)
	}
}

func requireMap(t *testing.T, data map[string]any, key string) map[string]any {
	t.Helper()
	value, ok := data[key].(map[string]any)
	if !ok || value == nil {
		t.Fatalf("expected %s map, got %#v", key, data[key])
	}
	return value
}
