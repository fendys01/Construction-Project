package model

import (
	"context"
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
func (c *Contract) GetTcIDLeastWorkByID(db *pgxpool.Conn, ctx context.Context, id int32) (int32, error) {
	var idTc, totalItin int32

	err := db.QueryRow(ctx, `
			select 
				u.id, count(o.member_itin_id) a
			from users as u
			left join orders o on o.tc_id = u.id
			join log_visit_app l on l.user_id = u.id
			where is_active = true and u.role = 'tc' and u.id not in($1)
			group by paid_by, u.id
			order by a asc limit 1`, id).Scan(&idTc, &totalItin)

	return idTc, err
}

func (c *Contract) GetTcIDLeastWork(db *pgxpool.Conn, ctx context.Context) (int32, string, error) {
	var idTc, totalItin int32
	var name string

	startDate := fmt.Sprintf("%v", time.Now().Format("2006-01-02")) + " 00:00:00"
	endDate := fmt.Sprintf("%v", time.Now().Format("2006-01-02")) + " 23:59:59"

	err := db.QueryRow(ctx, `
			select 
				u.id, count(o.member_itin_id) a, u.name
			from users as u
			left join orders o on o.tc_id = u.id
			join log_visit_app l on l.user_id = u.id
			where is_active = true and u.role = 'tc' and l.role = 'tc' 
			and l.last_active_date between $1 AND $2
			group by paid_by, u.id 
			order by a asc limit 1`, startDate, endDate).Scan(&idTc, &totalItin, &name)

	return idTc, name, err
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

// GetMemberItinByTcID ...
func (c *Contract) GetMemberItinByTcID(db *pgxpool.Conn, ctx context.Context, tcID int32) ([]int32, error) {

	var arrID []int32

	sql := `
		select 
			mi.id 
		from orders as o
		join member_itins mi on mi.id = o.member_itin_id
		where o.tc_id = $1 and order_status = 'P'`

	rows, err := db.Query(ctx, sql, tcID)
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
func (c *Contract) UpdateTcIdOrder(db *pgxpool.Conn, ctx context.Context, tx pgx.Tx, newTcID int32, arrItinID int32) error {

	sql := `UPDATE orders SET  tc_id=$1 WHERE member_itin_id = $2 and order_status = 'P' `

	stmt, err := tx.Query(ctx, sql, newTcID, arrItinID)

	defer stmt.Close()

	return err
}
