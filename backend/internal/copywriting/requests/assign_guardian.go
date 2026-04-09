package requests

var assignGuardianTemplate = EmailTemplate{
	Kind:    "requests.assign_guardian",
	Subject: "Mentoring request: {{full_name}}",
	Body: emailLayout(
		"New mentoring request",
		titleFor("{{full_name}}"),
		`You have been assigned as guardian of <strong>{{group_name}}</strong>. Continue with mentoring.`,
		[][2]string{
			row("Region", "{{region}}"),
			row("Age", "{{age}}"),
			row("Marital status", "{{marital_status}}"),
			row("Children", "{{children}}"),
			row("Motivation", "{{motivation}}"),
		},
	),
}
