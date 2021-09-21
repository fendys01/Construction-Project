package handler

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"panorama/services/api/handler/response"
	"panorama/services/api/model"
)

// Get Dashboard
func (h *Contract) GetDashboardAct(w http.ResponseWriter, r *http.Request) {
	role := h.GetUserRole(r.Context())

	if role != "admin" && role != "tc" {
		h.SendUnAuthorizedData(w)
		return
	}

	param := map[string]interface{}{
		"start_date": "",
		"end_date":   "",
		"created_by": "false",
	}

	if start_date, ok := r.URL.Query()["start_date"]; ok && len(start_date[0]) > 0 {
		parseStartTime, _ := time.Parse("2006-01-02", start_date[0])
		param["start_date"] = parseStartTime.Format("2006-01-02")
	}

	if end_date, ok := r.URL.Query()["end_date"]; ok && len(end_date[0]) > 0 {
		parseEndTime, _ := time.Parse("2006-01-02", end_date[0])
		param["end_date"] = parseEndTime.Format("2006-01-02")
	}

	if param["start_date"] != "" && param["end_date"] != "" {
		parseStartTime, _ := time.Parse("2006-01-02", param["start_date"].(string))
		parseEndTime, _ := time.Parse("2006-01-02", param["end_date"].(string))
		if parseStartTime.After(parseEndTime) {
			h.SendBadRequest(w, "Start date should not be more end date")
			return
		}
	}

	m := model.Contract{App: h.App}
	ctx := context.Background()
	db, err := h.DB.Acquire(ctx)
	if err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}
	defer db.Release()

	var result interface{}
	if role == "admin" {
		var res response.DashboardAdminResponse
		dashboard, err := m.GetDashboardAdmin(db, ctx, param)
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
		result = res.Transform(dashboard)
	} else if role == "tc" {
		var res response.DashboardTCResponse

		if c, ok := r.URL.Query()["created_by"]; ok && c[0] == "true" {
			createdByCode := h.GetUserCode(r.Context())
			createdBy, _ := m.GetUserByCode(db, ctx, createdByCode)
			if createdBy.ID == 0 {
				h.SendNotfound(w, fmt.Sprintf("User %s not found.", createdByCode))
				return
			}
			param["created_by"] = strconv.Itoa(int(createdBy.ID))
		} else {
			param["created_by"] = ""
		}

		dashboard, err := m.GetDashboardTc(db, ctx, param)
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
		result = res.Transform(dashboard)
	}

	h.SendSuccess(w, result, nil)
}
