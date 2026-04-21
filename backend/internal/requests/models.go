package requests

type SignupFieldConfig struct {
	Field string `json:"field"`
}

type SignupSettingsConfig struct {
	Steps []SignupFieldConfig `json:"steps"`
}

type ProfileFieldConfig struct {
	Key      string   `json:"key"`
	Type     string   `json:"type"`
	Required bool     `json:"required"`
	Unique   *bool    `json:"unique,omitempty"`
	Options  []string `json:"options"`
	Min      int      `json:"min"`
	Max      int      `json:"max"`
}

type ProfileSchemaConfig struct {
	Fields []ProfileFieldConfig `json:"fields"`
}

type OnboardingPageConfig struct {
	Title  string `json:"title"`
	Text   string `json:"text"`
	Button string `json:"button"`
}

type OnboardingStepConfig struct {
	Field string `json:"field"`
	Title string `json:"title"`
	Label string `json:"label,omitempty"`
	Unique *bool  `json:"unique,omitempty"`
}

type OnboardingSettingsConfig struct {
	StartPage    *OnboardingPageConfig  `json:"start_page,omitempty"`
	Steps        []OnboardingStepConfig `json:"steps"`
	Confirmation *OnboardingPageConfig  `json:"confirmation,omitempty"`
}

type ActionPayload struct {
	Action         string `json:"action"`
	Reason         string `json:"reason"`
	GroupID        string `json:"group"`
	GuardianID     string `json:"guardian"`
	MentoringNotes string `json:"mentoring_notes"`
}

type RequestFlowSnapshot struct {
	Version int        `json:"version"`
	Steps   []FlowStep `json:"steps"`
}

type ListItem struct {
	ID          string         `json:"id"`
	Email       string         `json:"email"`
	Status      string         `json:"status"`
	Rejected    bool           `json:"rejected"`
	GroupID     string         `json:"group"`
	Guardian    string         `json:"guardian"`
	Created     string         `json:"created"`
	Updated     string         `json:"updated"`
	Data        map[string]any `json:"data"`
	FlowVersion int            `json:"flow_version"`
	Workflow    map[string]any `json:"workflow"`
}

func DisplayName(data map[string]any, email string, fallbackID string) string {
	if fullName, ok := data["full_name"].(string); ok && fullName != "" {
		return fullName
	}
	if name, ok := data["name"].(string); ok && name != "" {
		return name
	}
	if email != "" {
		return email
	}
	return fallbackID
}

type GuardianRequestItem struct {
	ID          string `json:"id"`
	FullName    string `json:"full_name"`
	Email       string `json:"email"`
	Status      string `json:"status"`
	StatusLabel string `json:"status_label"`
	GroupID     string `json:"group"`
	GroupName   string `json:"group_name"`
	Created     string `json:"created"`
	AssignedAt  string `json:"assigned_at"`
}

const (
	ActionReject           = "reject"
	ActionPromote          = "promote"
	ActionSetGroup         = "set_group"
	ActionSetGuardian      = "set_guardian"
	ActionAddMentoringNote = "add_mentoring_note"
	ActionSetMentoring     = "set_mentoring_done"
	ActionSetGroupApprove  = "set_group_approved"
	ActionSetAdminApprove  = "set_admin_approved"
)
