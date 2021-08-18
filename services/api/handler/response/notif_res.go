package response

import (
	"panorama/lib/utils"
	"panorama/services/api/model"
	"strconv"
	"time"
)

// MemberResponse ...
type NotifResponse struct {
	Code       string `json:"code"`
	MemberCode string `json:"member_code"`
	Type       int32  `json:"type"`
	Title      string `json:"title"`
	Content    string `json:"content"`
	Link       string `json:"link"`
}

// Transform from member model to member response
func (r NotifResponse) Transform(m model.NotificationEnt) NotifResponse {

	r.Code = m.Code
	r.MemberCode = m.MemberCode
	r.Type = m.Type
	r.Title = m.Title
	r.Content = m.Content
	r.Link = m.Link

	return r
}

type NotifItinResponse struct {
	Code              string    `json:"code"`
	MemberCode        string    `json:"member_code"`
	MemberName        string    `json:"member_name"`
	Type              int32     `json:"type"`
	Title             string    `json:"title"`
	Content           string    `json:"content"`
	Link              string    `json:"link"`
	IsRead            bool      `json:"is_read"`
	CreatedDate       time.Time `json:"created_date"`
	ItinCode          string    `json:"itin_code"`
	ItinTitle         string    `json:"itin_title"`
	ItinStartDate     time.Time `json:"itin_start_date"`
	ItinEndDate       time.Time `json:"itin_end_date"`
	ItinTcCode        string    `json:"itin_tc_code"`
	ItinTcName        string    `json:"itin_tc_name"`
	ItinTcCodeChanged string    `json:"itin_tc_code_changed"`
	ItinTcNameChanged string    `json:"itin_tc_name_changed"`
	TimeElapsed       string    `json:"time_elapsed"`
	DayPeriod         string    `json:"day_period"`
}

// Transform from member model to member response
func (r NotifItinResponse) Transform(m model.NotificationEnt) NotifItinResponse {
	r.Code = m.Code
	r.MemberCode = m.MemberCode
	r.MemberName = m.Member.Name
	r.Type = m.Type
	r.Title = m.Title
	r.Content = m.Content
	r.Link = m.Link
	r.IsRead = m.IsRead != 0
	r.ItinCode = m.MemberItin.ItinCode
	r.ItinTitle = m.MemberItin.Title
	r.ItinStartDate = m.MemberItin.StartDate
	r.ItinEndDate = m.MemberItin.EndDate
	r.ItinTcCode = m.User.UserCode
	r.ItinTcName = m.User.Name
	r.ItinTcCodeChanged = m.MemberItinChanges.User.UserCode
	r.ItinTcNameChanged = m.MemberItinChanges.User.Name
	r.CreatedDate = m.CreatedDate
	r.TimeElapsed = utils.TimeElapsed(m.CreatedDate)
	r.DayPeriod = "(" + strconv.Itoa(int(m.MemberItin.DayPeriod)) + "D" + strconv.Itoa(int(m.MemberItin.DayPeriod-1)) + "N)"

	return r
}
