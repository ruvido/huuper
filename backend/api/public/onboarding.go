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

var allowedAvatarExts = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".webp": true,
	".gif":  true,
	".heic": true,
	".heif": true,
}

var allowedAvatarMimes = map[string]bool{
	"image/jpeg": true,
	"image/jpg":  true,
	"image/png":  true,
	"image/webp": true,
	"image/gif":  true,
	"image/heic": true,
	"image/heif": true,
}

func validateAvatarUpload(fh *multipart.FileHeader) error {
	if fh == nil {
		return apis.NewBadRequestError("invalid_avatar", nil)
	}
	ext := strings.ToLower(filepath.Ext(fh.Filename))
	if !allowedAvatarExts[ext] {
		return apis.NewBadRequestError("invalid_avatar_type", nil)
	}
	contentType := strings.ToLower(strings.TrimSpace(fh.Header.Get("Content-Type")))
	if contentType != "" && !allowedAvatarMimes[contentType] {
		return apis.NewBadRequestError("invalid_avatar_type", nil)
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

func OnboardingCompleteHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		token := strings.TrimSpace(e.Request.PathValue("token"))
		if token == "" {
			return apis.NewBadRequestError("missing_token", nil)
		}

		_, user, err := backendrequests.OnboardingUserForToken(app, token)
		if err != nil {
			return err
		}

		userData := backendinternal.ParseJSONMap(user.Get("data"))

		var payload struct {
			Password        string `json:"password"`
			PasswordConfirm string `json:"password_confirm"`
		}
		if err := e.BindBody(&payload); err != nil {
			return apis.NewBadRequestError("invalid_payload", err)
		}

		password := strings.TrimSpace(payload.Password)
		passwordConfirm := strings.TrimSpace(payload.PasswordConfirm)
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

		user.Set("password", password)
		user.Set("passwordConfirm", passwordConfirm)
		userData["onboarding_password_set_at"] = time.Now().UTC().Format(time.RFC3339)
		user.Set("data", userData)
		if err := app.Save(user); err != nil {
			app.Logger().Error("[onboarding.complete] save failed", "user_id", user.Id, "error", err)
			return apis.NewBadRequestError("failed_to_set_password", err)
		}

		return e.JSON(http.StatusOK, map[string]any{
			"email":   strings.TrimSpace(user.GetString("email")),
			"user_id": user.Id,
			"success": true,
		})
	}
}

func OnboardingProfileHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		token := strings.TrimSpace(e.Request.PathValue("token"))
		if token == "" {
			return apis.NewBadRequestError("missing_token", nil)
		}

		_, user, err := backendrequests.OnboardingUserForToken(app, token)
		if err != nil {
			return err
		}

		data := map[string]any{}
		rawData := strings.TrimSpace(e.Request.FormValue("data"))
		if rawData != "" {
			if err := json.Unmarshal([]byte(rawData), &data); err != nil {
				return apis.NewBadRequestError("invalid_payload", err)
			}
		}

		if _, fh, err := e.Request.FormFile("avatar"); err == nil && fh != nil {
			if err := validateAvatarUpload(fh); err != nil {
				return err
			}
			file, err := filesystem.NewFileFromMultipart(fh)
			if err != nil {
				return apis.NewBadRequestError("invalid_avatar", err)
			}
			user.Set("avatar", file)
		}

		existingData := backendinternal.ParseJSONMap(user.Get("data"))
		user.Set("data", backendrequests.MergeUserData(existingData, data))
		if err := app.Save(user); err != nil {
			app.Logger().Error("[onboarding.profile] save failed", "user_id", user.Id, "error", err)
			return apis.NewBadRequestError("failed_to_save_profile", err)
		}

		return e.JSON(http.StatusOK, map[string]any{
			"email":   strings.TrimSpace(user.GetString("email")),
			"user_id": user.Id,
			"avatar":  strings.TrimSpace(user.GetString("avatar")),
			"success": true,
		})
	}
}

func OnboardingFinalizeHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		token := strings.TrimSpace(e.Request.PathValue("token"))
		if token == "" {
			return apis.NewBadRequestError("missing_token", nil)
		}

		_, user, err := backendrequests.OnboardingUserForToken(app, token)
		if err != nil {
			return err
		}

		app.Logger().Info(
			"[onboarding.finalize] start",
			"token", token,
			"user_id", user.Id,
			"content_type", e.Request.Header.Get("Content-Type"),
			"content_length", e.Request.ContentLength,
		)

		data := map[string]any{}
		rawData := strings.TrimSpace(e.Request.FormValue("data"))
		if rawData != "" {
			if err := json.Unmarshal([]byte(rawData), &data); err != nil {
				return apis.NewBadRequestError("invalid_payload", err)
			}
		}
		app.Logger().Info(
			"[onboarding.finalize] parsed payload",
			"token", token,
			"user_id", user.Id,
			"raw_data_bytes", len(rawData),
			"data_keys", len(data),
		)

		avatarPresent := false
		avatarName := ""
		if _, fh, err := e.Request.FormFile("avatar"); err == nil && fh != nil {
			avatarPresent = true
			avatarName = fh.Filename
			if err := validateAvatarUpload(fh); err != nil {
				app.Logger().Warn(
					"[onboarding.finalize] rejected avatar upload",
					"token", token,
					"user_id", user.Id,
					"filename", fh.Filename,
					"content_type", fh.Header.Get("Content-Type"),
				)
				return err
			}
			file, err := filesystem.NewFileFromMultipart(fh)
			if err != nil {
				app.Logger().Error(
					"[onboarding.finalize] invalid avatar upload",
					"token", token,
					"user_id", user.Id,
					"filename", fh.Filename,
					"error", err,
				)
				return apis.NewBadRequestError("invalid_avatar", err)
			}
			user.Set("avatar", file)
		} else {
			app.Logger().Warn(
				"[onboarding.finalize] no avatar file received",
				"token", token,
				"user_id", user.Id,
			)
		}

		existingData := backendinternal.ParseJSONMap(user.Get("data"))
		mergedData := backendrequests.MergeUserData(existingData, data)
		mergedData["onboarding_completed_at"] = time.Now().UTC().Format(time.RFC3339)
		user.Set("data", mergedData)
		if err := app.Save(user); err != nil {
			app.Logger().Error(
				"[onboarding.finalize] save failed",
				"token", token,
				"user_id", user.Id,
				"avatar_present", avatarPresent,
				"avatar_name", avatarName,
				"error", err,
			)
			return apis.NewBadRequestError("failed_to_save_profile", err)
		}

		app.Logger().Info(
			"[onboarding.finalize] saved user",
			"token", token,
			"user_id", user.Id,
			"avatar_present", avatarPresent,
			"avatar_name", avatarName,
			"avatar_stored", strings.TrimSpace(user.GetString("avatar")),
		)

		if err := backendrequests.DeleteOnboardingToken(app, token); err != nil {
			app.Logger().Error(
				"[onboarding.finalize] token delete failed",
				"token", token,
				"user_id", user.Id,
				"error", err,
			)
			return apis.NewBadRequestError("failed_to_finalize_onboarding", err)
		}

		app.Logger().Info(
			"[onboarding.finalize] token deleted",
			"token", token,
			"user_id", user.Id,
		)

		return e.JSON(http.StatusOK, map[string]any{
			"success": true,
			"avatar":  strings.TrimSpace(user.GetString("avatar")),
		})
	}
}
