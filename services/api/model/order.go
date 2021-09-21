package model

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"panorama/lib/utils"
	"strings"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

const (
	ORDER_STATUS_COMPLETED = "C"
	ORDER_STATUS_PENDING   = "P"
	ORDER_STATUS_CANCEL    = "X"
	ORDER_TYPE_REGULER     = "R"
	ORDER_TYPE_CUSTOM      = "C"
)

type OrderEnt struct {
	ID                     int32
	MemberItinID           int32
	PaidBy                 int32
	OrderCode              string
	OrderStatus            string
	TotalPrice             int64
	OrderType              string
	TcID                   int32
	CreatedDate            time.Time
	DayPeriod              int32
	MemberItin             MemberItinEnt
	UserEnt                UserEnt
	MemberEnt              MemberEnt
	OrderPayment           OrderPaymentEnt
	OrderStatusDescription string
}

func (c *Contract) SetOrderCode() string {
	rand.Seed(time.Now().UnixNano())
	code, _ := utils.Generate(`[a-z0-9]{6}`)
	return fmt.Sprintf("OD-%s-%s", time.Now().In(time.Local).Format("060102"), code)
}

// AddOrder add new orders
func (c *Contract) AddOrder(tx pgx.Tx, ctx context.Context, o OrderEnt) (OrderEnt, error) {
	var lastInsID int32
	timeStamp := time.Now().In(time.UTC)

	sql := `INSERT INTO orders(member_itin_id, paid_by, order_code, order_status, total_price, tc_id, order_type, created_date) VALUES($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`

	err := tx.QueryRow(ctx, sql, o.MemberItinID, o.PaidBy, o.OrderCode, o.OrderStatus, o.TotalPrice, o.TcID, o.OrderType, timeStamp).Scan(&lastInsID)

	o.ID = lastInsID
	o.CreatedDate = timeStamp

	return o, err
}

// UpdateOrder update orders
func (c *Contract) UpdateOrderByMemberItinID(tx pgx.Tx, ctx context.Context, o OrderEnt, memberItinID int32) (OrderEnt, error) {
	var ID int32

	sql := `UPDATE orders SET paid_by=$1, order_status=$2, total_price=$3, tc_id=$4, order_type=$5 WHERE member_itin_id=$6 RETURNING id`

	err := tx.QueryRow(ctx, sql, o.PaidBy, o.OrderStatus, o.TotalPrice, o.TcID, o.OrderType, memberItinID).Scan(&ID)

	o.ID = ID

	return o, err
}

// Get Order List by Member_Code
func (c *Contract) GetOrderByCode(db *pgxpool.Conn, ctx context.Context, code string) (OrderEnt, error) {
	var o OrderEnt
	var paymentType, paymentStatus, paymentURL sql.NullString
	var paymentAmount sql.NullInt64
	var paymentExpiredDate sql.NullTime

	sqlM := `select 
		mi.itin_code, 
		m.member_code, 
		u.name, 
		m.name, 
		mi.title, 
		mi.details, 
		order_code, 
		order_status, 
		order_type, 
		paid_by, 
		total_price,
		op.payment_type,
		op.amount payment_amount,
		op.payment_status,
		op.expired_date,
		op.payment_url,
		op.payloads,
		orders.created_date 
	from orders
	join member_itins mi on mi.id = orders.member_itin_id 
	join users u on u.id = orders.tc_id 
	join members m on m.id = orders.paid_by
	left join order_payments op on op.order_id = orders.id
	where order_code = $1 limit 1`

	for _, v := range o.MemberItin.Details {
		c, err := json.Marshal(v["visit_list"])
		if err != nil {
			return o, err
		}
		o.MemberItin.DayPeriod = int32(strings.Count(string(c), "]"))
	}

	err := db.QueryRow(ctx, sqlM, code).Scan(&o.MemberItin.ItinCode, &o.MemberEnt.MemberCode, &o.UserEnt.Name, &o.MemberEnt.Name, &o.MemberItin.Title, &o.MemberItin.Details, &o.OrderCode, &o.OrderStatus,
		&o.OrderType, &o.PaidBy, &o.TotalPrice, &paymentType, &paymentAmount, &paymentStatus, &paymentExpiredDate, &paymentURL, &o.OrderPayment.Payloads, &o.CreatedDate)

	o.OrderPayment.PaymentType = paymentType.String
	o.OrderPayment.Amount = paymentAmount.Int64
	o.OrderPayment.PaymentStatus = paymentStatus.String
	o.OrderPayment.ExpiredDate = paymentExpiredDate.Time
	o.OrderPayment.PaymentURL = paymentURL.String

	return o, err
}

