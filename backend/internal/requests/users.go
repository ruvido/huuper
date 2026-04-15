package requests

import (
	"fmt"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

func UserPasswordMin(app *pocketbase.PocketBase) (int, error) {
	users, err := app.FindCollectionByNameOrId("users")
	if err != nil {
		return 0, err
	}

	passwordField, ok := users.Fields.GetByName(core.FieldNamePassword).(*core.PasswordField)
	if !ok || passwordField == nil {
		return 0, fmt.Errorf("users password field not found")
	}

	if passwordField.Min < 0 {
		return 0, fmt.Errorf("invalid users password min")
	}

	return passwordField.Min, nil
}

func PasswordTooShortMessage(min int) string {
	if min <= 0 {
		return ""
	}
	if min == 1 {
		return "Password must be at least 1 character."
	}
	return fmt.Sprintf("Password must be at least %d characters.", min)
}
