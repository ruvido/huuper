package requests

var newRequestTemplate = EmailTemplate{
	Kind:    "requests.new_request",
	Subject: "New request from {{full_name}}",
	Body: emailLayout(
		"New request from {{full_name}}",
		titleFor("{{full_name}}"),
		"A new request is ready for review.",
		[][2]string{
			row("Region", "{{region}}"),
			row("Age", "{{age}}"),
			row("Marital status", "{{marital_status}}"),
			row("Children", "{{children}}"),
			row("Motivation", "{{motivation}}"),
		},
	),
}
