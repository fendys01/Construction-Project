package model

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type CallLogEnt struct {
	ID          int32
	TrxID       string
	Provider    string
	CallType    string
	BillPrice   sql.NullInt64
	Payloads    map[string]interface{}
	CreatedDate time.Time
}

const (
	PROVIDER_CITCALL = "citcall"
	TYPE_CALL        = "call"
	TYPE_SMS         = "sms"
	TYPE_OTP         = "otp"
)

var timeStamp = time.Now().In(time.UTC)

// AddCallLog add new call logs
func (c *Contract) AddCallLog(tx pgx.Tx, ctx context.Context, cl CallLogEnt) (CallLogEnt, error) {
	var lastInsID int32

	sql := `INSERT INTO call_logs(trx_id, provider, call_type, bill_price, payloads, created_date) VALUES($1, $2, $3, $4, $5, $6) RETURNING id`

	err := tx.QueryRow(ctx, sql, cl.TrxID, cl.Provider, cl.CallType, cl.BillPrice.Int64, cl.Payloads, timeStamp).Scan(&lastInsID)

	cl.ID = lastInsID
	cl.CreatedDate = timeStamp

	return cl, err
}

// Get Call Logs List
func (c *Contract) GetListCallLogs(db *pgxpool.Conn, ctx context.Context, param map[string]interface{}) ([]CallLogEnt, error) {
	list := []CallLogEnt{}
	var where []string
	var paramQuery []interface{}

	sql := `select id, trx_id, provider, call_type, bill_price, payloads, created_date from call_logs`

	var q string = sql

	if len(param["trx_id"].(string)) > 0 {
		var orWhere []string
		orWhere = append(orWhere, " trx_id = '"+param["trx_id"].(string)+"'")

		where = append(where, strings.Join(orWhere, " AND "))
	}

	if len(param["call_type"].(string)) > 0 {
		var orWhere []string
		orWhere = append(orWhere, " call_type = '"+param["call_type"].(string)+"'")

		where = append(where, strings.Join(orWhere, " AND "))
	}

	if len(param["provider"].(string)) > 0 {
		var orWhere []string
		orWhere = append(orWhere, " provider = '"+param["provider"].(string)+"'")

		where = append(where, strings.Join(orWhere, " AND "))
	}

	if len(where) > 0 {
		q += " WHERE " + strings.Join(where, " AND ")
	}

	rows, err := db.Query(ctx, q, paramQuery...)
	if err != nil {
		return list, err
	}

	defer rows.Close()

	for rows.Next() {
		var c CallLogEnt
		err = rows.Scan(&c.ID, &c.TrxID, &c.Provider, &c.CallType, &c.BillPrice, &c.Payloads, &c.CreatedDate)
		if err != nil {
			return list, err
		}

		list = append(list, c)
	}
	return list, err
}
