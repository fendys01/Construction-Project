package model

import (
	"context"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

const (
	PAYMENT_TYPE_DEFAULT   = "default"
	PAYMENT_STATUS_PROCESS = "PROC"
	PAYMENT_STATUS_PAID    = "PAID"
	PAYMENT_STATUS_CANCEL  = "CANC"
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
	expiredTime := time.Now().AddDate(0, 0, 1).In(time.UTC)

	sql := `INSERT INTO order_payments(order_id, payment_type, amount, payment_status, expired_date, created_date, payment_url) VALUES($1, $2, $3, $4, $5, $6, $7) RETURNING id`

	err := tx.QueryRow(ctx, sql, oP.OrderID, oP.PaymentType, oP.Amount, oP.PaymentStatus, expiredTime, timeStamp, oP.PaymentURL).Scan(&lastInsID)

	oP.ID = lastInsID
	oP.CreatedDate = timeStamp
	oP.ExpiredDate = expiredTime

	return oP, err
}

// UpdateOrderPayment update order payments
func (c *Contract) UpdateOrderPayment(tx pgx.Tx, ctx context.Context, oP OrderPaymentEnt, orderID int32) (OrderPaymentEnt, error) {
	var ID int32

	sql := `UPDATE order_payments SET order_id=$1, payment_type=$2, amount=$3, payment_status=$4, payment_url=$5, payloads=$6 WHERE order_id=$7 RETURNING id`

	err := tx.QueryRow(ctx, sql, oP.OrderID, oP.PaymentType, oP.Amount, oP.PaymentStatus, oP.PaymentURL, oP.Payloads, orderID).Scan(&ID)

	oP.ID = ID

	return oP, err
}

// GetPaymentOrderByOrderID Get Order payment by Order ID
func (c *Contract) GetPaymentOrderByOrderID(db *pgxpool.Conn, ctx context.Context, orderID int32) (OrderPaymentEnt, error) {
	var oP OrderPaymentEnt

	sqlM := `SELECT id, order_id, payment_type, amount, payment_status, expired_date, created_date, payment_url, payloads 
			FROM order_payments
			WHERE order_id = $1`

	err := db.QueryRow(ctx, sqlM, orderID).Scan(&oP.ID, &oP.OrderID, &oP.PaymentType, &oP.Amount, &oP.PaymentStatus, &oP.ExpiredDate, &oP.CreatedDate, &oP.PaymentURL, &oP.Payloads)

	return oP, err
}
