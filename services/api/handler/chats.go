package handler

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"panorama/lib/agora"
	"panorama/lib/array"
	"panorama/lib/psql"
	"panorama/lib/utils"
	"panorama/services/api/handler/request"
	"panorama/services/api/handler/response"
	"panorama/services/api/model"

	rtctokenbuilder "github.com/AgoraIO/Tools/DynamicKey/AgoraDynamicKey/go/src/RtcTokenBuilder"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

// ChatAct ...
func (h *Contract) ChatAct(w http.ResponseWriter, r *http.Request) {
	token, err := agora.Contract{
		AppID:   h.Config.GetString("agora.app_id"),
		AppCert: h.Config.GetString("agora.app_cert"),
	}.GenerateRtcToken(1, "test channel", rtctokenbuilder.RoleAdmin)
	if err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}

	h.SendSuccess(w, map[string]interface{}{"token": token}, nil)
}

// CreateChatGroup ...
func (h *Contract) CreateChatGroup(w http.ResponseWriter, r *http.Request) {

	var err error
	var tcID int32
	var mItin model.MemberItinEnt
	var mTemp, mRel []string
	var res response.ChatGroupRes

	req := request.NewChatGroupReq{}
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

	if h.GetUserRole(r.Context()) != "customer" {
		h.SendBadRequest(w, "Only customer cant create chat group")
		return
	}

	m := model.Contract{App: h.App}

	// get member id
	member, err := m.GetMemberBy(db, ctx, "member_code", h.GetUserCode(r.Context()))
	if member.ID == 0 {
		h.SendNotfound(w, fmt.Sprintf("Customer %s not found.", h.GetUserCode(r.Context())))
		return
	}
	if err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}

	// find tc id yang plg sedikit orderannya
	if req.TcCode != "" {
		userTc, _ := m.GetUserByCode(db, ctx, req.TcCode)
		tcID = userTc.ID
	}
	if tcID == 0 {
		// tcID, _, err = m.GetTcIDLeastWork(db, ctx)
		// if err == sql.ErrNoRows {
		// 	h.SendBadRequest(w, "Tidak ada tc yang bisa di assign")
		// 	return
		// }
		// if err != nil {
		// 	h.SendBadRequest(w, err.Error())
		// 	return
		// }
	}

	// itin member id
	if len(req.MemberItinCode) > 0 {
		mItin, err = m.GetMemberItinByCode(db, ctx, req.MemberItinCode)
		if err != nil && err != sql.ErrNoRows {
			h.SendBadRequest(w, err.Error())
			return
		}
	}

	tx, err := db.Begin(ctx)
	if err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}

	// generat random code
	rand.Seed(time.Now().UnixNano())
	code, _ := utils.Generate(`CG-[a-z0-9]{8}`)

	cg, err := m.CreateChatGroup(tx, ctx, model.ChatGroupEnt{
		Member:        model.MemberEnt{ID: member.ID, Name: member.Name},
		MemberItin:    mItin,
		User:          model.UserEnt{ID: tcID},
		ChatGroupType: req.ChatGroupType,
		ChatGroupCode: code,
		Name:          req.Name,
		Status:        true,
	})
	if err != nil {
		h.SendBadRequest(w, psql.ParseErr(err))
		tx.Rollback(ctx)
		return
	}

	// process add member to chat group relation
	if len(req.ChatGroupRelation) > 0 {
		for _, v := range req.ChatGroupRelation {
			mbr, err := m.GetMemberByEmail(db, ctx, v)
			if err != nil && mbr.ID != 0 {
				h.SendBadRequest(w, err.Error())
				return
			}
			if mbr.ID <= 0 {
				mTemp = append(mTemp, "('"+v+"',"+strconv.Itoa(int(cg.ID))+",current_timestamp)")

			} else {
				mRel = append(mRel, "("+strconv.Itoa(int(mbr.ID))+","+strconv.Itoa(int(cg.ID))+",current_timestamp)")
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
		cg.ChatGroupRelation = req.ChatGroupRelation
	}

	// Activity user
	log := model.LogActivityUserEnt{
		UserID:    int64(member.ID),
		Role:      "customer",
		Title:     "Create new chat room",
		Activity:  "Room " + cg.Name,
		EventType: r.Method,
	}
	_, err = m.AddLogActivity(tx, ctx, log)
	if err != nil {
		h.SendBadRequest(w, err.Error())
		tx.Rollback(ctx)
		return
	}

	if tcID > 0 {

		// Activity user
		logtc := model.LogActivityUserEnt{
			UserID:    int64(tcID),
			Role:      "tc",
			Title:     "Assigned to a chat room",
			Activity:  "Client " + member.Name,
			EventType: r.Method,
		}
		_, err = m.AddLogActivity(tx, ctx, logtc)
		if err != nil {
			h.SendBadRequest(w, err.Error())
			tx.Rollback(ctx)
			return
		}

		// Send Notifications
		players, err := m.GetListPlayerByUserCodeAndRole(db, ctx, req.TcCode, "tc")
		if err != nil {
			h.SendBadRequest(w, err.Error())
			tx.Rollback(ctx)
			return
		}
		notifContent := model.NotificationContent{
			Subject:  model.NOTIF_SUBJ_CHAT_ROOM_ASSIGNED,
			RoomName: req.Name,
		}
		_, err = m.SendNotifications(tx, db, ctx, players, notifContent)
		if err != nil {
			h.SendBadRequest(w, psql.ParseErr(err))
			tx.Rollback(ctx)
			return
		}
	}

	// Commit transaction
	err = tx.Commit(ctx)
	if err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}

	h.SendSuccess(w, res.Transform(cg), nil)
}

