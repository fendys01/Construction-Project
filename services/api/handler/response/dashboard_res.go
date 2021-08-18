package response

import (
	"panorama/services/api/model"
)

// Dashboard Admin Response ...
type DashboardAdminResponse struct {
	DashboardBookedTripsResponse	int32	`json:"booked_trips"` 
	DashboardActiveTripsResponse	int32	`json:"active_trips"`
	DashboardUsersOnlineResponse	int32	`json:"users_online"`
	DashboardTcOnlineResponse		int32	`json:"tc_online"`
	DashboardDailyVisitsResponse	[]DashboardDailyVisitsResponse	`json:"daily_visits"`
}

// Transform dashboard Admin
func (r DashboardAdminResponse) Transform(i model.DashboardEnt) DashboardAdminResponse {	
	r.DashboardBookedTripsResponse = i.Order.MemberItinID
	r.DashboardActiveTripsResponse = i.MemberItin.ID
	r.DashboardUsersOnlineResponse = i.LogApp.LastActiveDate
	r.DashboardTcOnlineResponse    = i.LogApp.LastActiveDate
	
	var listResponse []DashboardDailyVisitsResponse
	for _, g := range i.DailyVisitsEnt {
		var res DashboardDailyVisitsResponse
		res = res.Transform(g)
		listResponse = append(listResponse, res)
	}

	r.DashboardDailyVisitsResponse = listResponse

   return r
}

// Dashboard TC Response ...
type DashboardTCResponse struct {
	DashboardUsersOnlineResponse	int32	`json:"users_online"`
	DashboardTcOnlineResponse		int32	`json:"tc_online"`
	DashboardDailyVisitsResponse	[]DashboardDailyVisitsResponse	`json:"daily_visits"`
}

// Transform dashboard TC
func (r DashboardTCResponse) Transform(i model.DashboardEnt) DashboardTCResponse {	
	r.DashboardUsersOnlineResponse = i.LogApp.LastActiveDate
	r.DashboardTcOnlineResponse    = i.LogApp.LastActiveDate
	
	var listResponse []DashboardDailyVisitsResponse
	for _, g := range i.DailyVisitsEnt {
		var res DashboardDailyVisitsResponse
		res = res.Transform(g)
		listResponse = append(listResponse, res)
	}

	r.DashboardDailyVisitsResponse = listResponse

   return r
}

// Daily Visits
type DashboardDailyVisitsResponse struct {
	LastActiveDate 	string		`json:"date"`
	TotalVisited	int32  		`json:"total_visited"`
}

// Transform dashboard Daily Visits
func (r DashboardDailyVisitsResponse) Transform(i model.DailyVisitsEnt) DashboardDailyVisitsResponse {
	var date string
	if i.LastActiveDate.Valid {
		date = i.LastActiveDate.Time.Format("2006-01-02")
	}

	r.LastActiveDate = date
	r.TotalVisited = i.TotalVisited.Int32
	
   return r
}

