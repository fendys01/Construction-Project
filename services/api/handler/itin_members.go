package handler

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	// "fmt"
	"net/http"
	"panorama/lib/array"
	"panorama/lib/psql"
	"panorama/lib/utils"
	"panorama/services/api/handler/request"
	"panorama/services/api/handler/response"
	"panorama/services/api/model"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	// "github.com/jackc/pgx/v4/pgxpool"
)

//GetMemberItAct
func (h *Contract) GetMemberItinAct(w http.ResponseWriter, r *http.Request) {
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
		s, err := m.GetMemberItinWithGroupsByCode(db, ctx, code)
		if err != nil {
			h.SendBadRequest(w, err.Error())
			return
		}

		var res response.ItinMemberResponse
		res = res.Transform(s)

		h.SendSuccess(w, res, nil)
		return
	}

	h.SendSuccess(w, h.EmptyJSONArr(), nil)
}

// GetItinMemberList ...
func (h *Contract) GetItinMemberList(w http.ResponseWriter, r *http.Request) {
	param := map[string]interface{}{
		"member_code": "",
		"keyword":     "",
		"page":        1,
		"limit":       10,
		"offset":      0,
		"sort":        "desc",
		"order":       "mi.created_date",
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

	if limit, ok := r.URL.Query()["limit"]; ok {
		if l, err := strconv.Atoi(limit[0]); err == nil {
			param["limit"] = l
		}
	}

	if c, ok := r.URL.Query()["created_by"]; ok && c[0] == "true" {
		param["created_by"] = h.GetUserCode(r.Context())
	} else {
		param["created_by"] = ""
	}

	param["offset"] = (param["page"].(int) - 1) * param["limit"].(int)

	m := model.Contract{App: h.App}
	ctx := context.Background()
	db, err := h.DB.Acquire(ctx)
	if err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}
	defer db.Release()

	members, err := m.GetListItinMember(db, ctx, param)
	if err != nil && sql.ErrNoRows != nil {
		h.SendBadRequest(w, err.Error())
		return
	}

	var listResponse []response.ItinMemberResponse
	for _, a := range members {
		var res response.ItinMemberResponse
		res = res.Transform(a)

		listResponse = append(listResponse, res)
	}

	h.SendSuccess(w, listResponse, param)
}

