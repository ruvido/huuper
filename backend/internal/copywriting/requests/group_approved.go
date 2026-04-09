package requests

var groupApprovedTemplate = EmailTemplate{
	Kind:    "requests.group_approved",
	Subject: "Final review: {{full_name}}",
	Body: emailLayout(
		"Request ready for final review",
		titleFor("{{full_name}}"),
		`This request for <strong>{{group_name}}</strong> has been approved at group level by <strong>{{actor_name}}</strong> and is ready for final review.`,
		nil,
	),
}
