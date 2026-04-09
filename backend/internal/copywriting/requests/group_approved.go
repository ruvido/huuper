package requests

var groupApprovedTemplate = EmailTemplate{
	Kind:    "requests.group_approved",
	Subject: "Request ready for final review",
	Body: emailLayout(
		"Request ready for final review",
		titleFor("{{full_name}}"),
		"This request has been approved at group level and is ready for final review.",
		[][2]string{
			row("Group", "{{group_name}}"),
			row("Approved by", "{{actor_name}}"),
		},
	),
}
