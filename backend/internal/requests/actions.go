package requests

import (
	"database/sql"
	"errors"
	"log"
	"strings"
	"time"

	backendinternal "members/backend/internal"
	tginternal "members/backend/internal/telegram"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

type PromoteResult struct {
	UserID  string
	Created bool
}

func expectedCurrentStep(app *pocketbase.PocketBase, actor *core.Record, record *core.Record, data map[string]any, expectedAction string) (FlowConfig, FlowStep, error) {
	if record.GetBool("rejected") {
		return FlowConfig{}, FlowStep{}, apis.NewBadRequestError("request_rejected", nil)
	}

	flow, err := LoadFlowForRequest(app, data)
	if err != nil {
		return FlowConfig{}, FlowStep{}, apis.NewBadRequestError("invalid_requests_flow_settings", err)
	}

	stepIndex := EffectiveStepIndex(record, data, flow)
	nextStep, hasNext := FlowStepAt(flow, stepIndex)
	if !hasNext {
		return FlowConfig{}, FlowStep{}, apis.NewBadRequestError("last_step_reached", nil)
	}
	if nextStep.Action != expectedAction {
		return FlowConfig{}, FlowStep{}, apis.NewBadRequestError("invalid_current_step", nil)
	}

	ok, err := backendinternal.HasRoleForRequest(app, actor, record, nextStep.Role, RoleAdmin, RoleGuardian, RoleAssistant)
	if err != nil {
		return FlowConfig{}, FlowStep{}, apis.NewBadRequestError("role_resolution_failed", err)
	}
	if !ok {
		return FlowConfig{}, FlowStep{}, apis.NewForbiddenError("forbidden_transition", nil)
	}

	return flow, nextStep, nil
}

func applyExpectedStepAction(app *pocketbase.PocketBase, actor *core.Record, record *core.Record, data map[string]any, payload ActionPayload, expectedAction string) error {
	flow, nextStep, err := expectedCurrentStep(app, actor, record, data, expectedAction)
	if err != nil {
		return err
	}

	if err := ApplyStepAction(app, actor, record, data, payload, nextStep); err != nil {
		return err
	}

	data = SetRequestFlowSnapshot(data, flow)
	record.Set("data", data)
	return nil
}

func ApplySetMentoringAction(app *pocketbase.PocketBase, actor *core.Record, record *core.Record, data map[string]any, payload ActionPayload) error {
	return applyExpectedStepAction(app, actor, record, data, payload, FlowActionMentoring)
}

func ApplySetGroupApprovedAction(app *pocketbase.PocketBase, actor *core.Record, record *core.Record, data map[string]any, payload ActionPayload) error {
	return applyExpectedStepAction(app, actor, record, data, payload, FlowActionGroupApproved)
}

func ApplySetAdminApprovedAction(app *pocketbase.PocketBase, actor *core.Record, record *core.Record, data map[string]any, payload ActionPayload) error {
	return applyExpectedStepAction(app, actor, record, data, payload, FlowActionAdminApproved)
}

func ApplyRejectAction(actor *core.Record, record *core.Record, data map[string]any, reason string) error {
	if reason == "" {
		return apis.NewBadRequestError("missing_reject_reason", nil)
	}

	if !actor.GetBool("admin") {
		return apis.NewForbiddenError("forbidden_reject", nil)
	}

	data["rejected"] = map[string]any{
		"reason":      reason,
		"rejected_at": time.Now().UTC().Format(time.RFC3339),
		"rejected_by": actorDisplayName(actor),
	}
	record.Set("rejected", true)
	record.Set("data", data)
	return nil
}

func ApplySetGuardianAction(app *pocketbase.PocketBase, actor *core.Record, record *core.Record, data map[string]any, guardianID string) error {
	flow, step, err := expectedCurrentStep(app, actor, record, data, FlowActionAssignGuardian)
	if err != nil {
		return err
	}

	if guardianID == "" {
		record.Set("guardian", "")
		delete(data, "guardian")
		ResetStepsAfterAction(flow, record, data, FlowActionAssignGuardian)
		data = SetRequestFlowSnapshot(data, flow)
		record.Set("data", data)
		return nil
	}

	if err := applyGuardianAssignment(app, record, data, actor, guardianID, step.Filter); err != nil {
		return err
	}
	ResetStepsAfterAction(flow, record, data, FlowActionAssignGuardian)
	data = SetRequestFlowSnapshot(data, flow)
	record.Set("data", data)
	return nil
}

func ApplySetGroupAction(app *pocketbase.PocketBase, actor *core.Record, record *core.Record, data map[string]any, groupID string) error {
	flow, step, err := expectedCurrentStep(app, actor, record, data, FlowActionAssignGroup)
	if err != nil {
		return err
	}

	if err := applyGroupAssignment(app, record, data, actor, strings.TrimSpace(groupID), step.Filter); err != nil {
		return err
	}

	ResetStepsAfterAction(flow, record, data, FlowActionAssignGroup)
	data = SetRequestFlowSnapshot(data, flow)
	record.Set("data", data)
	return nil
}

func ApplyPromoteAction(app *pocketbase.PocketBase, actor *core.Record, record *core.Record, data map[string]any) (PromoteResult, error) {
	if !actor.GetBool("admin") {
		return PromoteResult{}, apis.NewForbiddenError("forbidden_promote", nil)
	}
	if record.GetBool("rejected") {
		return PromoteResult{}, apis.NewBadRequestError("request_rejected", nil)
	}

	flow, err := LoadFlowForRequest(app, data)
	if err != nil {
		return PromoteResult{}, apis.NewBadRequestError("invalid_requests_flow_settings", err)
	}
	stepIndex := EffectiveStepIndex(record, data, flow)
	if stepIndex < len(flow.Steps) {
		nextStep, hasNext := FlowStepAt(flow, stepIndex)
		if !hasNext || nextStep.Action != FlowActionAdminApproved {
			return PromoteResult{}, apis.NewBadRequestError("invalid_promote_status", nil)
		}
		ok, err := backendinternal.HasRoleForRequest(app, actor, record, nextStep.Role, RoleAdmin, RoleGuardian, RoleAssistant)
		if err != nil {
			return PromoteResult{}, apis.NewBadRequestError("role_resolution_failed", err)
		}
		if !ok {
			return PromoteResult{}, apis.NewForbiddenError("forbidden_promote", nil)
		}
	}

	email, err := backendinternal.NormalizeEmail(record.GetString("email"))
	if err != nil {
		return PromoteResult{}, apis.NewBadRequestError("invalid_email", nil)
	}

	existing, err := app.FindFirstRecordByFilter(
		"users",
		"email = {:email}",
		map[string]any{"email": email},
	)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return PromoteResult{}, apis.NewBadRequestError("failed_user_lookup", err)
	}
	if existing != nil {
		// Idempotent promote: if the user already exists for this email,
		// treat the promote action as already completed.
		return PromoteResult{UserID: existing.Id, Created: false}, nil
	}

	users, err := app.FindCollectionByNameOrId("users")
	if err != nil {
		return PromoteResult{}, apis.NewNotFoundError("users_collection_not_found", err)
	}

	user := core.NewRecord(users)
	tempPassword := backendinternal.RandomToken()
	if tempPassword == "" {
		return PromoteResult{}, apis.NewBadRequestError("failed_to_generate_password", nil)
	}
	user.Set("email", email)
	user.Set("password", tempPassword)
	user.Set("passwordConfirm", tempPassword)
	user.Set("status", "approved")

	userData := BuildUserData(data)
	if len(userData) > 0 {
		user.Set("data", userData)
	}

	if err := app.Save(user); err != nil {
		return PromoteResult{}, apis.NewBadRequestError("failed_to_create_user", err)
	}

	targetGroupIDs, err := promoteTargetGroupIDs(app, record)
	if err != nil {
		_ = app.Delete(user)
		return PromoteResult{}, err
	}

	createdInviteGroupIDs := generatePromoteInvites(
		user.Id,
		targetGroupIDs,
		func(id string) (*core.Record, error) { return app.FindRecordById("groups", id) },
		func(group *core.Record) bool {
			_, err := tginternal.TelegramChatIDForGroup(group)
			return err == nil
		},
		func(group *core.Record) error {
			_, err := tginternal.GenerateGroupInvite(app, user, group)
			return err
		},
	)

	onboardingToken, err := GenerateOnboardingToken(app, user.Id)
	if err != nil {
		rollbackPromotedUser(app, user.Id, createdInviteGroupIDs)
		return PromoteResult{}, apis.NewBadRequestError("failed_to_create_onboarding_token", err)
	}
	onboardingURL := BuildOnboardingURL(app, onboardingToken)
	if strings.TrimSpace(onboardingURL) == "" {
		rollbackPromotedUser(app, user.Id, createdInviteGroupIDs)
		return PromoteResult{}, apis.NewBadRequestError("invalid_onboarding_url", nil)
	}
	data["onboarding_url"] = onboardingURL

	return PromoteResult{UserID: user.Id, Created: true}, nil
}

