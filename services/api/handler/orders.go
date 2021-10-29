package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"panorama/lib/array"
	"panorama/lib/payment"
	"panorama/lib/psql"
	"panorama/services/api/handler/request"
	"panorama/services/api/handler/response"
	"panorama/services/api/model"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

func (h *Contract) GetDetailItinOrderMember(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	db, err := h.DB.Acquire(ctx)
	if err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}
	defer db.Release()

	m := model.Contract{App: h.App}
	code := chi.URLParam(r, "code")

	s, err := m.GetOrderByCode(db, ctx, code)
	if err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}

	var res response.DetailOrderMemberResponse
	res = res.Transform(s)

	h.SendSuccess(w, res, nil)
}

// GetItinOrderMember ...
func (h *Contract) GetListItinOrderMember(w http.ResponseWriter, r *http.Request) {
	param := map[string]interface{}{
		"page":        1,
		"keyword":    "",
		"limit":       10,
		"offset":      0,
		"sort":        "desc",
		"order":       "o.id",
		"order_code":  "",
		"member_code": "",
		"order_type": "",
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

	if limit, ok := r.URL.Query()["limit"]; ok {
		if l, err := strconv.Atoi(limit[0]); err == nil {
			param["limit"] = l
		}
	}

	if keyword, ok := r.URL.Query()["keyword"]; ok && len(keyword[0]) > 0 {
		param["keyword"] = keyword[0]
	}

	if order_code, ok := r.URL.Query()["order_code"]; ok && len(order_code[0]) > 0 {
		param["order_code"] = order_code[0]
	}

	if member_code, ok := r.URL.Query()["member_code"]; ok && len(member_code[0]) > 0 {
		param["member_code"] = member_code[0]
	}

	if order_type, ok := r.URL.Query()["order_type"]; ok && len(order_type[0]) > 0 {
		param["order_type"] = order_type[0]
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
	orders, err := m.GetListItinOrderMember(db, ctx, param)
	if err != nil && sql.ErrNoRows != nil {
		h.SendBadRequest(w, err.Error())
		return
	}

	var listResponse []response.ItinOrderMemberResponse
	for _, a := range orders {
		var res response.ItinOrderMemberResponse
		res = res.Transform(a)
		listResponse = append(listResponse, res)
	}

	h.SendSuccess(w, listResponse, param)
}

// AddOrderAct add / update (tc itin member) order from itinerary
func (h *Contract) AddOrderAct(w http.ResponseWriter, r *http.Request) {
	// Initial response handler
	var res response.ItinOrderMemberResponse
	var orderSetter model.OrderEnt

	// Binding request
	req := request.OrderReq{}
	if err := h.Bind(r, &req); err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}

	// Validate request of struct request
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

	// get chat id
	chat, _ := m.GetGroupChatByCode(db, ctx, req.ChatGroupCode)
	if chat.ID == 0 {
		h.SendNotfound(w, fmt.Sprintf("Room Chat %s not found.", req.ChatGroupCode))
		tx.Rollback(ctx)
		return
	}

	// Assign member paid by into order setter
	member, _ := m.GetMemberByCode(db, ctx, req.PaidByCode)
	if member.ID == 0 {
		h.SendNotfound(w, fmt.Sprintf("Member paid by %s not found.", req.PaidByCode))
		tx.Rollback(ctx)
		return
	}

	// Assign tc into order setter
	tcCode := h.GetUserCode(r.Context())
	userTc, _ := m.GetUserByCode(db, ctx, tcCode)
	if userTc.ID == 0 {
		h.SendNotfound(w, fmt.Sprintf("User TC %s not found.", tcCode))
		tx.Rollback(ctx)
		return
	}

	// Order setter request
	orderSetter.ChatID = chat.ID
	orderSetter.TotalPrice = int64(req.TotalPrice)
	orderSetter.TotalPricePpn = int64(req.TotalPricePPN)
	orderSetter.Description = req.Description
	orderSetter.OrderStatus = model.ORDER_STATUS_PENDING
	orderSetter.Title = req.Title
	orderSetter.PaidBy = member.ID
	orderSetter.MemberEnt = member
	orderSetter.TcID = userTc.ID
	orderSetter.UserEnt = userTc
	orderSetter.Details = req.Details
	orderSetter.OrderType = req.OrderType
	orderSetter.OrderCode = m.SetOrderCode()

	// Create or Order
	orderSaved, err := m.AddOrder(tx, ctx, orderSetter)
	if err != nil {
		h.SendBadRequest(w, err.Error())
		tx.Rollback(ctx)
		return
	}

	// Activity user logging in process
	log := model.LogActivityUserEnt{
		UserID:    int64(userTc.ID),
		Role:      h.GetUserRole(r.Context()),
		Title:     "Add New Order",
		Activity:  fmt.Sprintf("Add New Order Trip Itin For %s", member.Name),
		EventType: r.Method,
	}
	_, err = m.AddLogActivity(tx, ctx, log)
	if err != nil {
		h.SendBadRequest(w, err.Error())
		tx.Rollback(ctx)
		return
	}

	// Send Notifications - To Member (Customer)
	memberPlayers, err := m.GetListPlayerByUserCodeAndRole(db, ctx, member.MemberCode, "customer")
	if err != nil {
		h.SendBadRequest(w, err.Error())
		tx.Rollback(ctx)
		return
	}
	notifContentMember := model.NotificationContent{
		Subject:   model.NOTIF_SUBJ_ORDER_INCOME,
		TripName:  req.Title,
		OrderCode: orderSaved.OrderCode,
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
		TripName:      req.Title,
		StatusPayment: model.PAYMENT_STATUS_PROCESS_DESC,
	}
	_, err = m.SendNotifications(tx, db, ctx, userPlayers, notifContentUser)
	if err != nil {
		h.SendBadRequest(w, psql.ParseErr(err))
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

	h.SendSuccess(w, res.Transform(orderSaved), nil)
}

// UpdateOrderAct update payment process order from itinerary (tc)
func (h *Contract) UpdateOrderAct(w http.ResponseWriter, r *http.Request) {
	// Initial response handler
	code := chi.URLParam(r, "code")

	// Binding request
	req := request.OrderReqUpdate{}
	if err := h.Bind(r, &req); err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}

	// Validate request of struct request
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

	// Check paid by order
	memberCode := req.PaidByCode
	member, _ := m.GetMemberByCode(db, ctx, memberCode)
	if member.ID == 0 {
		h.SendNotfound(w, fmt.Sprintf("Member %s not found.", memberCode))
		tx.Rollback(ctx)
		return
	}

	// Check order data exist
	orderExist, _ := m.GetOrderByCodeDetail(db, ctx, code)
	if orderExist.ID == 0 {
		h.SendNotfound(w, fmt.Sprintf("Order %s not found.", code))
		tx.Rollback(ctx)
		return
	}

	checkPay, err := m.GetPaymentOrderByOrderID(db, ctx, orderExist.ID)
	if checkPay.ID > 1 {
		h.SendNotfound(w, fmt.Sprintf("payment order %s must be paid.", code))
		tx.Rollback(ctx)
		return
	}

	// transform request
	order := req.Transform(orderExist)

	order, err = m.UpdateOrderByCode(tx, ctx, order)
	if err != nil {
		h.SendBadRequest(w, err.Error())
		tx.Rollback(ctx)
		return
	}

	// Activity user logging in process
	log := model.LogActivityUserEnt{
		UserID:    int64(member.ID),
		Role:      h.GetUserRole(r.Context()),
		Title:     "Update Order",
		Activity:  fmt.Sprintf("Update Order Payment Trip Itin For %s", member.Name),
		EventType: r.Method,
	}
	_, err = m.AddLogActivity(tx, ctx, log)
	if err != nil {
		h.SendBadRequest(w, err.Error())
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
		TripName:      orderExist.MemberItin.Title,
		StatusPayment: model.PAYMENT_STATUS_PROCESS_METHOD_DESC,
	}
	_, err = m.SendNotifications(tx, db, ctx, userPlayers, notifContentUser)
	if err != nil {
		h.SendBadRequest(w, psql.ParseErr(err))
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

	h.SendSuccess(w, nil, nil)
}

// AddMidtransNotificationAct update payment process from midtrans notification
func (h *Contract) AddMidtransNotificationAct(w http.ResponseWriter, r *http.Request) {
	// Binding request
	req := request.OrderPaymentMidtransNotification{}
	if err := h.Bind(r, &req); err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}

	// Validate request of struct request
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

	// Check order data exist
	order, _ := m.GetOrderByOrderCode(db, ctx, req.OrderID)
	if order.ID == 0 {
		h.SendNotfound(w, fmt.Sprintf("Order %s not found.", req.OrderID))
		tx.Rollback(ctx)
		return
	}

	// Check order payment data exist
	orderPayment, _ := m.GetPaymentOrderByOrderID(db, ctx, order.ID)
	if orderPayment.ID == 0 {
		h.SendNotfound(w, fmt.Sprintf("Order payment %s not found.", order.OrderCode))
		tx.Rollback(ctx)
		return
	}

	// Adjust order & order payment status
	paymentService := payment.New(h.App)
	midtransStatus := paymentService.GetMidtransStatus(req.PaymentType, req.TransactionStatus, req.FraudStatus)
	orderStatus := midtransStatus["order_status"]
	paymentStatus := midtransStatus["payment_status"]

	// Update order status
	orderSetter := model.OrderEnt{
		OrderCode:     order.OrderCode,
		PaidBy:        order.PaidBy,
		OrderStatus:   orderStatus,
		TotalPrice:    order.TotalPrice,
		TcID:          order.TcID,
		OrderType:     order.OrderType,
		Details:       order.Details,
		Title:         order.Title,
		ChatID:        order.ChatID,
		Description:   order.Description,
		TotalPricePpn: order.TotalPricePpn,
	}
	orderUpdated, err := m.UpdateOrderByCode(tx, ctx, orderSetter)
	if err != nil {
		h.SendBadRequest(w, err.Error())
		tx.Rollback(ctx)
		return
	}

	// Setter payment Or update order payment
	orderPaymentSetter := model.OrderPaymentEnt{
		OrderID:       orderUpdated.ID,
		PaymentType:   req.PaymentType,
		Amount:        orderPayment.Amount,
		PaymentStatus: paymentStatus,
		PaymentURL:    orderPayment.PaymentURL,
		ExpiredDate:   orderPayment.ExpiredDate,
	}
	if req.TransactionStatus == payment.MIDTRANS_TRANSACTION_STATUS_PENDING {
		orderPaymentSetter.ExpiredDate = time.Now().Add(time.Minute * payment.DURATION_EXPIRED).In(time.UTC)
	}

	// Assign payload from midtrans webhook notification
	encode, _ := json.Marshal(req)
	resArray := make(map[string]interface{})
	err = json.Unmarshal(encode, &resArray)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		tx.Rollback(ctx)
		return
	}
	orderPaymentSetter.Payloads = resArray

	// Processing payment Or Update Order Payment
	_, err = m.UpdateOrderPayment(tx, ctx, orderPaymentSetter, orderUpdated.ID)
	if err != nil {
		h.SendBadRequest(w, err.Error())
		tx.Rollback(ctx)
		return
	}

	// Send Notifications
	if paymentStatus != model.PAYMENT_STATUS_PROCESS {
		// Send Notifications - Assign subject with payment status
		var subjectCust string
		var subjectTC string
		var itinTitle string
		var chatRoomName string
		var paymentStatusDesc string

		if paymentStatus == model.PAYMENT_STATUS_PAID {
			subjectCust = model.NOTIF_SUBJ_ORDER_VERIF
			subjectTC = model.NOTIF_SUBJ_ORDER_CLIENT_COMPLETE
			paymentStatusDesc = model.PAYMENT_STATUS_PAID_DESC
		} else if paymentStatus == model.PAYMENT_STATUS_CANCEL {
			subjectCust = model.NOTIF_SUBJ_ORDER_FAIL
			subjectTC = model.NOTIF_SUBJ_ORDER_CLIENT_FAIL
			paymentStatusDesc = model.PAYMENT_STATUS_CANCEL_DESC
		}

		// Send Notifications - To Member (Customer)
		itinGroups, err := m.GetListMemberItinRelationByItinID(db, ctx, order.MemberItinID)
		if err != nil {
			h.SendBadRequest(w, err.Error())
			tx.Rollback(ctx)
			return
		}
		if len(itinGroups) > 0 && len(subjectCust) > 0 {
			for _, itinRelation := range itinGroups {
				role := "customer"
				itinTitle = itinRelation.MemberItinEnt.Title
				chatRoomName = itinRelation.ChatGroup.Name
				memberPlayers, err := m.GetListPlayerByUserCodeAndRole(db, ctx, itinRelation.MemberEnt.MemberCode, role)
				if err != nil {
					h.SendBadRequest(w, err.Error())
					tx.Rollback(ctx)
					return
				}
				notifContentMember := model.NotificationContent{
					Subject:       subjectCust,
					TripName:      itinTitle,
					OrderCode:     order.OrderCode,
					PaymentMethod: orderPaymentSetter.PaymentType,
				}
				_, err = m.SendNotifications(tx, db, ctx, memberPlayers, notifContentMember)
				if err != nil {
					h.SendBadRequest(w, psql.ParseErr(err))
					tx.Rollback(ctx)
					return
				}
			}
		}

		// Send Notifications - To User (TC)
		role := "tc"
		tcPlayers, err := m.GetListPlayerByUserCodeAndRole(db, ctx, order.UserEnt.UserCode, role)
		if err != nil {
			h.SendBadRequest(w, err.Error())
			tx.Rollback(ctx)
			return
		}
		notifContentTC := model.NotificationContent{
			Subject:    subjectTC,
			RoomName:   chatRoomName,
			ClientName: order.MemberEnt.Name,
			OrderCode:  order.OrderCode,
		}
		_, err = m.SendNotifications(tx, db, ctx, tcPlayers, notifContentTC)
		if err != nil {
			h.SendBadRequest(w, psql.ParseErr(err))
			tx.Rollback(ctx)
			return
		}

		// Send Notifications - To User (Admin, TC)
		if len(paymentStatusDesc) > 0 {
			userPlayers, err := m.GetListPlayerByUserCodeAndRole(db, ctx, "", "")
			if err != nil {
				h.SendBadRequest(w, err.Error())
				tx.Rollback(ctx)
				return
			}
			notifContentUser := model.NotificationContent{
				Subject:       model.NOTIF_SUBJ_ORDER_HISTORY,
				TripName:      itinTitle,
				StatusPayment: paymentStatusDesc,
			}
			_, err = m.SendNotifications(tx, db, ctx, userPlayers, notifContentUser)
			if err != nil {
				h.SendBadRequest(w, psql.ParseErr(err))
				tx.Rollback(ctx)
				return
			}
		}
	}

	// Commit transaction
	err = tx.Commit(ctx)
	if err != nil {
		h.SendBadRequest(w, err.Error())
		tx.Rollback(ctx)
		return
	}

	h.SendSuccess(w, req, nil)
}

