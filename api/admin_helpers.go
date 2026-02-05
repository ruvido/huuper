package api

import (
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

func requireAdmin(e *core.RequestEvent) (*core.Record, error) {
	authRecord := e.Auth
	if authRecord == nil {
		return nil, apis.NewUnauthorizedError("Unauthorized", nil)
	}
	if !authRecord.GetBool("admin") {
		return nil, apis.NewForbiddenError("Forbidden", nil)
	}
	return authRecord, nil
}
