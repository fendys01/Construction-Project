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

type MemberItinEnt struct {
	ID            int32
	ItinCode      string
	Title         string
	CreatedBy     int32
	EstPrice      sql.NullInt64
	StartDate     time.Time
	EndDate       time.Time
	Details       []map[string]interface{}
	CreatedDate   time.Time
	UpdatedDate   sql.NullTime
	DeletedDate   sql.NullTime
	Destination   string
	Img           sql.NullString
	DayPeriod     int32
	MemberEnt     MemberEnt
	GroupMembers  []map[string]interface{}
	ChatGroupCode string
}

// GetMemberItinID get member itinerary by itenerary code
func (c *Contract) GetMemberItinID(db *pgxpool.Conn, ctx context.Context, code string) (int32, error) {
	var id int32
	err := pgxscan.Get(ctx, db, &id, "select id from member_itins where itin_code=$1", code)

	return id, err
}

// GetMemberItinByID ...
func (c *Contract) GetMemberItinByID(db *pgxpool.Conn, ctx context.Context, id int32) (MemberItinEnt, error) {
	var member MemberItinEnt
	err := pgxscan.Get(ctx, db, &member,
		`select * from member_itins where id=$1 limit 1`, id)

	return member, err
}

// GetMemberItinByCode
func (c *Contract) GetMemberItinByCode(db *pgxpool.Conn, ctx context.Context, code string) (MemberItinEnt, error) {
	var m MemberItinEnt
	var dest sql.NullString
	sql := `select * from member_itins where itin_code=$1 limit 1`
	err := db.QueryRow(ctx, sql, code).Scan(&m.ID, &m.ItinCode, &m.Title, &m.CreatedBy, &m.EstPrice, &m.StartDate, &m.EndDate, &m.Details, &m.CreatedDate, &m.UpdatedDate, &m.DeletedDate, &dest, &m.Img)
	if err != nil {
		return m, err
	}
	for _, v := range m.Details {
		c, err := json.Marshal(v["visit_list"])
		if err != nil {
			return m, err
		}
		m.DayPeriod = int32(strings.Count(string(c), "]"))
	}
	m.Destination = dest.String

	return m, err
}

// AddMemberItin add new itinerary by members
func (c *Contract) AddMemberItin(tx pgx.Tx, ctx context.Context, m MemberItinEnt) (MemberItinEnt, error) {
	var lastInsID int32
	timeStamp := time.Now().In(time.UTC)

	sql := `INSERT INTO member_itins(itin_code, title, destination, created_by, est_price, start_date, end_date, details, created_date, img) 
		VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id`

	err := tx.QueryRow(ctx, sql, m.ItinCode, m.Title, m.Destination, m.CreatedBy, m.EstPrice, m.StartDate, m.EndDate, m.Details, timeStamp, m.Img).Scan(&lastInsID)

	m.ID = lastInsID

	return m, err
}

// DelMemberItin ...
func (c *Contract) DelMemberItin(tx pgx.Tx, ctx context.Context, code string) error {
	_, err := tx.Exec(ctx,
		"update member_itins set deleted_date=$1 where itin_code=$2", time.Now().In(time.UTC), code)

	return err
}

// GetListItinmember ...
func (c *Contract) GetListItinMember(db *pgxpool.Conn, ctx context.Context, param map[string]interface{}) ([]MemberItinEnt, error) {
	list := []MemberItinEnt{}
	var where []string
	var paramQuery []interface{}
	var destination sql.NullString
	var groupMembersTemp []map[string]interface{}

	query := `select
		mi.itin_code, 
		mi.title, 
		mi.destination, 
		mi.est_price, 
		mi.start_date, 
		mi.end_date, 
		mi.img,
		mi.details,
		mi.created_date,
		mg.member_name, 
		mg.member_code,
		mg.group_members
	from member_itins mi
	join (
		select
			m2.id,
			m2.name member_name,
			m2.member_code,
			(select 
				array_to_json(array_agg(row_to_json(groups_itin)))
			from (
				select
					mi.itin_code,
					m.member_code, 
					m.name member_name, 
					m.username member_username, 
					m.email member_email,
					m.img member_img
				from members m
				join member_itins mi on mi.created_by = m.id 
				union
				select
					mi.itin_code,
					m.member_code, 
					m.name member_name, 
					m.username member_username, 
					m.email member_email,
					m.img member_img
				from member_itin_relations mir
				join member_itins mi on mi.id = mir.member_itin_id 
				join members m on m.id = mir.member_id 
				where mir.deleted_date is null
				union
				select
					mi.itin_code,
					m.member_code, 
					m.name member_name, 
					m.username member_username, 
					mt.email member_email,
					m.img member_img
				from member_temporaries mt
				left join member_itins mi on mi.id = mt.member_itin_id 
				left join members m on m.email = mt.email 
			) groups_itin
		) group_members
		from members m2
	) mg on mg.id = mi.created_by `

	var queryString string = query

	if len(param["member_code"].(string)) > 0 {
		var orWhere []string
		orWhere = append(orWhere, " mg.member_code = '"+param["member_code"].(string)+"'")

		where = append(where, strings.Join(orWhere, " AND "))
	}

	if len(param["keyword"].(string)) > 0 {
		var orWhere []string
		orWhere = append(orWhere, "lower(mg.member_name) like lower('%"+param["keyword"].(string)+"%')")
		orWhere = append(orWhere, "lower(mi.title) like lower('%"+param["keyword"].(string)+"%')")
		orWhere = append(orWhere, "lower(mi.itin_code) like lower('%"+param["keyword"].(string)+"%')")
		orWhere = append(orWhere, "lower(mi.destination) like lower('%"+param["keyword"].(string)+"%')")

		where = append(where, "("+strings.Join(orWhere, " OR ")+")")
	}

	if len(param["created_by"].(string)) > 0 {
		var orWhere []string
		orWhere = append(orWhere, "mi.created_by = "+param["created_by"].(string))

		where = append(where, "("+strings.Join(orWhere, " AND ")+")")
	}

	if len(where) > 0 {
		queryString += " WHERE " + strings.Join(where, " AND ")
	}

	{
		var count int
		newQcount := `SELECT COUNT(*) FROM ( ` + queryString + ` ) AS data`

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
		queryString += " ORDER BY " + param["order"].(string) + " " + param["sort"].(string)
	} else {
		queryString += " ORDER BY " + param["order"].(string) + " " + param["sort"].(string) + " offset $1 limit $2 "
		paramQuery = append(paramQuery, param["offset"])
		paramQuery = append(paramQuery, param["limit"])
	}

	rows, err := db.Query(ctx, queryString, paramQuery...)
	if err != nil {
		return list, err
	}
	defer rows.Close()

	for rows.Next() {
		var m MemberItinEnt
		err = rows.Scan(&m.ItinCode, &m.Title, &destination, &m.EstPrice, &m.StartDate, &m.EndDate, &m.Img, &m.Details, &m.CreatedDate, &m.MemberEnt.Name, &m.MemberEnt.MemberCode, &groupMembersTemp)
		if err != nil {
			return list, err
		}

		// mapping visit details
		for _, v := range m.Details {
			c, err := json.Marshal(v["visit_list"])
			if err != nil {
				return list, err
			}
			m.DayPeriod = int32(strings.Count(string(c), "]"))
		}

		// mapping group member
		for _, member := range groupMembersTemp {
			if m.ItinCode == member["itin_code"] {
				member["is_owner"] = false
				if member["member_code"] == m.MemberEnt.MemberCode {
					member["is_owner"] = true
				}
				m.GroupMembers = append(m.GroupMembers, member)
			}
		}

		m.Destination = destination.String

		list = append(list, m)
	}

	return list, nil
}

