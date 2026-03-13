package requests

type SignupFieldConfig struct {
	Field string `json:"field"`
}

type SignupSettingsConfig struct {
	Steps []SignupFieldConfig `json:"steps"`
}

type ProfileFieldConfig struct {
	Key string `json:"key"`
}

type ProfileSchemaConfig struct {
	Fields []ProfileFieldConfig `json:"fields"`
}

type ActionPayload struct {
	Action         string `json:"action"`
	Reason         string `json:"reason"`
	GroupID        string `json:"group"`
	GuardianID     string `json:"guardian"`
	MentoringNotes string `json:"mentoring_notes"`
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
	StepIndex   int            `json:"step_index"`
	Workflow    map[string]any `json:"workflow"`
}

type GuardianRequestItem struct {
	ID         string `json:"id"`
	FullName   string `json:"full_name"`
	Email      string `json:"email"`
	Status     string `json:"status"`
	GroupID    string `json:"group"`
	GroupName  string `json:"group_name"`
	Created    string `json:"created"`
	AssignedAt string `json:"assigned_at"`
}

const (
	ActionAdvance     = "advance"
	ActionReject      = "reject"
	ActionPromote     = "promote"
	ActionSetGuardian = "set_guardian"
)