// generatePromoteInvites tries to generate a telegram invite link for every
// target group. Invite generation is best-effort: if a group cannot be
// resolved, has no telegram chat_id, or the telegram API returns an error,
// the loop logs the incident and moves on. This prevents a misconfigured
// chat from blocking the whole admin promote flow.
func generatePromoteInvites(
	userID string,
	targetGroupIDs []string,
	loadGroup func(id string) (*core.Record, error),
	hasChatID func(group *core.Record) bool,
	generateInvite func(group *core.Record) error,
) []string {
	created := make([]string, 0, len(targetGroupIDs))
	for _, groupID := range targetGroupIDs {
		group, err := loadGroup(groupID)
		if err != nil || group == nil {
			log.Printf("[promote] user=%s group=%s load failed (non-blocking): %v", userID, groupID, err)
			continue
		}
		if !hasChatID(group) {
			continue
		}
		if err := generateInvite(group); err != nil {
			log.Printf("[promote] user=%s group=%s invite generation failed (non-blocking): %v", userID, groupID, err)
			continue
		}
		created = append(created, groupID)
	}
	return created
}

func promoteTargetGroupIDs(app *pocketbase.PocketBase, record *core.Record) ([]string, error) {
	groupIDs := []string{}

	requestGroupID := strings.TrimSpace(record.GetString("group"))
	if requestGroupID == "" {
		return nil, apis.NewBadRequestError("missing_group_assignment", nil)
	}
	groupIDs = append(groupIDs, requestGroupID)

	defaultGroup, err := app.FindFirstRecordByFilter(
		"groups",
		"type = 'default'",
		map[string]any{},
	)
	if err == nil && defaultGroup != nil && strings.TrimSpace(defaultGroup.Id) != requestGroupID {
		groupIDs = append(groupIDs, defaultGroup.Id)
	}

	return groupIDs, nil
}