// PostPaymentAct post payment process order (cust_app)
func (h *Contract) PostPaymentAct(w http.ResponseWriter, r *http.Request) {
	// Initial response handler
	var res response.OrderPaymentResponse

	// Binding request
	req := request.OrderPaymentReq{}
	if err := h.Bind(r, &req); err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}

	// Validate request of struct request
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

	// Check paid by order
	memberCode := h.GetUserCode(r.Context())
	member, _ := m.GetMemberByCode(db, ctx, memberCode)
	if member.ID == 0 {
		h.SendNotfound(w, fmt.Sprintf("Member %s not found.", memberCode))
		tx.Rollback(ctx)
		return
	}

	// Check order data exist
	orderExist, _ := m.GetOrderByOrderCode(db, ctx, req.OrderCode)
	if orderExist.ID == 0 {
		h.SendNotfound(w, fmt.Sprintf("Order %s not found.", req.OrderCode))
		tx.Rollback(ctx)
		return
	}

	// Check paid by order by request auth
	if member.ID != orderExist.PaidBy {
		h.SendNotfound(w, fmt.Sprintf("Order paid by %s is invalid.", memberCode))
		tx.Rollback(ctx)
		return
	}

	// Create order payment default set expired date
	orderPaymentSetter := model.OrderPaymentEnt{
		OrderID:       orderExist.ID,
		Amount:        int64(req.Amount),
		PaymentStatus: model.PAYMENT_STATUS_PROCESS,
	}

	// Update url snap url midtrans
	paymentService := payment.New(h.App)
	orderCode := orderExist.OrderCode
	orderAmount := int64(req.Amount)
	paramMidtrans := paymentService.SetMidtransParam(member.Email, member.Name, orderCode, orderAmount)
	midtransResponse, err := paymentService.GetMidtransPaymentURL(paramMidtrans)
	if err != nil {
		h.SendBadRequest(w, err.Error())
		tx.Rollback(ctx)
		return
	}
	if midtransResponse["redirect_url"] == "" {
		h.SendBadRequest(w, "failed to get snap url midtrans.")
		tx.Rollback(ctx)
		return
	}
	orderPaymentSetter.PaymentURL = midtransResponse["redirect_url"]

	// Check order payment exist then saved, if exist = renew payment
	orderPaymentSaved := model.OrderPaymentEnt{}
	orderPayment, _ := m.GetPaymentOrderByOrderID(db, ctx, orderExist.ID)
	if orderPayment.ID != 0 {
		orderPaymentSetter.PaymentType = orderPayment.PaymentType
		orderPaymentSetter.ExpiredDate = orderPayment.ExpiredDate
		orderPaymentSetter.Payloads = orderPayment.Payloads
		orderPaymentSaved, err = m.UpdateOrderPayment(tx, ctx, orderPaymentSetter, orderExist.ID)
		if err != nil {
			h.SendBadRequest(w, err.Error())
			tx.Rollback(ctx)
			return
		}
		orderPaymentSaved.CreatedDate = orderPayment.CreatedDate
		orderPaymentSaved.ExpiredDate = orderPayment.ExpiredDate
	} else {
		orderPaymentSaved, err = m.AddOrderPayment(db, ctx, tx, orderPaymentSetter)
		if err != nil {
			h.SendBadRequest(w, err.Error())
			tx.Rollback(ctx)
			return
		}
	}
	orderPaymentSaved.OrderCode = orderExist.OrderCode

	// Activity user logging in process
	log := model.LogActivityUserEnt{
		UserID:    int64(member.ID),
		Role:      h.GetUserRole(r.Context()),
		Title:     "Update Order",
		Activity:  fmt.Sprintf("Update Order Payment Trip Itin For %s", member.Name),
		EventType: r.Method,
	}
	_, err = m.AddLogActivity(tx, ctx, log)
	if err != nil {
		h.SendBadRequest(w, err.Error())
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
		TripName:      orderExist.MemberItin.Title,
		StatusPayment: model.PAYMENT_STATUS_PROCESS_METHOD_DESC,
	}
	_, err = m.SendNotifications(tx, db, ctx, userPlayers, notifContentUser)
	if err != nil {
		h.SendBadRequest(w, psql.ParseErr(err))
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

	h.SendSuccess(w, res.Transform(orderPaymentSaved), nil)
}
