package handler

import (
	"context"
	"database/sql"
	"net/http"

	"panorama/services/api/handler/response"
	"panorama/services/api/model"
)

// Get Dashboard Admin
func (h *Contract) GetDashboardAdminAct(w http.ResponseWriter, r *http.Request) {
	param := map[string]interface{}{
		"start_date": "",
		"end_date":   "",
	}
	
	if start_date, ok := r.URL.Query()["start_date"]; ok && len(start_date[0]) > 0 {
		param["start_date"] = start_date[0]
	}

	if end_date, ok := r.URL.Query()["end_date"]; ok && len(end_date[0]) > 0 {
		param["end_date"] = end_date[0]
	}

	ctx := context.Background()
	db, err := h.DB.Acquire(ctx)
	if err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}

	defer db.Release()

	m := model.Contract{App: h.App}
	dashboard, err := m.GetDashboardAdmin(db, ctx)
	if err != nil && sql.ErrNoRows != nil {
		h.SendBadRequest(w, err.Error())
		return
	}

	log, err := m.GetDailyVisitsAct(db, ctx, param)
	if err != nil && err == sql.ErrNoRows {
		h.SendBadRequest(w, err.Error())
		return
	}

	dashboard.DailyVisitsEnt = log
	var res response.DashboardAdminResponse
	res = res.Transform(dashboard)

	h.SendSuccess(w, res, nil)
}

// Get All Dashboard TC
func (h *Contract) GetDashboardTcAct(w http.ResponseWriter, r *http.Request) {
	param := map[string]interface{}{
		"start_date": "",
		"end_date":   "",
	}

	if start_date, ok := r.URL.Query()["start_date"]; ok && len(start_date[0]) > 0 {
		param["start_date"] = start_date[0]
	}

	if end_date, ok := r.URL.Query()["end_date"]; ok && len(end_date[0]) > 0 {
		param["end_date"] = end_date[0]
	}

	ctx := context.Background()
	db, err := h.DB.Acquire(ctx)
	if err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}

	defer db.Release()

	m := model.Contract{App: h.App}
	dashboard, err := m.GetDashboardTc(db, ctx)
	if err != nil && sql.ErrNoRows != nil {
		h.SendBadRequest(w, err.Error())
		return
	}

	log, err := m.GetDailyVisitsAct(db, ctx, param)
	if err != nil && err == sql.ErrNoRows {
		h.SendBadRequest(w, err.Error())
		return
	}

	dashboard.DailyVisitsEnt = log
	var res response.DashboardTCResponse
	res = res.Transform(dashboard)

	h.SendSuccess(w, res, nil)
}