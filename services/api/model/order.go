package model

import (
	"context"
	"database/sql"
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
	Title                  string
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
	Details                string
	ChatID                 int32
	Description            string
	TotalPricePpn          int64
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

	sql := `INSERT INTO orders(title, paid_by, order_code, order_status, total_price, tc_id, order_type, created_date, details, chat_id, description, total_price_ppn) VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11,$12) RETURNING id`

	err := tx.QueryRow(ctx, sql, o.Title, o.PaidBy, o.OrderCode, o.OrderStatus, o.TotalPrice, o.TcID, o.OrderType, timeStamp, o.Details, o.ChatID, o.Description, o.TotalPricePpn).Scan(&lastInsID)

	o.ID = lastInsID
	o.CreatedDate = timeStamp

	return o, err
}

// UpdateOrder update orders
func (c *Contract) UpdateOrderByCode(tx pgx.Tx, ctx context.Context, o OrderEnt) (OrderEnt, error) {
	var ID int32

	sql := `UPDATE orders SET paid_by=$1, order_status=$2, total_price=$3, tc_id=$4, order_type=$5, details=$6, title=$7 WHERE order_code=$8 RETURNING id`

	err := tx.QueryRow(ctx, sql, o.PaidBy, o.OrderStatus, o.TotalPrice, o.TcID, o.OrderType, o.Details, o.Title, o.OrderCode).Scan(&ID)

	o.ID = ID

	return o, err
}

// Get Order List by Member_Code
func (c *Contract) GetOrderByCode(db *pgxpool.Conn, ctx context.Context, code string) (OrderEnt, error) {
	var o OrderEnt

	sqlM := `select 
		m.member_code, 
		u.name, 
		m.name, 
		title, 
		order_code, 
		order_status, 
		order_type, 
		paid_by, 
		total_price,
		orders.created_date 
	from orders
	join users u on u.id = orders.tc_id 
	join members m on m.id = orders.paid_by
	where order_code = $1 limit 1`

	err := db.QueryRow(ctx, sqlM, code).Scan(&o.MemberEnt.MemberCode, &o.UserEnt.Name, &o.MemberEnt.Name, &o.Title, &o.OrderCode, &o.OrderStatus,
		&o.OrderType, &o.PaidBy, &o.TotalPrice, &o.CreatedDate)

	return o, err
}

// Get Order List
func (c *Contract) GetListItinOrderMember(db *pgxpool.Conn, ctx context.Context, param map[string]interface{}) ([]OrderEnt, error) {
	list := []OrderEnt{}
	var where []string
	var paramQuery []interface{}
	var title, description, memberImg, paymentStatus, detail sql.NullString
	var totalPPN sql.NullInt64

	sql := `select
		title,
		m.name,
		order_code,
		description,
		order_type, 
		details,
		total_price,
		total_price_ppn, 
		m.img,
		op.payment_status,
		o.created_date
	from orders o
	join members m on m.id = o.paid_by
	left join order_payments op on o.id = op.order_id `

	var q string = sql

	if len(param["member_code"].(string)) > 0 {
		var orWhere []string
		orWhere = append(orWhere, " m.member_code = '"+param["member_code"].(string)+"'")

		where = append(where, strings.Join(orWhere, " AND "))
	}

	if len(param["order_type"].(string)) > 0 {
		var orWhere []string
		orWhere = append(orWhere, " order_type = '"+param["order_type"].(string)+"'")

		where = append(where, strings.Join(orWhere, " AND "))
	}

	if len(param["keyword"].(string)) > 0 {
		var orWhere []string
		orWhere = append(orWhere, "lower(order_code) like lower('%"+param["keyword"].(string)+"%')")

		where = append(where, "("+strings.Join(orWhere, " OR ")+")")
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
		err = rows.Scan(&title, &o.MemberEnt.Name, &o.OrderCode, &description, &o.OrderType, &detail, &o.TotalPrice, &totalPPN, &memberImg, &paymentStatus, &o.CreatedDate)
		if err != nil {
			return list, err
		}

		o.Title = title.String
		o.Details = detail.String
		o.Description = description.String
		o.TotalPricePpn = totalPPN.Int64
		o.MemberEnt.Img = memberImg
		o.OrderPayment.PaymentStatus = paymentStatus.String

		list = append(list, o)
	}
	return list, err
}

// Get Order by Code
func (c *Contract) GetOrderByCodeDetail(db *pgxpool.Conn, ctx context.Context, code string) (OrderEnt, error) {
	var o OrderEnt

	sqlM := `SELECT id, paid_by, order_code, order_status, total_price, tc_id, order_type, created_date, details, title
			FROM orders
			WHERE order_code = $1`

	err := db.QueryRow(ctx, sqlM, code).Scan(&o.ID, &o.PaidBy, &o.OrderCode, &o.OrderStatus, &o.TotalPrice, &o.TcID, &o.OrderType, &o.CreatedDate, &o.Details, &o.Title)
	return o, err
}

// Get Order List History by chat id
func (c *Contract) GetListHistoryOrdByChatCode(db *pgxpool.Conn, ctx context.Context, code string) ([]OrderEnt, error) {
	list := []OrderEnt{}
	var desc sql.NullString
	var totalPricePPN sql.NullInt64

	sql := `
		select 
			title, description, order_code, total_price, total_price_ppn, order_type, details , o.created_date 
		from orders as o
		join chat_groups cg on cg.id = o.chat_id
		where cg.chat_group_code = $1
		`

	var q string = sql

	rows, err := db.Query(ctx, q, code)
	if err != nil {
		return list, err
	}

	defer rows.Close()

	for rows.Next() {
		var o OrderEnt
		err = rows.Scan(&o.Title, &desc, &o.OrderCode, &o.TotalPrice, &totalPricePPN, &o.OrderType, &o.Details, &o.CreatedDate)
		if err != nil {
			return list, err
		}
		o.Description = desc.String
		o.TotalPricePpn = totalPricePPN.Int64

		list = append(list, o)
	}
	return list, err
}

// GetOrderByOrderCode Get Order by Order Code
func (c *Contract) GetOrderByOrderCode(db *pgxpool.Conn, ctx context.Context, code string) (OrderEnt, error) {
	var o OrderEnt
	var tcID, memberID, itinID sql.NullInt32
	var itinCode, itinTitle, itinDestination, tcRole, tcCode, tcName, memberCode, memberName sql.NullString

	sqlM := `select 
		o.id, 
		cg.member_itin_id, 
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
	left join chat_groups cg on cg.member_itin_id = o.chat_id 
	left join member_itins mi on mi.id = cg.member_itin_id and mi.deleted_date is null
	left join users u on u.id = o.tc_id and u.deleted_date is null
	left join members m on m.id = o.paid_by and m.deleted_date is null
	where o.order_code = $1`

	err := db.QueryRow(ctx, sqlM, code).Scan(&o.ID, &itinID, &o.PaidBy, &o.OrderCode, &o.OrderStatus, &o.TotalPrice, &o.TcID, &o.OrderType, &o.CreatedDate, &itinCode, &itinTitle, &itinDestination, &tcID, &tcRole, &tcCode, &tcName, &memberID, &memberCode, &memberName)

	o.MemberItinID = itinID.Int32
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
