package response

import (
	"Contruction-Project/services/api/model"
)

// MemberResponse ...
type MemberResponse struct {
	MemberCode     string `json:"member_code"`
	Username       string `json:"username"`
	Name           string `json:"name"`
	Email          string `json:"email"`
	Phone          string `json:"phone"`
	Img            string `json:"image"`
	LastActiveDate string `json:"last_active_date"`
	TotalVisited   int32  `json:"total_visited"`
}

// Transform from member model to member response
func (r MemberResponse) Transform(m model.MemberEnt) MemberResponse {

	r.MemberCode = m.MemberCode
	r.Username = m.Username
	r.Name = m.Name
	r.Email = m.Email
	r.Phone = m.Phone
	r.Img = m.Img.String
	r.LastActiveDate = m.LastActiveDate.Time.Format("Jan 02,2006")
	r.TotalVisited = m.TotalVisited

	return r
}

// MemberResponse ...
type MemberDetailResponse struct {
	MemberCode           string                 `json:"member_code"`
	Username             string                 `json:"username"`
	Name                 string                 `json:"name"`
	Email                string                 `json:"email"`
	Phone                string                 `json:"phone"`
	Img                  string                 `json:"image"`
	DetailActivityMember DetailActivityMember   `json:"detail_activity"`
	RecentActivityMember []RecentActivityMember `json:"recent_activity"`
}

// Transform from member model to member response
func (r MemberDetailResponse) Transform(m model.MemberEnt) MemberDetailResponse {

	r.MemberCode = m.MemberCode
	r.Username = m.Username
	r.Name = m.Name
	r.Email = m.Email
	r.Phone = m.Phone
	r.Img = m.Img.String
	r.DetailActivityMember = r.DetailActivityMember.Transform(m)

	// var listResponse []MoreActivityMember
	// for _, g := range r.MoreActivityMember {
	// 	var res MoreActivityMember
	// 	res = res.Transform(g)
	// 	listResponse = append(listResponse, res)
	// }
	var l []RecentActivityMember
	var res RecentActivityMember
	l = append(l, res.Transform())
	r.RecentActivityMember = l

	return r
}

// DetailActivityMember ...
type DetailActivityMember struct {
	LastActiveDate     string `json:"last_active_date"`
	TotalVisited       int32  `json:"total_visited"`
	TotalComplateTrips int32  `json:"total_complate_trips"`
	TotalTc            int32  `json:"total_tc"`
}

// Transform from member model to member response
func (r DetailActivityMember) Transform(m model.MemberEnt) DetailActivityMember {

	r.LastActiveDate = m.LastActiveDate.Time.Format("Jan 02,2006")
	r.TotalVisited = m.TotalVisited
	r.TotalComplateTrips = 2
	r.TotalTc = 2

	return r
}

// RecentActivityMember ...
type RecentActivityMember struct {
	Title       string `json:"title"`
	TagActivity string `json:"tag_activity"`
	TagContent  string `json:"tag_content"`
	Date        string `json:"date"`
	LastActive  string `json:"last_active"`
}

// Transform response
func (r RecentActivityMember) Transform() RecentActivityMember {

	r.Title = "Request For assistan"
	r.TagActivity = "Trip"
	r.TagContent = "Summer 2021"
	r.Date = "2021-08-01"
	r.LastActive = "Now"

	return r
}
