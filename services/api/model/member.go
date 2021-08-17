package model

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"math/rand"
	"Contruction-Project/lib/utils"
	"strings"
	"time"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgx/v4/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

// MemberEnt ...
type MemberEnt struct {
	ID             int32
	MemberCode     string
	Name           string
	Username       string
	Email          string
	Phone          string
	Password       string
	Img            sql.NullString
	IsEmailValid   bool `db:"is_valid_email"`
	IsPhoneValid   bool `db:"is_valid_phone"`
	IsActive       bool `db:"is_active"`
	CreatedDate    time.Time
	UpdatedDate    time.Time
	LastActiveDate sql.NullTime
	TotalVisited   int32
}

func (c *Contract) createMemberCode() string {
	rand.Seed(time.Now().UnixNano())
	code, _ := utils.Generate(`[a-z0-9]{6}`)
	return fmt.Sprintf("m-%s-%s", time.Now().In(time.Local).Format("20060102"), code)
}

// ActivateAndSetPhoneValid ...
func (c *Contract) ActivateAndSetPhoneValid(db *pgxpool.Conn, ctx context.Context, phone string) error {
	sql := `update members set is_active=$1, is_valid_phone=$2, updated_date=$3 where phone=$4;`
	_, err := db.Exec(ctx, sql, true, true, time.Now().In(time.UTC), phone)

	return err
}

// AddMember registering the member and add to database
func (c *Contract) AddMember(db *pgxpool.Conn, ctx context.Context, m MemberEnt) (MemberEnt, error) {
	// TODO: need to check is email, @username, phone is exist or not

	var lastInsID int32
	pass, _ := bcrypt.GenerateFromPassword([]byte(m.Password), 10)
	m.MemberCode = c.createMemberCode()

	err := db.QueryRow(ctx, `insert into members (member_code, name, username, email, phone, password, img,is_active, created_date,updated_date) 
		values($1, $2, $3, $4, $5, $6, $7,$8,$9,$10) RETURNING id`,
		m.MemberCode, m.Name, m.Username, m.Email, m.Phone, pass, m.Img, m.IsActive, time.Now().In(time.UTC), time.Now().In(time.UTC),
	).Scan(&lastInsID)

	m.ID = lastInsID

	return m, err
}

func (c *Contract) GetMemberBy(db *pgxpool.Conn, ctx context.Context, field, username string) (MemberEnt, error) {
	var m MemberEnt
	err := pgxscan.Get(ctx, db, &m,
		fmt.Sprintf("select * from members where %s=$1 limit 1", field),
		username,
	)

	return m, err
}

func (c *Contract) GetMemberByCode(db *pgxpool.Conn, ctx context.Context, code string) (MemberEnt, error) {
	var m MemberEnt
	err := pgxscan.Get(ctx, db, &m, `select 
									 members.id, member_code, name, username, email, phone, img, is_valid_email, 
									 is_valid_phone, is_active, l.last_active_date, total_visited
									 from members 
									 JOIN log_visit_app l on l.member_id = members.id where member_code=$1 limit 1 `, code)

	return m, err
}

func (c *Contract) GetMemberByEmail(db *pgxpool.Conn, ctx context.Context, email string) (MemberEnt, error) {
	var m MemberEnt
	err := pgxscan.Get(ctx, db, &m, "select * from members where email=$1 limit 1", email)

	return m, err
}

func (c *Contract) GetMemberByPhone(db *pgxpool.Conn, ctx context.Context, phone string) (MemberEnt, error) {
	var m MemberEnt
	err := pgxscan.Get(ctx, db, &m, "select * from members where phone=$1 limit 1", phone)

	return m, err
}

func (c *Contract) GetMemberByUsername(db *pgxpool.Conn, ctx context.Context, username string) (MemberEnt, error) {
	var m MemberEnt
	err := pgxscan.Get(ctx, db, &m, "select * from members where username=$1 limit 1", username)

	return m, err
}

func (c *Contract) isMemberExists(db *pgxpool.Conn, ctx context.Context, field, username string) bool {
	_, err := c.GetMemberBy(db, ctx, field, username)
	if err != nil {
		return false
	}

	return true
}

