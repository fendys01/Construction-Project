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

	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type UserEnt struct {
	ID                    int32
	UserCode              string
	Name                  string
	Email                 string
	Phone                 string
	Password              string
	Role                  string // admin, tc
	Img                   sql.NullString
	IsActive              bool
	CreatedDate           time.Time
	LogActivityUser       []LogActivityUserEnt
	UpdatedDate           sql.NullTime
	DeletedDate           sql.NullTime
	LastVisit             sql.NullTime
	TotalClient           sql.NullInt32
	TotalOrd              sql.NullInt32
	TotalCustomOrder      sql.NullInt32
	TotalItinSugView      sql.NullInt32
	TotalItinSug          sql.NullInt32
	ActiveClientConsultan []ActiveClientConsultan
}

func (c *Contract) createUserCode(role string) string {
	rand.Seed(time.Now().UnixNano())

	u := "u"

	if role == "tc" {
		u = "TC"
	} else if role == "admin" {
		u = "ADM"
	}
	code, _ := utils.Generate(u + `-[a-z0-9]{7}`)
	return code
}

// GetUser ...
func (c *Contract) GetUser(db *pgxpool.Conn, ctx context.Context, param map[string]interface{}) ([]UserEnt, error) {

	list := []UserEnt{}
	var where []string
	var paramQuery []interface{}

	if len(param["keyword"].(string)) > 0 {
		var orWhere []string
		orWhere = append(orWhere, "name like '%"+param["keyword"].(string)+"%'")
		orWhere = append(orWhere, "role like '%"+param["keyword"].(string)+"%'")
		orWhere = append(orWhere, "phone like '%"+param["keyword"].(string)+"%'")
		orWhere = append(orWhere, "email like '%"+param["keyword"].(string)+"%'")

		where = append(where, "("+strings.Join(orWhere, " OR ")+")")
	}

	sql := `
		select role, user_code, name,img, phone, email, count(distinct(total_client)), MAX(last_active_date) 
			from (
				select 
					u.id, l.role, u.user_code,
					u.name, u.img, u.phone, u.email,
					paid_by as total_client, 
					l.last_active_date
				from users as u
				left join orders o on o.tc_id = u.id
				join log_visit_app l on l.user_id = u.id
				where is_active = true 
				group by paid_by, u.id, l.id 
			) a `

	var q string = sql

	if len(param["role"].(string)) > 0 {

		var orWhere []string
		orWhere = append(orWhere, " role = '"+param["role"].(string)+"'")

		where = append(where, strings.Join(orWhere, " AND "))

	}

	if len(where) > 0 {
		q += " WHERE " + strings.Join(where, " AND ")
	}

	{
		var count int
		newQcount := `SELECT COUNT(*) FROM ( ` + q + ` group by id, role, name, img, phone, user_code, email ) AS data`

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
		q += " group by id, role, name, img, phone, email, user_code ORDER BY " + param["order"].(string) + " " + param["sort"].(string)
	} else {
		q += " group by id, role, name, img, phone, email, user_code ORDER BY " + param["order"].(string) + " " + param["sort"].(string) + " offset $1 limit $2 "
		paramQuery = append(paramQuery, param["offset"])
		paramQuery = append(paramQuery, param["limit"])
	}

	rows, err := db.Query(ctx, q, paramQuery...)
	if err != nil {
		return list, err
	}

	defer rows.Close()
	for rows.Next() {
		var a UserEnt
		err = rows.Scan(&a.Role, &a.UserCode, &a.Name, &a.Img, &a.Phone, &a.Email, &a.TotalClient, &a.LastVisit)
		if err != nil {
			return list, err
		}

		list = append(list, a)
	}
	return list, err
}

// GetUserByCode ...
func (c *Contract) GetUserByCode(db *pgxpool.Conn, ctx context.Context, code string) (UserEnt, error) {
	var u UserEnt

	err := pgxscan.Get(ctx, db, &u, "select * from users where is_active = true and user_code=$1 limit 1", code)
	return u, err
}

