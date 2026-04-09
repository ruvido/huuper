package requests

var adminApprovedTemplate = EmailTemplate{
	Kind:    "requests.admin_approved",
	Subject: "Your request has been approved",
	Body: emailLayout(
		"Your request has been approved",
		titleFor("{{full_name}}"),
		"Your request has been approved. We will contact you soon with the next steps.",
		nil,
	),
}
