package groups

type MemberItem struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	FullName      string `json:"full_name"`
	Avatar        string `json:"avatar,omitempty"`
	Age           *int   `json:"age,omitempty"`
	Region        string `json:"region,omitempty"`
	IsAssistant   bool   `json:"is_assistant,omitempty"`
	IsGuardian    bool   `json:"is_guardian"`
	ProtegesCount int    `json:"proteges_count"`
}

type GuardianItem struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	FullName      string `json:"full_name"`
	Avatar        string `json:"avatar,omitempty"`
	ProtegesCount int    `json:"proteges_count"`
}

type GuardianGroupItem struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	ProtegesCount int    `json:"proteges_count"`
}

type GroupListItem struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Type          string `json:"type"`
	Description   string `json:"description"`
	Assistant     string `json:"assistant,omitempty"`
	MembersCount  int    `json:"members_count"`
	RequestsCount *int   `json:"requests_count,omitempty"`
}

type PendingRequestItem struct {
	ID          string         `json:"id"`
	FullName    string         `json:"full_name"`
	Email       string         `json:"email"`
	Status      string         `json:"status"`
	StatusLabel string         `json:"status_label"`
	Created     string         `json:"created"`
	AssignedAt  string         `json:"assigned_at,omitempty"`
	Data        map[string]any `json:"data"`
	Workflow    map[string]any `json:"workflow"`
}

type MembersResponse struct {
	GroupID string       `json:"group_id"`
	Items   []MemberItem `json:"items"`
}

type GuardiansResponse struct {
	GroupID string         `json:"group_id"`
	Items   []GuardianItem `json:"items"`
}
