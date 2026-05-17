package public

import (
	"encoding/json"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	backendinternal "members/backend/internal"
	backendrequests "members/backend/internal/requests"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/filesystem"
)

var allowedOnboardingFileExts = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".webp": true,
	".gif":  true,
	".heic": true,
	".heif": true,
}

var allowedOnboardingFileMimes = map[string]bool{
	"image/jpeg": true,
	"image/jpg":  true,
	"image/png":  true,
	"image/webp": true,
	"image/gif":  true,
	"image/heic": true,
	"image/heif": true,
}

func validateOnboardingFileUpload(fh *multipart.FileHeader) error {
	if fh == nil {
		return apis.NewBadRequestError("invalid_file", nil)
	}
	ext := strings.ToLower(filepath.Ext(fh.Filename))
	if !allowedOnboardingFileExts[ext] {
		return apis.NewBadRequestError("invalid_file_type", nil)
	}
	contentType := strings.ToLower(strings.TrimSpace(fh.Header.Get("Content-Type")))
	if contentType != "" && !allowedOnboardingFileMimes[contentType] {
		return apis.NewBadRequestError("invalid_file_type", nil)
	}
	return nil
}

func OnboardingGetHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		_, user, err := backendrequests.OnboardingUserForToken(app, e.Request.PathValue("token"))
		if err != nil {
			return err
		}

		passwordMin, err := backendrequests.UserPasswordMin(app)
		if err != nil {
			return err
		}

		onboarding, err := backendrequests.LoadOnboardingSettings(app)
		if err != nil {
			return err
		}

		data := backendinternal.ParseJSONMap(user.Get("data"))
		return e.JSON(http.StatusOK, map[string]any{
			"email":        strings.TrimSpace(user.GetString("email")),
			"full_name":    backendrequests.DisplayName(data, user.GetString("email"), user.Id),
			"user_id":      user.Id,
			"data":         data,
			"password_min": passwordMin,
			"onboarding":   onboarding,
		})
	}
}

// OnboardingFinalizeHandler is the atomic terminal step of onboarding.
// It validates the password and that ALL required profile fields are
// populated, then sets the user password, marks completion, and deletes
// the onboarding token. Until this succeeds the user has no usable
// password and cannot log in (defense-in-depth also enforced by the
// users auth-gate hook).
func OnboardingFinalizeHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		token := strings.TrimSpace(e.Request.PathValue("token"))
		if token == "" {
			return apis.NewBadRequestError("missing_token", nil)
		}

		data := map[string]any{}
		rawData := strings.TrimSpace(e.Request.FormValue("data"))
		if rawData != "" {
			if err := json.Unmarshal([]byte(rawData), &data); err != nil {
				return apis.NewBadRequestError("invalid_payload", err)
			}
		}

		password := strings.TrimSpace(e.Request.FormValue("password"))
		passwordConfirm := strings.TrimSpace(e.Request.FormValue("password_confirm"))
		if password == "" || passwordConfirm == "" {
			return apis.NewBadRequestError("missing_password", nil)
		}
		if password != passwordConfirm {
			return apis.NewBadRequestError("password_mismatch", nil)
		}
		passwordMin, err := backendrequests.UserPasswordMin(app)
		if err != nil {
			return err
		}
		if passwordMin > 0 && len([]rune(password)) < passwordMin {
			return apis.NewBadRequestError(backendrequests.PasswordTooShortMessage(passwordMin), nil)
		}

		onboarding, err := backendrequests.LoadOnboardingSettings(app)
		if err != nil {
			return err
		}

		var finalizedUser *core.Record
		txErr := app.RunInTransaction(func(txApp core.App) error {
			_, txUser, err := backendrequests.OnboardingUserForToken(txApp, token)
			if err != nil {
				return err
			}

			existingData := backendinternal.ParseJSONMap(txUser.Get("data"))
			mergedData := backendrequests.MergeUserData(existingData, data)
			if err := applyOnboardingFileUploads(e, txUser, onboarding); err != nil {
				return err
			}
			if missing := backendrequests.MissingOnboardingFields(mergedData, onboarding, txUser); len(missing) > 0 {
				return apis.NewApiError(http.StatusBadRequest, "missing_onboarding_fields", map[string]any{"missing": missing})
			}

			now := time.Now().UTC().Format(time.RFC3339)
			mergedData["onboarding_password_set_at"] = now
			mergedData["onboarding_completed_at"] = now

			txUser.Set("password", password)
			txUser.Set("passwordConfirm", passwordConfirm)
			txUser.Set("data", mergedData)

			if err := txApp.Save(txUser); err != nil {
				return err
			}
			if err := backendrequests.DeleteOnboardingToken(txApp, token); err != nil {
				return err
			}
			finalizedUser = txUser
			return nil
		})
		if txErr != nil {
			app.Logger().Error("[onboarding.finalize] commit failed", "token", token, "error", txErr)
			return apis.NewBadRequestError("failed_to_finalize_onboarding", txErr)
		}

		app.Logger().Info("[onboarding.finalize] completed", "user_id", finalizedUser.Id, "email", finalizedUser.GetString("email"))

		return e.JSON(http.StatusOK, map[string]any{
			"success": true,
			"avatar":  strings.TrimSpace(finalizedUser.GetString("avatar")),
		})
	}
}

func applyOnboardingFileUploads(e *core.RequestEvent, user *core.Record, onboarding backendrequests.OnboardingSettingsConfig) error {
	if e == nil || user == nil || user.Collection() == nil {
		return nil
	}

	seen := map[string]bool{}
	for _, step := range onboarding.Steps {
		field := strings.TrimSpace(step.Field)
		if field == "" || seen[field] {
			continue
		}
		seen[field] = true

		if _, ok := user.Collection().Fields.GetByName(field).(*core.FileField); !ok {
			continue
		}

		_, fh, err := e.Request.FormFile(field)
		if err != nil || fh == nil {
			continue
		}
		if err := validateOnboardingFileUpload(fh); err != nil {
			return err
		}
		file, err := filesystem.NewFileFromMultipart(fh)
		if err != nil {
			return apis.NewBadRequestError("invalid_file", err)
		}
		user.Set(field, file)
	}

	return nil
}
