package model

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
)

type DashboardEnt struct {
	Order      OrderEnt
	MemberItin MemberItinEnt
	User	   UserEnt
	DailyVisitsEnt	[]DailyVisitsEnt
	LogApp 	LogAppEnt
}

type LogAppEnt struct {
	LastActiveDate  	int32
}

// Get Dashboard Admin
func (c *Contract) GetDashboardAdmin(db *pgxpool.Conn, ctx context.Context) (DashboardEnt, error) {
	var d DashboardEnt

	sql := `select (select
						count(itin_code) as boooked_trips
					from member_itins 
					join orders o on o.member_itin_id = member_itins.id),

					(select
						count(itin_code) as active_trips
					from member_itins 
					where now() between start_date and end_date),

					(select 
						count(last_active_date) as users_online
						from log_visit_app 
						where role = 'admin'),
	
					(select 
						count(last_active_date) as tc_online
						from log_visit_app 
						where role = 'tc')`

	err := db.QueryRow(ctx, sql).Scan(&d.Order.MemberItinID, &d.MemberItin.ID, &d.LogApp.LastActiveDate, &d.LogApp.LastActiveDate)

	return d, err
}

// Get Dashboard TC
func (c *Contract) GetDashboardTc(db *pgxpool.Conn, ctx context.Context) (DashboardEnt, error) {
	var d DashboardEnt

	sql := `select	(select 
						count(last_active_date) as users_online
						from log_visit_app 
						where role = 'admin'),

					(select 
						count(last_active_date) as tc_online
						from log_visit_app 
						where role = 'tc')`

	err := db.QueryRow(ctx, sql).Scan(&d.LogApp.LastActiveDate, &d.LogApp.LastActiveDate)

	return d, err
}