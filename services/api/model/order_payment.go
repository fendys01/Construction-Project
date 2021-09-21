package model

import (
	"context"
	"database/sql"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

const (
	PAYMENT_STATUS_PROCESS             = "PROC"
	PAYMENT_STATUS_PAID                = "PAID"
	PAYMENT_STATUS_CANCEL              = "CANC"
	PAYMENT_STATUS_PROCESS_DESC        = "Waiting For Payment"
	PAYMENT_STATUS_PROCESS_METHOD_DESC = "Waiting For Payment Method"
	PAYMENT_STATUS_PAID_DESC           = "Completed"
	PAYMENT_STATUS_CANCEL_DESC         = "Cancel"
)

type OrderPaymentEnt struct {
	ID            int32
	OrderID       int32
	PaymentType   string
	Amount        int64
	PaymentStatus string
	ExpiredDate   time.Time
	CreatedDate   time.Time
	PaymentURL    string
	Payloads      map[string]interface{}
	OrderCode     string
}

// AddOrderPayment add new order payments
func (c *Contract) AddOrderPayment(db *pgxpool.Conn, ctx context.Context, tx pgx.Tx, oP OrderPaymentEnt) (OrderPaymentEnt, error) {
	var lastInsID int32
	timeStamp := time.Now().In(time.UTC)

	sql := `INSERT INTO order_payments(order_id, payment_type, amount, payment_status, expired_date, created_date, payment_url) VALUES($1, $2, $3, $4, $5, $6, $7) RETURNING id`

	err := tx.QueryRow(ctx, sql, oP.OrderID, nil, oP.Amount, oP.PaymentStatus, nil, timeStamp, oP.PaymentURL).Scan(&lastInsID)

	oP.ID = lastInsID
	oP.CreatedDate = timeStamp

	return oP, err
}

// UpdateOrderPayment update order payments
func (c *Contract) UpdateOrderPayment(tx pgx.Tx, ctx context.Context, oP OrderPaymentEnt, orderID int32) (OrderPaymentEnt, error) {
	var ID int32
	var paramQuery []interface{}

	sql := `UPDATE order_payments SET order_id=$1, payment_type=$2, amount=$3, payment_status=$4, expired_date=$5, payment_url=$6, payloads=$7 WHERE order_id=$8 RETURNING id`

	if oP.PaymentType == "" && oP.ExpiredDate.IsZero() && oP.Payloads == nil {
		paramQuery = append(paramQuery, oP.OrderID, nil, oP.Amount, oP.PaymentStatus, nil, oP.PaymentURL, nil, orderID)
	} else if oP.ExpiredDate.IsZero() {
		paramQuery = append(paramQuery, oP.OrderID, oP.PaymentType, oP.Amount, oP.PaymentStatus, nil, oP.PaymentURL, oP.Payloads, orderID)
	} else {
		paramQuery = append(paramQuery, oP.OrderID, oP.PaymentType, oP.Amount, oP.PaymentStatus, oP.ExpiredDate, oP.PaymentURL, oP.Payloads, orderID)
	}

	err := tx.QueryRow(ctx, sql, paramQuery...).Scan(&ID)

	oP.ID = ID

	return oP, err
}

// GetPaymentOrderByOrderID Get Order payment by Order ID
func (c *Contract) GetPaymentOrderByOrderID(db *pgxpool.Conn, ctx context.Context, orderID int32) (OrderPaymentEnt, error) {
	var oP OrderPaymentEnt
	var paymentType, paymentURL sql.NullString
	var orderNullID sql.NullInt32
	var expiredDate sql.NullTime

	sqlM := `SELECT id, order_id, payment_type, amount, payment_status, expired_date, created_date, payment_url, payloads FROM order_payments WHERE order_id = $1`

	err := db.QueryRow(ctx, sqlM, orderID).Scan(&oP.ID, &orderNullID, &paymentType, &oP.Amount, &oP.PaymentStatus, &expiredDate, &oP.CreatedDate, &paymentURL, &oP.Payloads)

	oP.PaymentType = paymentType.String
	oP.PaymentURL = paymentURL.String
	oP.OrderID = orderNullID.Int32
	oP.ExpiredDate = expiredDate.Time

	return oP, err
}
