package payment

import (
	"fmt"
	"panorama/bootstrap"
	"panorama/services/api/model"

	midtrans "github.com/veritrans/go-midtrans"
)

const (
	MIDTRANS_PAYMENT_TYPE_CREDIT_CARD      = "credit_card"
	MIDTRANS_TRANSACTION_STATUS_CAPTURE    = "capture"
	MIDTRANS_TRANSACTION_STATUS_ACCEPT     = "accept"
	MIDTRANS_TRANSACTION_STATUS_SETTLEMENT = "settlement"
	MIDTRANS_TRANSACTION_STATUS_DENY       = "deny"
	MIDTRANS_TRANSACTION_STATUS_EXPIRE     = "expire"
	MIDTRANS_TRANSACTION_STATUS_CANCEL     = "cancel"
	MIDTRANS_FRAUD_STATUS_ACCEPT           = "accept"
)

type service struct {
	app *bootstrap.App
}

func New(app *bootstrap.App) *service {
	return &service{app}
}

func (s *service) SetMidtransParam(email string, name string, code string, amount int64) map[string]interface{} {
	return map[string]interface{}{
		"user_email":   email,
		"user_name":    name,
		"order_code":   code,
		"order_amount": amount,
	}
}

func (s *service) GetMidtransStatus(paymentType string, transactionStatus string, fraudStatus string) map[string]string {
	orderStatus := model.ORDER_STATUS_PENDING
	paymentStatus := model.PAYMENT_STATUS_PROCESS

	if paymentType == MIDTRANS_PAYMENT_TYPE_CREDIT_CARD && transactionStatus == MIDTRANS_TRANSACTION_STATUS_CAPTURE && fraudStatus == MIDTRANS_FRAUD_STATUS_ACCEPT {
		orderStatus = model.ORDER_STATUS_COMPLETED
		paymentStatus = model.PAYMENT_STATUS_PAID
	} else if transactionStatus == MIDTRANS_TRANSACTION_STATUS_SETTLEMENT {
		orderStatus = model.ORDER_STATUS_COMPLETED
		paymentStatus = model.PAYMENT_STATUS_PAID
	} else if transactionStatus == MIDTRANS_TRANSACTION_STATUS_DENY || transactionStatus == MIDTRANS_TRANSACTION_STATUS_EXPIRE || transactionStatus == MIDTRANS_TRANSACTION_STATUS_CANCEL {
		orderStatus = model.ORDER_STATUS_CANCEL
		paymentStatus = model.PAYMENT_STATUS_CANCEL
	}

	result := map[string]string{
		"order_status":   orderStatus,
		"payment_status": paymentStatus,
	}

	return result
}

func (s *service) GetMidtransPaymentURL(r map[string]interface{}) (map[string]string, error) {
	if r["user_email"] == nil || r["user_name"] == nil || r["order_code"] == nil || r["order_amount"] == nil {
		return nil, fmt.Errorf("invalid param %v", r)
	}

	midtransClient := midtrans.NewClient()
	midtransClient.ServerKey = s.app.Config.GetString("midtrans.server_key")
	midtransClient.ClientKey = s.app.Config.GetString("midtrans.client_key")

	// Check midtrans environment type
	midtransClient.APIEnvType = midtrans.Sandbox
	if s.app.Config.GetString("midtrans.env_type") != "" && s.app.Config.GetString("midtrans.env_type") == "prod" {
		midtransClient.APIEnvType = midtrans.Production
	}

	snapGateway := midtrans.SnapGateway{
		Client: midtransClient,
	}

	snapRequest := &midtrans.SnapReq{
		CustomerDetail: &midtrans.CustDetail{
			Email: fmt.Sprintf("%v", r["user_email"]),
			FName: fmt.Sprintf("%v", r["user_name"]),
		},
		TransactionDetails: midtrans.TransactionDetails{
			OrderID:  fmt.Sprintf("%v", r["order_code"]),
			GrossAmt: r["order_amount"].(int64),
		},
	}

	snapResponse, err := snapGateway.GetToken(snapRequest)
	if err != nil {
		return nil, err
	}
	result := map[string]string{
		"token":        snapResponse.Token,
		"redirect_url": snapResponse.RedirectURL,
		"status_code":  snapResponse.StatusCode,
	}

	return result, nil

}
