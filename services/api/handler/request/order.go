package request

type OrderReq struct {
	MemberItinCode    string                   `json:"member_itin_code" validate:"required"`
	PaidByCode        string                   `json:"paid_by_code" validate:"required"`
	TotalPrice        int                      `json:"total_price" validate:"required"`
	AdditionalDetails []map[string]interface{} `json:"additional_details"`
}

type OrderPaymentReq struct {
	Amount int `json:"amount" validate:"required"`
}

// VANumber : bank virtual account number
type VANumberMidtrans struct {
	Bank     string `json:"bank"`
	VANumber string `json:"va_number"`
}

// Action represents response action
type ActionMidtrans struct {
	Name   string `json:"name"`
	Method string `json:"method"`
	URL    string `json:"url"`
}

// Refund Details
type RefundMidtrans struct {
	RefundChargebackID int    `json:"refund_chargeback_id"`
	RefundAmount       string `json:"refund_amount"`
	Reason             string `json:"reason"`
	RefundKey          string `json:"refund_key"`
	RefundMethod       string `json:"refund_method"`
	BankConfirmedAt    string `json:"bank_confirmed_at"`
	CreatedAt          string `json:"created_at"`
}

type OrderPaymentMidtransNotification struct {
	StatusCode           string             `json:"status_code"`
	StatusMessage        string             `json:"status_message"`
	PermataVaNumber      string             `json:"permata_va_number"`
	SignKey              string             `json:"signature_key"`
	CardToken            string             `json:"token_id"`
	SavedCardToken       string             `json:"saved_token_id"`
	SavedTokenExpAt      string             `json:"saved_token_id_expired_at"`
	SecureToken          bool               `json:"secure_token"`
	Bank                 string             `json:"bank"`
	BillerCode           string             `json:"biller_code"`
	BillKey              string             `json:"bill_key"`
	XlTunaiOrderID       string             `json:"xl_tunai_order_id"`
	BIIVaNumber          string             `json:"bii_va_number"`
	ReURL                string             `json:"redirect_url"`
	ECI                  string             `json:"eci"`
	ValMessages          []string           `json:"validation_messages"`
	Page                 int                `json:"page"`
	TotalPage            int                `json:"total_page"`
	TotalRecord          int                `json:"total_record"`
	FraudStatus          string             `json:"fraud_status"`
	PaymentType          string             `json:"payment_type"`
	OrderID              string             `json:"order_id"`
	TransactionID        string             `json:"transaction_id"`
	TransactionTime      string             `json:"transaction_time"`
	TransactionStatus    string             `json:"transaction_status"`
	GrossAmount          string             `json:"gross_amount"`
	VANumbers            []VANumberMidtrans `json:"va_numbers"`
	PaymentCode          string             `json:"payment_code"`
	Store                string             `json:"store"`
	MerchantID           string             `json:"merchant_id"`
	MaskedCard           string             `json:"masked_card"`
	Currency             string             `json:"currency"`
	CardType             string             `json:"card_type"`
	Actions              []ActionMidtrans   `json:"actions"`
	RefundChargebackID   int                `json:"refund_chargeback_id"`
	RefundAmount         string             `json:"refund_amount"`
	RefundKey            string             `json:"refund_key"`
	Refunds              []RefundMidtrans   `json:"refunds"`
	ChannelResponseCode  string             `json:"channel_response_code"`
	ChannelStatusMessage string             `json:"channel_status_message"`
}
