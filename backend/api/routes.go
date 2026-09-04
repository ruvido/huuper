package api

import (
	adminroutes "members/backend/api/admin"
	meroutes "members/backend/api/me"
	publicroutes "members/backend/api/public"
	staticroutes "members/backend/api/static"
	eventinternal "members/backend/internal/events"
	paymentsinternal "members/backend/internal/payments"
	retreatsinternal "members/backend/internal/retreats"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

func RegisterRoutes(app *pocketbase.PocketBase, se *core.ServeEvent) {
	paymentsinternal.RegisterConfirmHandler("event_registration", eventinternal.ConfirmPayment)
	paymentsinternal.RegisterConfirmHandler("retreat_registration", retreatsinternal.ConfirmPayment)

	publicroutes.Register(se, publicroutes.Handlers{
		Settings:           publicroutes.SettingsHandler(app),
		EventsAccept:       publicroutes.AcceptEventHandler(app),
		EventDetails:       publicroutes.EventDetailsHandler(app),
		EventsRegister:     publicroutes.RegisterEventHandler(app),
		RetreatDetails:     publicroutes.RetreatDetailsHandler(app),
		RetreatsRegister:   publicroutes.RegisterRetreatHandler(app),
		RequestsEmailOTP:   publicroutes.RequestEmailOTPHandler(app),
		RequestsOTPVerify:  publicroutes.VerifyRequestEmailOTPHandler(app),
		RequestsCreate:     publicroutes.SubmitRequestHandler(app),
		OnboardingGet:      publicroutes.OnboardingGetHandler(app),
		OnboardingFinalize: publicroutes.OnboardingFinalizeHandler(app),
		StripeWebhook:      publicroutes.StripeWebhookHandler(app),
	})

	meroutes.Register(se, meroutes.Handlers{
		Settings:              meroutes.SettingsHandler(app),
		GroupsList:            meroutes.GroupsListHandler(app),
		GroupGet:              meroutes.GroupGetHandler(app),
		GroupAssistant:        meroutes.GroupAssistantHandler(app),
		UserGet:               meroutes.UserGetHandler(app),
		EventsList:            meroutes.ListEventsHandler(app),
		EventGet:              meroutes.EventGetHandler(app),
		EventStatus:           meroutes.EventStatusHandler(app),
		EventCreate:           meroutes.CreateEventHandler(app),
		EventUpdate:           meroutes.UpdateEventHandler(app),
		EventReschedule:       meroutes.RescheduleEventHandler(app),
		EventCancel:           meroutes.CancelEventHandler(app),
		EventCancelOccurrence: meroutes.CancelOccurrenceHandler(app),
		EventRegister:         meroutes.RegisterEventHandler(app),
		EventUnregister:       meroutes.UnregisterEventHandler(app),
		EventUnsubscribe:      meroutes.EventUnsubscribeHandler(app),
		EventAttendance:       meroutes.MarkAttendanceHandler(app),
		TelegramToken:         meroutes.GenerateTelegramTokenHandler(app),
		RequestsList:          meroutes.ListRequestsHandler(app),
		RequestGet:            meroutes.GetRequestHandler(app),
		RequestAction:         meroutes.RequestActionHandler(app),
		GroupMembers:          meroutes.GroupMembersHandler(app),
		GroupGuardians:        meroutes.GroupGuardiansHandler(app),
		GroupRequestsCount:    meroutes.GroupRequestsCountHandler(app),
		DefaultInvite:         meroutes.DefaultGroupInviteHandler(app),
		BattleplansList:       meroutes.ListBattleplansHandler(app),
		BattleplanGet:         meroutes.GetBattleplanHandler(app),
		BattleplanCreate:      meroutes.CreateBattleplanHandler(app),
		BattleplanUpdate:      meroutes.UpdateBattleplanHandler(app),
		BattleplanStatus:      meroutes.BattleplanStatusHandler(app),
		BattleplanActivate:    meroutes.ActivateBattleplanHandler(app),
		BattleplanNote:        meroutes.BattleplanNoteHandler(app),
		BattleplanDelete:      meroutes.DeleteBattleplanHandler(app),
		BattleplanAccess:      meroutes.BattleplanAccessHandler(app),
		EventsAccess:          meroutes.EventsAccessHandler(app),
	})

	adminroutes.Register(se, adminroutes.Handlers{
		Settings:              adminroutes.SettingsHandler(app),
		Summary:               adminroutes.SummaryHandler(app),
		GroupsList:            adminroutes.GroupsListHandler(app),
		GroupGet:              adminroutes.GroupGetHandler(app),
		GroupAssistant:        adminroutes.GroupAssistantHandler(app),
		EventsList:            adminroutes.EventsListHandler(app),
		EventCreate:           adminroutes.CreateEventHandler(app),
		EventUpdate:           adminroutes.UpdateEventHandler(app),
		EventReschedule:       adminroutes.RescheduleEventHandler(app),
		EventCancel:           adminroutes.CancelEventHandler(app),
		EventCancelOccurrence: adminroutes.CancelOccurrenceEventHandler(app),
		EventSetActive:        adminroutes.SetEventActiveHandler(app),
		EventAttendance:       adminroutes.MarkAttendanceHandler(app),
		RequestsList:          meroutes.ListRequestsHandler(app),
		RequestCreate:         adminroutes.CreateRequestHandler(app),
		RequestGet:            meroutes.GetRequestHandler(app),
		RequestAction:         meroutes.RequestActionHandler(app),
		EventDetails:          adminroutes.EventDetailsHandler(app),
		UserGet:               adminroutes.UserGetHandler(app),
		EventEmail:            adminroutes.EventEmailHandler(app),
		RegistrationApprove:   adminroutes.ApproveRegistrationHandler(app),
		RegistrationReject:    adminroutes.RejectRegistrationHandler(app),
		RegistrationCancel:    adminroutes.CancelRegistrationHandler(app),
		GroupSyncMemberships:  adminroutes.SyncGroupMembershipsHandler(app),
		UsersList:             adminroutes.UsersListHandler(app),
		UserCancel:            adminroutes.CancelUserHandler(app),
		UserDelete:            adminroutes.DeleteUserHandler(app),

		RetreatsList:               adminroutes.RetreatsListHandler(app),
		RetreatCreate:              adminroutes.RetreatCreateHandler(app),
		RetreatUpdate:              adminroutes.RetreatUpdateHandler(app),
		RetreatCancel:              adminroutes.RetreatCancelHandler(app),
		RetreatDetails:             adminroutes.RetreatDetailsAdminHandler(app),
		RetreatUploadGallery:       adminroutes.RetreatUploadGalleryHandler(app),
		RetreatRemoveGalleryFile:   adminroutes.RetreatRemoveGalleryFileHandler(app),
		RetreatBroadcastEmail:      adminroutes.RetreatBroadcastEmailHandler(app),
		RetreatRegistrationApprove: adminroutes.RetreatRegistrationApproveHandler(app),
		RetreatRegistrationReject:  adminroutes.RetreatRegistrationRejectHandler(app),
		RetreatRegistrationCancel:  adminroutes.RetreatRegistrationCancelHandler(app),
	})

	staticroutes.Register(se)
}
