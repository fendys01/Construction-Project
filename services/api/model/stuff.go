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

// NotificationEnt ...
type StuffEnt struct {
	ID               int32
	Code        	 string
	Name        	 string
	Image            sql.NullString
	Description      string
	Price            string
	Type             int32
	IsActive         bool
	CreatedDate      time.Time
	UpdatedDate      sql.NullTime
	DeletedDate      sql.NullTime
}

func (c *Contract) SetStuffCode() string {
	rand.Seed(time.Now().UnixNano())
	code, _ := utils.Generate(`[a-z0-9]{6}`)
	return fmt.Sprintf("ST-%s-%s", time.Now().In(time.Local).Format("060102"), code)
}

func (c *Contract) AddStuff(db *pgxpool.Conn, ctx context.Context, s StuffEnt) (StuffEnt, error) {
	var lastInsID int32
	s.Code = c.SetStuffCode()
	err := db.QueryRow(ctx, `insert into stuff (code, name, image, description , price, type, created_date) 
		values($1, $2, $3, $4, $5, $6, $7) RETURNING id`,
		s.Code, s.Name, s.Image, s.Description, s.Price, s.Type, time.Now().In(time.UTC),
	).Scan(&lastInsID)
	
	s.ID = lastInsID

	return s, err
}

// GetListStuff ...
func (c *Contract) GetListStuff(db *pgxpool.Conn, ctx context.Context, param map[string]interface{}) ([]StuffEnt, error) {

	list := []StuffEnt{}
	var where []string
	var paramQuery []interface{}

	if len(param["stuff"].(string)) > 0 {
		var orWhere []string

		orWhere = append(orWhere, "s.name like '%"+param["stuff"].(string)+"%'")

		where = append(where, "("+strings.Join(orWhere, " OR ")+")")
	}

	if len(param["type"].(string)) > 0 {
		var orWhere []string
		orWhere = append(orWhere, " s.type = '"+param["type"].(string)+"'")

		where = append(where, strings.Join(orWhere, " AND "))
	}

	sql := `select id, code, name, image, description, price, type, created_date from stuff s`

	var q string = sql

	if len(where) > 0 {
		q += " WHERE " + strings.Join(where, " AND ")
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
		var a StuffEnt
		err = rows.Scan(&a.ID, &a.Code, &a.Name, &a.Image, &a.Description, &a.Price, &a.Type, &a.CreatedDate)
		if err != nil {
			return list, err
		}

		list = append(list, a)
	}

	return list, err
}

func (c *Contract) GetStuffCode(db *pgxpool.Conn, ctx context.Context, code string) (StuffEnt, error) {
	var s StuffEnt
	sql := `select
				code, name, image, description, price
			from stuff
			where code = $1 limit 1`
	err := db.QueryRow(ctx, sql, code).Scan(&s.Code, &s.Name, &s.Image, &s.Description, &s.Price)

	return s, err
}

func (c *Contract) GetStuffByCode(db *pgxpool.Conn, ctx context.Context, code string) (StuffEnt, error) {
	var s StuffEnt

	err := pgxscan.Get(ctx, db, &s, "select * from stuff where is_active = true and code=$1 limit 1", code)
	return s, err
}

func (c *Contract) UpdateStuff(tx pgx.Tx, ctx context.Context, code string, s StuffEnt) error {

	var ID int32

	sql := `update stuff set name=$1, image=$2, description=$3, price=$4, type=$5, is_active=$6, updated_date=$7 where code=$8 RETURNING id`

	err := tx.QueryRow(ctx, sql, s.Name, s.Image.String, s.Description, s.Price, s.Type, s.IsActive, time.Now().In(time.UTC), code).Scan(&ID)

	s.ID = ID

	return err
}

// Update IsActive Member.
func (c *Contract) UpdateIsActiveData(tx pgx.Tx, ctx context.Context, code string, s StuffEnt) error {

	var ID int32

	sql := `update stuff set is_active=$1, deleted_date=$2 where code=$3 RETURNING id`

	err := tx.QueryRow(ctx, sql, s.IsActive, time.Now().In(time.UTC), code).Scan(&ID)

	s.ID = ID

	return err
}

func (c *Contract) GetIsActiveStuff(db *pgxpool.Conn, ctx context.Context, code string) (StuffEnt, error) {
	var s StuffEnt

	err := pgxscan.Get(ctx, db, &s, "select * from stuff where is_active = true and code=$1 limit 1", code)
	return s, err
}