package handler

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"net/http"
	"panorama/lib/psql"
	"panorama/lib/utils"
	"panorama/services/api/handler/request"
	"panorama/services/api/handler/response"
	"panorama/services/api/model"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

// RegisterAct add new member (cust) to databse
func (h *Contract) RegisterAct(w http.ResponseWriter, r *http.Request) {
	var err error
	var responseMessage string
	var res response.MemberResponse

	req := request.AddMemberReq{}
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

	// Check via name send token
	viaEmail := m.Via(req.Email)
	if viaEmail == "" {
		h.SendBadRequest(w, "email requests is invalid.")
		return
	}

	viaPhone := m.Via(req.Phone)
	if viaPhone == "" {
		h.SendBadRequest(w, "phone requests is invalid.")
		return
	}

	tx, err := db.Begin(ctx)
	if err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}
	defer db.Release()

	if req.Password != req.RetypePassword {
		h.SendBadRequest(w, "Password is not match")
		return
	}

	member, _ := m.GetMemberByEmailUsernamePhone(db, ctx, req.Email, req.Username, req.Phone)

	if member.ID != 0 {
		h.SendBadRequest(w, "user has been registered and need activated")
		return
	} else {
		member, err = m.AddMember(tx, context.Background(), model.MemberEnt{
			Name:     req.Name,
			Gender:   req.Gender,
			Username: req.Username,
			Email:    req.Email,
			Phone:    req.Phone,
			Password: req.Password,
			IsActive: false, // hardcode sementara sebelum implement otp
		})

		if err != nil {
			h.SendBadRequest(w, psql.ParseErr(err))
			return
		}

		err = m.AddNewLogVisitApp(tx, ctx, member.ID, "customer")
		if err != nil {
			tx.Rollback(ctx)
			h.SendBadRequest(w, err.Error())
			return
		}

		// add member temporary to member itin relation
		mTemp, err := m.GetListMemberTemporaryByEmail(db, ctx, req.Email)
		if err != nil && err != sql.ErrNoRows {
			h.SendBadRequest(w, err.Error())
			return
		}

		if len(mTemp) > 0 {
			var arrstr []string
			for _, v := range mTemp {
				arrstr = append(arrstr, "("+strconv.Itoa(int(v.MemberItinID))+","+strconv.Itoa(int(member.ID))+",current_timestamp)")
			}
			err = m.AddMemberItinRelationBatch(ctx, tx, strings.Join(arrstr, ","))
			if err != nil {
				tx.Rollback(ctx)
				h.SendBadRequest(w, err.Error())
				return
			}

			err = m.DeleteMemberTempByEmail(ctx, tx, req.Email)
			if err != nil {
				tx.Rollback(ctx)
				h.SendBadRequest(w, err.Error())
				return
			}
		}

		// add chat member temporary to chat group relation
		chatMemberTemp, err := m.GetListChatMemberTempByEmail(db, ctx, req.Email)
		if err != nil && err != sql.ErrNoRows {
			h.SendBadRequest(w, err.Error())
			return
		}

		if len(chatMemberTemp) > 0 {
			var arrstr []string
			for _, v := range chatMemberTemp {
				arrstr = append(arrstr, "("+strconv.Itoa(int(member.ID))+","+strconv.Itoa(int(v.ChatGroupID))+",current_timestamp)")
			}
			err = m.AddChatGroupRelationBatch(ctx, tx, strings.Join(arrstr, ","))
			if err != nil {
				tx.Rollback(ctx)
				h.SendBadRequest(w, err.Error())
				return
			}

			err = m.DeleteChatMemberTempByEmail(ctx, tx, req.Email)
			if err != nil {
				tx.Rollback(ctx)
				h.SendBadRequest(w, err.Error())
				return
			}
		}
	}

	role := "customer"
	rand.Seed(time.Now().UnixNano())
	token, _ := utils.Generate(`[\d]{4}`)

	// 1. need send token for phone validations
	if _, err := m.SendToken(db, ctx, h.GetChannel(r), model.ActRegPhone, model.TokenViaPhone, req.Phone, role, token); err != nil {
		fmt.Printf("error send token phone: %s", err.Error())
	}

	// 2. need send token for email validations
	if _, err = m.SendToken(db, ctx, h.GetChannel(r), model.ActRegEmail, model.TokenViaEmail, req.Email, role, token); err != nil {
		fmt.Printf("error send token email: %s", err.Error())
	}

	res = res.Transform(member)
	res.Token = token

	// Commit transaction
	err = tx.Commit(ctx)
	if err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}

	h.SendSuccessCustomMsg(w, res, nil, responseMessage)
}