func (h *Contract) DelMemberItinAct(w http.ResponseWriter, r *http.Request) {
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

	// Assign created by member code member id
	memberCode := h.GetUserCode(r.Context())
	memberOwner, _ := m.GetMemberByCode(db, ctx, memberCode)
	if memberOwner.ID == 0 {
		h.SendNotfound(w, "Member not found.")
		tx.Rollback(ctx)
		return
	}

	err = m.DelMemberItin(tx, ctx, code)
	if err != nil {
		h.SendBadRequest(w, err.Error())
		tx.Rollback(ctx)
		return
	}

	// Activity user logging in process
	log := model.LogActivityUserEnt{
		UserID:    int64(memberOwner.ID),
		Role:      h.GetUserRole(r.Context()),
		Title:     fmt.Sprintf("Delete %s", code),
		Activity:  fmt.Sprintf("Delete Trip Itin %s", code),
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

// AddMemberItinAct add new member itinerary
func (h *Contract) AddMemberItinAct(w http.ResponseWriter, r *http.Request) {
	role := h.GetUserRole(r.Context())
	if role == "admin" {
		h.SendUnAuthorizedData(w)
		return
	}

	// Initial response handler
	var res response.ItinMemberResponse

	// Binding request
	req := request.MemberItinReq{}
	if err := h.Bind(r, &req); err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}

	// Validate request of struct request
	if err := h.Validator.Driver.Struct(req); err != nil {
		h.SendRequestValidationError(w, err.(validator.ValidationErrors))
		return
	}

	// Check request start date & end date
	if req.StartDate != "" && req.EndDate != "" {
		startDate, err := time.Parse("2006-01-02 15:04:05", req.StartDate)
		if err != nil {
			h.SendBadRequest(w, err.Error())
			return
		}
		endDate, err := time.Parse("2006-01-02 15:04:05", req.EndDate)
		if err != nil {
			h.SendBadRequest(w, err.Error())
			return
		}
		if startDate.After(endDate) {
			h.SendBadRequest(w, "Start date should not be more end date")
			return
		}
	}

	// Check db context
	ctx := context.Background()
	db, err := h.DB.Acquire(ctx)
	if err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}
	defer db.Release()

	// Formatting request member itin
	memberItinFormatted, err := req.ToMemberItinEnt(true)
	if err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}

	// Model db transaction
	m := model.Contract{App: h.App}
	tx, err := db.Begin(ctx)
	if err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}

	// Assign created by member code member id
	memberCode := h.GetUserCode(r.Context())
	if req.MemberCode != "" {
		memberCode = req.MemberCode
	}
	memberOwner, _ := m.GetMemberByCode(db, ctx, memberCode)
	if memberOwner.ID == 0 {
		h.SendNotfound(w, "Member not found.")
		tx.Rollback(ctx)
		return
	}
	memberItinFormatted.CreatedBy = memberOwner.ID

	// Set order type
	orderType := model.ORDER_TYPE_REGULER
	if req.OrderType != "" {
		if req.OrderType != model.ORDER_TYPE_REGULER && req.OrderType != model.ORDER_TYPE_CUSTOM {
			h.SendBadRequest(w, fmt.Sprintf("invalid order type %s or %s", model.ORDER_TYPE_REGULER, model.ORDER_TYPE_CUSTOM))
			tx.Rollback(ctx)
			return
		}
		orderType = req.OrderType
	}

	// Create member itin
	memberItinCreated, err := m.AddMemberItin(tx, ctx, memberItinFormatted, orderType)
	if err != nil {
		h.SendBadRequest(w, psql.ParseErr(err))
		tx.Rollback(ctx)
		return
	}
	memberItinCreated.MemberEnt = memberOwner
	userID := memberOwner.ID
	activityProcess := fmt.Sprintf("Add New Trip Itin %s", memberItinCreated.Title)

	// Assign order if user tc input member itin
	if role == "tc" {
		// Validation if tc create member itin for member
		if len(req.GroupChatCode) <= 0 {
			h.SendBadRequest(w, "Group chat code required")
			return
		}

		// Check TC input
		tcCode := h.GetUserCode(r.Context())
		userTc, _ := m.GetUserByCode(db, ctx, tcCode)
		if userTc.ID == 0 {
			h.SendNotfound(w, "User TC not found.")
			tx.Rollback(ctx)
			return
		}
		userID = userTc.ID
		activityProcess = fmt.Sprintf("Add New Trip Itin %s for %s", memberItinCreated.Title, memberOwner.Name)

		// Assign order itin from tc
		orderFormatted := model.OrderEnt{
			MemberItinID: memberItinCreated.ID,
			PaidBy:       memberItinCreated.CreatedBy,
			OrderCode:    m.SetOrderCode(),
			OrderStatus:  model.ORDER_STATUS_PENDING,
			TotalPrice:   memberItinCreated.EstPrice.Int64,
			OrderType:    orderType,
			TcID:         userTc.ID,
		}
		order, err := m.AddOrder(tx, ctx, orderFormatted)
		if err != nil {
			h.SendBadRequest(w, psql.ParseErr(err))
			tx.Rollback(ctx)
			return
		}

		// Assign member itin to chat group
		if len(req.GroupChatCode) > 0 {
			err = m.UpdateItinMemberToChat(ctx, tx, memberItinCreated.ID, userTc.ID, req.GroupChatCode)
			if err != nil {
				h.SendBadRequest(w, err.Error())
				tx.Rollback(ctx)
				return
			}
			memberItinCreated.ChatGroupCode = req.GroupChatCode
		}

		// Send Notifications - To Member (Customer)
		memberPlayers, err := m.GetListPlayerByUserCodeAndRole(db, ctx, memberOwner.MemberCode, "customer")
		if err != nil {
			h.SendBadRequest(w, err.Error())
			tx.Rollback(ctx)
			return
		}
		notifContentMember := model.NotificationContent{
			Subject:   model.NOTIF_SUBJ_ORDER_INCOME,
			TripName:  memberItinCreated.Title,
			OrderCode: order.OrderCode,
		}
		_, err = m.SendNotifications(tx, db, ctx, memberPlayers, notifContentMember)
		if err != nil {
			h.SendBadRequest(w, psql.ParseErr(err))
			tx.Rollback(ctx)
			return
		}

		// Send Notifications - To User (Admin, TC)
		userPlayers, err := m.GetListPlayerByUserCodeAndRole(db, ctx, "", "")
		if err != nil {
			h.SendBadRequest(w, err.Error())
			tx.Rollback(ctx)
			return
		}
		notifContentUser := model.NotificationContent{
			Subject:       model.NOTIF_SUBJ_ORDER_HISTORY,
			TripName:      memberItinCreated.Title,
			StatusPayment: model.PAYMENT_STATUS_PROCESS_DESC,
		}
		_, err = m.SendNotifications(tx, db, ctx, userPlayers, notifContentUser)
		if err != nil {
			h.SendBadRequest(w, psql.ParseErr(err))
			tx.Rollback(ctx)
			return
		}
	}

	// Assign member itin relation group member or assign member temporary
	memberItinGroups := []map[string]interface{}{}
	memberItinGroups = append(memberItinGroups, map[string]interface{}{
		"member_code":     memberOwner.MemberCode,
		"member_name":     memberOwner.Name,
		"member_username": memberOwner.Username,
		"member_email":    memberOwner.Email,
		"member_img":      memberOwner.Img.String,
		"is_owner":        true,
		"itin_code":       memberItinCreated.ItinCode,
	})
	if len(req.GroupMembers) > 0 {
		for _, groupMember := range req.GroupMembers {
			memberGroupEmail := fmt.Sprintf("%s", groupMember["member_email"])
			if memberGroupEmail != "" && memberOwner.Email != memberGroupEmail {
				if !utils.IsEmail(memberGroupEmail) {
					h.SendBadRequest(w, fmt.Sprintf("Email %s is invalid.", memberGroupEmail))
					tx.Rollback(ctx)
					return
				}
				memberGroup, _ := m.GetMemberByEmail(db, ctx, fmt.Sprintf("%v", memberGroupEmail))

				if memberGroup.ID != 0 {
					// Assign member itin relation group member
					memberItinRelationFormatted := model.MemberItinRelationEnt{
						MemberItinID: memberItinCreated.ID,
						MemberID:     memberGroup.ID,
					}
					memberItinRelationCreated, err := m.AddMemberItinRelation(tx, ctx, memberItinRelationFormatted)
					if err != nil {
						h.SendBadRequest(w, psql.ParseErr(err))
						tx.Rollback(ctx)
						return
					}
					memberItinRelationCreated.MemberEnt = memberGroup
					memberItinRelationCreated.MemberItinEnt = memberItinCreated
					memberItinGroups = append(memberItinGroups, map[string]interface{}{
						"member_code":     memberItinRelationCreated.MemberEnt.MemberCode,
						"member_name":     memberItinRelationCreated.MemberEnt.Name,
						"member_username": memberItinRelationCreated.MemberEnt.Username,
						"member_email":    memberItinRelationCreated.MemberEnt.Email,
						"member_img":      memberItinRelationCreated.MemberEnt.Img.String,
						"is_owner":        false,
						"itin_code":       memberItinCreated.ItinCode,
					})
				} else {
					// Assign member temporary
					memberTempFormatted := model.MemberTemporaryEnt{
						Email:        memberGroupEmail,
						MemberItinID: memberItinCreated.ID,
					}
					memberTempCreated, err := m.AddMemberTemporary(tx, ctx, memberTempFormatted)
					if err != nil {
						h.SendBadRequest(w, psql.ParseErr(err))
						tx.Rollback(ctx)
						return
					}
					memberTempCreated.MemberItin = memberItinCreated
					memberItinGroups = append(memberItinGroups, map[string]interface{}{
						"member_code":     "",
						"member_name":     "",
						"member_username": "",
						"member_email":    memberTempCreated.Email,
						"member_img":      "",
						"is_owner":        false,
						"itin_code":       memberItinCreated.ItinCode,
					})
					dataEmail := model.DataEmailInviteItinMember{
						Sender:        memberOwner.Name,
						URL:           "https://panoramatest.page.link/test",
						ItineraryName: memberItinCreated.Title,
						EmailInvite:   memberTempCreated.Email,
					}
					subject := fmt.Sprintf("[Panorama] Invitation Trip %s", dataEmail.ItineraryName)
					err = m.SendingMail(model.ActInviteGroupItinMember, subject, dataEmail.EmailInvite, dataEmail)
					if err != nil {
						fmt.Printf("error send email to %s : %s", memberTempCreated.Email, err.Error())
						tx.Rollback(ctx)
						return
					}
				}
			}
		}
	}
	if len(memberItinGroups) > 0 {
		memberItinCreated.GroupMembers = memberItinGroups
	}

	// Activity user logging in process
	log := model.LogActivityUserEnt{
		UserID:    int64(userID),
		Role:      h.GetUserRole(r.Context()),
		Title:     "Add New Itin",
		Activity:  activityProcess,
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

	h.SendSuccess(w, res.Transform(memberItinCreated), nil)
}

