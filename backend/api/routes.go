package api

import (
	adminroutes "members/backend/api/admin"
	meroutes "members/backend/api/me"
	publicroutes "members/backend/api/public"
	staticroutes "members/backend/api/static"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

func RegisterRoutes(app *pocketbase.PocketBase, se *core.ServeEvent) {
	publicroutes.Register(se, publicroutes.Handlers{
		Settings:       publicroutes.SettingsHandler(app),
		EventsAccept:   publicroutes.AcceptEventHandler(app),
		EventsRegister: publicroutes.RegisterEventHandler(app),
		RequestsCreate: publicroutes.SubmitRequestHandler(app),
	})

	meroutes.Register(se, meroutes.Handlers{
		Settings:           meroutes.SettingsHandler(app),
		GroupsList:         meroutes.GroupsListHandler(app),
		GroupGet:           meroutes.GroupGetHandler(app),
		GroupAssistant:     meroutes.GroupAssistantHandler(app),
		UserGet:            meroutes.UserGetHandler(app),
		EventGet:           meroutes.EventGetHandler(app),
		EventStatus:        meroutes.EventStatusHandler(app),
		EventUnsubscribe:   meroutes.EventUnsubscribeHandler(app),
		TelegramToken:      meroutes.GenerateTelegramTokenHandler(app),
		RequestsList:       meroutes.ListRequestsHandler(app),
		RequestGet:         meroutes.GetRequestHandler(app),
		RequestAction:      meroutes.RequestActionHandler(app),
		GroupMembers:       meroutes.GroupMembersHandler(app),
		GroupGuardians:     meroutes.GroupGuardiansHandler(app),
		GroupRequestsCount: meroutes.GroupRequestsCountHandler(app),
		DefaultInvite:      meroutes.DefaultGroupInviteHandler(app),
	})

	adminroutes.Register(se, adminroutes.Handlers{
		Settings:             adminroutes.SettingsHandler(app),
		Summary:              adminroutes.SummaryHandler(app),
		GroupsList:           adminroutes.GroupsListHandler(app),
		GroupGet:             adminroutes.GroupGetHandler(app),
		GroupAssistant:       adminroutes.GroupAssistantHandler(app),
		EventsList:           adminroutes.EventsListHandler(app),
		RequestsList:         meroutes.ListRequestsHandler(app),
		RequestGet:           meroutes.GetRequestHandler(app),
		RequestAction:        meroutes.RequestActionHandler(app),
		EventDetails:         adminroutes.EventDetailsHandler(app),
		UserGet:              adminroutes.UserGetHandler(app),
		EventEmail:           adminroutes.EventEmailHandler(app),
		RegistrationApprove:  adminroutes.ApproveRegistrationHandler(app),
		RegistrationReject:   adminroutes.RejectRegistrationHandler(app),
		RegistrationCancel:   adminroutes.CancelRegistrationHandler(app),
		GroupSyncMemberships: adminroutes.SyncGroupMembershipsHandler(app),
		UserDelete:           adminroutes.DeleteUserHandler(app),
	})

	staticroutes.Register(se)
}
