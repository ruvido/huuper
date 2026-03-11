package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		settings, err := app.FindCollectionByNameOrId("settings")
		if err != nil {
			return err
		}

		existing, err := app.FindFirstRecordByFilter(
			"settings",
			"name = 'password_reset'",
			map[string]any{},
		)
		if err == nil && existing != nil {
			return nil
		}

		data := map[string]any{
			"request": map[string]any{
				"title":                     "Reset your password",
				"helper":                    "Enter your email and we will send you a reset link.",
				"email_label":               "Email",
				"submit_label":              "Send reset link",
				"submitting_label":          "Sending...",
				"footer_prompt":             "Remembered your password?",
				"footer_action":             "Back to login",
				"confirmation_title":        "Password reset sent!",
				"confirmation_message":      "Check your email.",
				"confirmation_hint":         "If you do not see it, check spam or promotions.",
				"confirmation_back_to_login": "Back to login",
			},
			"reset": map[string]any{
				"title":               "Set a new password",
				"helper":              "Choose a strong password you do not use elsewhere.",
				"password_label":       "New password",
				"confirm_label":        "Confirm password",
				"submit_label":         "Update password",
				"submitting_label":     "Updating...",
				"success_title":        "Password updated",
				"success_message":      "You can now log in with the new password.",
				"success_back_to_login": "Back to login",
			},
			"errors": map[string]any{
				"required_email":   "Email is required",
				"send_failed":      "Unable to send reset email. Please try again.",
				"load_failed":      "Unable to load configuration.",
				"invalid_token":    "Invalid or missing reset token.",
				"required_passwords": "Password and confirmation are required",
				"reset_failed":     "Unable to reset password. Please try again.",
				"reset_invalid":    "Reset link is invalid or expired.",
			},
		}

		record := core.NewRecord(settings)
		record.Set("name", "password_reset")
		record.Set("data", data)
		return app.Save(record)
	}, func(app core.App) error {
		record, err := app.FindFirstRecordByFilter(
			"settings",
			"name = 'password_reset'",
			map[string]any{},
		)
		if err == nil && record != nil {
			return app.Delete(record)
		}
		return nil
	})
}
