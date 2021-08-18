package response

import (
	"net/url"
	"panorama/services/api/model"
	"strconv"
	"time"

	"github.com/spf13/viper"
)

// ItinSugResponse ...
type DetailItinSugResponse struct {
	ItinCode     string                   `json:"itin_code"`
	Title        string                   `json:"title"`
	Content      string                   `json:"content"`
	DayPeriod    string                   `json:"day_period"`
	TotalVisited int32                    `json:"total_visited"`
	Img          string                   `json:"img"`
	Destination  string                   `json:"destination"`
	Details      []map[string]interface{} `json:"detail"`
}

// Transform from itin suggetion model to itin suggetion response
func (r DetailItinSugResponse) Transform(i model.ItinSugEnt) DetailItinSugResponse {

	r.ItinCode = i.ItinCode
	r.Title = i.Title
	r.Content = i.Content
	r.DayPeriod = strconv.Itoa(int(i.DayPeriod)) + "D" + strconv.Itoa(int(i.DayPeriod-1)) + "N"
	r.TotalVisited = i.View.Int32
	r.Destination = i.Destination
	r.Details = i.Details

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

// ItinSugResponse ...
type ItinSugResponse struct {
	ItinCode     string    `json:"itin_code"`
	Name         string    `json:"name"`
	Title        string    `json:"title"`
	CreatedDate  time.Time `json:"created_date"`
	DayPeriod    string    `json:"day_period"`
	TotalVisited int32     `json:"total_visited"`
	Img          string    `json:"img"`
	Destination  string    `json:"destination"`
}

// Transform from itin suggetion model to itin suggetion response
func (r ItinSugResponse) Transform(i model.ItinSugEnt) ItinSugResponse {

	r.ItinCode = i.ItinCode
	r.Name = i.UserEnt.Name
	r.Title = i.Title
	r.CreatedDate = i.CreatedDate
	r.DayPeriod = strconv.Itoa(int(i.DayPeriod)) + "D" + strconv.Itoa(int(i.DayPeriod-1)) + "N"
	r.TotalVisited = i.View.Int32
	r.Destination = i.Destination

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

// validate is url
func IsUrl(str string) bool {
	u, err := url.Parse(str)
	return err == nil && u.Scheme != "" && u.Host != ""
}
