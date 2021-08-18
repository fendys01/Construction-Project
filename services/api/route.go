package api

import (
	"panorama/bootstrap"
	"panorama/services/api/handler"

	"github.com/go-chi/chi/v5"
)

// RegisterRoutes all routes for the apps
func RegisterRoutes(r *chi.Mux, app *bootstrap.App) {
	r.Route("/v1", func(r chi.Router) {
		r.Get("/ping", app.PingAction)

		RegisterSubsRoute(r, app)
	})
}

func RegisterSubsRoute(r chi.Router, app *bootstrap.App) {
	h := handler.Contract{App: app}

	r.Route("/auth", func(r chi.Router) {
		r.Post("/login", h.LoginAct)
		r.Post("/register", h.RegisterAct)
		r.Post("/register/phone-validation", h.RegisterPhoneValidateTokenAct)
		r.Post("/forgotpass", h.ForgotPassAct)
		r.Post("/change-pass", h.ForgotChagePassAct)
		r.Post("/forgotpass/validate-token", h.ForgotPassTokenValidationAct)
	})

	r.Route("/order", func(r chi.Router) {
		r.Post("/midtrans/notification", h.AddMidtransNotificationAct)
	})

	r.Route("/call-logs", func(r chi.Router) {
		r.Get("/", h.GetLogsList)
		r.Post("/citcall/{type}", h.AddCitcallLogs)
	})

	r.Group(func(r chi.Router) {
		r.Use(app.VerifyJwtToken)

		r.Post("/uploads", h.UploadAct)

		r.Route("/chats", func(r chi.Router) {
			r.Post("/", h.ChatAct)
			r.Post("/room", h.CreateChatGroup)
			r.Post("/invite-tc", h.InviteTcToGroupChat)
			r.Post("/message", h.ChatMessage)
		})

		r.Route("/sug-itin", func(r chi.Router) {
			r.Get("/", h.GetItinSugList)
			r.Post("/", h.AddSugItinAct)
			r.Get("/{code}", h.GetSugItinAct)
			r.Put("/{code}", h.UpdateSugItinAct)
			r.Delete("/{code}", h.DelSugItinAct)
		})

		r.Route("/member-itin", func(r chi.Router) {
			r.Get("/", h.GetItinMemberList)
			r.Post("/", h.AddMemberItinAct)
			r.Get("/{code}", h.GetMemberItinAct)
			r.Put("/{code}", h.UpdateMemberItinAct)
			r.Delete("/{code}", h.DelMemberItinAct)
		})

		r.Route("/users", func(r chi.Router) {
			r.Get("/", h.GetUserListAct)
			r.Get("/{code}", h.GetUserListAct)
			r.Get("/{code}/activity", h.GetDetailAdminAndTc)
			r.Post("/", h.AddUserAct)
			r.Put("/{code}", h.UpdateUserAct)
			r.Put("/{code}/pass", h.UpdateUserPassAct)
			r.Delete("/{code}", h.DeleteUser)
		})

		r.Route("/members", func(r chi.Router) {
			r.Get("/", h.GetMemberList)
			r.Get("/{code}", h.GetMember)
			r.Get("/{code}/activity", h.GetMemberStatistik)
			r.Post("/", h.AddMemberAct)
			r.Put("/{code}", h.UpdateMember)
			r.Put("/{code}/pass", h.UpdateMemberPassPhoneAct)
			r.Post("/pass", h.AddMemberPassHandlerAct)
		})

		r.Route("/orders", func(r chi.Router) {
			r.Get("/", h.GetListItinOrderMember)
			r.Get("/{code}/detail", h.GetDetailItinOrderMember)
			r.Post("/", h.AddOrderAct)
			r.Put("/{code}", h.UpdateOrderAct)
		})

		// create push notification
		r.Route("/notification", func(r chi.Router) {
			r.Post("/", h.AddPushNotifAct)
			r.Get("/", h.GetListNotifAct)
		})

		// add device
		r.Route("/players", func(r chi.Router) {
			r.Post("/", h.AddPlayersAct)
		})

		// Dashboard Admin
		r.Route("/dashboard", func(r chi.Router) {
			r.Get("/admin", h.GetDashboardAdminAct)
			r.Get("/tc", h.GetDashboardTcAct)
		})
	})
}
