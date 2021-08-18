package handler

import (
	"context"
	"fmt"
	"net/http"
	"panorama/services/api/handler/request"
	"panorama/services/api/model"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/go-playground/validator/v10"
)

// AddSettingAct ...
func (h *Contract) AddSettingAct(w http.ResponseWriter, r *http.Request) {
	var err error
	req := request.SettingReq{}
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
	if err = m.AddSetting(db, ctx, model.SettingEnt{
		SetGroup:     req.Group,
		SetKey:       req.Key,
		SetLabel:     req.Label,
		ContentType:  req.SetType,
		ContentValue: req.SetContent,
	}); err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}

	h.SendSuccess(w, h.EmptyJSONArr(), nil)
}

// UpdateSettingAct ...
func (h *Contract) UpdateSettingAct(w http.ResponseWriter, r *http.Request) {
	var err error
	req := request.SettingReq{}
	if err = h.Bind(r, &req); err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}
	if err = h.Validator.Driver.Struct(req); err != nil {
		h.SendRequestValidationError(w, err.(validator.ValidationErrors))
		return
	}

	id := chi.URLParam(r, "id")
	id32, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		h.SendBadRequest(w, fmt.Errorf("%s", "invalid parameter").Error())
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
	if err = m.UpdateSetting(db, ctx, int32(id32), model.SettingEnt{
		SetGroup:     req.Group,
		SetKey:       req.Key,
		SetLabel:     req.Label,
		ContentType:  req.SetType,
		ContentValue: req.SetContent,
	}); err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}

	h.SendSuccess(w, h.EmptyJSONArr(), nil)
}
