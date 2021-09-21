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
	OrderCode     string
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
	var dest, orderCode sql.NullString
	sql := `select mi.*, o.order_code
			from member_itins mi
			left join orders o on o.member_itin_id = mi.id where mi.itin_code=$1 limit 1`
	err := db.QueryRow(ctx, sql, code).Scan(&m.ID, &m.ItinCode, &m.Title, &m.CreatedBy, &m.EstPrice, &m.StartDate, &m.EndDate, &m.Details, &m.CreatedDate, &m.UpdatedDate, &m.DeletedDate, &dest, &m.Img, &orderCode)
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
	m.OrderCode = orderCode.String

	return m, err
}

// AddMemberItin add new itinerary by members
func (c *Contract) AddMemberItin(tx pgx.Tx, ctx context.Context, m MemberItinEnt, orderType string) (MemberItinEnt, error) {
	var lastInsID int32
	var paramQuery []interface{}

	timeStamp := time.Now().In(time.UTC)

	sql := `insert into member_itins(itin_code, title, destination, created_by, est_price, start_date, end_date, details, created_date, img) values($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id`

	if orderType == ORDER_TYPE_CUSTOM {
		paramQuery = append(paramQuery, m.ItinCode, m.Title, nil, m.CreatedBy, m.EstPrice, nil, nil, m.Details, timeStamp, m.Img)
	} else {
		paramQuery = append(paramQuery, m.ItinCode, m.Title, m.Destination, m.CreatedBy, m.EstPrice, m.StartDate, m.EndDate, m.Details, timeStamp, m.Img)
	}

	err := tx.QueryRow(ctx, sql, paramQuery...).Scan(&lastInsID)

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
	var destination, chatGroupCode, memberName, memberCode sql.NullString
	var startDate, endDate sql.NullTime

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
		ig.name member_name,
		ig.member_code,
		cg.chat_group_code,
		mg.groups_itin_member
	from (
		select
			mi.itin_code,
			mi.created_by,
			m.member_code, 
			m.name, 
			m.username, 
			m.email,
			m.img member_img
		from members m
		join member_itins mi on mi.created_by = m.id 
		union
		select
			mi.itin_code,
			mi.created_by,
			m.member_code, 
			m.name, 
			m.username, 
			m.email,
			m.img member_img
		from member_itin_relations mir
		join member_itins mi on mi.id = mir.member_itin_id 
		join members m on m.id = mir.member_id 
		where mir.deleted_date is null
		union
		select
			mi.itin_code,
			mi.created_by,
			m.member_code, 
			m.name, 
			m.username, 
			mt.email,
			m.img member_img
		from member_temporaries mt
		left join member_itins mi on mi.id = mt.member_itin_id 
		left join members m on m.email = mt.email
	) ig
	join member_itins mi on mi.itin_code = ig.itin_code and mi.deleted_date is null
	join members m on m.id = ig.created_by and m.deleted_date is null
	left join chat_groups cg on cg.member_itin_id = mi.id
	left join (
		select
			groups_itin.itin_code,
			array_to_json(array_agg(row_to_json(groups_itin))) groups_itin_member
		from (
			select
				mi.itin_code,
				m.member_code, 
				m.name member_name, 
				m.username member_username, 
				m.email member_email,
				m.img member_img,
				case 
					when m.member_code is not null then true
					else true 
				end is_owner,
				case 
					when m.member_code is null then false
					else false 
				end is_temporary
			from members m
			join member_itins mi on mi.created_by = m.id 
			union
			select
				mi.itin_code,
				m.member_code, 
				m.name member_name, 
				m.username member_username, 
				m.email member_email,
				m.img member_img,
				case 
					when m.member_code is not null then false
					else false 
				end is_owner,
				case 
					when m.member_code is null then false
					else false 
				end is_temporary
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
				m.img member_img,
				case 
					when m.member_code is not null then false
					else false 
				end is_owner,
				case 
					when mt.email is not null then true
					else true 
				end is_temporary
			from member_temporaries mt
			left join member_itins mi on mi.id = mt.member_itin_id 
			left join members m on m.email = mt.email 
		) groups_itin
		group by groups_itin.itin_code
	) mg on mg.itin_code = ig.itin_code `

	var queryString string = query

	if len(param["member_code"].(string)) > 0 {
		var orWhere []string
		orWhere = append(orWhere, " ig.member_code = '"+param["member_code"].(string)+"'")

		where = append(where, strings.Join(orWhere, " AND "))
	}

	if len(param["keyword"].(string)) > 0 {
		var orWhere []string
		orWhere = append(orWhere, " lower(ig.name) like lower('%"+param["keyword"].(string)+"%')")
		orWhere = append(orWhere, " lower(mi.title) like lower('%"+param["keyword"].(string)+"%')")
		orWhere = append(orWhere, " lower(mi.itin_code) like lower('%"+param["keyword"].(string)+"%')")
		orWhere = append(orWhere, " lower(mi.destination) like lower('%"+param["keyword"].(string)+"%')")

		where = append(where, "("+strings.Join(orWhere, " OR ")+")")
	}

	if len(param["created_by"].(string)) > 0 {
		var orWhere []string
		orWhere = append(orWhere, " m.member_code = '"+param["created_by"].(string)+"'")
		orWhere = append(orWhere, " ig.member_code = m.member_code")

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
		err = rows.Scan(&m.ItinCode, &m.Title, &destination, &m.EstPrice, &startDate, &endDate, &m.Img, &m.Details, &m.CreatedDate, &memberName, &memberCode, &chatGroupCode, &m.GroupMembers)
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

		m.MemberEnt.Name = memberName.String
		m.MemberEnt.MemberCode = memberCode.String
		m.Destination = destination.String
		m.ChatGroupCode = chatGroupCode.String
		m.StartDate = startDate.Time
		m.EndDate = endDate.Time

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
	var dest, cgCode, orderCode sql.NullString
	var startDate, endDate sql.NullTime

	query := `select
		mi.id,
		mi.itin_code, 
		mi.title, 
		mi.destination, 
		mi.est_price, 
		mi.start_date, 
		mi.end_date, 
		mi.img,
		mi.details,
		mi.created_date,
		mi.updated_date,
		mi.deleted_date,
		m.name member_name,
		m.member_code,
		cg.chat_group_code,
		o.order_code,
		mg.groups_itin_member
	from member_itins mi 
	join members m on m.id = mi.created_by and m.deleted_date is null
	left join chat_groups cg on cg.member_itin_id = mi.id
	left join orders o on o.member_itin_id = mi.id
	left join (
		select
			groups_itin.itin_code,
			array_to_json(array_agg(row_to_json(groups_itin))) groups_itin_member
		from (
			select
				mi.itin_code,
				m.member_code, 
				m.name member_name, 
				m.username member_username, 
				m.email member_email,
				m.img member_img,
				case 
					when m.member_code is not null then true
					else true 
				end is_owner,
				case 
					when m.member_code is null then false
					else false 
				end is_temporary
			from members m
			join member_itins mi on mi.created_by = m.id 
			union
			select
				mi.itin_code,
				m.member_code, 
				m.name member_name, 
				m.username member_username, 
				m.email member_email,
				m.img member_img,
				case 
					when m.member_code is not null then false
					else false 
				end is_owner,
				case 
					when m.member_code is null then false
					else false 
				end is_temporary
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
				m.img member_img,
				case 
					when m.member_code is not null then false
					else false 
				end is_owner,
				case 
					when mt.email is not null then true
					else true 
				end is_temporary
			from member_temporaries mt
			left join member_itins mi on mi.id = mt.member_itin_id 
			left join members m on m.email = mt.email 
		) groups_itin
		group by groups_itin.itin_code
	) mg on mg.itin_code = mi.itin_code
	where mi.itin_code=$1 limit 1`

	err := db.QueryRow(ctx, query, code).Scan(&m.ID, &m.ItinCode, &m.Title, &dest, &m.EstPrice, &startDate, &endDate, &m.Img, &m.Details, &m.CreatedDate, &m.UpdatedDate, &m.DeletedDate, &m.MemberEnt.Name, &m.MemberEnt.MemberCode, &cgCode, &orderCode, &m.GroupMembers)
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

	m.Destination = dest.String
	m.ChatGroupCode = cgCode.String
	m.OrderCode = orderCode.String
	m.StartDate = startDate.Time
	m.EndDate = endDate.Time

	return m, err
}
