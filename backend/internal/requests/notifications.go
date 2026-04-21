package requests

import (
	"net/mail"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"members/backend/bot"
	backendinternal "members/backend/internal"
	copywritingrequests "members/backend/internal/copywriting/requests"
	groupinternal "members/backend/internal/groups"
	backendsettings "members/backend/internal/settings"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

const (
	templateKindNewRequest       = "requests.new_request"
	templateKindRequestSubmitted = "requests.request_submitted"
	templateKindAssignGroup      = "requests.assign_group"
	templateKindAssignGuardian   = "requests.assign_guardian"
	templateKindMentoring        = "requests.mentoring"
	templateKindGroupApproved    = "requests.group_approved"
	templateKindAdminApproved    = "requests.admin_approved"
	templateKindUnknownTemplate  = ""
)

type NotificationTemplate struct {
	Email struct {
		Subject string `json:"subject"`
		Body    string `json:"body"`
	} `json:"email"`
	Telegram struct {
		Body string `json:"body"`
	} `json:"telegram"`
}

func NotifyNewRequest(app *pocketbase.PocketBase, record *core.Record, data map[string]any) {
	template, found, err := loadNotificationTemplate(app, templateKindNewRequest)
	if err != nil {
		app.Logger().Warn("Failed to load request template", "kind", templateKindNewRequest, "error", err)
		return
	}
	if !found {
		return
	}

	recipient, ok := adminNotificationRecipient(app)
	if !ok {
		app.Logger().Warn("Missing request admin email", "kind", templateKindNewRequest)
		return
	}

	values := requestNotificationValues(app, record, data, nil, "", "", nil, nil, nil)
	_, _ = sendNotificationEmail(app, []mail.Address{recipient}, template, values)
}

func NotifyRequestSubmitted(app *pocketbase.PocketBase, record *core.Record, data map[string]any) {
	if record == nil {
		return
	}

	template, found, err := loadNotificationTemplate(app, templateKindRequestSubmitted)
	if err != nil {
		app.Logger().Warn("Failed to load request template", "kind", templateKindRequestSubmitted, "error", err)
		return
	}
	if !found {
		return
	}

	recipient, ok := backendinternal.ParseAddress(strings.TrimSpace(record.GetString("email")))
	if !ok {
		app.Logger().Warn("Missing request candidate email", "kind", templateKindRequestSubmitted, "request", record.Id)
		return
	}

	values := requestNotificationValues(app, record, data, nil, "", "", nil, nil, nil)
	_, _ = sendNotificationEmail(app, []mail.Address{recipient}, template, values)
}

func NotifyRequestStep(app *pocketbase.PocketBase, actor *core.Record, record *core.Record, data map[string]any, step FlowStep) {
	templateKind := notificationTemplateKind(step.Action)
	if templateKind == templateKindUnknownTemplate {
		return
	}

	template, found, err := loadNotificationTemplate(app, templateKind)
	if err != nil {
		app.Logger().Warn("Failed to load request template", "kind", templateKind, "error", err)
		return
	}
	if !found {
		return
	}

	group, assistant, guardian := requestNotificationRelations(app, record)
	values := requestNotificationValues(app, record, data, group, step.Action, step.Label, actor, assistant, guardian)

	if step.EmailTo != "" {
		recipient, ok := requestEmailRecipient(app, record, group, assistant, guardian, step.EmailTo)
		if !ok {
			app.Logger().Warn("Missing request notification recipient", "kind", templateKind, "email_to", step.EmailTo, "request", record.Id)
		} else {
			_, _ = sendNotificationEmail(app, []mail.Address{recipient}, template, values)
		}
	}

	if step.TelegramMessage {
		sendNotificationTelegram(app, group, template, values, templateKind, record.Id)
	}
}

func notificationTemplateKind(action string) string {
	switch strings.TrimSpace(action) {
	case FlowActionAssignGroup:
		return templateKindAssignGroup
	case FlowActionAssignGuardian:
		return templateKindAssignGuardian
	case FlowActionMentoring:
		return templateKindMentoring
	case FlowActionGroupApproved:
		return templateKindGroupApproved
	case FlowActionAdminApproved:
		return templateKindAdminApproved
	default:
		return templateKindUnknownTemplate
	}
}

func loadNotificationTemplate(app *pocketbase.PocketBase, kind string) (NotificationTemplate, bool, error) {
	if strings.TrimSpace(kind) == "" {
		return NotificationTemplate{}, false, nil
	}

	defaultTemplate, defaultFound, err := defaultNotificationTemplate(kind)
	if err != nil || !defaultFound {
		return NotificationTemplate{}, false, err
	}

	records, err := app.FindRecordsByFilter(
		"templates",
		"kind = {:kind}",
		"",
		1,
		0,
		map[string]any{"kind": kind},
	)
	if err != nil {
		return NotificationTemplate{}, false, err
	}
	if len(records) == 0 || records[0] == nil {
		return defaultTemplate, true, nil
	}
	record := records[0]

	var template NotificationTemplate
	if err := record.UnmarshalJSONField("data", &template); err != nil {
		return NotificationTemplate{}, false, err
	}
	template = mergeNotificationTemplate(defaultTemplate, template)
	return template, true, nil
}

func defaultNotificationTemplate(kind string) (NotificationTemplate, bool, error) {
	copy, ok := copywritingrequests.ByKind(strings.TrimSpace(kind))
	if !ok {
		return NotificationTemplate{}, false, nil
	}

	template := NotificationTemplate{}
	template.Email.Subject = copy.Subject
	template.Email.Body = copy.Body
	template.Telegram.Body = copy.TelegramBody
	return template, true, nil
}

func mergeNotificationTemplate(defaultTemplate NotificationTemplate, template NotificationTemplate) NotificationTemplate {
	if strings.TrimSpace(template.Email.Subject) == "" {
		template.Email.Subject = defaultTemplate.Email.Subject
	}
	if strings.TrimSpace(template.Email.Body) == "" {
		template.Email.Body = defaultTemplate.Email.Body
	}
	if strings.TrimSpace(template.Telegram.Body) == "" {
		template.Telegram.Body = defaultTemplate.Telegram.Body
	}
	return template
}

func sendNotificationEmail(app *pocketbase.PocketBase, recipients []mail.Address, template NotificationTemplate, values map[string]string) (int, int) {
	subject := strings.TrimSpace(backendinternal.RenderTemplate(template.Email.Subject, values))
	body := strings.TrimSpace(backendinternal.RenderTemplate(template.Email.Body, values))
	if subject == "" || body == "" {
		app.Logger().Warn("Skipping request email notification with empty content", "subject", subject)
		return 0, len(recipients)
	}
	sent, failed := backendinternal.SendPlainEmailToRecipients(app, recipients, subject, body, backendinternal.EmailSenderKeyGeneral)
	if failed > 0 {
		app.Logger().Warn("Failed to send request email notification", "subject", subject, "sent", sent, "failed", failed)
	}
	return sent, failed
}

func sendNotificationTelegram(app *pocketbase.PocketBase, group *core.Record, template NotificationTemplate, values map[string]string, kind string, requestID string) {
	if group == nil {
		app.Logger().Warn("Missing request group for telegram notification", "kind", kind, "request", requestID)
		return
	}

	chatID, ok := requestTelegramChatID(group)
	if !ok {
		app.Logger().Warn("Missing telegram chat id for request group", "kind", kind, "request", requestID, "group", group.Id)
		return
	}

	body := strings.TrimSpace(backendinternal.RenderTemplate(template.Telegram.Body, values))
	if body == "" {
		return
	}
	if err := bot.SendMessage(chatID, body); err != nil {
		app.Logger().Warn("Failed to send request telegram notification", "kind", kind, "request", requestID, "error", err)
	}
}

func requestNotificationValues(app *pocketbase.PocketBase, record *core.Record, data map[string]any, group *core.Record, action string, actionLabel string, actor *core.Record, assistant *core.Record, guardian *core.Record) map[string]string {
	values := map[string]string{
		"request_id":      "",
		"full_name":       "",
		"name":            "",
		"email":           "",
		"mobile":          "",
		"region":          "",
		"age":             "",
		"marital_status":  "",
		"children":        "",
		"motivation":      "",
		"mentoring_notes": "",
		"action":          strings.TrimSpace(action),
		"action_label":    strings.TrimSpace(actionLabel),
		"group_id":        "",
		"group_name":      "",
		"actor_name":      "",
		"actor_email":     "",
		"assistant_name":  "",
		"assistant_email": "",
		"guardian_name":   "",
		"guardian_email":  "",
		"request_url":     "",
		"onboarding_url":  "",
		"data":            renderNotificationData(BuildUserData(data)),
	}

	if record != nil {
		values["request_id"] = record.Id
		values["email"] = strings.TrimSpace(record.GetString("email"))
		values["full_name"] = strings.TrimSpace(DisplayName(data, record.GetString("email"), record.Id))
		values["name"] = values["full_name"]
		values["request_url"] = "https://branco.realmen.it/me/request/?id=" + url.QueryEscape(record.Id)
	}
	if onboardingURL, ok := data["onboarding_url"].(string); ok {
		values["onboarding_url"] = strings.TrimSpace(onboardingURL)
	}
	if mobile, ok := data["mobile"].(string); ok {
		values["mobile"] = strings.TrimSpace(mobile)
	}
	if region, ok := data["region"].(string); ok {
		values["region"] = strings.TrimSpace(region)
	}
	if motivation, ok := data["motivation"].(string); ok {
		values["motivation"] = strings.TrimSpace(motivation)
	}
	if maritalStatus, ok := data["marital_status"].(string); ok {
		values["marital_status"] = strings.TrimSpace(maritalStatus)
	}
	if children, ok := data["children"].(string); ok {
		values["children"] = strings.TrimSpace(children)
	}
	values["mentoring_notes"] = MentoringNotesJoined(data)
	values["age"] = requestAge(data)
	if group != nil {
		values["group_id"] = group.Id
		values["group_name"] = strings.TrimSpace(group.GetString("name"))
	}
	if actor != nil {
		values["actor_name"] = groupinternal.UserDisplayName(actor)
		values["actor_email"] = strings.TrimSpace(actor.GetString("email"))
	}
	if assistant != nil {
		values["assistant_name"] = groupinternal.UserDisplayName(assistant)
		values["assistant_email"] = strings.TrimSpace(assistant.GetString("email"))
	}
	if guardian != nil {
		values["guardian_name"] = groupinternal.UserDisplayName(guardian)
		values["guardian_email"] = strings.TrimSpace(guardian.GetString("email"))
	}
	return values
}

func requestAge(data map[string]any) string {
	rawBirthYear, ok := data["birth_year"].(string)
	if !ok {
		return ""
	}
	birthYear, err := strconv.Atoi(strings.TrimSpace(rawBirthYear))
	if err != nil {
		return ""
	}
	currentYear := time.Now().UTC().Year()
	if birthYear < 1900 || birthYear > currentYear {
		return ""
	}
	return strconv.Itoa(currentYear - birthYear)
}

func renderNotificationData(data map[string]any) string {
	if len(data) == 0 {
		return ""
	}

	lines := make([]string, 0, len(data))
	for _, key := range sortedNotificationKeys(data) {
		lines = append(lines, "- "+humanizeNotificationKey(key)+": "+strings.TrimSpace(backendinternal.AnyToString(data[key])))
	}
	return strings.Join(lines, "\n")
}

func sortedNotificationKeys(data map[string]any) []string {
	keys := make([]string, 0, len(data))
	for key := range data {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func humanizeNotificationKey(key string) string {
	parts := strings.Fields(strings.ReplaceAll(strings.TrimSpace(key), "_", " "))
	for i, part := range parts {
		if part == "" {
			continue
		}
		parts[i] = strings.ToUpper(part[:1]) + part[1:]
	}
	return strings.Join(parts, " ")
}

func requestNotificationRelations(app *pocketbase.PocketBase, record *core.Record) (*core.Record, *core.Record, *core.Record) {
	group, _ := requestGroup(app, record)
	var assistant *core.Record
	if group != nil {
		assistant = requestGroupAssistant(app, group)
	}
	guardian := requestGuardian(app, record)
	return group, assistant, guardian
}

func requestGroup(app *pocketbase.PocketBase, record *core.Record) (*core.Record, bool) {
	if record == nil {
		return nil, false
	}
	groupID := strings.TrimSpace(record.GetString("group"))
	if groupID == "" {
		return nil, false
	}
	group, err := app.FindRecordById("groups", groupID)
	if err != nil || group == nil {
		return nil, false
	}
	return group, true
}

func requestGroupAssistant(app *pocketbase.PocketBase, group *core.Record) *core.Record {
	if group == nil {
		return nil
	}
	assistantID := strings.TrimSpace(group.GetString("assistant"))
	if assistantID == "" {
		return nil
	}
	assistant, err := app.FindRecordById("users", assistantID)
	if err != nil || assistant == nil {
		return nil
	}
	return assistant
}

func requestGuardian(app *pocketbase.PocketBase, record *core.Record) *core.Record {
	if record == nil {
		return nil
	}
	guardianID := strings.TrimSpace(record.GetString("guardian"))
	if guardianID == "" {
		return nil
	}
	guardian, err := app.FindRecordById("users", guardianID)
	if err != nil || guardian == nil {
		return nil
	}
	return guardian
}

func requestEmailRecipient(app *pocketbase.PocketBase, record *core.Record, group *core.Record, assistant *core.Record, guardian *core.Record, target string) (mail.Address, bool) {
	switch strings.TrimSpace(target) {
	case EmailToAdmin:
		return adminNotificationRecipient(app)
	case EmailToAssistant:
		if assistant == nil {
			if group == nil {
				return mail.Address{}, false
			}
			assistant = requestGroupAssistant(app, group)
		}
		if assistant == nil {
			return mail.Address{}, false
		}
		return backendinternal.ParseAddress(strings.TrimSpace(assistant.GetString("email")))
	case EmailToGuardian:
		if guardian == nil {
			guardian = requestGuardian(app, record)
		}
		if guardian == nil {
			return mail.Address{}, false
		}
		return backendinternal.ParseAddress(strings.TrimSpace(guardian.GetString("email")))
	case EmailToCandidate:
		if record == nil {
			return mail.Address{}, false
		}
		return backendinternal.ParseAddress(strings.TrimSpace(record.GetString("email")))
	default:
		return mail.Address{}, false
	}
}

func adminNotificationRecipient(app *pocketbase.PocketBase) (mail.Address, bool) {
	settingsData, err := backendsettings.FindSettingData(app, "email")
	if err != nil || settingsData == nil {
		return mail.Address{}, false
	}
	adminRaw, _ := settingsData[EmailToAdmin].(string)
	return backendinternal.ParseAddress(adminRaw)
}

func requestTelegramChatID(group *core.Record) (int64, bool) {
	if group == nil {
		return 0, false
	}
	telegramData := backendinternal.ParseJSONMap(group.Get("telegram"))
	chatIDRaw := strings.TrimSpace(backendinternal.AnyToString(telegramData["chat_id"]))
	if chatIDRaw == "" {
		return 0, false
	}
	chatID, ok := backendinternal.AnyToInt64(chatIDRaw)
	if !ok || chatID == 0 {
		return 0, false
	}
	return chatID, true
}