// Get Order List
func (c *Contract) GetListItinOrderMember(db *pgxpool.Conn, ctx context.Context, param map[string]interface{}) ([]OrderEnt, error) {
	list := []OrderEnt{}
	var where []string
	var paramQuery []interface{}
	var paymentType, paymentStatus, paymentURL sql.NullString
	var expiredDate sql.NullTime
	var paymentAmount sql.NullInt64

	sql := `select 
			mi.Title, 
			m.name,
			o.order_code, 
			o.order_status,
			o.order_type,
			o.total_price,
			o.created_date,
			op.payment_type,
			op.amount payment_amount,
			op.payment_status,
			op.expired_date,
			op.payment_url,
			op.payloads
		from orders o
		join member_itins mi on mi.id = o.member_itin_id and mi.created_by = o.paid_by 
		join users u on u.id = o.tc_id 
		join members m on m.id = o.paid_by
		left join order_payments op on op.order_id = o.id`

	var q string = sql

	if len(param["member_code"].(string)) > 0 {
		var orWhere []string
		orWhere = append(orWhere, " m.member_code = '"+param["member_code"].(string)+"'")

		where = append(where, strings.Join(orWhere, " AND "))
	}

	if len(param["expired_date"].(string)) > 0 {
		var orWhere []string
		orWhere = append(orWhere, " op.expired_date = '"+param["expired_date"].(string)+"'")

		where = append(where, strings.Join(orWhere, " AND "))
	}

	if len(param["order_code"].(string)) > 0 {
		var orWhere []string
		orWhere = append(orWhere, " order_code = '"+param["order_code"].(string)+"'")

		where = append(where, strings.Join(orWhere, " AND "))
	}

	if len(where) > 0 {
		q += " WHERE " + strings.Join(where, " AND ")
	}

	{
		var count int
		newQcount := `SELECT COUNT(*) FROM ( ` + q + ` ) AS data`

		err := db.QueryRow(ctx, newQcount, paramQuery...).Scan(&count)
		if err != nil {
			return list, err
		}
		param["count"] = count
	}

	// Select Max Page
	if param["count"].(int) > param["limit"].(int) && param["page"].(int) > int(param["count"].(int)/param["limit"].(int)) {
		param["page"] = int(math.Ceil(float64(param["count"].(int)) / float64(param["limit"].(int))))
	}

	param["offset"] = (param["page"].(int) - 1) * param["limit"].(int)

	if param["limit"].(int) == -1 {
		q += " ORDER BY " + param["order"].(string) + " " + param["sort"].(string)
	} else {
		q += " ORDER BY " + param["order"].(string) + " " + param["sort"].(string) + " offset $1 limit $2 "
		paramQuery = append(paramQuery, param["offset"])
		paramQuery = append(paramQuery, param["limit"])
	}

	rows, err := db.Query(ctx, q, paramQuery...)
	if err != nil {
		return list, err
	}

	defer rows.Close()

	for rows.Next() {
		var o OrderEnt
		err = rows.Scan(&o.MemberItin.Title, &o.MemberEnt.Name, &o.OrderCode, &o.OrderStatus, &o.OrderType, &o.TotalPrice, &o.CreatedDate, &paymentType, &paymentAmount, &paymentStatus, &expiredDate, &paymentURL, &o.OrderPayment.Payloads)
		if err != nil {
			return list, err
		}

		o.OrderPayment.PaymentType = paymentType.String
		o.OrderPayment.Amount = paymentAmount.Int64
		o.OrderPayment.PaymentStatus = paymentStatus.String
		o.OrderPayment.ExpiredDate = expiredDate.Time
		o.OrderPayment.PaymentURL = paymentURL.String

		list = append(list, o)
	}
	return list, err
}

// Get Order by Member Itin ID
func (c *Contract) GetOrderByMemberItinID(db *pgxpool.Conn, ctx context.Context, memberItinID int32) (OrderEnt, error) {
	var o OrderEnt

	sqlM := `SELECT id, member_itin_id, paid_by, order_code, order_status, total_price, tc_id, order_type, created_date 
			FROM orders
			WHERE member_itin_id = $1`

	err := db.QueryRow(ctx, sqlM, memberItinID).Scan(&o.ID, &o.MemberItinID, &o.PaidBy, &o.OrderCode, &o.OrderStatus, &o.TotalPrice, &o.TcID, &o.OrderType, &o.CreatedDate)

	return o, err
}

// GetOrderByOrderCode Get Order by Order Code
func (c *Contract) GetOrderByOrderCode(db *pgxpool.Conn, ctx context.Context, code string) (OrderEnt, error) {
	var o OrderEnt
	var tcID, memberID sql.NullInt32
	var itinCode, itinTitle, itinDestination, tcRole, tcCode, tcName, memberCode, memberName sql.NullString

	sqlM := `select 
		o.id, 
		o.member_itin_id, 
		o.paid_by, 
		o.order_code, 
		o.order_status, 
		o.total_price, 
		o.tc_id, 
		o.order_type, 
		o.created_date,
		mi.itin_code,
		mi.title,
		mi.destination,
		u.id tc_id,
		u.role tc_role,
		u.user_code tc_code,
		u.name tc_name,
		m.id member_name,
		m.member_code,
		m.name member_name
	from orders o 
	left join member_itins mi on mi.id = o.member_itin_id and mi.deleted_date is null
	left join users u on u.id = o.tc_id and u.deleted_date is null
	left join members m on m.id = o.paid_by and m.deleted_date is null
	where o.order_code = $1`

	err := db.QueryRow(ctx, sqlM, code).Scan(&o.ID, &o.MemberItinID, &o.PaidBy, &o.OrderCode, &o.OrderStatus, &o.TotalPrice, &o.TcID, &o.OrderType, &o.CreatedDate, &itinCode, &itinTitle, &itinDestination, &tcID, &tcRole, &tcCode, &tcName, &memberID, &memberCode, &memberName)

	o.MemberItin.ItinCode = itinCode.String
	o.MemberItin.Title = itinTitle.String
	o.MemberItin.Destination = itinDestination.String
	o.UserEnt.ID = tcID.Int32
	o.UserEnt.Role = tcRole.String
	o.UserEnt.UserCode = tcCode.String
	o.UserEnt.Name = tcName.String
	o.MemberEnt.ID = memberID.Int32
	o.MemberEnt.MemberCode = memberCode.String
	o.MemberEnt.Name = memberName.String

	return o, err
}
