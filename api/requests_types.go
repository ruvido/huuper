package api

type signupFieldConfig struct {
	Field string `json:"field"`
}

type signupSettingsConfig struct {
	Steps []signupFieldConfig `json:"steps"`
}

type profileFieldConfig struct {
	Key string `json:"key"`
}

type profileSchemaConfig struct {
	Fields []profileFieldConfig `json:"fields"`
}

type requestActionPayload struct {
	Action         string `json:"action"`
	Reason         string `json:"reason"`
	GroupID        string `json:"group"`
	GuardianID     string `json:"guardian"`
	MentoringNotes string `json:"mentoring_notes"`
}

type requestListItem struct {
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

const (
	requestActionAdvance     = "advance"
	requestActionReject      = "reject"
	requestActionPromote     = "promote"
	requestActionSetGuardian = "set_guardian"
)
