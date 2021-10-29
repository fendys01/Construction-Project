package response

import (
	"panorama/services/api/model"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// ItinOrderMember Response Detail

type DetailOrderMemberResponse struct {
	MemberCode  string    `json:"member_code"`
	TcName      string    `json:"tc_name"`
	MemberName  string    `json:"member_name"`
	OrderCode   string    `json:"order_code"`
	Title       string    `json:"title"`
	Details     string    `json:"detail_pdf"`
	PaidBy      int32     `json:"paid_by"`
	TotalPrice  int64     `json:"total_price"`
	OrderType   string    `json:"order_type"`
	CreatedDate time.Time `json:"created_date"`
}

// Transform from order model to detail order member response
func (r DetailOrderMemberResponse) Transform(i model.OrderEnt) DetailOrderMemberResponse {
	r.MemberCode = i.MemberEnt.MemberCode
	r.TcName = i.UserEnt.Name
	r.MemberName = i.MemberEnt.Name
	r.OrderCode = i.OrderCode
	r.Title = i.Title

	if len(strings.TrimSpace(i.Details)) > 0 {
		if IsUrl(i.Details) {
			r.Details = i.Details
		} else {
			r.Details = viper.GetString("aws.s3.public_url") + i.Details
		}

	} else {
		r.Details = ""
	}

	r.PaidBy = i.PaidBy
	r.TotalPrice = i.TotalPrice
	r.OrderType = i.OrderType
	r.CreatedDate = i.CreatedDate

	return r

}

// ItinOrderMember Response List
type ItinOrderMemberResponse struct {
	Title         string `json:"title"`
	MemberName    string `json:"customer_name"`
	Details       string `json:"detail_pdf"`
	OrderCode     string `json:"order_code"`
	OrderType     string `json:"order_type"`
	Description   string `josn:"description"`
	TotalPrice    int64  `json:"total_price"`
	TotalPricePpn int64  `json:"total_price_ppn"`
	CreatedDate   string `json:"created_date"`
	MemberImage   string `json:"member_image"`
	PaymentStatus string `json:"payment_status"`
}

// Transform from order model to itin order member response
func (r ItinOrderMemberResponse) Transform(i model.OrderEnt) ItinOrderMemberResponse {
	time := i.CreatedDate
	timeday := time.Format("Jan 02, 2006")
	r.Title = i.Title
	r.MemberName = i.MemberEnt.Name
	r.Description = i.Description

	if len(strings.TrimSpace(i.Details)) > 0 {
		if IsUrl(i.Details) {
			r.Details = i.Details
		} else {
			r.Details = viper.GetString("aws.s3.public_url") + i.Details
		}

	} else {
		r.Details = ""
	}

	r.OrderCode = i.OrderCode
	r.OrderType = i.OrderType
	r.Description = i.Description
	r.CreatedDate = timeday
	r.TotalPrice = i.TotalPrice
	r.TotalPricePpn = i.TotalPricePpn

	var memberImage string
	if len(i.MemberEnt.Img.String) > 0 {
		if IsUrl(i.MemberEnt.Img.String) {
			memberImage = i.MemberEnt.Img.String
		} else {
			memberImage = viper.GetString("aws.s3.public_url") + i.MemberEnt.Img.String
		}

	}
	r.MemberImage = memberImage
	r.PaymentStatus = i.OrderPayment.PaymentStatus

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
