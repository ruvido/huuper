package groups

type MemberItem struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	FullName      string `json:"full_name"`
	IsGuardian    bool   `json:"is_guardian"`
	ProtegesCount int    `json:"proteges_count"`
}

type GuardianItem struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	FullName      string `json:"full_name"`
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
	MembersCount  int    `json:"members_count"`
	RequestsCount *int   `json:"requests_count,omitempty"`
}

type MembersResponse struct {
	GroupID string       `json:"group_id"`
	Items   []MemberItem `json:"items"`
}

type GuardiansResponse struct {
	GroupID string         `json:"group_id"`
	Items   []GuardianItem `json:"items"`
}
