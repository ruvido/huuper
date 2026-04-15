package requests

var requestSubmittedTemplate = EmailTemplate{
	Kind:    "requests.request_submitted",
	Subject: "We received your request",
	Body: emailLayoutWithCTAURL(
		"Request received",
		titleFor("{{full_name}}"),
		`Thanks, we received your request. A member of the group will call you as soon as possible.`,
		[][2]string{
			row("Email", "{{email}}"),
			row("Mobile", "{{mobile}}"),
			row("Region", "{{region}}"),
		},
		"Go to website",
		"https://branco.realmen.it",
	),
}
