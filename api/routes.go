package api

import (
	adminroutes "members/api/admin"
	meroutes "members/api/me"
	publicroutes "members/api/public"
	staticroutes "members/api/static"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

func RegisterRoutes(app *pocketbase.PocketBase, se *core.ServeEvent) {
	publicroutes.Register(se, publicroutes.Handlers{
		Settings:       GetSettingsHandler(app),
		EventsAccept:   AcceptEventHandler(app),
		EventsRegister: RegisterEventHandler(app),
		RequestsCreate: publicroutes.SubmitRequestHandler(app),
	})

	meroutes.Register(se, meroutes.Handlers{
		Settings:           GetSettingsHandler(app),
		EventStatus:        EventStatusHandler(app),
		EventUnsubscribe:   EventUnsubscribeHandler(app),
		TelegramToken:      GenerateTelegramTokenHandler(app),
		RequestsList:       meroutes.ListRequestsHandler(app),
		RequestGet:         meroutes.GetRequestHandler(app),
		RequestAction:      meroutes.RequestActionHandler(app),
		GroupMembers:       GroupMembersHandler(app),
		GroupGuardians:    GroupGuardiansHandler(app),
		GroupRequestsCount: GroupRequestsCountHandler(app),
		DefaultInvite:      DefaultGroupInviteHandler(app),
	})

	adminroutes.Register(se, adminroutes.Handlers{
		Settings:             GetSettingsHandler(app),
		Summary:              adminroutes.SummaryHandler(app),
		EventDetails:         AdminEventDetailsHandler(app),
		EventEmail:           AdminEventEmailHandler(app),
		RegistrationApprove:  AdminApproveRegistrationHandler(app),
		RegistrationReject:   AdminRejectRegistrationHandler(app),
		RegistrationCancel:   AdminCancelRegistrationHandler(app),
		GroupSyncMemberships: adminroutes.SyncGroupMembershipsHandler(app),
		UserDelete:           adminroutes.DeleteUserHandler(app),
	})

	staticroutes.Register(se)
}
