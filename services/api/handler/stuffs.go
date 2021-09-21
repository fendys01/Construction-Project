package handler

import (
	"context"
	"database/sql"
	"net/http"
	"panorama/lib/array"
	"panorama/services/api/handler/request"
	"panorama/services/api/handler/response"
	"panorama/services/api/model"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

// AddStuff ...
func (h *Contract) AddStuffAct(w http.ResponseWriter, r *http.Request) {

	var err error
	req := request.AddStuffReq{}
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
	stuff, err := m.AddStuff(db, ctx, model.StuffEnt{
		Name:  		  req.Name,
		Image:        sql.NullString{String: req.Image, Valid: true},
		Description:  req.Description,
		Price:     	  req.Price,
		Type:         req.Type,
		CreatedDate:  time.Time{},
	})
	if err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}

	// Commit transaction
	err = tx.Commit(ctx)
	if err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}

	var res response.StuffResponse
	res = res.Transform(stuff)

	h.SendSuccess(w, res, nil)
}

// GetStuffList ...
func (h *Contract) GetListStuffAct(w http.ResponseWriter, r *http.Request) {

	param := map[string]interface{}{
		"stuff": "",
		"type": "",
		"page":    1,
		"limit":   10,
		"offset":  0,
		"sort":    "desc",
		"order":   "s.id",
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
	
	if keyword, ok := r.URL.Query()["stuff"]; ok && len(keyword[0]) > 0 {
		param["stuff"] = keyword[0]
	}

	if keyword, ok := r.URL.Query()["type"]; ok && len(keyword[0]) > 0 {
		param["type"] = keyword[0]
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
	stuff, err := m.GetListStuff(db, ctx, param)
	if err != nil && sql.ErrNoRows != nil {
		h.SendBadRequest(w, err.Error())
		return
	}

	var listResponse []response.StuffListResponse
	for _, a := range stuff {
		var res response.StuffListResponse
		res = res.Transform(a)
		listResponse = append(listResponse, res)
	}

	h.SendSuccess(w, listResponse, param)
}

func (h *Contract) GeDetailStuffAct(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	db, err := h.DB.Acquire(ctx)
	if err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}
	defer db.Release()

	m := model.Contract{App: h.App}
	code := chi.URLParam(r, "code")
	if len(code) > 0 {
		s, err := m.GetStuffCode(db, ctx, code)
		if err != nil {
			h.SendBadRequest(w, err.Error())
			return
		}

		var res response.StuffDetailResponse
		res = res.Transform(s)

		h.SendSuccess(w, res, nil)
		return
	}

	h.SendSuccess(w, h.EmptyJSONArr(), nil)
}

func (h *Contract) UpdateDataStuffAct(w http.ResponseWriter, r *http.Request) {
	var err error
	req := request.StuffReqUpdate{}
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
	tx, err := db.Begin(ctx)
	if err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}

	data, err := m.GetStuffByCode(db, ctx, code)
	if err == sql.ErrNoRows {
		h.SendNotfound(w, err.Error())
		return
	}
	if err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}

	// Commit transaction
	err = tx.Commit(ctx)
	if err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}

	var res response.StuffResponse
	res = res.Transform(data)

	h.SendSuccess(w, res, nil)
}

// Deleted Stuff
func (h *Contract) DeleteStuffAct(w http.ResponseWriter, r *http.Request) {
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

	dataStuff, err := m.GetIsActiveStuff(db, ctx, code)
	if err == sql.ErrNoRows {
		h.SendNotfound(w, err.Error())
		return
	}
	if err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}

	// edit status to false
	dataStuff.IsActive = false
	err = m.UpdateIsActiveData(tx, ctx, code, dataStuff)
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

	h.SendSuccess(w, h.EmptyJSONArr(), nil)
}
