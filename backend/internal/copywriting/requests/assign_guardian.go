package requests

var assignGuardianTemplate = EmailTemplate{
	Kind:    "requests.assign_guardian",
	Subject: "New mentoring request",
	Body: emailLayout(
		"New mentoring request",
		titleFor("{{full_name}}"),
		"You have been assigned as guardian for this request. Continue with mentoring.",
		[][2]string{
			row("Group", "{{group_name}}"),
			row("Age", "{{age}}"),
			row("Region", "{{region}}"),
			row("Marital status", "{{marital_status}}"),
			row("Children", "{{children}}"),
			row("Motivation", "{{motivation}}"),
		},
	),
}
