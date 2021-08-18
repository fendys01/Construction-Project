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
	"time"

	"github.com/go-playground/validator/v10"
)

// AddMemberAct ...
func (h *Contract) AddPushNotifAct(w http.ResponseWriter, r *http.Request) {

	var err error
	req := request.AddNotifReq{}
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
	notification, err := m.AddNotif(db, ctx, model.NotificationEnt{
		MemberCode:  req.MemberCode,
		Type:        req.Type,
		Title:       req.Title,
		Content:     req.Content,
		Link:        req.Link,
		CreatedDate: time.Time{},
	})
	if err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}

	var res response.NotifResponse
	res = res.Transform(notification)

	h.SendSuccess(w, res, nil)
}

// GetListNotifAct List notification
func (h *Contract) GetListNotifAct(w http.ResponseWriter, r *http.Request) {
	// Check db context
	ctx := context.Background()
	db, err := h.DB.Acquire(ctx)
	if err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}
	defer db.Release()
	m := model.Contract{App: h.App}

	param := map[string]interface{}{
		"member_code": "",
		"keyword":     "",
		"page":        1,
		"limit":       10,
		"offset":      0,
		"sort":        "desc",
		"order":       "n.created_date",
		"created_by":  "false",
	}

	if member_code, ok := r.URL.Query()["member_code"]; ok && len(member_code[0]) > 0 {
		param["member_code"] = member_code[0]
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
		memberCode := h.GetUserCode(r.Context())
		member, _ := m.GetMemberByCode(db, ctx, memberCode)
		if member.ID == 0 {
			h.SendNotfound(w, fmt.Sprintf("Member %s not found.", memberCode))
			return
		}
		memberID := strconv.Itoa(int(member.ID))
		param["created_by"] = memberID
	} else {
		param["created_by"] = ""
	}

	if limit, ok := r.URL.Query()["limit"]; ok {
		if l, err := strconv.Atoi(limit[0]); err == nil {
			param["limit"] = l
		}
	}

	param["offset"] = (param["page"].(int) - 1) * param["limit"].(int)

	// Get list notification by params
	notifications, err := m.GetListNotification(db, ctx, param)
	if err != nil && sql.ErrNoRows != nil {
		h.SendBadRequest(w, err.Error())
		return
	}

	// Fetch formatting response list notification from query
	var result []response.NotifItinResponse
	for _, notif := range notifications {
		var res response.NotifItinResponse
		result = append(result, res.Transform(notif))
	}

	h.SendSuccess(w, result, param)
}
