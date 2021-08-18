package handler

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"panorama/lib/agora"
	"panorama/lib/psql"
	"panorama/lib/utils"
	"panorama/services/api/handler/request"
	"panorama/services/api/handler/response"
	"panorama/services/api/model"

	rtctokenbuilder "github.com/AgoraIO/Tools/DynamicKey/AgoraDynamicKey/go/src/RtcTokenBuilder"
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
	var idTc int32
	var mItin model.MemberItinEnt

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

	if req.ChatWithTc {

		// find tc id yang plg sedikit orderannya
		idTc, err = m.GetTcIDLeastWork(db, ctx)
		if err == sql.ErrNoRows {
			h.SendBadRequest(w, "Tidak ada tc yang bisa di assign")
			return
		}
		if err != nil {
			h.SendBadRequest(w, err.Error())
			return
		}
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

	_, err = m.CreateChatGroup(tx, ctx, model.ChatGroupEnt{
		Member:        model.MemberEnt{ID: member.ID},
		MemberItin:    mItin,
		User:          model.UserEnt{ID: idTc},
		ChatGroupType: req.ChatGroupType,
		ChatGroupCode: code,
		Name:          req.Name,
	})
	if err != nil {
		h.SendBadRequest(w, psql.ParseErr(err))
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
	idTc, err := m.GetTcIDLeastWork(db, ctx)
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
	var role, name string
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
