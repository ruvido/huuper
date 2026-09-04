package public

import "github.com/pocketbase/pocketbase/core"

type Handlers struct {
	Settings           func(e *core.RequestEvent) error
	EventsAccept       func(e *core.RequestEvent) error
	EventDetails       func(e *core.RequestEvent) error
	EventsRegister     func(e *core.RequestEvent) error
	RetreatDetails     func(e *core.RequestEvent) error
	RetreatsRegister   func(e *core.RequestEvent) error
	RequestsEmailOTP   func(e *core.RequestEvent) error
	RequestsOTPVerify  func(e *core.RequestEvent) error
	RequestsCreate     func(e *core.RequestEvent) error
	OnboardingGet      func(e *core.RequestEvent) error
	OnboardingFinalize func(e *core.RequestEvent) error
	StripeWebhook      func(e *core.RequestEvent) error
}

func Register(se *core.ServeEvent, h Handlers) {
	se.Router.GET("/api/public/settings/{name}", h.Settings)
	se.Router.GET("/api/public/events/accept", h.EventsAccept)
	se.Router.GET("/api/public/events/{slug}", h.EventDetails)
	se.Router.POST("/api/public/events/{slug}/register", h.EventsRegister)
	se.Router.GET("/api/public/retreats/{slug}", h.RetreatDetails)
	se.Router.POST("/api/public/retreats/{slug}/register", h.RetreatsRegister)
	se.Router.POST("/api/public/requests/email-otp", h.RequestsEmailOTP)
	se.Router.POST("/api/public/requests/email-otp/verify", h.RequestsOTPVerify)
	se.Router.POST("/api/public/requests", h.RequestsCreate)
	se.Router.GET("/api/public/onboarding/{token}", h.OnboardingGet)
	se.Router.POST("/api/public/onboarding/{token}/finalize", h.OnboardingFinalize)
	se.Router.POST("/api/public/payments/webhook", h.StripeWebhook)
}
