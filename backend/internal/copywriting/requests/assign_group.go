package requests

var assignGroupTemplate = EmailTemplate{
	Kind:    "requests.assign_group",
	Subject: "{{full_name}} for {{group_name}}",
	Body: emailLayout(
		"New request for {{group_name}}",
		titleFor("{{full_name}}"),
		`This request has been assigned to <strong>{{group_name}}</strong>. Review it and assign a guardian.`,
		[][2]string{
			row("Region", "{{region}}"),
			row("Age", "{{age}}"),
			row("Marital status", "{{marital_status}}"),
			row("Children", "{{children}}"),
			row("Motivation", "{{motivation}}"),
		},
	),
}
