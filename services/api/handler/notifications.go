package handler

import (
	"context"
	"database/sql"
	"net/http"
	"panorama/lib/array"
	"panorama/services/api/handler/response"
	"panorama/services/api/model"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
)

// GetListNotifAct List notification
func (h *Contract) GetListNotifAct(w http.ResponseWriter, r *http.Request) {
	userCode := h.GetUserCode(r.Context())

	param := map[string]interface{}{
		"keyword":   "",
		"page":      1,
		"limit":     10,
		"offset":    0,
		"sort":      "desc",
		"order":     "n.id",
		"is_paging": "false",
	}

	param["user_code"] = userCode

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

	if c, ok := r.URL.Query()["is_paging"]; ok && c[0] == "true" {
		param["is_paging"] = true
		param["offset"] = (param["page"].(int) - 1) * param["limit"].(int)
		if page, ok := r.URL.Query()["page"]; ok && len(page[0]) > 0 {
			if p, err := strconv.Atoi(page[0]); err == nil && p > 1 {
				param["page"] = p
			}
		}
	} else {
		param["is_paging"] = false
		if offet, ok := r.URL.Query()["offset"]; ok {
			if l, err := strconv.Atoi(offet[0]); err == nil {
				param["offset"] = l
			}
		}
		param["page"] = ""
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

	// Get list notification by params
	notifications, err := m.GetListNotification(db, ctx, param)
	if err != nil && sql.ErrNoRows != nil {
		h.SendBadRequest(w, err.Error())
		return
	}
	// Fetch formatting response list notification from query
	var result []response.NotifResponse
	for _, notif := range notifications {
		var res response.NotifResponse
		result = append(result, res.Transform(notif))
	}

	h.SendSuccess(w, result, param)
}

// GetNotifAct Get detail notif
func (h *Contract) GetNotifAct(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")

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

	// Get detail notif
	notif, err := m.GetNotificationByCode(db, ctx, code)
	if err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}

	// Update flag read notif
	notif.IsRead = true
	_, err = m.UpdateNotificationByCode(tx, ctx, notif, code)
	if err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}

	// Commit transaction
	err = tx.Commit(ctx)
	if err != nil {
		h.SendBadRequest(w, err.Error())
		tx.Rollback(ctx)
		return
	}

	// Set Response
	var res response.NotifResponse
	response := res.Transform(notif)

	h.SendSuccess(w, response, nil)
}

// GetCounterNotifAct Counter total notif user
func (h *Contract) GetCounterNotifAct(w http.ResponseWriter, r *http.Request) {
	userCode := h.GetUserCode(r.Context())

	param := map[string]interface{}{
		"is_read": "false",
	}
	param["is_read"] = false
	if c, ok := r.URL.Query()["is_read"]; ok && c[0] == "true" {
		param["is_read"] = true
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

	// Get counter notif
	counterNotif, err := m.GetCounterNotifByUserCode(db, ctx, userCode, param["is_read"].(bool))
	if err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}

	h.SendSuccess(w, counterNotif, param)
}
