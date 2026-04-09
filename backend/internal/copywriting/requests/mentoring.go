package requests

var mentoringTemplate = EmailTemplate{
	Kind:    "requests.mentoring",
	Subject: "Mentoring completed: {{full_name}}",
	Body: emailLayout(
		"Mentoring completed",
		titleFor("{{full_name}}"),
		"Mentoring is complete for this request. Review it for group approval.",
		[][2]string{
			row("Mentoring notes", "{{mentoring_notes}}"),
		},
	),
}
