package handler

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"net/http"
	"panorama/lib/psql"
	"panorama/lib/utils"
	"panorama/services/api/handler/request"
	"panorama/services/api/handler/response"
	"panorama/services/api/model"
	"strconv"
	"strings"
	"time"

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

	tx, err := db.Begin(ctx)
	if err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}
	defer db.Release()

	if req.Password != req.RetypePassword {
		h.SendBadRequest(w, "Password is not match")
		return
	}

	member, err := m.AddMember(tx, context.Background(), model.MemberEnt{
		Name:     req.Name,
		Username: req.Username,
		Email:    req.Email,
		Phone:    req.Phone,
		Password: req.Password,
		IsActive: true, // hardcode sementara sebelum implement otp
	})
	if err != nil {
		h.SendBadRequest(w, psql.ParseErr(err))
		return
	}

	err = m.AddNewLogVisitApp(tx, ctx, member.ID, "customer")
	if err != nil {
		tx.Rollback(ctx)
		h.SendBadRequest(w, err.Error())
		return
	}

	mTemp, err := m.GetListMemberTemporaryByEmail(db, ctx, req.Email)
	if err != nil && err != sql.ErrNoRows {
		h.SendBadRequest(w, err.Error())
		return
	}

	if len(mTemp) > 0 {
		var arrstr []string
		for _, v := range mTemp {
			arrstr = append(arrstr, "("+strconv.Itoa(int(v.MemberItinID))+","+strconv.Itoa(int(member.ID))+",current_timestamp)")
		}
		err = m.AddMemberItinRelationBatch(ctx, tx, strings.Join(arrstr, ","))
		if err != nil {
			tx.Rollback(ctx)
			h.SendBadRequest(w, err.Error())
			return
		}

		err = m.DeleteMemberTempByEmail(ctx, tx, req.Email)
		if err != nil {
			tx.Rollback(ctx)
			h.SendBadRequest(w, err.Error())
			return
		}
	}

	// Commit transaction
	err = tx.Commit(ctx)
	if err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}

	var res response.MemberResponse
	res = res.Transform(member)

	rand.Seed(time.Now().UnixNano())
	token, _ := utils.Generate(`[\d]{4}`)

	// 1. need send token for phone validations
	if _, err := m.SendToken(db, ctx, h.GetChannel(r), model.ActRegPhone, model.TokenViaPhone, req.Phone, token); err != nil {
		fmt.Printf("error send token phone: %s", err.Error())
	}

	// // 2. need send token for email validations
	if _, err = m.SendToken(db, ctx, h.GetChannel(r), model.ActRegEmail, model.TokenViaEmail, req.Email, token); err != nil {
		fmt.Printf("error send token email: %s", err.Error())
	}

	res.Token = token

	// 3. need to create jwt token

	h.SendSuccess(w, res, nil)
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

	userTokenCredential, err := m.AuthLogin(db, context.Background(), h.GetChannel(r), req.Username, req.Password)
	if err != nil {
		h.SendAuthError(w, "invalid user or credential")
		return
	}

	h.SendSuccess(w, map[string]interface{}{
		"token":     userTokenCredential["token"],
		"user_code": userTokenCredential["user_code"],
		"user_role": userTokenCredential["user_role"],
		"user_name": userTokenCredential["user_name"],
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

	via := m.Via(req.Username)
	viaName := model.TokenViaEmail
	if via == "" {
		h.SendBadRequest(w, "email / phone requests is invalid.")
		return
	}
	if via == model.TokenViaPhone {
		viaName = model.TokenViaPhone
	}

	token, err := m.SendToken(db, ctx, h.GetChannel(r), model.ActForgotPass, viaName, req.Username, "")
	if err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}

	h.SendSuccess(w, map[string]interface{}{
		"token": token,
	}, nil)
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
	if req.Password != req.RetypePassword {
		h.SendBadRequest(w, "Password is not match")
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
