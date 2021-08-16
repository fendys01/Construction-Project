package handler

import (
	"Contruction-Project/lib/array"
	"Contruction-Project/services/api/handler/request"
	"Contruction-Project/services/api/handler/response"
	"Contruction-Project/services/api/model"
	"context"
	"database/sql"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"
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

	m := model.Contract{App: h.App}
	member, err := m.AddMember(db, ctx, model.MemberEnt{
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
	// return single user (get by user code)
	u, err := m.GetMemberByCode(db, ctx, code)
	if err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}

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
		h.SendBadRequest(w, err.Error())
		return
	}

	var res response.MemberResponse
	res = res.Transform(member)

	h.SendSuccess(w, res, nil)

}

// UpdateUserPass ...
func (h *Contract) UpdateMemberPass(w http.ResponseWriter, r *http.Request) {
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
	err = m.UpdateMemberPass(db, ctx, code, string(pass))
	if err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}

	h.SendSuccess(w, h.EmptyJSONArr(), nil)
}
