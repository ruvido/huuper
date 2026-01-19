package api

import (
	"log"
	"math/rand"
	"time"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

func BindRequestHooks(app *pocketbase.PocketBase) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	app.OnRecordUpdateRequest("requests").BindFunc(func(e *core.RecordRequestEvent) error {
		record := e.Record
		if record == nil {
			return e.Next()
		}

		oldStatus := record.Original().GetString("status")
		newStatus := record.GetString("status")
		if oldStatus != newStatus {
			log.Printf("requests hook: status change (id=%s old=%s new=%s)", record.Id, oldStatus, newStatus)
		}

		if newStatus != "1-accepted" {
			return e.Next()
		}
		if record.GetString("group") != "" {
			return e.Next()
		}

		regionID := record.GetString("region")
		if regionID == "" {
			log.Printf("requests hook: accepted but missing region (id=%s)", record.Id)
			return e.Next()
		}

		groupsFilter := "regions:each ?= {:region} && is_open = true"
		if groupsCollection, err := app.FindCollectionByNameOrId("groups"); err == nil {
			if groupsCollection.Fields.GetByName("regions") == nil {
				groupsFilter = "region = {:region} && is_open = true"
			}
		}

		groups, err := app.FindRecordsByFilter(
			"groups",
			groupsFilter,
			"",
			0,
			0,
			map[string]any{"region": regionID},
		)
		if err != nil || len(groups) == 0 {
			log.Printf("requests hook: no groups matched (id=%s region=%s err=%v)", record.Id, regionID, err)
			return e.Next()
		}

		groupID := groups[rng.Intn(len(groups))].Id
		log.Printf("requests hook: assigning group (id=%s region=%s group=%s matches=%d)", record.Id, regionID, groupID, len(groups))
		record.Set("group", groupID)

		return e.Next()
	})
}