// UpdateMemberItinAct edit member itinerary
func (h *Contract) UpdateMemberItinAct(w http.ResponseWriter, r *http.Request) {

	var mTemp, mRel []string

	role := h.GetUserRole(r.Context())
	if role == "admin" {
		h.SendUnAuthorizedData(w)
		return
	}

	// Initial response handler and param code
	var res response.ItinMemberResponse
	code := chi.URLParam(r, "code")

	// Binding request
	req := request.MemberItinReq{}
	if err := h.Bind(r, &req); err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}

	// Validate request of struct request
	if err := h.Validator.Driver.Struct(req); err != nil {
		h.SendRequestValidationError(w, err.(validator.ValidationErrors))
		return
	}

	// validate requeired if want invite friend to itin
	if len(req.GroupMembers) > 0 {
		if len(req.GroupChatCode) <= 0 {
			h.SendBadRequest(w, "Group chat code required")
			return
		}
	}

	// Check request start date & end date
	if req.StartDate != "" && req.EndDate != "" {
		startDate, err := time.Parse("2006-01-02 15:04:05", req.StartDate)
		if err != nil {
			h.SendBadRequest(w, err.Error())
			return
		}
		endDate, err := time.Parse("2006-01-02 15:04:05", req.EndDate)
		if err != nil {
			h.SendBadRequest(w, err.Error())
			return
		}
		if startDate.After(endDate) {
			h.SendBadRequest(w, "Start date should not be more end date")
			return
		}
	}

	// Check db context
	ctx := context.Background()
	db, err := h.DB.Acquire(ctx)
	if err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}
	defer db.Release()

	// Formatting request member itin
	memberItinFormatted, err := req.ToMemberItinEnt(false)
	if err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}

	// Model db transaction
	m := model.Contract{App: h.App}
	tx, err := db.Begin(ctx)
	if err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}

	// Assign created by member code member id
	memberCode := h.GetUserCode(r.Context())
	if req.MemberCode != "" {
		memberCode = req.MemberCode
	}
	memberOwner, _ := m.GetMemberByCode(db, ctx, memberCode)
	if memberOwner.ID == 0 {
		h.SendNotfound(w, "Member not found.")
		tx.Rollback(ctx)
		return
	}

	// Check created by member itin existing
	memberItinExist, _ := m.GetMemberItinByCode(db, ctx, code)
	if memberItinExist.ID == 0 {
		h.SendNotfound(w, "Member itin not found.")
		tx.Rollback(ctx)
		return
	}
	if memberItinExist.CreatedBy != memberOwner.ID {
		h.SendNotfound(w, fmt.Sprintf("Itinerary %s member %s not found.", memberItinExist.ItinCode, memberOwner.Name))
		tx.Rollback(ctx)
		return
	}

	// Set order type
	orderType := model.ORDER_TYPE_REGULER
	if req.OrderType != "" {
		if req.OrderType != model.ORDER_TYPE_REGULER && req.OrderType != model.ORDER_TYPE_CUSTOM {
			h.SendBadRequest(w, fmt.Sprintf("invalid order type %s or %s", model.ORDER_TYPE_REGULER, model.ORDER_TYPE_CUSTOM))
			tx.Rollback(ctx)
			return
		}
		orderType = req.OrderType
	}

	// Update member itin
	memberItinUpdated, err := m.UpdateMemberItin(tx, ctx, memberItinFormatted, code)
	if err != nil {
		h.SendBadRequest(w, err.Error())
		tx.Rollback(ctx)
		return
	}
	memberItinUpdated.CreatedDate = memberItinExist.CreatedDate
	memberItinUpdated.MemberEnt = memberOwner
	memberItinUpdated.ItinCode = code
	userID := memberOwner.ID
	activityProcess := fmt.Sprintf("Update Trip Itin %s", memberItinUpdated.Title)

	// ReAssign order if user tc edit member itin
	if role == "tc" {
		// Validation if tc create member itin for member
		if len(req.GroupChatCode) <= 0 {
			h.SendBadRequest(w, "Group chat code required")
			return
		}

		// Check TC input
		tcCode := h.GetUserCode(r.Context())
		userTc, _ := m.GetUserByCode(db, ctx, tcCode)
		if userTc.ID == 0 {
			h.SendNotfound(w, "User TC not found.")
			tx.Rollback(ctx)
			return
		}
		userID = userTc.ID
		activityProcess = fmt.Sprintf("Update Trip Itin %s for %s", memberItinUpdated.Title, memberOwner.Name)

		// Assign order itin from tc
		orderFormatted := model.OrderEnt{
			PaidBy:      memberOwner.ID,
			OrderStatus: model.ORDER_STATUS_PENDING,
			TotalPrice:  memberItinUpdated.EstPrice.Int64,
			OrderType:   orderType,
			TcID:        userTc.ID,
		}
		updatedOrder, err := m.UpdateOrderByMemberItinID(tx, ctx, orderFormatted, memberItinUpdated.ID)
		if updatedOrder.ID == 0 {
			orderFormatted.MemberItinID = memberItinUpdated.ID
			orderFormatted.OrderCode = m.SetOrderCode()
			_, err = m.AddOrder(tx, ctx, orderFormatted)
			if err != nil {
				h.SendBadRequest(w, psql.ParseErr(err))
				tx.Rollback(ctx)
				return
			}
		}
		if err != nil {
			h.SendBadRequest(w, err.Error())
			tx.Rollback(ctx)
			return
		}

		// Assign member itin to chat group
		if len(req.GroupChatCode) > 0 {
			err = m.UpdateItinMemberToChat(ctx, tx, memberItinUpdated.ID, userTc.ID, req.GroupChatCode)
			if err != nil {
				h.SendBadRequest(w, err.Error())
				tx.Rollback(ctx)
				return
			}
		}
	}

	// Assign member itin relation group member or assign member temporary
	memberItinGroups := []map[string]interface{}{}
	memberItinGroups = append(memberItinGroups, map[string]interface{}{
		"member_code":     memberOwner.MemberCode,
		"member_name":     memberOwner.Name,
		"member_username": memberOwner.Username,
		"member_email":    memberOwner.Email,
		"member_img":      memberOwner.Img.String,
		"is_owner":        true,
		"itin_code":       memberItinUpdated.ItinCode,
	})
	if len(req.GroupMembers) > 0 {

		// get chat group id
		chatGroupID, err := m.GetIDGroupChatByCode(db, ctx, req.GroupChatCode)
		if err != nil && chatGroupID <= 0 {
			h.SendNotfound(w, fmt.Sprintf("Chat Group with code %s not found.", req.GroupChatCode))
			tx.Rollback(ctx)
			return
		}

		for _, groupMember := range req.GroupMembers {
			memberGroupEmail := fmt.Sprintf("%s", groupMember["member_email"])
			if memberGroupEmail != "" && memberOwner.Email != memberGroupEmail {
				if !utils.IsEmail(memberGroupEmail) {
					h.SendBadRequest(w, fmt.Sprintf("Email %s is invalid.", memberGroupEmail))
					tx.Rollback(ctx)
					return
				}
				memberGroup, _ := m.GetMemberByEmail(db, ctx, fmt.Sprintf("%v", memberGroupEmail))

				if memberGroup.ID != 0 {
					// Assign member itin relation group member
					memberRelationExist, _ := m.GetMemberItinRelationByMemberIDAndMemberItinID(db, ctx, memberGroup.ID, memberItinUpdated.ID)
					// Check relation itin member exist
					if memberRelationExist.ID == 0 {
						memberItinRelationFormatted := model.MemberItinRelationEnt{
							MemberItinID: memberItinUpdated.ID,
							MemberID:     memberGroup.ID,
						}
						memberItinRelationCreated, err := m.AddMemberItinRelation(tx, ctx, memberItinRelationFormatted)
						if err != nil {
							h.SendBadRequest(w, psql.ParseErr(err))
							tx.Rollback(ctx)
							return
						}
						memberItinRelationCreated.MemberEnt = memberGroup
						memberItinRelationCreated.MemberItinEnt = memberItinUpdated
						memberItinGroups = append(memberItinGroups, map[string]interface{}{
							"member_code":     memberItinRelationCreated.MemberEnt.MemberCode,
							"member_name":     memberItinRelationCreated.MemberEnt.Name,
							"member_username": memberItinRelationCreated.MemberEnt.Username,
							"member_email":    memberItinRelationCreated.MemberEnt.Email,
							"member_img":      memberItinRelationCreated.MemberEnt.Img.String,
							"is_owner":        false,
							"itin_code":       memberItinUpdated.ItinCode,
						})

						// append email member for query add to chat group relation
						mRel = append(mRel, "("+strconv.Itoa(int(memberGroup.ID))+","+strconv.Itoa(int(chatGroupID))+",current_timestamp)")
					} else {
						memberRelationExist.MemberEnt = memberGroup
						memberItinGroups = append(memberItinGroups, map[string]interface{}{
							"member_code":     memberRelationExist.MemberEnt.MemberCode,
							"member_name":     memberRelationExist.MemberEnt.Name,
							"member_username": memberRelationExist.MemberEnt.Username,
							"member_email":    memberRelationExist.MemberEnt.Email,
							"member_img":      memberRelationExist.MemberEnt.Img.String,
							"is_owner":        false,
							"itin_code":       memberItinUpdated.ItinCode,
						})
					}
				} else {
					// Assign member temporary
					// Check member temp exist
					memberTempExist, _ := m.GetMemberTemporaryByEmailAndItinID(db, ctx, memberGroupEmail, memberItinUpdated.ID)
					if memberTempExist.ID == 0 {
						memberTempFormatted := model.MemberTemporaryEnt{
							Email:        memberGroupEmail,
							MemberItinID: memberItinUpdated.ID,
						}
						memberTempCreated, err := m.AddMemberTemporary(tx, ctx, memberTempFormatted)
						if err != nil {
							h.SendBadRequest(w, psql.ParseErr(err))
							tx.Rollback(ctx)
							return
						}
						memberTempCreated.MemberItin = memberItinUpdated
						memberItinGroups = append(memberItinGroups, map[string]interface{}{
							"member_code":     "",
							"member_name":     "",
							"member_username": "",
							"member_email":    memberTempCreated.Email,
							"member_img":      "",
							"is_owner":        false,
							"itin_code":       memberItinUpdated.ItinCode,
						})
						dataEmail := model.DataEmailInviteItinMember{
							Sender:        memberOwner.Name,
							URL:           "https://panoramatest.page.link/test",
							ItineraryName: memberItinUpdated.Title,
							EmailInvite:   memberTempCreated.Email,
						}
						subject := fmt.Sprintf("[Panorama] Invitation Trip %s", dataEmail.ItineraryName)
						err = m.SendingMail(model.ActInviteGroupItinMember, subject, dataEmail.EmailInvite, dataEmail)
						if err != nil {
							fmt.Printf("error send email to %s : %s", memberTempCreated.Email, err.Error())
							tx.Rollback(ctx)
							return
						}

						// append email member for query add to chat group temporary
						mTemp = append(mTemp, "('"+memberTempCreated.Email+"',"+strconv.Itoa(int(chatGroupID))+",current_timestamp)")

					} else {
						memberTempExist.MemberItin = memberItinUpdated
						memberItinGroups = append(memberItinGroups, map[string]interface{}{
							"member_code":     "",
							"member_name":     "",
							"member_username": "",
							"member_email":    memberTempExist.Email,
							"member_img":      "",
							"is_owner":        false,
							"itin_code":       memberItinUpdated.ItinCode,
						})
					}

				}
			}
		}

		// add member temporary to chat member temporary
		if len(mTemp) > 0 {
			err = m.AddChatMemberTempBatch(ctx, tx, strings.Join(mTemp, ","))
			if err != nil {
				h.SendBadRequest(w, psql.ParseErr(err))
				tx.Rollback(ctx)
				return
			}
		}

		// add member is active to chat group relation
		if len(mRel) > 0 {
			err = m.AddChatGroupRelationBatch(ctx, tx, strings.Join(mRel, ","))
			if err != nil {
				h.SendBadRequest(w, psql.ParseErr(err))
				tx.Rollback(ctx)
				return
			}
		}

	}
	if len(memberItinGroups) > 0 {
		memberItinUpdated.GroupMembers = memberItinGroups
	}

	// Activity user logging in process
	log := model.LogActivityUserEnt{
		UserID:    int64(userID),
		Role:      h.GetUserRole(r.Context()),
		Title:     "Update Itin",
		Activity:  activityProcess,
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
		return
	}

	h.SendSuccess(w, res.Transform(memberItinUpdated), nil)
}