// InviteTcToGroupChat ...
func (h *Contract) InviteTcToGroupChat(w http.ResponseWriter, r *http.Request) {

	var err error

	req := request.ChatGroupReq{}
	if err = h.Bind(r, &req); err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}
	if err = h.Validator.Driver.Struct(req); err != nil {
		h.SendRequestValidationError(w, err.(validator.ValidationErrors))
		return
	}

	// validation if tc create member itin for member
	if h.GetUserRole(r.Context()) == "tc" {
		h.SendBadRequest(w, "Only role customer can access")
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

	// find tc id yang plg sedikit orderannya
	idTc, name, err := m.GetTcIDLeastWork(db, ctx)
	if err == sql.ErrNoRows {
		h.SendBadRequest(w, "Tidak ada tc yang bisa di assign")
		return
	}
	if err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}

	tx, err := db.Begin(ctx)
	if err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}

	err = m.InviteTcToChatGroup(ctx, tx, idTc, req.ChatGroupCode)
	if err != nil {
		tx.Rollback(ctx)
		h.SendBadRequest(w, err.Error())
		return
	}

	member, _ := m.GetMemberBy(db, ctx, "member_code", h.GetUserCode(r.Context()))

	// Activity user
	log := model.LogActivityUserEnt{
		UserID:    int64(member.ID),
		Role:      "customer",
		Title:     "Invited a Travel Consultant",
		Activity:  "Assignee TC" + name,
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

	h.SendSuccess(w, h.EmptyJSONArr(), nil)
}