// AuthCheckTokenPhoneAct is a phone validation token for register customer
func (h *Contract) AuthCheckTokenPhoneAct(w http.ResponseWriter, r *http.Request) {
	var err error
	req := request.CheckPhoneReq{}
	if err := h.Bind(r, &req); err != nil {
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
	if err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}

	if err = m.ValidateToken(db, ctx, h.GetChannel(r), model.ActRegPhone, req.Phone, req.Token); err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}

	// activate and set the phone validation flag to true
	if err = m.ActivateAndSetPhoneValid(db, ctx, req.Phone); err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}

	h.SendSuccess(w, h.EmptyJSONArr(), nil)
}

// Login ...
func (h *Contract) LoginAct(w http.ResponseWriter, r *http.Request) {
	var err error
	var responseMessage string

	req := request.LoginReq{}
	if err := h.Bind(r, &req); err != nil {
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

	userTokenCredential, err := m.AuthLogin(db, ctx, h.GetChannel(r), req.Username, req.Password)
	if err != nil {
		h.SendAuthError(w, "invalid user or credential")
		return
	}

	// Define with cast user credential
	channel := h.GetChannel(r)
	xPlayerID := h.GetPlayer(r)
	userToken := userTokenCredential["token"].(string)
	userRole := userTokenCredential["user_role"].(string)
	userEmail := userTokenCredential["user_email"].(string)
	userPhone := userTokenCredential["user_phone"].(string)
	userCode := userTokenCredential["user_code"].(string)
	userName := userTokenCredential["user_name"].(string)
	userImage  := userTokenCredential["user_image"].(string)
	userStatus := userTokenCredential["user_status"].(bool)
	userID, err := strconv.Atoi(fmt.Sprintf("%v", userTokenCredential["user_id"]))
	if err != nil {
		h.SendBadRequest(w, err.Error())
		tx.Rollback(ctx)
		return
	}

	// Cek user non active
	var otp string
	if !userStatus {
		userToken = ""
		responseMessage = fmt.Sprintf("Account %s has been registered and need activated", userEmail)

		// 1. need send token for phone validations again only app
		if channel == model.ChannelCustApp {
			rand.Seed(time.Now().UnixNano())
			otp, _ = utils.Generate(`[\d]{4}`)

			if _, err = m.SendToken(db, ctx, channel, model.ActRegPhone, model.TokenViaPhone, userPhone, userRole, otp); err != nil {
				fmt.Printf("error send token phone: %s", err.Error())
			}
			// 2. need send token for email validations again both of channel app & cms
			if _, err = m.SendToken(db, ctx, channel, model.ActRegEmail, model.TokenViaEmail, userEmail, userRole, otp); err != nil {
				fmt.Printf("error send token email: %s", err.Error())
			}
		} else if channel == model.ChannelCMS {
			if otp, err = m.SendToken(db, ctx, channel, model.ActRegEmail, model.TokenViaEmail, userEmail, userRole, otp); err != nil {
				fmt.Printf("error send token email: %s", err.Error())
			}
		}
	}

	// Check x-player if device app
	if channel == model.ChannelCustApp && len(xPlayerID) <= 0 {
		h.SendBadRequest(w, "header X-Player is required for app channel")
		tx.Rollback(ctx)
		return
	}

	// Add player
	player, err := m.AddPlayer(tx, db, ctx, int64(userID), userRole, xPlayerID, channel)
	if err != nil {
		h.SendBadRequest(w, psql.ParseErr(err))
		tx.Rollback(ctx)
		return
	}

	// Commit process
	err = tx.Commit(ctx)
	if err != nil {
		h.SendBadRequest(w, err.Error())
		tx.Rollback(ctx)
		return
	}

	response := map[string]interface{}{
		"token":             userToken,
		"user_code":         userCode,
		"user_role":         userRole,
		"user_name":         userName,
		"user_image": 		 userImage,
		"user_status":       userStatus,
		"user_phone":        userPhone,
		"user_email":        userEmail,
		"user_otp":          otp,
		"user_xplayer":      player.PlayerID,
		"user_device_type":  player.DeviceType,
		"user_device_model": player.DeviceModel,
	}

	h.SendSuccessCustomMsg(w, response, nil, responseMessage)
}

// ForgotPassAct
func (h *Contract) ForgotPassAct(w http.ResponseWriter, r *http.Request) {
	var err error
	var responseMessage string

	req := request.AuthReq{}
	if err := h.Bind(r, &req); err != nil {
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

	// Check via name send token
	via := m.Via(req.Username)
	viaName := model.TokenViaEmail
	channel := h.GetChannel(r)
	if via == "" {
		h.SendBadRequest(w, "email / phone requests is invalid.")
		return
	}
	if channel == model.ChannelCMS && via != model.TokenViaEmail {
		h.SendBadRequest(w, "email is invalid.")
		return
	}
	if via == model.TokenViaPhone {
		viaName = model.TokenViaPhone
	}

	// Define user auth from different channel
	var userID int32
	var userStatus bool
	var userEmail, userCode, userPhone, userRole string

	if channel == model.ChannelCustApp {
		member, _ := m.GetMemberBy(db, ctx, viaName, req.Username)
		userID = member.ID
		userEmail = member.Email
		userCode = member.MemberCode
		userPhone = member.Phone
		userStatus = member.IsActive
		userRole = "customer"
	} else if channel == model.ChannelCMS {
		user, _ := m.GetUserBy(db, ctx, viaName, req.Username)
		userID = user.ID
		userEmail = user.Email
		userCode = user.UserCode
		userPhone = user.Phone
		userStatus = user.IsActive
		userRole = user.Role
	}
	if userID == 0 {
		h.SendBadRequest(w, "user not found.")
		return
	}

	// Generate token default
	var token string
	if channel == model.ChannelCustApp {
		rand.Seed(time.Now().UnixNano())
		token, _ = utils.Generate(`[\d]{4}`)
	}

	// Check user is non active
	if userID != 0 && !userStatus {
		// 1. need send token for phone validations again only app
		if channel == model.ChannelCustApp {
			if _, err = m.SendToken(db, ctx, channel, model.ActRegPhone, model.TokenViaPhone, userPhone, userRole, token); err != nil {
				fmt.Printf("error send token phone: %s", err.Error())
			}
			// 2. need send token for email validations again both of channel app & cms
			if _, err = m.SendToken(db, ctx, channel, model.ActRegEmail, model.TokenViaEmail, userEmail, userRole, token); err != nil {
				fmt.Printf("error send token email: %s", err.Error())
			}
		} else if channel == model.ChannelCMS {
			if token, err = m.SendToken(db, ctx, channel, model.ActRegEmail, model.TokenViaEmail, userEmail, userRole, token); err != nil {
				fmt.Printf("error send token email: %s", err.Error())
			}
		}
		responseMessage = fmt.Sprintf("Account %s has been registered and need activated", userEmail)
	} else {
		token, err = m.SendToken(db, ctx, h.GetChannel(r), model.ActForgotPass, viaName, req.Username, userRole, token)
		if err != nil {
			h.SendBadRequest(w, err.Error())
			return
		}
	}

	response := map[string]interface{}{
		"user_code":   userCode,
		"user_email":  userEmail,
		"user_status": userStatus,
		"user_phone":  userPhone,
		"token":       token,
	}

	h.SendSuccessCustomMsg(w, response, nil, responseMessage)
}

// AuthCheckTokenForgotPassAct validate the token that user input when they receive token
func (h *Contract) AuthCheckTokenForgotPassAct(w http.ResponseWriter, r *http.Request) {
	var err error
	req := request.CheckTokenReq{}
	if err := h.Bind(r, &req); err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}
	if err = h.Validator.Driver.Struct(req); err != nil {
		h.SendRequestValidationError(w, err.(validator.ValidationErrors))
		return
	}

	db, err := h.DB.Acquire(context.Background())
	if err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}
	defer db.Release()

	m := model.Contract{App: h.App}
	if err = m.ValidateToken(db, context.Background(), h.GetChannel(r), model.ActForgotPass, req.Username, req.Token); err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}

	h.SendSuccess(w, h.EmptyJSONArr(), nil)
}

