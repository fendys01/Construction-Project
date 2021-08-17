package response

import (
	"Contruction-Project/services/api/model"
)

// UsersResponse ...
type UsersResponse struct {
	UserCode      string `json:"user_code"`
	Name          string `json:"name"`
	Email         string `json:"email"`
	Phone         string `json:"phone"`
	Img           string `json:"image"`
	Role          string `json:"role"`
	LastVisitDate string `json:"last_visit_date"`
	TotalClient   int32  `json:"total_client"`
}

// Transform from member model to member response
func (r UsersResponse) Transform(m model.UserEnt) UsersResponse {

	r.UserCode = m.UserCode
	r.Name = m.Name
	r.Email = m.Email
	r.Phone = m.Phone
	r.Img = m.Img.String
	r.LastVisitDate = m.LastVisit.Format("Jan 02,2006")
	r.TotalClient = m.TotalClient

	return r
}

// MemberResponse ...
type UsersDetailResponse struct {
	UserCode           string               `json:"user_code"`
	Name               string               `json:"name"`
	Email              string               `json:"email"`
	Phone              string               `json:"phone"`
	Img                string               `json:"image"`
	DetailActivityUser DetailActivityUser   `json:"detail_activity"`
	RecentActivityUser []RecentActivityUser `json:"recent_activity"`
}

// Transform from member model to member response
func (r UsersDetailResponse) Transform(m model.UserEnt) UsersDetailResponse {

	r.UserCode = m.UserCode
	r.Name = m.Name
	r.Email = m.Email
	r.Phone = m.Phone
	r.Img = m.Img.String
	r.DetailActivityUser = r.DetailActivityUser.Transform(m)

	// var listResponse []MoreActivityMember
	// for _, g := range r.MoreActivityMember {
	// 	var res MoreActivityMember
	// 	res = res.Transform(g)
	// 	listResponse = append(listResponse, res)
	// }
	var l []RecentActivityUser
	var res RecentActivityUser
	l = append(l, res.Transform())
	r.RecentActivityUser = l

	return r
}

// DetailActivityMember ...
type DetailActivityUser struct {
	TotalClient      int32 `json:"total_client"`
	OverAllView      int32 `json:"overall_view"`
	ItineraryCreated int32 `json:"itin_created"`
}

// Transform from member model to member response
func (r DetailActivityUser) Transform(m model.UserEnt) DetailActivityUser {

	r.TotalClient = 30
	r.OverAllView = 1200
	r.ItineraryCreated = 100

	return r
}

// RecentActivityUser ...
type RecentActivityUser struct {
	Title       string `json:"title"`
	TagActivity string `json:"tag_activity"`
	TagContent  string `json:"tag_content"`
	Date        string `json:"date"`
	LastActive  string `json:"last_active"`
}

// Transform from member model to member response
func (r RecentActivityUser) Transform() RecentActivityUser {

	r.Title = "Request For assistan"
	r.TagActivity = "Trip"
	r.TagContent = "Summer 2021"
	r.Date = "2021-08-01"
	r.LastActive = "Now"

	return r
}
