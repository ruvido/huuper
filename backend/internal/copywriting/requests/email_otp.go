package requests

var emailOTPTemplate = EmailTemplate{
	Kind:    "requests.email_otp",
	Subject: "Your verification code",
	Body: emailLayoutNoCTA(
		"Your verification code",
		"Verify your email",
		`<span style="display:block;text-align:center;margin:8px 0 4px 0;font-size:15px;line-height:22px;color:#374151;">Use this 4-digit code to complete your request</span><strong style="display:block;text-align:center;margin:10px 0 0 0;font-size:34px;line-height:40px;letter-spacing:8px;color:#111827;">{{otp_code}}</strong>`,
		nil,
	),
}
