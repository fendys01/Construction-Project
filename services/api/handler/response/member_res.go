package response

import (
	"panorama/services/api/model"

	"github.com/andanhm/go-prettytime"
	"github.com/spf13/viper"
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
	Token          string `json:"token"`
	IsActive       bool   `json:"is_active"`
}

// Transform from member model to member response
func (r MemberResponse) Transform(m model.MemberEnt) MemberResponse {

	var date string
	if m.MemberStatistik.LastActiveDate.Valid {
		date = m.MemberStatistik.LastActiveDate.Time.Format("Jan 02,2006")
	}

	if len(m.Img.String) > 0 {
		if IsUrl(m.Img.String) {
			r.Img = m.Img.String
		} else {
			r.Img = viper.GetString("aws.s3.public_url") + m.Img.String
		}

	} else {
		r.Img = ""
	}

	r.MemberCode = m.MemberCode
	r.Username = m.Username
	r.Name = m.Name
	r.Email = m.Email
	r.Phone = m.Phone
	r.LastActiveDate = date
	r.IsActive = m.IsActive
	r.TotalVisited = m.MemberStatistik.TotalVisited.Int32

	return r
}

// MemberDetailResponse ...
type MemberDetailResponse struct {
	MemberCode           string               `json:"member_code"`
	Name                 string               `json:"name"`
	Img                  string               `json:"image"`
	LastSeen             string               `json:"last_seen"`
	DetailActivityMember DetailActivityMember `json:"summary_activity"`
	RecentActivityUser   []RecentActivityUser `json:"log_activity"`
}

// Transform from member model to member response
func (r MemberDetailResponse) Transform(m model.MemberEnt) MemberDetailResponse {

	if len(m.Img.String) > 0 {
		if IsUrl(m.Img.String) {
			r.Img = m.Img.String
		} else {
			r.Img = viper.GetString("aws.s3.public_url") + m.Img.String
		}

	} else {
		r.Img = ""
	}

	var date string
	if m.MemberStatistik.LastActiveDate.Valid {
		date = prettytime.Format(m.MemberStatistik.LastActiveDate.Time)
	}

	r.LastSeen = date

	r.MemberCode = m.MemberCode
	r.Name = m.Name
	r.DetailActivityMember = r.DetailActivityMember.Transform(m)

	var listResponse []RecentActivityUser
	for _, g := range m.LogActivityUser {
		var res RecentActivityUser
		res = res.Transform(g)
		listResponse = append(listResponse, res)
	}

	r.RecentActivityUser = listResponse

	return r
}

// DetailActivityMember ...
type DetailActivityMember struct {
	TotalVisited       int32 `json:"total_app_visited"`
	TotalComplateTrips int32 `json:"total_complate_trips"`
	TotalTc            int32 `json:"total_tc"`
}

// DetailActivityMember ...
func (r DetailActivityMember) Transform(m model.MemberEnt) DetailActivityMember {

	r.TotalVisited = m.MemberStatistik.TotalVisited.Int32
	r.TotalComplateTrips = m.MemberStatistik.TotalCompletedItinerary.Int32
	r.TotalTc = m.MemberStatistik.TotalTc.Int32

	return r
}
