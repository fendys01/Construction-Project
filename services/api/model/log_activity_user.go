package model

import (
	"context"
	"database/sql"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

// LogActivityUserEnt ...
type LogActivityUserEnt struct {
	ID          int64
	UserID      int64
	Role        string
	Title       string
	Activity    string
	EventType   string
	CreatedDate sql.NullTime
}

// ActiveClientConsultan ...
type ActiveClientConsultan struct {
	Title               string
	Name                string
	TcReplacementStatus string
	ItinDate            time.Time
	TcReplacementDate   sql.NullTime
	ItinCode            string
}

// GetListLogActivity ...
func (c *Contract) GetListLogActivity(db *pgxpool.Conn, ctx context.Context, role string, code string) ([]LogActivityUserEnt, error) {
	list := []LogActivityUserEnt{}
	var sql string
	if role == "customer" {
		sql = `
		select 
			l.title, l.activity, l.created_date FROM log_activity_users l
		join members m on m.id = l.user_id
		where l.role = $1 and m.member_code = $2`
	} else {
		sql = `
		select 
			l.title, l.activity, l.created_date FROM log_activity_users l
		join users u on u.id = l.user_id
		where l.role = $1 and u.user_code = $2`
	}

	rows, err := db.Query(ctx, sql, role, code)
	if err != nil {
		return list, err
	}

	defer rows.Close()
	for rows.Next() {
		var a LogActivityUserEnt
		err = rows.Scan(&a.Title, &a.Activity, &a.CreatedDate)
		if err != nil {
			return list, err
		}

		list = append(list, a)
	}
	return list, err
}

// GetListLogActivity ...
func (c *Contract) GetActiveClient(db *pgxpool.Conn, ctx context.Context, tcID int32) ([]ActiveClientConsultan, error) {
	list := []ActiveClientConsultan{}

	sql := `
		select 
		m.name, 
		mi.itin_code,
		mi.title, 
		mi.created_date, 
		a.created_date 
	from member_itins as mi
	join chat_groups cg on cg.member_itin_id = mi.id
	join orders o on o.chat_id = cg.id 
	join members m on m.id = mi.created_by
	left join (
		select 
			mic.member_itin_id, 
			mic.created_date
		from member_itin_changes as mic 
		where mic.changed_user_id = $1
	) a on a.member_itin_id = mi.id
	where o.tc_id = $2
	group by 
		m.name, 
		mi.itin_code,
		mi.title, 
		mi.created_date, 
		a.created_date 
	`

	rows, err := db.Query(ctx, sql, tcID, tcID)
	if err != nil {
		return list, err
	}

	defer rows.Close()
	for rows.Next() {
		var ac ActiveClientConsultan
		err = rows.Scan(&ac.Name, &ac.ItinCode, &ac.Title, &ac.ItinDate, &ac.TcReplacementDate)
		if err != nil {
			return list, err
		}

		list = append(list, ac)
	}
	return list, err
}

// AddLogActivity ...
func (c *Contract) AddLogActivity(tx pgx.Tx, ctx context.Context, u LogActivityUserEnt) (int32, error) {

	sql := "insert into log_activity_users (user_id, role, title, activity, event_type, created_date) values($1,$2,$3,$4,$5,$6) RETURNING id"

	var lastInsID int32

	err := tx.QueryRow(ctx, sql,
		u.UserID, u.Role, u.Title, u.Activity, u.EventType, time.Now().In(time.UTC),
	).Scan(&lastInsID)

	return lastInsID, err
}
