package requests

var adminApprovedTemplate = EmailTemplate{
	Kind:    "requests.admin_approved",
	Subject: "Set your password",
	Body: emailLayoutWithCTAURL(
		"Your request has been approved",
		titleFor("{{full_name}}"),
		`Your request for <strong>{{group_name}}</strong> has been approved. Set your password to continue.`,
		nil,
		"Set password",
		"{{onboarding_url}}",
	),
}
