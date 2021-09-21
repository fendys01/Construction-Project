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
	tx, err := db.Begin(ctx)
	if err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}

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

	// Update member
	member, err := m.UpdateMember(tx, mcode, data)
	if err != nil {
		h.SendBadRequest(w, psql.ParseErr(err))
		tx.Rollback(ctx)
		return
	}

	// Send Notifications
	players, err := m.GetListPlayerByUserCodeAndRole(db, ctx, mcode, h.GetUserRole(r.Context()))
	if err != nil {
		h.SendBadRequest(w, err.Error())
		tx.Rollback(ctx)
		return
	}
	notifContent := model.NotificationContent{
		Subject: model.NOTIF_SUBJ_PROFILE_CHANGE,
	}
	_, err = m.SendNotifications(tx, db, ctx, players, notifContent)
	if err != nil {
		h.SendBadRequest(w, psql.ParseErr(err))
		tx.Rollback(ctx)
		return
	}

	// Set response
	var res response.MemberResponse
	res = res.Transform(member)

	// Commit process
	err = tx.Commit(ctx)
	if err != nil {
		h.SendBadRequest(w, err.Error())
		tx.Rollback(ctx)
		return
	}

	h.SendSuccess(w, res, nil)
}

// UpdateMemberPassTokenAct is handler for send token member auth
func (h *Contract) UpdateMemberPassTokenAct(w http.ResponseWriter, r *http.Request) {
	// Check each request
	req := request.CheckTokenReq{}
	if err := h.Bind(r, &req); err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}
	if err := h.Validator.Driver.Struct(req); err != nil {
		h.SendRequestValidationError(w, err.(validator.ValidationErrors))
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

	if err = m.ValidateToken(db, context.Background(), h.GetChannel(r), model.ActChangePass, req.Username, req.Token); err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}

	// Response
	response := map[string]interface{}{
		"member_code": h.GetUserCode(r.Context()),
		"username":    req.Username,
		"token":       req.Token,
	}

	h.SendSuccess(w, response, nil)
}

// UpdateMemberPassAct is handler for handling update pass
func (h *Contract) UpdateMemberPassAct(w http.ResponseWriter, r *http.Request) {
	var err error

	// Check each request
	req := request.ChangePassReq{}
	if err = h.Bind(r, &req); err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}
	if err = h.Validator.Driver.Struct(req); err != nil {
		h.SendRequestValidationError(w, err.(validator.ValidationErrors))
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

	// Model db transaction
	m := model.Contract{App: h.App}
	tx, err := db.Begin(ctx)
	if err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}

	// Response
	response := map[string]interface{}{
		"member_code": code,
	}

	if req.Password != req.RetypePassword {
		h.SendBadRequest(w, "Password is not match")
		tx.Rollback(ctx)
		return
	}
	err = m.UpdatePassword(tx, ctx, h.GetChannel(r), code, req.Password)
	if err != nil {
		h.SendBadRequest(w, err.Error())
		tx.Rollback(ctx)
		return
	}

	// Commit transaction
	err = tx.Commit(ctx)
	if err != nil {
		h.SendBadRequest(w, err.Error())
		tx.Rollback(ctx)
		return
	}

	h.SendSuccess(w, response, nil)
}

// UpdateMemberPhoneAct is handler for handling update pass
func (h *Contract) UpdateMemberPhoneAct(w http.ResponseWriter, r *http.Request) {
	var err error

	// Check each request
	req := request.CheckPhoneReq{}
	if err = h.Bind(r, &req); err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}
	if err = h.Validator.Driver.Struct(req); err != nil {
		h.SendRequestValidationError(w, err.(validator.ValidationErrors))
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
	tx, err := db.Begin(ctx)
	if err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}

	member, _ := m.GetMemberByCode(db, ctx, code)
	if member.ID == 0 {
		h.SendNotfound(w, fmt.Sprintf("Member %s not found.", code))
		tx.Rollback(ctx)
		return
	}

	if err = m.ValidateToken(db, context.Background(), h.GetChannel(r), model.ActChangePhone, req.Phone, req.Token); err != nil {
		h.SendBadRequest(w, err.Error())
		tx.Rollback(ctx)
		return
	}

	phoneUpdated, err := m.UpdatePhoneMember(tx, ctx, req.Phone, code)
	if err != nil {
		h.SendBadRequest(w, err.Error())
		tx.Rollback(ctx)
		return
	}
	member.Phone = phoneUpdated

	// Response
	var res response.MemberResponse
	response := res.Transform(member)
	response.Token = req.Token

	err = tx.Commit(ctx)
	if err != nil {
		h.SendBadRequest(w, err.Error())
		tx.Rollback(ctx)
		return
	}

	h.SendSuccess(w, response, nil)
}

// Deleted member
func (h *Contract) DeleteMember(w http.ResponseWriter, r *http.Request) {
	var err error
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
	tx, err := db.Begin(ctx)
	if err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}

	data, _ := m.GetMemberDelByCode(db, ctx, code)
	if data.ID == 0 {
		h.SendNotfound(w, "Member not found.")
		tx.Rollback(ctx)
		return
	}
	data.IsActive = false

	// edit is active to false
	err = m.UpdateIsActive(tx, ctx, code, data)
	if err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}

	// Activity user logging in process
	log := model.LogActivityUserEnt{
		UserID:    int64(data.ID),
		Role:      h.GetUserRole(r.Context()),
		Title:     fmt.Sprintf("Delete %s", code),
		Activity:  fmt.Sprintf("Delete member %s", code),
		EventType: r.Method,
	}
	_, err = m.AddLogActivity(tx, ctx, log)
	if err != nil {
		h.SendBadRequest(w, err.Error())
		tx.Rollback(ctx)
		return
	}

	// Send Notifications - To User (Admin)
	players, err := m.GetListPlayerByUserCodeAndRole(db, ctx, "", "admin")
	if err != nil {
		h.SendBadRequest(w, err.Error())
		tx.Rollback(ctx)
		return
	}
	notifContent := model.NotificationContent{
		CustomerName: data.Name,
		Subject:      model.NOTIF_SUBJ_CUSTOMER_BANNED,
	}
	_, err = m.SendNotifications(tx, db, ctx, players, notifContent)
	if err != nil {
		h.SendBadRequest(w, psql.ParseErr(err))
		tx.Rollback(ctx)
		return
	}

	// Commit transaction
	err = tx.Commit(ctx)
	if err != nil {
		h.SendBadRequest(w, err.Error())
		tx.Rollback(ctx)
		return
	}

	h.SendSuccess(w, h.EmptyJSONArr(), nil)
}
