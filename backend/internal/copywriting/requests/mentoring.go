package requests

var mentoringTemplate = EmailTemplate{
	Kind:    "requests.mentoring",
	Subject: "Mentoring completed: {{full_name}}",
	Body: emailLayout(
		"Mentoring completed",
		titleFor("{{full_name}}"),
		`Mentoring for <strong>{{group_name}}</strong> is complete. Review it for group approval.`,
		[][2]string{
			row("Mentoring notes", "{{mentoring_notes}}"),
		},
	),
}
