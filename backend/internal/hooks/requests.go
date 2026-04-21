package hooks

import (
	backendinternal "members/backend/internal"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

// RegisterRequestsNormalization keeps request profile data normalized on every save.
func RegisterRequestsNormalization(app *pocketbase.PocketBase) {
	app.OnRecordValidate("requests").BindFunc(func(e *core.RecordEvent) error {
		if err := normalizeRequestProfileData(e.Record); err != nil {
			return err
		}
		return e.Next()
	})
}

func normalizeRequestProfileData(record *core.Record) error {
	if record == nil {
		return nil
	}

	data := backendinternal.ParseJSONMap(record.Get("data"))
	rawMobile, hasMobile := data["mobile"]
	if !hasMobile {
		record.Set("data", data)
		return nil
	}

	mobile := backendinternal.AnyToString(rawMobile)
	if mobile == "" {
		record.Set("data", data)
		return nil
	}

	normalized, err := backendinternal.NormalizePhone(mobile)
	if err != nil {
		return apis.NewBadRequestError("invalid_mobile", err)
	}

	data["mobile"] = normalized
	record.Set("data", data)
	return nil
}

