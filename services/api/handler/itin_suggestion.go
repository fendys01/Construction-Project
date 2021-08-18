package handler

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"panorama/lib/array"
	"panorama/services/api/handler/request"
	"panorama/services/api/handler/response"
	"panorama/services/api/model"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v4/pgxpool"
)

// GetSugItinAct ...
func (h *Contract) GetSugItinAct(w http.ResponseWriter, r *http.Request) {
	// Check db context
	ctx := context.Background()
	db, err := h.DB.Acquire(ctx)
	if err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}
	defer db.Release()

	code := chi.URLParam(r, "code")

	if len(code) > 0 {
		// Model db transaction
		m := model.Contract{App: h.App}
		tx, err := db.Begin(ctx)
		if err != nil {
			h.SendBadRequest(w, err.Error())
			return
		}

		s, err := m.GetSugItinByCode(db, ctx, code)
		if err != nil {
			h.SendBadRequest(w, err.Error())
			tx.Rollback(ctx)
			return
		}

		if h.GetUserRole(r.Context()) == "customer" {
			s.View.Int32 = s.View.Int32 + 1
			err = m.UpdateSugItin(tx, ctx, s, code)
			if err != nil {
				h.SendBadRequest(w, err.Error())
				tx.Rollback(ctx)
				return
			}
		}

		// Commit transaction
		err = tx.Commit(ctx)
		if err != nil {
			h.SendBadRequest(w, err.Error())
			tx.Rollback(ctx)
			return
		}

		var res response.DetailItinSugResponse
		res = res.Transform(s)

		h.SendSuccess(w, res, nil)
		return
	}

	h.SendSuccess(w, h.EmptyJSONArr(), nil)
}

// AddSugItinAct add new suggestion itinerary
func (h *Contract) AddSugItinAct(w http.ResponseWriter, r *http.Request) {
	req := request.SugItinReq{}
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

	// Model db transaction
	m := model.Contract{App: h.App}
	tx, err := db.Begin(ctx)
	if err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}

	// Assign tc id
	tcCode := h.GetUserCode(r.Context())
	userTc, _ := m.GetUserByCode(db, ctx, tcCode)
	if userTc.ID == 0 {
		h.SendNotfound(w, fmt.Sprintf("User Admin %s not found.", tcCode))
		tx.Rollback(ctx)
		return
	}

	// Create suggestion itin
	sugItinReq, _ := req.ToSugItinEnt(true)
	sugItin, err := m.AddSugItin(tx, ctx, sugItinReq, userTc.ID)
	if err != nil {
		h.SendBadRequest(w, err.Error())
		tx.Rollback(ctx)
		return
	}

	// Process the tags
	if err = processTags(db, ctx, m, sugItin.ID, req); err != nil {
		h.SendBadRequest(w, err.Error())
		tx.Rollback(ctx)
		return
	}

	// Activity user logging in process
	log := model.LogActivityUserEnt{
		UserID:    int64(userTc.ID),
		Role:      userTc.Role,
		Title:     sugItin.Title,
		Activity:  "Add New Suggest Itin",
		EventType: r.Method,
	}
	_, err = m.AddLogActivity(tx, ctx, log)
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

	var res response.DetailItinSugResponse
	res = res.Transform(sugItin)

	h.SendSuccess(w, res, nil)
}