// GetDetailTcByCode ...
func (c *Contract) GetDetailTcByCode(db *pgxpool.Conn, ctx context.Context, code string) (UserEnt, error) {
	var u UserEnt
	sql := `
			select 
				u.ID, u.user_code, l.role,
				u.name, u.phone, u.img, u.phone,	
				count(distinct(paid_by)) as total_client, 
				count(o.id) as  total_ord,
				count(case when o.order_type = 'C' then o.order_type end) as total_cust_ord,
				l.last_active_date
			from users as u
			left join orders o on o.tc_id = u.id
			left join log_visit_app l on l.user_id = u.id
			left join member_itins mi on mi.id = o.member_itin_id
			left join members m on m.id = mi.created_by
			where l.role = 'tc' and user_code = $1
			group by paid_by, u.id, l.id`

	err := db.QueryRow(ctx, sql, code).Scan(
		&u.ID, &u.UserCode, &u.Role, &u.Name, &u.Phone, &u.Img, &u.Phone, &u.TotalClient,
		&u.TotalOrd, &u.TotalCustomOrder, &u.LastVisit)

	return u, err
}

// GetDetailAdminByCode ...
func (c *Contract) GetDetailAdminByCode(db *pgxpool.Conn, ctx context.Context, code string) (UserEnt, error) {
	var u UserEnt
	sql := `
		select 
			u.id, u.user_code,u.email, u.name, u.img, u.phone, SUM(view), count(itin.id) ,l.last_active_date
		from users as u
		left join itin_suggestions itin on itin.created_by = u.id
		left join log_visit_app l on l.user_id = u.id
		where user_code = $1
		group by u.id, l.id order by l.id desc limit 1`

	err := db.QueryRow(ctx, sql, code).Scan(
		&u.ID, &u.UserCode, &u.Email, &u.Name, &u.Img, &u.Phone, &u.TotalItinSugView,
		&u.TotalItinSug, &u.LastVisit)

	return u, err
}

// GetUserByEmail ...
func (c *Contract) GetUserByEmail(db *pgxpool.Conn, ctx context.Context, email string) (UserEnt, error) {
	var u UserEnt

	err := pgxscan.Get(ctx, db, &u, "select * from users where email=$1 limit 1", email)
	return u, err
}

func (c *Contract) isUserExists(db *pgxpool.Conn, ctx context.Context, email string) bool {
	_, err := c.GetUserByEmail(db, ctx, email)
	if err != nil {
		return false
	}

	return true
}

// GetUserByID ...
func (c *Contract) GetUserByID(db *pgxpool.Conn, ctx context.Context, id int32) (UserEnt, error) {
	var u UserEnt

	err := pgxscan.Get(ctx, db, &u, "select * from users where id=$1", id)
	return u, err
}

// AddUser ...
func (c *Contract) AddUser(tx pgx.Tx, ctx context.Context, u UserEnt) (UserEnt, error) {

	sql := "insert into users (name, user_code, phone, email, password, role, is_active, created_date) values($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id"

	var lastInsID int32
	code := c.createUserCode(u.Role)

	err := tx.QueryRow(context.Background(), sql,
		u.Name, code, u.Phone, u.Email, u.Password, u.Role, u.IsActive, time.Now().In(time.UTC),
	).Scan(&lastInsID)

	u.ID = lastInsID
	u.UserCode = code

	return u, err
}

// UpdateUser ...
func (c *Contract) UpdateUser(tx pgx.Tx, ctx context.Context, code string, u UserEnt) error {

	var ID int32

	sql := `update users set name=$1, role=$2, is_active=$3, phone=$4, img=$5, email=$6, password=$7, updated_date=$8 where user_code=$9 RETURNING id`

	err := tx.QueryRow(ctx, sql, u.Name, u.Role, u.IsActive, u.Phone, u.Img.String, u.Email, u.Password, time.Now().In(time.UTC), code).Scan(&ID)

	u.ID = ID

	return err
}

// UpdateEmail ...
func (c *Contract) UpdateEmail(db *pgxpool.Conn, ctx context.Context, id int32, email string) error {
	_, err := db.Exec(context.Background(),
		"update users set email=$1, updated_date=$2 where id=$3",
		email, time.Now().In(time.UTC), id,
	)

	if err != nil {
		return err
	}

	// TODO: need to send email confirmation for change email

	return nil
}

func (c *Contract) GetUserBy(db *pgxpool.Conn, ctx context.Context, field, username string) (UserEnt, error) {
	var u UserEnt
	err := pgxscan.Get(ctx, db, &u,
		fmt.Sprintf("select * from users where %s=$1 limit 1", field),
		username,
	)

	return u, err
}