// UpdateMemberItin edit itinerary
func (c *Contract) UpdateMemberItin(tx pgx.Tx, ctx context.Context, m MemberItinEnt, code string) (MemberItinEnt, error) {
	var ID int32
	timeStamp := time.Now().In(time.UTC)

	sql := `UPDATE member_itins SET title=$1, destination=$2, est_price=$3, start_date=$4, end_date=$5, details=$6, updated_date=$7, img=$8 WHERE itin_code=$9 RETURNING id`

	err := tx.QueryRow(ctx, sql, m.Title, m.Destination, m.EstPrice, m.StartDate, m.EndDate, m.Details, timeStamp, m.Img, code).Scan(&ID)

	m.ID = ID

	return m, err
}

// GetMemberItinWithGroupsByCode
func (c *Contract) GetMemberItinWithGroupsByCode(db *pgxpool.Conn, ctx context.Context, code string) (MemberItinEnt, error) {
	var m MemberItinEnt
	var dest sql.NullString
	var groupMembersTemp []map[string]interface{}

	query := `select
		mi.id,
		mi.itin_code, 
		mi.title, 
		mi.destination, 
		mi.created_by,
		mi.est_price, 
		mi.start_date, 
		mi.end_date, 
		mi.img,
		mi.details,
		mi.created_date,
		mi.updated_date,
		mi.deleted_date,
		mg.member_name, 
		mg.member_code,
		mg.group_members
	from member_itins mi
	join (
		select
			m2.id,
			m2.name member_name,
			m2.member_code,
			(select 
				array_to_json(array_agg(row_to_json(groups_itin)))
			from (
				select
					mi.itin_code,
					m.member_code, 
					m.name member_name, 
					m.username member_username, 
					m.email member_email,
					m.img member_img
				from members m
				join member_itins mi on mi.created_by = m.id 
				union
				select
					mi.itin_code,
					m.member_code, 
					m.name member_name, 
					m.username member_username, 
					m.email member_email,
					m.img member_img
				from member_itin_relations mir
				join member_itins mi on mi.id = mir.member_itin_id 
				join members m on m.id = mir.member_id 
				where mir.deleted_date is null
				union
				select
					mi.itin_code,
					m.member_code, 
					m.name member_name, 
					m.username member_username, 
					mt.email member_email,
					m.img member_img
				from member_temporaries mt
				left join member_itins mi on mi.id = mt.member_itin_id 
				left join members m on m.email = mt.email 
			) groups_itin
		) group_members
		from members m2
	) mg on mg.id = mi.created_by where mi.itin_code=$1 limit 1`

	err := db.QueryRow(ctx, query, code).Scan(&m.ID, &m.ItinCode, &m.Title, &dest, &m.CreatedBy, &m.EstPrice, &m.StartDate, &m.EndDate, &m.Img, &m.Details, &m.CreatedDate, &m.UpdatedDate, &m.DeletedDate, &m.MemberEnt.Name, &m.MemberEnt.MemberCode, &groupMembersTemp)
	if err != nil {
		return m, err
	}

	// mapping visit list
	for _, v := range m.Details {
		c, err := json.Marshal(v["visit_list"])
		if err != nil {
			return m, err
		}
		m.DayPeriod = int32(strings.Count(string(c), "]"))
	}

	// mapping group member
	for _, member := range groupMembersTemp {
		if m.ItinCode == member["itin_code"] {
			member["is_owner"] = false
			if member["member_code"] == m.MemberEnt.MemberCode {
				member["is_owner"] = true
			}
			m.GroupMembers = append(m.GroupMembers, member)
		}
	}

	m.Destination = dest.String

	return m, err
}
