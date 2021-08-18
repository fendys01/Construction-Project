package model

import (
	"context"
	"database/sql"
	"encoding/json"
	"math"
	"strings"
	"time"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type ItinSugEnt struct {
	ID          int32
	ItinCode    string
	Title       string
	Content     string
	Img         sql.NullString
	Details     []map[string]interface{}
	CreatedDate time.Time
	UpdatedDate sql.NullTime
	DeletedDate sql.NullTime
	CreatedBy   int64
	DayPeriod   int32
	UserEnt     UserEnt
	View        sql.NullInt32
	Destination string
}

// GetSugItinID get suggestion itinerary by itenerary code
func (c *Contract) GetSugItinID(db *pgxpool.Conn, ctx context.Context, code string) (int32, error) {
	var id int32
	err := pgxscan.Get(ctx, db, &id, "select id from itin_suggestions where itin_code=$1", code)

	return id, err
}

// GetSugItinByCode ...
func (c *Contract) GetSugItinByCode(db *pgxpool.Conn, ctx context.Context, code string) (ItinSugEnt, error) {
	var sug ItinSugEnt
	var destination sql.NullString

	sql := `select * from itin_suggestions where itin_code=$1 limit 1`
	err := db.QueryRow(ctx, sql, code).Scan(&sug.ID, &sug.ItinCode, &sug.CreatedBy, &sug.Title, &sug.Content, &sug.Img, &sug.Details, &sug.CreatedDate, &sug.UpdatedDate, &sug.DeletedDate, &sug.View, &destination)
	if err != nil {
		return sug, err
	}

	for _, v := range sug.Details {
		c, err := json.Marshal(v["visit_list"])
		if err != nil {
			return sug, err
		}
		sug.DayPeriod = int32(strings.Count(string(c), "]"))
	}
	sug.Destination = destination.String

	return sug, err
}

// AddSugItin ...
func (c *Contract) AddSugItin(tx pgx.Tx, ctx context.Context, sug ItinSugEnt, userID int32) (ItinSugEnt, error) {
	var lastInsID int32

	sql := `insert into itin_suggestions 
	(itin_code, title, content,img, details, created_date, created_by, destination)
	values($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id;`

	err := tx.QueryRow(ctx, sql,
		sug.ItinCode, sug.Title, sug.Content, sug.Img, sug.Details, time.Now().In(time.UTC), userID, sug.Destination,
	).Scan(&lastInsID)

	sug.ID = lastInsID

	return sug, err
}

// DelSugItin ...
func (c *Contract) DelSugItin(tx pgx.Tx, ctx context.Context, code string) error {
	_, err := tx.Exec(ctx,
		"update itin_suggestions set deleted_date=$1 where itin_code=$2", time.Now().In(time.UTC), code)

	return err
}

// UpdateSugItin ...
func (c *Contract) UpdateSugItin(tx pgx.Tx, ctx context.Context, sug ItinSugEnt, code string) error {
	sql := `update itin_suggestions set 
		title=$1, content=$2, img=$3,
		details=$4, updated_date=$5, view=$6, destination=$7 where itin_code=$8;`

	_, err := tx.Exec(ctx, sql,
		sug.Title, sug.Content, sug.Img,
		sug.Details, time.Now().In(time.UTC), sug.View.Int32, sug.Destination, code,
	)

	return err
}

// GetListItinSug ...
func (c *Contract) GetListItinSug(db *pgxpool.Conn, ctx context.Context, param map[string]interface{}) ([]ItinSugEnt, error) {

	list := []ItinSugEnt{}
	var where []string
	var paramQuery []interface{}
	var destination sql.NullString

	if len(param["keyword"].(string)) > 0 {
		var orWhere []string
		orWhere = append(orWhere, "name like ?")
		paramQuery = append(paramQuery, "%"+param["keyword"].(string)+"%")

		where = append(where, "("+strings.Join(orWhere, " OR ")+")")
	}

	sql := `select itin_code,title, content, itin_suggestions.img, details, itin_suggestions.created_date, name, view, destination
			from itin_suggestions 
			join users us on us.id = itin_suggestions.created_by`

	var q string = sql

	if len(param["created_by"].(string)) > 0 {

		var orWhere []string
		orWhere = append(orWhere, " user_code = '"+param["created_by"].(string)+"'")
		where = append(where, "("+strings.Join(orWhere, " AND ")+")")

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
		var a ItinSugEnt
		err = rows.Scan(&a.ItinCode, &a.Title, &a.Content, &a.Img, &a.Details, &a.CreatedDate, &a.UserEnt.Name, &a.View, &destination)
		if err != nil {
			return list, err
		}
		for _, v := range a.Details {
			c, err := json.Marshal(v["visit_list"])
			if err != nil {
				return list, err
			}
			a.DayPeriod = int32(strings.Count(string(c), "]"))
		}

		a.Destination = destination.String

		list = append(list, a)
	}
	return list, err
}
