package api

import (
	"Contruction-Project/bootstrap"
	"Contruction-Project/services/api/handler"

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
	})

	r.Group(func(r chi.Router) {
		r.Use(app.VerifyJwtToken)

		r.Route("/users", func(r chi.Router) {
			r.Get("/", h.GetUserAct)
			r.Get("/{code}", h.GetUserAct)
			r.Post("/", h.AddUserAct)
			r.Put("/{code}", h.UpdateUserAct)
			r.Put("/{code}/pass", h.UpdateUserPassAct)
		})

		r.Route("/members", func(r chi.Router) {
			r.Get("/", h.GetMemberList)
			r.Get("/{code}", h.GetMember)
			r.Post("/", h.AddMemberAct)
			r.Put("/{code}", h.UpdateMember)
			r.Put("/{code}/pass", h.UpdateMemberPass)
		})
	})

}
