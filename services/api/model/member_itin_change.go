package model

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type MemberItinChangesEnt struct {
	MemberItinID  int32
	Detail        []map[string]interface{}
	ChangedBy     string
	ChangedUserID int32
	CreatedDate   time.Time
	User          UserEnt
}

// GetTcIDLeastWorkByID ...
func (c *Contract) GetTcIDLeastWorkByID(db *pgxpool.Conn, ctx context.Context, id int32) (int32, string, string, error) {
	var idTc, totalItin int32
	var name, code string

	err := db.QueryRow(ctx, `
			select 
				u.id, count(o.chat_id) a, u.name, u.user_code
			from users as u
			left join orders o on o.tc_id = u.id
			join log_visit_app l on l.user_id = u.id
			where is_active = true and u.role = 'tc' and u.id not in($1)
			group by paid_by, u.id
			order by a asc limit 1`, id).Scan(&idTc, &totalItin, &name, &code)

	return idTc, name, code, err
}

func (c *Contract) GetTcIDLeastWork(db *pgxpool.Conn, ctx context.Context) (int32, string, string, error) {
	var idTc, totalItin int32
	var name, code string

	startDate := fmt.Sprintf("%v", time.Now().Format("2006-01-02")) + " 00:00:00"
	endDate := fmt.Sprintf("%v", time.Now().Format("2006-01-02")) + " 23:59:59"

	err := db.QueryRow(ctx, `
			select 
				u.id, count(o.chat_id) a, u.name, u.user_code
			from users as u
			left join orders o on o.tc_id = u.id
			join log_visit_app l on l.user_id = u.id
			where is_active = true and u.role = 'tc' and l.role = 'tc' 
			and l.last_active_date between $1 AND $2
			group by paid_by, u.id 
			order by a asc limit 1`, startDate, endDate).Scan(&idTc, &totalItin, &name, &code)

	return idTc, name, code, err
}

// ChangeTc ...
func (c *Contract) ChangeTc(db *pgxpool.Conn, ctx context.Context, tx pgx.Tx, mc MemberItinChangesEnt) error {

	timeStamp := time.Now().In(time.UTC)

	sql := `INSERT INTO member_itin_changes (member_itin_id, details, changed_by, changed_user_id, created_date) VALUES($1, $2, $3, $4, $5) RETURNING id`

	stmt, err := tx.Query(ctx, sql, mc.MemberItinID, mc.Detail, mc.ChangedBy, mc.ChangedUserID, timeStamp)

	defer stmt.Close()

	return err
}

// GetMemberItinIdByCustID ...
func (c *Contract) GetMemberItinIdByCustID(db *pgxpool.Conn, ctx context.Context) ([]int32, error) {
	var arrID []int32

	sql := `
		select
			mi.id 
		from member_itins as mi
		where mi.created_by = 1`

	rows, err := db.Query(ctx, sql)
	if err != nil {
		return arrID, err
	}

	defer rows.Close()
	for rows.Next() {
		var a int32
		err = rows.Scan(&a)
		if err != nil {
			return arrID, err
		}

		arrID = append(arrID, a)
	}
	return arrID, err
}

// UpdateTcIdOrder ...
func (c *Contract) UpdateTcIdOrder(db *pgxpool.Conn, ctx context.Context, tx pgx.Tx, newTcID int32, chatGroupID int32) error {

	sql := `UPDATE orders SET tc_id = $1 WHERE chat_id = $2 and order_status = $3 `

	stmt, err := tx.Query(ctx, sql, newTcID, chatGroupID, ORDER_STATUS_PENDING)

	defer stmt.Close()

	return err
}

// GetListChatGroupByTcID ...
func (c *Contract) GetListChatGroupByTcID(db *pgxpool.Conn, ctx context.Context, tcID int32) ([]ChatGroupEnt, error) {

	var chatGroupList []ChatGroupEnt
	var memberItinID sql.NullInt32

	sql := `select cg.id, cg.member_itin_id, cg.chat_group_code from orders o join chat_groups cg on cg.id = o.chat_id where o.tc_id = $1 and o.order_status = $2`

	rows, err := db.Query(ctx, sql, tcID, ORDER_STATUS_PENDING)
	if err != nil {
		return chatGroupList, err
	}

	defer rows.Close()
	for rows.Next() {
		var c ChatGroupEnt
		err = rows.Scan(&c.ID, &memberItinID, &c.ChatGroupCode)
		if err != nil {
			return chatGroupList, err
		}

		c.MemberItin.ID = memberItinID.Int32
		chatGroupList = append(chatGroupList, c)
	}
	return chatGroupList, err
}
