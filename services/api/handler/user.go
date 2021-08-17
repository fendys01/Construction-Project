package handler

import (
	"Contruction-Project/services/api/handler/response"
	"Contruction-Project/services/api/model"
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"
)

type (
	userReq struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Pass     string `json:"password"`
		Phone    string `json:"phone"`
		IsActive bool   `json:"is_active"`
		Role     string `json:"role"`
	}

	passReq struct {
		Pass string `json:"password"`
	}
)

// GetUserAct ...
func (h *Contract) GetUserAct(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")

	ctx := context.Background()
	db, err := h.DB.Acquire(ctx)
	if err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}
	defer db.Release()

	m := model.Contract{App: h.App}
	if len(code) > 0 {
		// return single user (get by user code)
		u, err := m.GetUserByCode(db, ctx, code)
		if err != nil {
			h.SendBadRequest(w, err.Error())
			return
		}

		var res response.UsersDetailResponse
		res = res.Transform(u)

		h.SendSuccess(w, res, nil)
		return
	}

	u, err := m.GetUser(db, ctx)
	if err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}

	var listResponse []response.UsersResponse
	for _, a := range u {
		var res response.UsersResponse
		res = res.Transform(a)
		listResponse = append(listResponse, res)
	}

	h.SendSuccess(w, listResponse, nil)
}

// AddUserAct ...
func (h *Contract) AddUserAct(w http.ResponseWriter, r *http.Request) {
	var err error
	req := userReq{}
	if err = h.Bind(r, &req); err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}
	if err = h.Validator.Driver.Struct(req); err != nil {
		h.SendRequestValidationError(w, err.(validator.ValidationErrors))
		return
	}

	pass, err := bcrypt.GenerateFromPassword([]byte(req.Pass), 10)
	if err != nil {
		h.SendBadRequest(w, "Error when generate password")
		return
	}

	ctx := context.Background()
	db, err := h.DB.Acquire(ctx)
	if err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}
	defer db.Release()

	m := model.Contract{App: h.App}
	err = m.AddUser(db, ctx, model.UserEnt{
		Name:     req.Name,
		Email:    req.Email,
		Phone:    req.Phone,
		Password: string(pass),
		Role:     req.Role,
		IsActive: req.IsActive,
	})
	if err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}

	h.SendSuccess(w, h.EmptyJSONArr(), nil)
}

// UpdateUserAct ...
func (h *Contract) UpdateUserAct(w http.ResponseWriter, r *http.Request) {
	var err error
	req := userReq{}
	if err = h.Bind(r, &req); err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}
	if err = h.Validator.Driver.Struct(req); err != nil {
		h.SendRequestValidationError(w, err.(validator.ValidationErrors))
		return
	}

	code := chi.URLParam(r, "code")
	if len(code) == 0 {
		h.SendBadRequest(w, "invalid code")
		return
	}

	ctx := context.Background()
	db, err := h.DB.Acquire(ctx)
	if err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}
	defer db.Release()

	m := model.Contract{App: h.App}
	err = m.UpdateUser(db, ctx, code, model.UserEnt{
		Name:     req.Name,
		Role:     req.Role,
		IsActive: req.IsActive,
	})
	if err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}

	h.SendSuccess(w, h.EmptyJSONArr(), nil)
}

// UpdateUserPassAct ...
func (h *Contract) UpdateUserPassAct(w http.ResponseWriter, r *http.Request) {
	var err error
	req := passReq{}
	if err = h.Bind(r, &req); err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}
	if err = h.Validator.Driver.Struct(req); err != nil {
		h.SendRequestValidationError(w, err.(validator.ValidationErrors))
		return
	}

	code := chi.URLParam(r, "code")
	if len(code) == 0 {
		h.SendBadRequest(w, "invalid code")
		return
	}

	pass, err := bcrypt.GenerateFromPassword([]byte(req.Pass), 10)
	if err != nil {
		h.SendBadRequest(w, "Error when generate password")
		return
	}

	ctx := context.Background()
	db, err := h.DB.Acquire(ctx)
	if err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}
	defer db.Release()

	m := model.Contract{App: h.App}
	err = m.UpdateUserPass(db, ctx, code, string(pass))
	if err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}

	h.SendSuccess(w, h.EmptyJSONArr(), nil)
}
