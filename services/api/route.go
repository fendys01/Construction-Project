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
		r.Post("/forgotpass", h.ForgotPassAct)
		r.Post("/change-pass", h.ForgotChangePassAct)
		r.Post("/checktoken-forgotpass", h.AuthCheckTokenForgotPassAct)
		r.Post("/checktoken-phone", h.AuthCheckTokenPhoneAct)
		r.Post("/token/{type}", h.SendTokenAct)
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
			r.Get("/", h.GetChatListAct)
			r.Post("/", h.ChatAct)
			r.Post("/room", h.CreateChatGroup)
			r.Put("/invite-tc", h.InviteTcToGroupChat)
			r.Post("/message", h.ChatMessage)
			r.Get("/{code}", h.GetHistoryChatByCode)
			r.Put("/{code}/leave-session", h.LeaveSessionChatAct)
			r.Put("/is-read", h.UpdateIsReadMessages)
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
			r.Post("/checktoken-pass", h.UpdateMemberPassTokenAct)
			r.Put("/{code}", h.UpdateMember)
			r.Put("/pass/{code}", h.UpdateMemberPassAct)
			r.Put("/phone/{code}", h.UpdateMemberPhoneAct)
			r.Delete("/{code}", h.DeleteMember)
			r.Delete("/{code}/force-delete", h.ForceDeleteMember)
		})

		r.Route("/orders", func(r chi.Router) {
			r.Get("/", h.GetListItinOrderMember)
			r.Get("/{code}/detail", h.GetDetailItinOrderMember)
			r.Post("/", h.AddOrderAct)
			r.Put("/{code}", h.UpdateOrderAct)
			r.Post("/payment", h.PostPaymentAct)
		})

		// create push notification
		r.Route("/notification", func(r chi.Router) {
			r.Get("/", h.GetListNotifAct)
			r.Get("/{code}", h.GetNotifAct)
			r.Get("/counter", h.GetCounterNotifAct)
			r.Put("/{code}/is-read", h.UpdateIsReadNotification)
			r.Delete("/{code}", h.DeleteNotificationAct)
			r.Delete("/", h.DeleteAllNotificationAct)
		})

		r.Route("/dashboard", func(r chi.Router) {
			r.Get("/", h.GetDashboardAct)
		})

		r.Route("/stuff", func(r chi.Router) {
			r.Post("/", h.AddStuffAct)
			r.Get("/", h.GetListStuffAct)
			r.Get("/{code}/detail", h.GeDetailStuffAct)
			r.Put("/{code}", h.UpdateDataStuffAct)
			r.Delete("/{code}", h.DeleteStuffAct)
		})
	})
}