// UpdateMemberPass ...
func (c *Contract) UpdateMemberPass(db *pgxpool.Conn, ctx context.Context, username, newPass string) error {
	pass, err := bcrypt.GenerateFromPassword([]byte(newPass), 10)
	if err != nil {
		return err
	}

	sql := fmt.Sprintf("update members set password=$1, updated_date=$2 where %s=$3", c.getMemberField(username))
	_, err = db.Exec(context.Background(), sql, string(pass), time.Now().In(time.UTC), username)

	return err
}

// GetListMember ...
func (c *Contract) GetListMember(db *pgxpool.Conn, ctx context.Context, param map[string]interface{}) ([]MemberEnt, error) {

	// TODO join to visited customer next

	list := []MemberEnt{}
	var where []string
	var paramQuery []interface{}

	if len(param["keyword"].(string)) > 0 {
		var orWhere []string
		orWhere = append(orWhere, "name like ?")
		paramQuery = append(paramQuery, "%"+param["keyword"].(string)+"%")
		orWhere = append(orWhere, "username like ?")
		paramQuery = append(paramQuery, "%"+param["keyword"].(string)+"%")
		orWhere = append(orWhere, "phone like ?")
		paramQuery = append(paramQuery, "%"+param["keyword"].(string)+"%")
		orWhere = append(orWhere, "member_code like ?")
		paramQuery = append(paramQuery, "%"+param["keyword"].(string)+"%")

		where = append(where, "("+strings.Join(orWhere, " OR ")+")")
	}

	sql := `SELECT members.id, member_code, name, username, email, phone, img, is_valid_email, 
			is_valid_phone, is_active, l.last_active_date, total_visited
			from members
			JOIN log_visit_app l on l.member_id = members.id`

	var q string = sql

	if len(where) > 0 {
		q += " AND " + strings.Join(where, " AND ")
	}

	{
		var count int
		newQcount := `SELECT COUNT(*) FROM (` + q + `) AS data`
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
		var a MemberEnt
		err = rows.Scan(&a.ID, &a.MemberCode, &a.Name, &a.Username, &a.Email, &a.Phone, &a.Img, &a.IsEmailValid, &a.IsPhoneValid, &a.IsActive, &a.LastActiveDate.Time, &a.TotalVisited)
		if err != nil {
			return list, err
		}

		list = append(list, a)
	}

	return list, err
}

// UpdateMember ...
func (c *Contract) UpdateMember(db *pgxpool.Conn, code string, m MemberEnt) (MemberEnt, error) {

	sql := `UPDATE members 
			SET 
			name = $1, 
			username = $2, 
			email = $3, 
			phone = $4, 
			img = $5,
			updated_date = $6
			WHERE member_code = $7;`
	_, err := db.Exec(context.Background(), sql, m.Name, m.Username, m.Email, m.Phone, m.Img, time.Now().In(time.UTC), code)
	if err != nil {
		return m, err
	}

	return m, err
}

// GetLogVisitApp ...
func (c *Contract) GetLogVisitApp(db *pgxpool.Conn, ctx context.Context, id int32) (time.Time, int32, error) {
	var a time.Time
	var i int32
	err := db.QueryRow(ctx, "SELECT last_active_date, total_visited FROM log_visit_app  WHERE member_id = $1 ", id).Scan(&a, &i)
	return a, i, err
}

// add new log visit app
func (c *Contract) AddNewLogVisitApp(db *pgxpool.Conn, ctx context.Context, id int32) error {

	sql := `insert into log_visit_app (member_id,total_visited, last_active_date) values($1,$2, $3);`
	_, err := db.Exec(ctx, sql, id, 1, time.Now().In(time.UTC))

	return err
}

// ActivateAndSetPhoneValid ...
func (c *Contract) UpdateTotalVisited(db *pgxpool.Conn, ctx context.Context, id int32, total int32) error {
	sql := `update log_visit_app set total_visited=$1, last_active_date=$2 where member_id=$3;`
	_, err := db.Exec(ctx, sql, total, time.Now().In(time.UTC), id)

	return err
}
