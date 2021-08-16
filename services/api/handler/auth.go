package handler

import (
	"Contruction-Project/services/api/handler/request"
	"Contruction-Project/services/api/model"
	"context"
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
)

// RegisterAct add new member (cust) to databse
func (h *Contract) RegisterAct(w http.ResponseWriter, r *http.Request) {
	var err error
	req := request.AddMemberReq{}
	if err = h.Bind(r, &req); err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}
	if err = h.Validator.Driver.Struct(req); err != nil {
		h.SendRequestValidationError(w, err.(validator.ValidationErrors))
		return
	}

	m := model.Contract{App: h.App}
	ctx := context.Background()
	db, err := h.DB.Acquire(ctx)
	if err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}
	defer db.Release()

	_, err = m.AddMember(db, context.Background(), model.MemberEnt{
		Name:     req.Name,
		Username: req.Username,
		Email:    req.Email,
		Phone:    req.Phone,
		Password: req.Password,
	})
	if err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}

	// 1. need send token for phone validations
	if err = m.SendToken(db, ctx, h.GetChannel(r), model.ActRegPhone, model.TokenViaPhone, req.Phone); err != nil {
		fmt.Printf("error send token phone: %s", err.Error())
	}

	// 2. need send token for email validations
	if err = m.SendToken(db, ctx, h.GetChannel(r), model.ActRegEmail, model.TokenViaEmail, req.Email); err != nil {
		fmt.Printf("error send token email: %s", err.Error())
	}

	// 3. need to create jwt token

	h.SendSuccess(w, h.EmptyJSONArr(), nil)
}

// RegisterPhoneValidateTokenAct is a phone validation token for register customer
func (h *Contract) RegisterPhoneValidateTokenAct(w http.ResponseWriter, r *http.Request) {
	var err error
	req := request.RegisterPhoneValidationReq{}
	if err := h.Bind(r, &req); err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}
	if err = h.Validator.Driver.Struct(req); err != nil {
		h.SendRequestValidationError(w, err.(validator.ValidationErrors))
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
	if err = m.ValidateToken(db, ctx, h.GetChannel(r), model.ActRegPhone, req.Phone, req.Token); err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}

	// activate and set the phone validation flag to true
	if err = m.ActivateAndSetPhoneValid(db, ctx, req.Phone); err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}

	h.SendSuccess(w, h.EmptyJSONArr(), nil)
}

// Login ...
func (h *Contract) LoginAct(w http.ResponseWriter, r *http.Request) {
	var err error

	req := request.LoginReq{}
	if err := h.Bind(r, &req); err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}
	if err = h.Validator.Driver.Struct(req); err != nil {
		h.SendRequestValidationError(w, err.(validator.ValidationErrors))
		return
	}

	m := model.Contract{App: h.App}
	db, err := h.DB.Acquire(context.Background())
	if err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}
	defer db.Release()

	var token string
	token, err = m.AuthLogin(db, context.Background(), h.GetChannel(r), req.Username, req.Password)
	if err != nil {
		h.SendAuthError(w, "invalid user or credential")
		return
	}

	h.SendSuccess(w, map[string]interface{}{
		"token": token,
	}, nil)
}

// ForgotPassReq
func (h *Contract) ForgotPassAct(w http.ResponseWriter, r *http.Request) {
	var err error
	req := request.ForgotPassReq{}
	if err := h.Bind(r, &req); err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}
	if err = h.Validator.Driver.Struct(req); err != nil {
		h.SendRequestValidationError(w, err.(validator.ValidationErrors))
		return
	}

	m := model.Contract{App: h.App}
	ctx := context.Background()
	db, err := h.DB.Acquire(ctx)
	if err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}
	defer db.Release()

	err = m.SendToken(db, ctx, h.GetChannel(r), model.ActForgotPass, model.TokenViaEmail, req.Username)
	if err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}

	h.SendSuccess(w, h.EmptyJSONArr(), nil)
}

// ForgotPassTokenValidation validate the token that user input when they receive token
func (h *Contract) ForgotPassTokenValidationAct(w http.ResponseWriter, r *http.Request) {
	var err error
	req := request.ForgotPassTokenReq{}
	if err := h.Bind(r, &req); err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}
	if err = h.Validator.Driver.Struct(req); err != nil {
		h.SendRequestValidationError(w, err.(validator.ValidationErrors))
		return
	}

	db, err := h.DB.Acquire(context.Background())
	if err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}
	defer db.Release()

	m := model.Contract{App: h.App}
	if err = m.ValidateToken(db, context.Background(), h.GetChannel(r), model.ActForgotPass, req.Username, req.Token); err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}

	h.SendSuccess(w, h.EmptyJSONArr(), nil)
}

// ForgotChagePass change password after token valid
func (h *Contract) ForgotChagePassAct(w http.ResponseWriter, r *http.Request) {
	var err error
	req := request.ChangePassForgotReq{}
	if err = h.Bind(r, &req); err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}
	if err = h.Validator.Driver.Struct(req); err != nil {
		h.SendRequestValidationError(w, err.(validator.ValidationErrors))
		return
	}

	// TODO: for now we get username from payload
	// must get username from the other (ex: jwt token, encrypted username)
	db, err := h.DB.Acquire(context.Background())
	if err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}
	defer db.Release()

	m := model.Contract{App: h.App}
	if err = m.UpdateMemberPass(db, context.Background(), req.Username, req.Password); err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}

	h.SendSuccess(w, h.EmptyJSONArr(), nil)
}
