package handler

import (
	"context"
	"database/sql"
	"net/http"
	"panorama/lib/psql"
	"panorama/services/api/handler/request"
	"panorama/services/api/handler/response"
	"panorama/services/api/model"

	"github.com/go-chi/chi/v5"
)

//AddCitcallLogs log handler otp from citcall provider
func (h *Contract) AddCitcallLogs(w http.ResponseWriter, r *http.Request) {
	citcallType := chi.URLParam(r, "type")

	// Check db context
	ctx := context.Background()
	db, err := h.DB.Acquire(ctx)
	if err != nil {
		h.SendSuccess(w, err.Error(), nil)
		return
	}
	defer db.Release()

	// Model db transaction
	m := model.Contract{App: h.App}
	tx, err := db.Begin(ctx)
	if err != nil {
		h.SendSuccess(w, err.Error(), nil)
		return
	}

	// Formatting to log by type
	log := model.CallLogEnt{}
	if citcallType == model.TYPE_CALL {
		req := request.CitCallLogCallReq{}
		if err := h.Bind(r, &req); err != nil {
			h.SendSuccess(w, err.Error(), nil)
			return
		}
		log, _ = request.CitCallLog(req.TrxID, citcallType, req.Price, req)
	} else {
		req := request.CitCallLogSMSOrOTPReq{}
		if err := h.Bind(r, &req); err != nil {
			h.SendSuccess(w, err.Error(), nil)
			return
		}
		log, _ = request.CitCallLog(req.TrxID, citcallType, req.Price, req)
	}

	// Save call log
	_, err = m.AddCallLog(tx, ctx, log)
	if err != nil {
		h.SendSuccess(w, psql.ParseErr(err), nil)
		tx.Rollback(ctx)
		return
	}

	// Commit transaction
	err = tx.Commit(ctx)
	if err != nil {
		h.SendSuccess(w, log.Payloads, nil)
		tx.Rollback(ctx)
		return
	}

	h.SendSuccess(w, log.Payloads, nil)
}

// GetLogsList ...
func (h *Contract) GetLogsList(w http.ResponseWriter, r *http.Request) {
	param := map[string]interface{}{
		"trx_id":    "",
		"call_type": "",
		"provider":  "",
	}

	if trx_id, ok := r.URL.Query()["trx_id"]; ok && len(trx_id[0]) > 0 {
		param["trx_id"] = trx_id[0]
	}

	if call_type, ok := r.URL.Query()["call_type"]; ok && len(call_type[0]) > 0 {
		param["call_type"] = call_type[0]
	}

	if provider, ok := r.URL.Query()["provider"]; ok && len(provider[0]) > 0 {
		param["provider"] = provider[0]
	}

	ctx := context.Background()
	db, err := h.DB.Acquire(ctx)
	if err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}

	defer db.Release()

	m := model.Contract{App: h.App}
	logs, err := m.GetListCallLogs(db, ctx, param)
	if err != nil && sql.ErrNoRows != nil {
		h.SendBadRequest(w, err.Error())
		return
	}

	var listResponse []response.CallLogResponse
	for _, l := range logs {
		var res response.CallLogResponse
		res = res.Transform(l)

		listResponse = append(listResponse, res)
	}

	h.SendSuccess(w, listResponse, param)
}