// ForgotChangePassAct change password after token valid
func (h *Contract) ForgotChangePassAct(w http.ResponseWriter, r *http.Request) {
	var err error
	req := request.ChangePassForgotReq{}
	if err = h.Bind(r, &req); err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}
	if err = h.Validator.Driver.Struct(req); err != nil {
		h.SendRequestValidationError(w, err.(validator.ValidationErrors))
		return
	}
	if req.Password != req.RetypePassword {
		h.SendBadRequest(w, "Password is not match")
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

	username := req.Username
	channel := h.GetChannel(r)

	// Set username from the jwt token
	if channel == model.ChannelCMS {
		tokenDecoded, err := m.DecodeTokenJWT(username)
		if err != nil {
			h.SendBadRequest(w, err.Error())
			tx.Rollback(ctx)
			return
		}
		username = tokenDecoded["member_code"].(string)
	}

	if err = m.UpdatePassword(tx, ctx, channel, username, req.Password); err != nil {
		h.SendBadRequest(w, err.Error())
		tx.Rollback(ctx)
		return
	}

	err = tx.Commit(ctx)
	if err != nil {
		h.SendBadRequest(w, err.Error())
		tx.Rollback(ctx)
		return
	}

	h.SendSuccess(w, h.EmptyJSONArr(), nil)
}

