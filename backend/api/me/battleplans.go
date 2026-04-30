package me

import (
	"net/http"
	"strconv"
	"strings"

	backendinternal "members/backend/internal"
	"members/backend/internal/battleplans"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

func ListBattleplansHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		actor, err := backendinternal.RequireAuthenticatedActor(e)
		if err != nil {
			return err
		}

		page, perPage := paginationFromQuery(e)
		items, err := battleplans.ListForUser(app, actor.Id, perPage, (page-1)*perPage)
		if err != nil {
			return apis.NewBadRequestError("failed_battleplans", err)
		}
		return e.JSON(http.StatusOK, map[string]any{"items": items, "page": page})
	}
}

func GetBattleplanHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		actor, err := backendinternal.RequireAuthenticatedActor(e)
		if err != nil {
			return err
		}
		record, err := loadBattleplan(app, e.Request.PathValue("id"))
		if err != nil {
			return err
		}
		ok, err := battleplans.CanView(app, actor, record)
		if err != nil {
			return apis.NewBadRequestError("failed_battleplan_visibility", err)
		}
		if !ok {
			return apis.NewForbiddenError("forbidden_battleplan", nil)
		}
		return e.JSON(http.StatusOK, battleplans.MapItem(record))
	}
}

func CreateBattleplanHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		actor, err := backendinternal.RequireAuthenticatedActor(e)
		if err != nil {
			return err
		}
		var input battleplans.Input
		if err := e.BindBody(&input); err != nil {
			return apis.NewBadRequestError("invalid_payload", err)
		}
		record, err := battleplans.Create(app, actor.Id, input)
		if err != nil {
			return apis.NewBadRequestError("failed_battleplan_create", err)
		}
		return e.JSON(http.StatusCreated, battleplans.MapItem(record))
	}
}

func UpdateBattleplanHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		actor, err := backendinternal.RequireAuthenticatedActor(e)
		if err != nil {
			return err
		}
		record, err := loadBattleplan(app, e.Request.PathValue("id"))
		if err != nil {
			return err
		}
		if record.GetString("user") != actor.Id {
			return apis.NewForbiddenError("forbidden_battleplan", nil)
		}
		var input battleplans.Input
		if err := e.BindBody(&input); err != nil {
			return apis.NewBadRequestError("invalid_payload", err)
		}
		if err := battleplans.Update(app, record, input); err != nil {
			return apis.NewBadRequestError("failed_battleplan_update", err)
		}
		return e.JSON(http.StatusOK, battleplans.MapItem(record))
	}
}

func BattleplanStatusHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		actor, err := backendinternal.RequireAuthenticatedActor(e)
		if err != nil {
			return err
		}
		record, err := loadBattleplan(app, e.Request.PathValue("id"))
		if err != nil {
			return err
		}
		if record.GetString("user") != actor.Id {
			return apis.NewForbiddenError("forbidden_battleplan", nil)
		}

		var payload struct {
			Status string `json:"status"`
		}
		if err := e.BindBody(&payload); err != nil {
			return apis.NewBadRequestError("invalid_payload", err)
		}
		next := strings.TrimSpace(payload.Status)
		if next != battleplans.StatusCompleted && next != battleplans.StatusArchived {
			return apis.NewBadRequestError("invalid_status_transition", nil)
		}
		if record.GetString("status") != battleplans.StatusActive {
			return apis.NewBadRequestError("battleplan_not_active", nil)
		}
		if err := battleplans.SetStatus(app, record, next); err != nil {
			return apis.NewBadRequestError("failed_battleplan_status", err)
		}
		return e.JSON(http.StatusOK, battleplans.MapItem(record))
	}
}

func loadBattleplan(app *pocketbase.PocketBase, id string) (*core.Record, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, apis.NewBadRequestError("invalid_battleplan", nil)
	}
	record, err := app.FindRecordById("battleplans", id)
	if err != nil || record == nil {
		return nil, apis.NewNotFoundError("battleplan_not_found", err)
	}
	return record, nil
}

func paginationFromQuery(e *core.RequestEvent) (int, int) {
	q := e.Request.URL.Query()
	page := 1
	if raw := strings.TrimSpace(q.Get("page")); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed >= 1 {
			page = parsed
		}
	}
	perPage := 50
	if raw := strings.TrimSpace(q.Get("per_page")); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed >= 1 && parsed <= 200 {
			perPage = parsed
		}
	}
	return page, perPage
}
