package response

import (
	"panorama/services/api/model"
	"strconv"
	"time"
)

// ItinOrderMember Response Detail

type DetailOrderMemberResponse struct {
	MemberItinCode string                   `json:"itin_code"`
	MemberCode     string                   `json:"member_code"`
	TcName         string                   `json:"tc_name"`
	MemberName     string                   `json:"member_name"`
	OrderCode      string                   `json:"order_code"`
	Title          string                   `json:"title"`
	Details        []map[string]interface{} `json:"detail"`
	DayPeriod      string                   `json:"day_period"`
	PaidBy         int32                    `json:"paid_by"`
	TotalPrice     int64                    `json:"total_price"`
	CreatedDate    time.Time                `json:"created_date"`
	OrderType      string                   `json:"order_type"`
}

// Transform from order model to detail order member response
func (r DetailOrderMemberResponse) Transform(i model.OrderEnt) DetailOrderMemberResponse {
	r.MemberItinCode = i.MemberItin.ItinCode
	r.MemberCode = i.MemberEnt.MemberCode
	r.TcName = i.UserEnt.Name
	r.MemberName = i.MemberEnt.Name
	r.OrderCode = i.OrderCode
	r.Title = i.MemberItin.Title
	r.Details = i.MemberItin.Details
	r.DayPeriod = strconv.Itoa(int(i.DayPeriod)) + "D" + strconv.Itoa(int(i.DayPeriod-1)) + "N"
	r.PaidBy = i.PaidBy
	r.TotalPrice = i.TotalPrice
	r.CreatedDate = i.CreatedDate
	r.OrderType = i.OrderType

	return r

}

// ItinOrderMember Response List
type ItinOrderMemberResponse struct {
	Title         string    `json:"title"`
	MemberName     string   `json:"customer_name"`
	OrderCode     string    `json:"order_code"`
	PaymentStatus string    `json:"payment_status"`
	PaymentType   string    `json:"payment_type"`
	OrderStatus   string    `json:"order_status"`
	OrderType     string    `json:"order_type"`
	TotalPrice    int64     `json:"total_price"`
	CreatedDate   string `json:"created_date"`
}

// Transform from order model to itin order member response
func (r ItinOrderMemberResponse) Transform(i model.OrderEnt) ItinOrderMemberResponse {
	time := i.CreatedDate
	timeday := time.Format("January 02, 2006")


	r.Title = i.MemberItin.Title
	r.MemberName = i.MemberEnt.Name
	r.OrderCode = i.OrderCode
	r.PaymentStatus = i.OrderPayment.PaymentStatus
	r.PaymentType = i.OrderPayment.PaymentType
	r.CreatedDate = timeday
	r.OrderStatus = i.OrderStatus
	r.OrderType = i.OrderType
	r.TotalPrice = i.TotalPrice

	return r

}

// OrderPayment Response
type OrderPaymentResponse struct {
	OrderCode     string                 `json:"order_code"`
	PaymentType   string                 `json:"payment_type"`
	Amount        int64                  `json:"amount"`
	PaymentStatus string                 `json:"payment_status"`
	ExpiredDate   time.Time              `json:"expired_date"`
	CreatedDate   time.Time              `json:"created_date"`
	PaymentURL    string                 `json:"payment_url"`
	Payloads      map[string]interface{} `json:"payloads"`
}

// Transform from order payment model
func (r OrderPaymentResponse) Transform(i model.OrderPaymentEnt) OrderPaymentResponse {
	r.OrderCode = i.OrderCode
	r.PaymentType = i.PaymentType
	r.Amount = i.Amount
	r.PaymentStatus = i.PaymentStatus
	r.ExpiredDate = i.ExpiredDate
	r.CreatedDate = i.CreatedDate
	r.PaymentURL = i.PaymentURL
	r.Payloads = i.Payloads

	return r
}