// SendTokenAct is handler for send token auth
func (h *Contract) SendTokenAct(w http.ResponseWriter, r *http.Request) {
	// Check each request
	req := request.AuthReq{}
	if err := h.Bind(r, &req); err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}
	if err := h.Validator.Driver.Struct(req); err != nil {
		h.SendRequestValidationError(w, err.(validator.ValidationErrors))
		return
	}

	// Define code param
	typeToken := chi.URLParam(r, "type")
	if len(typeToken) == 0 {
		h.SendBadRequest(w, "invalid type param")
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
	m := model.Contract{App: h.App}

	// Check via name send token
	via := m.Via(req.Username)
	viaName := model.TokenViaEmail
	channel := h.GetChannel(r)
	if via == "" {
		h.SendBadRequest(w, "email / phone requests is invalid.")
		return
	}
	if channel == model.ChannelCMS && via != model.TokenViaEmail {
		h.SendBadRequest(w, "email is invalid.")
		return
	}
	if via == model.TokenViaPhone {
		viaName = model.TokenViaPhone
	}

	// Define user auth from different channel
	var userID int32
	var userStatus bool
	var userEmail, userCode, userPhone, userRole string
	member, _ := m.GetMemberBy(db, ctx, viaName, req.Username)
	if channel == model.ChannelCustApp {
		if typeToken != "change-phone" {
			userID = member.ID
			userEmail = member.Email
			userCode = member.MemberCode
			userPhone = req.Username
			userStatus = member.IsActive
		}
		userRole = "customer"
	} else {
		user, _ := m.GetUserBy(db, ctx, viaName, req.Username)
		userID = user.ID
		userEmail = user.Email
		userCode = user.UserCode
		userPhone = user.Phone
		userStatus = user.IsActive
		userRole = user.Role
	}

	if member.Phone == req.Username && typeToken == "change-phone" {
		h.SendBadRequest(w, "user already exists.")
		return
	}

	if userID == 0 && typeToken != "change-phone" {
		h.SendBadRequest(w, "user not found.")
		return
	}

	// Response
	response := map[string]interface{}{
		"username":    req.Username,
		"user_code":   userCode,
		"user_email":  userEmail,
		"user_status": userStatus,
		"user_phone":  userPhone,
	}

	// Generate token default
	var token string
	if channel == model.ChannelCustApp {
		rand.Seed(time.Now().UnixNano())
		token, _ = utils.Generate(`[\d]{4}`)
	}

	// Act type
	if typeToken == "register" && userID != 0 && !userStatus {
		if channel == model.ChannelCustApp {
			if _, err = m.SendToken(db, ctx, channel, model.ActRegPhone, model.TokenViaPhone, userPhone, userRole, token); err != nil {
				fmt.Printf("error send token phone: %s", err.Error())
			}
			if _, err = m.SendToken(db, ctx, channel, model.ActRegEmail, model.TokenViaEmail, userEmail, userRole, token); err != nil {
				fmt.Printf("error send token email: %s", err.Error())
			}
		} else if channel == model.ChannelCMS {
			if token, err = m.SendToken(db, ctx, channel, model.ActRegEmail, model.TokenViaEmail, userEmail, userRole, token); err != nil {
				fmt.Printf("error send token email: %s", err.Error())
			}
		}
	} else {
		actType := model.ActChangePass
		if typeToken == "phone" {
			actType = model.ActChangePhone
		}

		if typeToken == "change-phone" {
			actType = model.ActChangePhone
		}

		if typeToken == "pass" {
			actType = model.ActChangePass
		}

		// Verification token
		token, err = m.SendToken(db, ctx, h.GetChannel(r), actType, viaName, req.Username, userRole, token)
		if err != nil {
			h.SendBadRequest(w, fmt.Sprintf("error send token %s: %s", viaName, err.Error()))
			return
		}
	}

	response["token"] = token

	h.SendSuccess(w, response, nil)
}
