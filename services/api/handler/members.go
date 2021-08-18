package handler

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"panorama/lib/array"
	"panorama/lib/psql"
	"panorama/services/api/handler/request"
	"panorama/services/api/handler/response"
	"panorama/services/api/model"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

// AddMemberAct ...
func (h *Contract) AddMemberAct(w http.ResponseWriter, r *http.Request) {

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

	if req.Password != req.RetypePassword {
		h.SendBadRequest(w, "Password is not match")
		return
	}

	m := model.Contract{App: h.App}
	tx, err := db.Begin(ctx)
	if err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}
	member, err := m.AddMember(tx, ctx, model.MemberEnt{
		Username: req.Username,
		Name:     req.Name,
		Email:    req.Email,
		Phone:    req.Phone,
		Img:      sql.NullString{String: req.Img, Valid: true},
		IsActive: true,
		Password: req.Password,
	})
	if err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}

	var res response.MemberResponse
	res = res.Transform(member)

	err = m.AddNewLogVisitApp(tx, ctx, member.ID, "customer")
	if err != nil {
		tx.Rollback(ctx)
		h.SendBadRequest(w, err.Error())
		return
	}

	// Commit transaction
	err = tx.Commit(ctx)
	if err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}

	h.SendSuccess(w, res, nil)
}

// GetMember ...
func (h *Contract) GetMember(w http.ResponseWriter, r *http.Request) {
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
	u, err := m.GetMemberByCode(db, ctx, code)
	if err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}

	var res response.MemberResponse
	res = res.Transform(u)

	h.SendSuccess(w, res, nil)
}

// GetMemberStatistik ...
func (h *Contract) GetMemberStatistik(w http.ResponseWriter, r *http.Request) {
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
	u, err := m.GetMemberStatistikByCode(db, ctx, code)
	if err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}

	log, err := m.GetListLogActivity(db, ctx, "customer", code)
	if err != nil && err == sql.ErrNoRows {
		h.SendBadRequest(w, err.Error())
		return
	}

	u.LogActivityUser = log

	var res response.MemberDetailResponse
	res = res.Transform(u)

	h.SendSuccess(w, res, nil)
}

// GetMemberList ...
func (h *Contract) GetMemberList(w http.ResponseWriter, r *http.Request) {

	param := map[string]interface{}{
		"keyword": "",
		"page":    1,
		"limit":   10,
		"offset":  0,
		"sort":    "desc",
		"order":   "id",
	}

	if page, ok := r.URL.Query()["page"]; ok && len(page[0]) > 0 {
		if p, err := strconv.Atoi(page[0]); err == nil && p > 1 {
			param["page"] = p
		}
	}

	if sort, ok := r.URL.Query()["sort"]; ok && len(sort[0]) > 0 && strings.ToLower(sort[0]) == "asc" {
		param["sort"] = "asc"
	}

	if order, ok := r.URL.Query()["order"]; ok && len(order[0]) > 0 {
		arrStr := new(array.ArrStr)
		if exist, _ := arrStr.InArray(order[0], []string{"id"}); exist {
			param["order"] = order[0]
		}
	}

	if keyword, ok := r.URL.Query()["keyword"]; ok && len(keyword[0]) > 0 {
		param["keyword"] = keyword[0]
	}

	if limit, ok := r.URL.Query()["limit"]; ok {
		if l, err := strconv.Atoi(limit[0]); err == nil {
			param["limit"] = l
		}
	}

	param["offset"] = (param["page"].(int) - 1) * param["limit"].(int)

	ctx := context.Background()
	db, err := h.DB.Acquire(ctx)
	if err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}

	defer db.Release()

	m := model.Contract{App: h.App}
	members, err := m.GetListMember(db, ctx, param)
	if err != nil && sql.ErrNoRows != nil {
		h.SendBadRequest(w, err.Error())
		return
	}

	var listResponse []response.MemberResponse
	for _, a := range members {
		var res response.MemberResponse
		res = res.Transform(a)
		listResponse = append(listResponse, res)
	}

	h.SendSuccess(w, listResponse, param)

}