// ChatMessage ...
func (h *Contract) ChatMessage(w http.ResponseWriter, r *http.Request) {

	var err error
	var role, name, code string
	var idUser int32

	req := request.ChatGroupMessagesReq{}
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

	// get user id for message user id
	if h.GetUserRole(r.Context()) == "customer" {
		member, err := m.GetMemberBy(db, ctx, "member_code", h.GetUserCode(r.Context()))
		if member.ID == 0 {
			h.SendNotfound(w, fmt.Sprintf("Customer %s not found.", h.GetUserCode(r.Context())))
			return
		}
		if err != nil {
			h.SendBadRequest(w, err.Error())
			return
		}
		role = "customer"
		idUser = member.ID
		name = member.Name
		code = member.MemberCode

	} else if h.GetUserRole(r.Context()) == "tc" {
		user, err := m.GetUserByCode(db, ctx, h.GetUserCode(r.Context()))
		if user.ID == 0 {
			h.SendNotfound(w, fmt.Sprintf("Admin or Tc %s not found.", h.GetUserCode(r.Context())))
			return
		}
		if err != nil {
			h.SendBadRequest(w, err.Error())
			return
		}
		role = user.Role
		idUser = user.ID
		name = user.Name
		code = user.UserCode
	}

	// get chat group id
	groupID, err := m.GetIDGroupChatByCode(db, ctx, req.ChatGroupCode)
	if groupID == 0 {
		h.SendNotfound(w, fmt.Sprintf("Group code %s not found.", req.ChatGroupCode))
		return
	}
	if err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}

	tx, err := db.Begin(ctx)
	if err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}

	// insert message chat
	cm, err := m.CreateChatMessage(tx, ctx, model.ChatMessagesEnt{
		ChatGroupID: groupID,
		UserID:      idUser,
		Role:        role,
		Message:     req.Message,
		IsRead:      false,
		Name:        name,
		UserCode:    code,
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

	var res response.ChatGroupMessageRes
	res = res.Transform(cm)

	h.SendSuccess(w, res, nil)
}

// GetChatListAct List of chat ...
func (h *Contract) GetChatListAct(w http.ResponseWriter, r *http.Request) {
	role := h.GetUserRole(r.Context())
	userCode := h.GetUserCode(r.Context())

	if role == "admin" {
		h.SendUnAuthorizedData(w)
		return
	}

	param := map[string]interface{}{
		"keyword":   "",
		"page":      1,
		"limit":     10,
		"offset":    0,
		"sort":      "desc",
		"order":     "cg.created_date",
		"is_paging": "false",
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

	if role == "customer" {
		param["tc_code"] = ""
		param["member_code"] = userCode
	} else if role == "tc" {
		param["member_code"] = ""
		param["tc_code"] = userCode
	}

	m := model.Contract{App: h.App}
	ctx := context.Background()
	db, err := h.DB.Acquire(ctx)
	if err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}
	defer db.Release()

	chatList, err := m.GetChatList(db, ctx, param)
	if err != nil && sql.ErrNoRows != nil {
		h.SendBadRequest(w, err.Error())
		return
	}

	var listResponse []response.ChatGroupOrderRes
	for _, chat := range chatList {
		var res response.ChatGroupOrderRes
		res = res.Transform(chat)

		listResponse = append(listResponse, res)
	}

	h.SendSuccess(w, listResponse, param)
}

// GetHistoryChatByCode history of chat ...
func (h *Contract) GetHistoryChatByCode(w http.ResponseWriter, r *http.Request) {

	code := chi.URLParam(r, "code")
	role := h.GetUserRole(r.Context())
	userCode := h.GetUserCode(r.Context())

	param := map[string]interface{}{
		"page":   1,
		"limit":  10,
		"offset": 0,
		"sort":   "desc",
		"order":  "cm.id",
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

	if limit, ok := r.URL.Query()["limit"]; ok {
		if l, err := strconv.Atoi(limit[0]); err == nil {
			param["limit"] = l
		}
	}

	if offet, ok := r.URL.Query()["offset"]; ok {
		if l, err := strconv.Atoi(offet[0]); err == nil {
			param["offset"] = l
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

	// validasi user yang bukan bagian dari chat group
	id, err := m.IsExistInGroupChat(db, ctx, code, h.GetUserCode(r.Context()))
	if id <= 0 {
		h.SendBadRequest(w, "Access denied for get history chat")
		return
	}
	if err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}

	chatList, err := m.GetChatHistoryByGroupCode(db, ctx, code, param)
	if err != nil && chatList.ID != 0 {
		h.SendBadRequest(w, err.Error())
		return
	}

	// get member id
	var userID int32
	if role == "customer" {
		member, _ := m.GetMemberBy(db, ctx, "member_code", userCode)
		userID = member.ID
	} else if role == "tc" {
		tc, _ := m.GetUserByCode(db, ctx, userCode)
		userID = tc.ID
	}
	if userID == 0 {
		h.SendNotfound(w, fmt.Sprintf("User %s not found.", h.GetUserCode(r.Context())))
		return
	}
	if err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}

	// get list user in group chat
	listUser, err := m.GetListUserGroupChat(db, ctx, strconv.Itoa(int(chatList.ID)))
	if err != nil {
		fmt.Println("err")

		h.SendBadRequest(w, err.Error())
		return
	}
	chatList.ListUser = listUser

	if chatList.ID > 0 {

		cm, err := m.CheckOtherMessages(db, ctx, chatList.ID, userID)
		if err != nil && cm != 0 {
			h.SendBadRequest(w, err.Error())
			return
		}

		if cm > 0 {
			tx, err := db.Begin(ctx)
			if err != nil {
				h.SendBadRequest(w, err.Error())
				return
			}

			// update is read other chat
			err = m.UpdateIsRead(ctx, tx, chatList.ID, userID)
			if err != nil {
				tx.Rollback(ctx)
				h.SendBadRequest(w, err.Error())
				return
			}

			// Commit transaction
			err = tx.Commit(ctx)
			if err != nil {
				h.SendBadRequest(w, err.Error())
				return
			}
		}

	}

	var res response.ChatGroupHistoryRes
	res = res.Transform(chatList)

	h.SendSuccess(w, res, param)
}

// Leave Season - Update Status to False
func (h *Contract) LeaveSesonChatAct(w http.ResponseWriter, r *http.Request) {
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

	data, err := m.GetStatusLeaveSeason(db, ctx, code)
	if err == sql.ErrNoRows {
		h.SendNotfound(w, err.Error())
		return
	}
	if err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}

	// edit status to false
	data.Status = false
	err = m.UpdateStatus(tx, ctx, code, data)
	if err != nil {
		h.SendBadRequest(w, err.Error())
		tx.Rollback(ctx)
		return
	}

	tc, _ := m.GetUserByCode(db, ctx, h.GetUserCode(r.Context()))

	// Activity user
	log := model.LogActivityUserEnt{
		UserID:    int64(tc.ID),
		Role:      h.GetUserRole(r.Context()),
		Title:     "TC has leaved chat room",
		Activity:  "Chat Room " + data.Name,
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

// UpdateIsReadMessages ...
func (h *Contract) UpdateIsReadMessages(w http.ResponseWriter, r *http.Request) {

	var err error
	var userID int32

	req := request.ChatGroupMessagesIsRead{}
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
	if err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}

	// check group is exist
	data, err := m.GetIDGroupChatByCode(db, ctx, req.ChatGroupCode)
	if data <= 0 {
		h.SendNotfound(w, fmt.Sprintf("Group chat with code %s not found.", req.ChatGroupCode))
		return
	}
	if err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}

	if req.Role == "customer" {
		member, err := m.GetMemberBy(db, ctx, "member_code", req.UserCode)
		if member.ID <= 0 {
			h.SendNotfound(w, fmt.Sprintf("Member with code %s not found.", req.UserCode))
			return
		}
		if err != nil {
			h.SendBadRequest(w, err.Error())
			return
		}

		userID = member.ID

	} else {
		us, err := m.GetMemberBy(db, ctx, "user_code", req.UserCode)
		if us.ID <= 0 {
			h.SendNotfound(w, fmt.Sprintf("User with code %s not found.", req.UserCode))
			return
		}
		if err != nil {
			h.SendBadRequest(w, err.Error())
			return
		}

		userID = us.ID
	}

	// edit is read message by user id
	err = m.UpdateIsReadMessageByUserID(tx, ctx, userID, data)
	if err != nil {
		tx.Rollback(ctx)
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

	h.SendSuccess(w, h.EmptyJSONArr(), nil)
}
