package model

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

type DashboardEnt struct {
	BookedTrips    map[string]interface{}
	ActiveTrips    map[string]interface{}
	ActiveChats    map[string]interface{}
	UsersOnline    map[string]interface{}
	TcOnline       map[string]interface{}
	DailyVisitsEnt []DailyVisitsEnt
}

// Get Dashboard Admin
func (c *Contract) GetDashboardAdmin(db *pgxpool.Conn, ctx context.Context, param map[string]interface{}) (DashboardEnt, error) {
	var d DashboardEnt
	var paramQuery []interface{}

	sql := `select (
		select row_to_json(booked_trips) booked_trips
		from (
			select
				count(o.id) as total,	
				(
					select 
						count(o.id)			
					from orders o 
					where o.order_type = 'R'
					and o.created_date between $1 and $2
				) total_date
				from orders o
		) booked_trips
	),
	(
		select row_to_json(active_trips) active_trips
		from (
			select
				count(mi.id) as total,	
				(
					select 
						count(mi.id)			
					from member_itins mi 
					where mi.deleted_date is null
					and now() between mi.start_date and mi.end_date
					and mi.created_date between $1 and $2
				) total_date
			from member_itins mi
			where mi.deleted_date is null
			and now() between mi.start_date and mi.end_date
		) active_trips
	),
	(
		select row_to_json(active_chats) active_chats
		from (
			select
				count(cg.id) as total,	
				(
					select 
						count(mi.id)			
					from chat_groups cg 
					join member_itins mi on mi.id = cg.member_itin_id and mi.deleted_date is null
					where now() between mi.start_date and mi.end_date
					and cg.created_date between $1 and $2
				) total_date
			from chat_groups cg
			join member_itins mi on mi.id = cg.member_itin_id and mi.deleted_date is null
			where now() between mi.start_date and mi.end_date
		) active_chats
	),
	(	
		select row_to_json(users_online) users_online
		from (
			select
				count(lva.last_active_date) as total,	
				(
					select 
						count(lva.last_active_date)			
					from log_visit_app lva 
					where lva.role = 'admin'
					and lva.last_active_date between $1 and $2
				) total_date
			from log_visit_app lva
			where lva.role = 'admin'
		) users_online
	),
	(
		select row_to_json(tc_online) tc_online
		from (
			select
				count(lva.last_active_date) as total,	
				(
					select 
						count(lva.last_active_date)			
					from log_visit_app lva 
					where lva.role = 'tc'
					and lva.last_active_date between $1 and $2
				) total_date
			from log_visit_app lva
			where lva.role = 'tc'
		) tc_online
	)`

	startDate := fmt.Sprintf("%v", time.Now().Format("2006-01-02")) + " 00:00:00"
	endDate := fmt.Sprintf("%v", time.Now().Format("2006-01-02")) + " 23:59:59"

	if len(param["start_date"].(string)) > 0 && len(param["end_date"].(string)) > 0 {
		startDate = fmt.Sprintf("%v %s", param["start_date"], "00:00:00")
		endDate = fmt.Sprintf("%v %s", param["end_date"], "23:59:59")
	}

	paramQuery = append(paramQuery, startDate)
	paramQuery = append(paramQuery, endDate)

	err := db.QueryRow(ctx, sql, paramQuery...).Scan(&d.BookedTrips, &d.ActiveTrips, &d.ActiveChats, &d.UsersOnline, &d.TcOnline)

	return d, err
}

// Get Dashboard TC
func (c *Contract) GetDashboardTc(db *pgxpool.Conn, ctx context.Context, param map[string]interface{}) (DashboardEnt, error) {
	var d DashboardEnt
	var paramQuery []interface{}

	whereTc := ``
	if len(param["created_by"].(string)) > 0 {
		whereTc = `and cg.tc_id = ` + param["created_by"].(string)
	}

	sql := `select (
		select row_to_json(active_chats) active_chats
		from (
			select
				count(cg.id) as total,	
				(
					select 
						count(mi.id)			
					from chat_groups cg 
					join member_itins mi on mi.id = cg.member_itin_id and mi.deleted_date is null
					where now() between mi.start_date and mi.end_date
					and cg.created_date between $1 and $2 
					` + whereTc + `
				) total_date
			from chat_groups cg
			join member_itins mi on mi.id = cg.member_itin_id and mi.deleted_date is null
			where now() between mi.start_date and mi.end_date
			` + whereTc + `
		) active_chats
	),
	(
		select row_to_json(users_online) users_online
		from (
			select
				count(lva.last_active_date) as total,	
				(
					select 
						count(lva.last_active_date)			
					from log_visit_app lva 
					where lva.role = 'admin'
					and lva.last_active_date between $1 and $2
				) total_date
			from log_visit_app lva
			where lva.role = 'admin'
		) users_online
	),
	(
		select row_to_json(tc_online) tc_online
		from (
			select
				count(lva.last_active_date) as total,	
				(
					select 
						count(lva.last_active_date)			
					from log_visit_app lva 
					where lva.role = 'tc'
					and lva.last_active_date between $1 and $2
				) total_date
			from log_visit_app lva
			where lva.role = 'tc'
		) tc_online
	)`

	startDate := fmt.Sprintf("%v", time.Now().Format("2006-01-02")) + " 00:00:00"
	endDate := fmt.Sprintf("%v", time.Now().Format("2006-01-02")) + " 23:59:59"

	if len(param["start_date"].(string)) > 0 && len(param["end_date"].(string)) > 0 {
		startDate = fmt.Sprintf("%v %s", param["start_date"], "00:00:00")
		endDate = fmt.Sprintf("%v %s", param["end_date"], "23:59:59")
	}

	paramQuery = append(paramQuery, startDate)
	paramQuery = append(paramQuery, endDate)

	err := db.QueryRow(ctx, sql, paramQuery...).Scan(&d.ActiveChats, &d.UsersOnline, &d.TcOnline)

	return d, err
}