// UpdateMember ...
func (h *Contract) UpdateMember(w http.ResponseWriter, r *http.Request) {

	mcode := chi.URLParam(r, "code")

	var err error
	req := request.UpdateMemberReq{}
	if err = h.Bind(r, &req); err != nil {
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

	data, err := m.GetMemberByCode(db, ctx, mcode)
	if err == sql.ErrNoRows {
		h.SendNotfound(w, err.Error())
		return
	}
	if err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}

	data = req.Transform(data)

	member, err := m.UpdateMember(db, mcode, data)
	if err != nil {
		h.SendBadRequest(w, psql.ParseErr(err))
		return
	}

	var res response.MemberResponse
	res = res.Transform(member)

	h.SendSuccess(w, res, nil)

}

// UpdateMemberPassPhoneAct is handler for handling update pass or phone member
func (h *Contract) UpdateMemberPassPhoneAct(w http.ResponseWriter, r *http.Request) {
	var err error

	// Check each request
	req := request.PassPhoneReq{}
	if err = h.Bind(r, &req); err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}
	if err = h.Validator.Driver.Struct(req); err != nil {
		h.SendRequestValidationError(w, err.(validator.ValidationErrors))
		return
	}
	if req.Password == "" && req.Phone == "" {
		h.SendBadRequest(w, "no request found.")
		return
	}
	if (req.Password != "" || req.RetypePassword != "") && req.Phone != "" {
		h.SendBadRequest(w, "invalid request.")
		return
	}

	// Define code param
	code := chi.URLParam(r, "code")
	if len(code) == 0 {
		h.SendBadRequest(w, "invalid code")
		return
	}

	// Check db context
	ctx := context.Background()
	db, err := h.DB.Acquire(ctx)
	if err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}
	defer db.Release()
	m := model.Contract{App: h.App}

	// Response
	response := map[string]interface{}{
		"member_code": code,
	}

	// Handle request
	if req.Password != "" { // Handle request password
		if req.Password != req.RetypePassword {
			h.SendBadRequest(w, "Password is not match")
			return
		}
		err = m.UpdateMemberPass(db, ctx, code, req.Password)
		if err != nil {
			h.SendBadRequest(w, err.Error())
			return
		}
	} else if req.Phone != "" { // Handle request phone
		err = m.UpdatePhoneMember(db, ctx, req.Phone, code)
		if err != nil {
			h.SendBadRequest(w, err.Error())
			return
		}
		response["member_phone"] = req.Phone
	}

	h.SendSuccess(w, response, nil)
}

// AddMemberPassHandlerAct is handler for handling update pass or phone member
func (h *Contract) AddMemberPassHandlerAct(w http.ResponseWriter, r *http.Request) {
	// Check each request
	req := request.ForgotPassReq{}
	if err := h.Bind(r, &req); err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}
	if err := h.Validator.Driver.Struct(req); err != nil {
		h.SendRequestValidationError(w, err.(validator.ValidationErrors))
		return
	}

	// Response
	response := map[string]interface{}{
		"member_code": h.GetUserCode(r.Context()),
	}

	// Check db context
	ctx := context.Background()
	db, err := h.DB.Acquire(ctx)
	if err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}
	defer db.Release()
	m := model.Contract{App: h.App}

	// Handle request
	via := m.Via(req.Username)
	viaName := model.TokenViaEmail
	if via == "" {
		h.SendBadRequest(w, "email / phone requests is invalid.")
		return
	}
	if via == model.TokenViaPhone {
		viaName = model.TokenViaPhone
	}
	response["username"] = req.Username

	// Verification token
	token, err := m.SendToken(db, ctx, h.GetChannel(r), model.ActChangePass, viaName, req.Username, "")
	if err != nil {
		h.SendBadRequest(w, fmt.Sprintf("error send token %s: %s", viaName, err.Error()))
		return
	}
	response["token"] = token

	h.SendSuccess(w, response, nil)
}
