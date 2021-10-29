package response

import (
	"panorama/services/api/model"
	"strings"

	"github.com/andanhm/go-prettytime"
	"github.com/spf13/viper"
)

// UsersResponse ...
type UsersResponse struct {
	UserCode string `json:"user_code"`
	Name     string `json:"name"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	// Password      string `json:"password"`
	Img           string `json:"image"`
	Role          string `json:"role"`
	LastVisitDate string `json:"last_visit_date"`
	TotalClient   int32  `json:"total_client"`
}

// Transform from member model to member response
func (r UsersResponse) Transform(m model.UserEnt) UsersResponse {

	var t string
	if m.LastVisit.Valid {
		t = m.LastVisit.Time.Format("Jan 02,2006")
	}

	if len(strings.TrimSpace(m.Img.String)) > 0 {
		if IsUrl(m.Img.String) {
			r.Img = m.Img.String
		} else {
			r.Img = viper.GetString("aws.s3.public_url") + m.Img.String
		}

	} else {
		r.Img = ""
	}

	r.UserCode = m.UserCode
	r.Name = m.Name
	r.Phone = m.Phone
	r.Email = m.Email
	// r.Password = m.Password
	r.Role = m.Role
	r.LastVisitDate = t
	r.TotalClient = m.TotalClient.Int32

	return r
}

// TcDetailResponse ...
type TcDetailResponse struct {
	UserCode              string                  `json:"user_code"`
	Name                  string                  `json:"name"`
	Email                 string                  `json:"email"`
	Phone                 string                  `json:"phone"`
	Img                   string                  `json:"image"`
	LastSeen              string                  `json:"last_seen"`
	SummaryActivityTc     SummaryActivityTc       `json:"summary_activity"`
	RecentActivityUser    []RecentActivityUser    `json:"log_activity_user"`
	ActiveClientConsultan []ActiveClientConsultan `json:"active_client_consultan"`
}

// TcDetailResponse ...
func (r TcDetailResponse) Transform(m model.UserEnt) TcDetailResponse {

	r.UserCode = m.UserCode
	r.Name = m.Name
	r.Email = m.Email
	r.Phone = m.Phone
	r.LastSeen = prettytime.Format(m.LastVisit.Time)
	r.SummaryActivityTc = r.SummaryActivityTc.Transform(m)

	if len(strings.TrimSpace(m.Img.String)) > 0 {
		if IsUrl(m.Img.String) {
			r.Img = m.Img.String
		} else {
			r.Img = viper.GetString("aws.s3.public_url") + m.Img.String
		}

	} else {
		r.Img = ""
	}

	var listResponse []RecentActivityUser
	for _, g := range m.LogActivityUser {
		var res RecentActivityUser
		res = res.Transform(g)
		listResponse = append(listResponse, res)
	}
	var act []ActiveClientConsultan
	for _, g := range m.ActiveClientConsultan {
		var resAct ActiveClientConsultan
		resAct = resAct.Transform(g)
		act = append(act, resAct)
	}

	r.RecentActivityUser = listResponse
	r.ActiveClientConsultan = act

	return r
}

// SummaryActivityTc ...
type SummaryActivityTc struct {
	TotalClient         int32 `json:"total_client"`
	TripBooked          int32 `json:"trip_booked"`
	CustomPackageBooked int32 `json:"custom_package_booked"`
}

// SummaryActivityTc ...
func (r SummaryActivityTc) Transform(m model.UserEnt) SummaryActivityTc {

	r.TotalClient = m.TotalClient.Int32
	r.TripBooked = m.TotalOrd.Int32
	r.CustomPackageBooked = m.TotalCustomOrder.Int32

	return r
}

// AdminDetailResponse ...
type AdminDetailResponse struct {
	UserCode             string               `json:"user_code"`
	Name                 string               `json:"name"`
	Email                string               `json:"email"`
	Img                  string               `json:"image"`
	Phone                string               `json:"phone"`
	LastSeen             string               `json:"last_seen"`
	SummaryActivityAdmin SummaryActivityAdmin `json:"summary_activity"`
	RecentActivityUser   []RecentActivityUser `json:"log_activity_user"`
}

// Transform from member model to member response
func (r AdminDetailResponse) Transform(m model.UserEnt) AdminDetailResponse {

	r.UserCode = m.UserCode
	r.Name = m.Name
	r.Email = m.Email
	r.Phone = m.Phone
	r.LastSeen = prettytime.Format(m.LastVisit.Time)
	r.SummaryActivityAdmin = r.SummaryActivityAdmin.Transform(m)

	if len(strings.TrimSpace(m.Img.String)) > 0 {
		if IsUrl(m.Img.String) {
			r.Img = m.Img.String
		} else {
			r.Img = viper.GetString("aws.s3.public_url") + m.Img.String
		}

	} else {
		r.Img = ""
	}

	var listResponse []RecentActivityUser
	for _, g := range m.LogActivityUser {
		var res RecentActivityUser
		res = res.Transform(g)
		listResponse = append(listResponse, res)
	}

	r.RecentActivityUser = listResponse

	return r
}

// SummaryActivityAdmin ...
type SummaryActivityAdmin struct {
	TotalItinSugView int32 `json:"total_itin_view"`
	TotalItinCreated int32 `json:"total_itin_created"`
}

// Transform from member model to member response
func (r SummaryActivityAdmin) Transform(m model.UserEnt) SummaryActivityAdmin {

	r.TotalItinSugView = m.TotalItinSugView.Int32
	r.TotalItinCreated = m.TotalItinSug.Int32

	return r
}

// RecentActivityUser ...
type RecentActivityUser struct {
	Title    string `json:"title"`
	Activity string `json:"activity"`
	Date     string `json:"date"`
}

// RecentActivityUser ...
func (r RecentActivityUser) Transform(log model.LogActivityUserEnt) RecentActivityUser {

	var t string
	if log.CreatedDate.Valid {
		t = log.CreatedDate.Time.Format("Jan 02,2006")
	}

	r.Title = log.Title
	r.Activity = log.Activity
	r.Date = t

	return r
}

// ActiveClientConsultan ...
type ActiveClientConsultan struct {
	Title      string `json:"title"`
	Name       string `json:"activity"`
	NewRequest string `json:"new_req_replacement"`
	ItinDate   string `json:"itin_date"`
}

// ActiveClientConsultan ...
func (r ActiveClientConsultan) Transform(v model.ActiveClientConsultan) ActiveClientConsultan {

	r.Title = v.Title
	r.Name = v.Name

	if v.TcReplacementDate.Valid {
		r.NewRequest = "New Request From Tc Replacement"
		r.ItinDate = v.TcReplacementDate.Time.Format("Jan 02,2006")
	} else {
		r.NewRequest = ""
		r.ItinDate = v.ItinDate.Format("Jan 02,2006")
	}

	return r
}
