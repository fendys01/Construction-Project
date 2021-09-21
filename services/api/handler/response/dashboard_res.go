package response

import (
	"panorama/services/api/model"
)

// Dashboard Admin Response ...
type DashboardAdminResponse struct {
	DashboardBookedTripsResponse map[string]interface{}         `json:"booked_trips"`
	DashboardActiveTripsResponse map[string]interface{}         `json:"active_trips"`
	DashboardActiveChatsResponse map[string]interface{}         `json:"active_chats"`
	DashboardUsersOnlineResponse map[string]interface{}         `json:"users_online"`
	DashboardTcOnlineResponse    map[string]interface{}         `json:"tc_online"`
	DashboardDailyVisitsResponse []DashboardDailyVisitsResponse `json:"daily_visits"`
}

// Transform dashboard Admin
func (r DashboardAdminResponse) Transform(i model.DashboardEnt) DashboardAdminResponse {
	r.DashboardBookedTripsResponse = i.BookedTrips
	r.DashboardActiveTripsResponse = i.ActiveTrips
	r.DashboardActiveChatsResponse = i.ActiveChats
	r.DashboardUsersOnlineResponse = i.UsersOnline
	r.DashboardTcOnlineResponse = i.TcOnline

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
	DashboardActiveChatsResponse map[string]interface{}         `json:"active_chats"`
	DashboardUsersOnlineResponse map[string]interface{}         `json:"users_online"`
	DashboardTcOnlineResponse    map[string]interface{}         `json:"tc_online"`
	DashboardDailyVisitsResponse []DashboardDailyVisitsResponse `json:"daily_visits"`
}

// Transform dashboard TC
func (r DashboardTCResponse) Transform(i model.DashboardEnt) DashboardTCResponse {
	r.DashboardActiveChatsResponse = i.ActiveChats
	r.DashboardUsersOnlineResponse = i.UsersOnline
	r.DashboardTcOnlineResponse = i.TcOnline

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
	LastActiveDate string `json:"date"`
	TotalVisited   int32  `json:"total_visited"`
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
