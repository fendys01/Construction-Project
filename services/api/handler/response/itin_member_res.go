package response

import (
	"panorama/services/api/model"
	"strconv"
	"time"

	"github.com/spf13/viper"
)

// ItinmemberResponse ...
type ItinMemberResponse struct {
	ItinCode     string                   `json:"itin_code"`
	MemberCode   string                   `json:"member_code"`
	Name         string                   `json:"name"`
	Destination  string                   `json:"destination"`
	Title        string                   `json:"title"`
	EstPrice     int64                    `json:"est_price"`
	StartDate    time.Time                `json:"start_date"`
	EndDate      time.Time                `json:"end_date"`
	CreatedDate  time.Time                `json:"created_date"`
	DayPeriod    string                   `json:"day_period"`
	Img          string                   `json:"img"`
	Details      []map[string]interface{} `json:"detail"`
	GroupMembers []map[string]interface{} `json:"group_members"`
}

// Transform from itin member model to itin member response
func (r ItinMemberResponse) Transform(i model.MemberItinEnt) ItinMemberResponse {

	r.ItinCode = i.ItinCode
	r.MemberCode = i.MemberEnt.MemberCode
	r.Name = i.MemberEnt.Name
	r.Title = i.Title
	r.Destination = i.Destination
	r.CreatedDate = i.CreatedDate
	r.DayPeriod = strconv.Itoa(int(i.DayPeriod)) + "D" + strconv.Itoa(int(i.DayPeriod-1)) + "N"
	r.EstPrice = i.EstPrice.Int64
	r.StartDate = i.StartDate
	r.EndDate = i.EndDate
	r.Details = i.Details
	r.GroupMembers = i.GroupMembers

	if len(i.Img.String) > 0 {
		if IsUrl(i.Img.String) {
			r.Img = i.Img.String
		} else {
			r.Img = viper.GetString("aws.s3.public_url") + i.Img.String
		}

	} else {
		r.Img = ""
	}

	return r
}
