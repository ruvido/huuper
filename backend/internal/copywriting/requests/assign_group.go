package requests

var assignGroupTemplate = EmailTemplate{
	Kind:    "requests.assign_group",
	Subject: "New request for {{group_name}}",
	Body: emailLayout(
		"New request for {{group_name}}",
		titleFor("{{full_name}}"),
		"A new request has been assigned to {{group_name}}. Review it and assign a guardian.",
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
