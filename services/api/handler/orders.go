package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"panorama/lib/array"
	"panorama/lib/payment"
	"panorama/services/api/handler/request"
	"panorama/services/api/handler/response"
	"panorama/services/api/model"
	"reflect"
	"strconv"
	"strings"

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
	if len(code) > 0 {
		s, err := m.GetOrderByCode(db, ctx, code)
		if err != nil {
			h.SendBadRequest(w, err.Error())
			return
		}

		var res response.DetailOrderMemberResponse
		res = res.Transform(s)

		h.SendSuccess(w, res, nil)
		return
	}

	h.SendSuccess(w, h.EmptyJSONArr(), nil)
}

// GetItinOrderMember ...
func (h *Contract) GetListItinOrderMember(w http.ResponseWriter, r *http.Request) {
	param := map[string]interface{}{
		"page":    1,
		"limit":   10,
		"offset":  0,
		"sort":    "desc",
		"order":   "o.id",
		"member_code":  "",
		"expired_date": "",
		"order_code":   "",
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

	if member_code, ok := r.URL.Query()["member_code"]; ok && len(member_code[0]) > 0 {
		param["member_code"] = member_code[0]
	}

	if expired_date, ok := r.URL.Query()["expired_date"]; ok && len(expired_date[0]) > 0 {
		param["expired_date"] = expired_date[0]
	}

	if order_code, ok := r.URL.Query()["order_code"]; ok && len(order_code[0]) > 0 {
		param["order_code"] = order_code[0]
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

	// Order setter request
	orderSetter := model.OrderEnt{
		TotalPrice:  int64(req.TotalPrice),
		OrderStatus: model.ORDER_STATUS_PENDING,
		OrderType:   model.ORDER_TYPE_REGULER,
	}

	// Assign member itin order into order setter
	memberItin, _ := m.GetMemberItinByCode(db, ctx, req.MemberItinCode)
	if memberItin.ID == 0 {
		h.SendNotfound(w, fmt.Sprintf("Member itin %s not found.", req.MemberItinCode))
		tx.Rollback(ctx)
		return
	}
	memberItin.EstPrice = sql.NullInt64{Int64: int64(req.TotalPrice), Valid: true}
	orderSetter.MemberItinID = memberItin.ID
	orderSetter.MemberItin = memberItin

	// Assign member paid by into order setter
	member, _ := m.GetMemberByCode(db, ctx, req.PaidByCode)
	if member.ID == 0 {
		h.SendNotfound(w, fmt.Sprintf("Member paid by %s not found.", req.PaidByCode))
		tx.Rollback(ctx)
		return
	}
	orderSetter.PaidBy = member.ID
	orderSetter.MemberEnt = member

	// Assign tc into order setter
	tcCode := h.GetUserCode(r.Context())
	userTc, _ := m.GetUserByCode(db, ctx, tcCode)
	if userTc.ID == 0 {
		h.SendNotfound(w, fmt.Sprintf("User TC %s not found.", tcCode))
		tx.Rollback(ctx)
		return
	}
	orderSetter.TcID = userTc.ID
	orderSetter.UserEnt = userTc

	// Append additional detail from request into member itin
	orderAdditionalDetails := req.AdditionalDetails
	if len(orderAdditionalDetails) > 0 {

		var m map[string]interface{}
		md, err := json.Marshal(memberItin.Details[0])
		if err != nil {
			h.SendBadRequest(w, err.Error())
			tx.Rollback(ctx)
			return
		}

		err = json.Unmarshal(md, &m)
		if err != nil {
			h.SendBadRequest(w, err.Error())
			tx.Rollback(ctx)
			return
		}

		// TODO tambah key object sesuai kebutuhan frontend
		if reflect.ValueOf(orderAdditionalDetails[0]["hotel"]).IsValid() {
			m["hotel"] = orderAdditionalDetails[0]["hotel"]
		}

		memberItin.Details[0] = m
	}

	// Update member itin detail
	_, err = m.UpdateMemberItin(tx, ctx, memberItin, memberItin.ItinCode)
	if err != nil {
		h.SendBadRequest(w, err.Error())
		tx.Rollback(ctx)
		return
	}

	// Create or update if exist orders
	orderExist, _ := m.GetOrderByMemberItinID(db, ctx, memberItin.ID)
	orderSaved := model.OrderEnt{}
	if orderExist.ID != 0 {
		orderSaved, err = m.UpdateOrderByMemberItinID(tx, ctx, orderSetter, memberItin.ID)
		if err != nil {
			h.SendBadRequest(w, err.Error())
			tx.Rollback(ctx)
			return
		}
		orderSaved.OrderCode = orderExist.OrderCode
		orderSaved.CreatedDate = orderExist.CreatedDate
	} else {
		orderSetter.MemberItinID = memberItin.ID
		orderSetter.OrderCode = m.SetOrderCode()
		orderSaved, err = m.AddOrder(tx, ctx, orderSetter)
		if err != nil {
			h.SendBadRequest(w, err.Error())
			tx.Rollback(ctx)
			return
		}
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

	// Commit transaction
	err = tx.Commit(ctx)
	if err != nil {
		h.SendBadRequest(w, err.Error())
		tx.Rollback(ctx)
		return
	}

	h.SendSuccess(w, res.Transform(orderSaved), nil)
}

// UpdateOrderAct update payment process order from itinerary
func (h *Contract) UpdateOrderAct(w http.ResponseWriter, r *http.Request) {
	// Initial response handler
	var res response.OrderPaymentResponse
	code := chi.URLParam(r, "code")

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
	orderExist, _ := m.GetOrderByOrderCode(db, ctx, code)
	if orderExist.ID == 0 {
		h.SendNotfound(w, fmt.Sprintf("Order %s not found.", code))
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
		PaymentType:   model.PAYMENT_TYPE_DEFAULT,
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

	// Commit transaction
	err = tx.Commit(ctx)
	if err != nil {
		h.SendBadRequest(w, err.Error())
		tx.Rollback(ctx)
		return
	}

	h.SendSuccess(w, res.Transform(orderPaymentSaved), nil)
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
		PaidBy:      order.PaidBy,
		OrderStatus: orderStatus,
		TotalPrice:  order.TotalPrice,
		TcID:        order.TcID,
		OrderType:   order.OrderType,
	}
	orderUpdated, err := m.UpdateOrderByMemberItinID(tx, ctx, orderSetter, order.MemberItinID)
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

	// Commit transaction
	err = tx.Commit(ctx)
	if err != nil {
		h.SendBadRequest(w, err.Error())
		tx.Rollback(ctx)
		return
	}

	h.SendSuccess(w, req, nil)
}