func (h *Contract) UpdateSugItinAct(w http.ResponseWriter, r *http.Request) {
	var err error

	code := chi.URLParam(r, "code")

	req := request.SugItinReq{}
	if err = h.Bind(r, &req); err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}
	if err = h.Validator.Driver.Struct(req); err != nil {
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

	// Model db transaction
	m := model.Contract{App: h.App}
	tx, err := db.Begin(ctx)
	if err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}

	// Assign tc id
	tcCode := h.GetUserCode(r.Context())
	userTc, _ := m.GetUserByCode(db, ctx, tcCode)
	if userTc.ID == 0 {
		h.SendNotfound(w, fmt.Sprintf("User Admin %s not found.", tcCode))
		tx.Rollback(ctx)
		return
	}

	sug, err := m.GetSugItinByCode(db, ctx, code)
	if err != nil {
		h.SendBadRequest(w, err.Error())
		tx.Rollback(ctx)
		return
	}

	sugItin := req.Transform(sug)
	sugItin.Destination = req.Destination
	sugItin.ItinCode = code
	err = m.UpdateSugItin(tx, ctx, sugItin, code)
	if err != nil {
		h.SendBadRequest(w, err.Error())
		tx.Rollback(ctx)
		return
	}

	// remove all tags that related with the itenerary
	sugID, err := m.GetSugItinID(db, ctx, code)
	if err != nil {
		h.SendBadRequest(w, err.Error())
		tx.Rollback(ctx)
		return
	}
	err = m.DelSugItinTagsBySugItinID(tx, ctx, sugID)
	if err != nil {
		h.SendBadRequest(w, err.Error())
		tx.Rollback(ctx)
		return
	}

	// process the tags
	if err = processTags(db, ctx, m, sugID, req); err != nil {
		h.SendBadRequest(w, err.Error())
		tx.Rollback(ctx)
		return
	}

	// Activity user logging in process
	log := model.LogActivityUserEnt{
		UserID:    int64(userTc.ID),
		Role:      userTc.Role,
		Title:     sugItin.Title,
		Activity:  fmt.Sprintf("Update Suggest Itin %s", code),
		EventType: r.Method,
	}
	_, err = m.AddLogActivity(tx, ctx, log)
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

	var res response.DetailItinSugResponse
	res = res.Transform(sugItin)

	h.SendSuccess(w, res, nil)
}

func (h *Contract) DelSugItinAct(w http.ResponseWriter, r *http.Request) {
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

	// Assign tc id
	tcCode := h.GetUserCode(r.Context())
	userTc, _ := m.GetUserByCode(db, ctx, tcCode)
	if userTc.ID == 0 {
		h.SendNotfound(w, fmt.Sprintf("User Admin %s not found.", tcCode))
		tx.Rollback(ctx)
		return
	}

	err = m.DelSugItin(tx, ctx, code)
	if err != nil {
		h.SendBadRequest(w, err.Error())
		tx.Rollback(ctx)
		return
	}

	// Activity user logging in process
	log := model.LogActivityUserEnt{
		UserID:    int64(userTc.ID),
		Role:      userTc.Role,
		Title:     fmt.Sprintf("Delete %s", code),
		Activity:  fmt.Sprintf("Delete Suggest Itin %s", code),
		EventType: r.Method,
	}
	_, err = m.AddLogActivity(tx, ctx, log)
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

func processTags(db *pgxpool.Conn, ctx context.Context, m model.Contract, sugID int32, req request.SugItinReq) error {
	var tagsID []int32

	tx, err := db.Begin(ctx)
	if err != nil {
		return err
	}

	if len(req.NewTags) > 0 {
		tagsID = m.AddMutiTag(tx, ctx, req.NewTags)
	}

	// add tags
	if len(req.Tags) > 0 {
		for _, t := range req.Tags {
			tagsID = append(tagsID, int32(t))
		}
	}

	// save to suggestion itinerary tags
	if err := m.AddMultiSugItinTags(tx, ctx, sugID, tagsID); err != nil {
		return fmt.Errorf("%s", "something error when adding the tags. But the itinerary has been added.")
	}

	return nil
}

// GetItinSugList ...
func (h *Contract) GetItinSugList(w http.ResponseWriter, r *http.Request) {

	param := map[string]interface{}{
		"keyword":    "",
		"page":       1,
		"limit":      10,
		"offset":     0,
		"sort":       "desc",
		"order":      "itin_suggestions.id",
		"created_by": "false",
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

	if c, ok := r.URL.Query()["created_by"]; ok && c[0] == "true" {
		param["created_by"] = h.GetUserCode(r.Context())
	} else {
		param["created_by"] = ""
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
	members, err := m.GetListItinSug(db, ctx, param)
	if err != nil && sql.ErrNoRows != nil {
		h.SendBadRequest(w, err.Error())
		return
	}
	var listResponse []response.ItinSugResponse
	for _, a := range members {
		var res response.ItinSugResponse
		res = res.Transform(a)
		listResponse = append(listResponse, res)
	}

	h.SendSuccess(w, listResponse, param)

}