func rollbackPromotedUser(app *pocketbase.PocketBase, userID string, inviteGroupIDs []string) {
	if app == nil || strings.TrimSpace(userID) == "" {
		return
	}

	for _, groupID := range inviteGroupIDs {
		group, err := app.FindRecordById("groups", strings.TrimSpace(groupID))
		if err != nil || group == nil {
			_ = tginternal.DeleteInviteToken(app, userID, groupID)
			continue
		}
		chatID, err := tginternal.TelegramChatIDForGroup(group)
		if err == nil {
			link, linkErr := tginternal.InviteLinkForUserGroup(app, userID, groupID)
			if linkErr == nil {
				_ = tginternal.RevokeInviteLink(chatID, link)
			}
		}
		_ = tginternal.DeleteInviteToken(app, userID, groupID)
	}
	deleteOnboardingTokensForUser(app, userID)
	user, err := app.FindRecordById("users", userID)
	if err == nil && user != nil {
		_ = app.Delete(user)
	}
}

func RollbackPromotedUser(app *pocketbase.PocketBase, userID string, record *core.Record) {
	if app == nil || strings.TrimSpace(userID) == "" {
		return
	}

	groupIDs := []string{}
	if record != nil {
		if ids, err := promoteTargetGroupIDs(app, record); err == nil {
			groupIDs = ids
		}
	}
	rollbackPromotedUser(app, userID, groupIDs)
}

func applyGroupAssignment(app *pocketbase.PocketBase, record *core.Record, data map[string]any, actor *core.Record, groupID string, filter string) error {
	if groupID == "" {
		return apis.NewBadRequestError("missing_group", nil)
	}
	group, err := app.FindRecordById("groups", groupID)
	if err != nil || group == nil {
		return apis.NewBadRequestError("invalid_group", err)
	}
	if filter == FilterLocal && strings.TrimSpace(group.GetString("type")) != FilterLocal {
		return apis.NewBadRequestError("invalid_group_filter", nil)
	}
	record.Set("group", groupID)
	if data != nil {
		assigned := map[string]any{
			"assigned_at": time.Now().UTC().Format(time.RFC3339),
		}
		if actor != nil {
			assigned["assigned_by"] = actorDisplayName(actor)
		}
		data["assign_group"] = assigned
	}
	return nil
}

func applyGuardianAssignment(app *pocketbase.PocketBase, record *core.Record, data map[string]any, actor *core.Record, guardianID string, filter string) error {
	groupID := strings.TrimSpace(record.GetString("group"))
	if groupID == "" {
		return apis.NewBadRequestError("missing_group_assignment", nil)
	}
	if guardianID == "" {
		return apis.NewBadRequestError("missing_guardian", nil)
	}

	guardian, err := app.FindRecordById("users", guardianID)
	if err != nil || guardian == nil {
		return apis.NewBadRequestError("invalid_guardian", err)
	}

	if filter == FilterGroupMembers {
		guardianMemberships, err := app.FindRecordsByFilter(
			"user_groups",
			"user = {:user} && group = {:group}",
			"",
			1,
			0,
			map[string]any{
				"user":  guardianID,
				"group": groupID,
			},
		)
		if err != nil {
			return apis.NewBadRequestError("guardian_membership_check_failed", err)
		}
		if len(guardianMemberships) == 0 {
			return apis.NewBadRequestError("guardian_not_in_group", nil)
		}
	}

	record.Set("guardian", guardianID)
	guardianPayload := map[string]any{
		"name":        actorDisplayName(guardian),
		"assigned_at": time.Now().UTC().Format(time.RFC3339),
	}
	if actor != nil {
		guardianPayload["assigned_by"] = actorDisplayName(actor)
	}
	data["guardian"] = guardianPayload
	return nil
}

func actorDisplayName(actor *core.Record) string {
	if actor == nil {
		return ""
	}

	data := backendinternal.ParseJSONMap(actor.Get("data"))
	if fullName, ok := data["full_name"].(string); ok && strings.TrimSpace(fullName) != "" {
		return strings.TrimSpace(fullName)
	}
	if name, ok := data["name"].(string); ok && strings.TrimSpace(name) != "" {
		return strings.TrimSpace(name)
	}

	email := strings.TrimSpace(actor.GetString("email"))
	if email != "" {
		return email
	}

	return actor.Id
}
