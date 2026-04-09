package requests

var adminApprovedTemplate = EmailTemplate{
	Kind:    "requests.admin_approved",
	Subject: "Your request has been approved",
	Body: emailLayoutWithCTA(
		"Your request has been approved",
		titleFor("{{full_name}}"),
		`Your request for <strong>{{group_name}}</strong> has been approved. Complete your registration by clicking the button below.`,
		nil,
		"Start",
	),
}
