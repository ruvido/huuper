package me

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	backendinternal "members/backend/internal"
	"members/backend/internal/battleplans"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

func requireBattleplanActor(app *pocketbase.PocketBase, e *core.RequestEvent) (*core.Record, error) {
	actor, err := backendinternal.RequireAuthenticatedActor(e)
	if err != nil {
		return nil, err
	}
	allowed, err := backendinternal.HasBattleplanAccess(app, actor)
	if err != nil {
		return nil, apis.NewBadRequestError("failed_battleplan_access", err)
	}
	if !allowed {
		return nil, apis.NewForbiddenError("forbidden_battleplan", nil)
	}
	return actor, nil
}

// requireOwnedBattleplan resolves the authenticated actor, loads the battleplan
// referenced by the request path, and verifies the actor owns it. Used by the
// mutation endpoints (update / status / delete).
func requireOwnedBattleplan(app *pocketbase.PocketBase, e *core.RequestEvent) (*core.Record, *core.Record, error) {
	actor, err := requireBattleplanActor(app, e)
	if err != nil {
		return nil, nil, err
	}
	record, err := loadBattleplan(app, e.Request.PathValue("id"))
	if err != nil {
		return nil, nil, err
	}
	if record.GetString("user") != actor.Id {
		return nil, nil, apis.NewForbiddenError("forbidden_battleplan", nil)
	}
	return actor, record, nil
}

func ListBattleplansHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		actor, err := requireBattleplanActor(app, e)
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
		actor, err := requireBattleplanActor(app, e)
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
		actor, err := requireBattleplanActor(app, e)
		if err != nil {
			return err
		}
		var input battleplans.Input
		if err := e.BindBody(&input); err != nil {
			return apis.NewBadRequestError("invalid_payload", err)
		}
		record, err := battleplans.Create(app, actor.Id, input)
		if err != nil {
			return battleplanBadRequest("failed_battleplan_create", err)
		}
		return e.JSON(http.StatusCreated, battleplans.MapItem(record))
	}
}

func UpdateBattleplanHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		_, record, err := requireOwnedBattleplan(app, e)
		if err != nil {
			return err
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
		_, record, err := requireOwnedBattleplan(app, e)
		if err != nil {
			return err
		}

		var payload struct {
			Status string `json:"status"`
		}
		if err := e.BindBody(&payload); err != nil {
			return apis.NewBadRequestError("invalid_payload", err)
		}
		next := strings.TrimSpace(payload.Status)
		if !battleplans.IsValidStatus(next) {
			return apis.NewBadRequestError("invalid_status_transition", nil)
		}
		if !battleplans.CanTransition(record.GetString("status"), next) {
			return apis.NewBadRequestError("invalid_status_transition", nil)
		}
		if err := battleplans.SetStatus(app, record, next); err != nil {
			return battleplanBadRequest("failed_battleplan_status", err)
		}
		return e.JSON(http.StatusOK, battleplans.MapItem(record))
	}
}

func ActivateBattleplanHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		_, record, err := requireOwnedBattleplan(app, e)
		if err != nil {
			return err
		}
		if err := battleplans.Activate(app, record); err != nil {
			return battleplanBadRequest("failed_battleplan_activate", err)
		}
		refreshed, err := loadBattleplan(app, record.Id)
		if err != nil {
			return err
		}
		return e.JSON(http.StatusOK, battleplans.MapItem(refreshed))
	}
}

func battleplanBadRequest(fallback string, err error) error {
	var collision battleplans.StatusCollisionError
	if errors.As(err, &collision) {
		return apis.NewBadRequestError(collision.Error(), nil)
	}
	return apis.NewBadRequestError(fallback, err)
}

func BattleplanNoteHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		actor, record, err := requireOwnedBattleplan(app, e)
		if err != nil {
			return err
		}

		var payload struct {
			Note string `json:"note"`
		}
		if err := e.BindBody(&payload); err != nil {
			return apis.NewBadRequestError("invalid_payload", err)
		}
		if err := battleplans.AddNote(app, actor, record, payload.Note); err != nil {
			return err
		}
		return e.JSON(http.StatusOK, battleplans.MapItem(record))
	}
}

func DeleteBattleplanHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		_, record, err := requireOwnedBattleplan(app, e)
		if err != nil {
			return err
		}
		if err := battleplans.Delete(app, record); err != nil {
			return apis.NewBadRequestError("failed_battleplan_delete", err)
		}
		return e.JSON(http.StatusOK, map[string]any{"deleted": true})
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

func BattleplanAccessHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		actor, err := backendinternal.RequireAuthenticatedActor(e)
		if err != nil {
			return err
		}
		access, err := backendinternal.HasBattleplanAccess(app, actor)
		if err != nil {
			return apis.NewBadRequestError("failed_battleplan_access", err)
		}
		return e.JSON(http.StatusOK, map[string]any{"access": access})
	}
}
