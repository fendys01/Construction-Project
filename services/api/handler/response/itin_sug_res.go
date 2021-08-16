package response

import (
	"panorama/services/api/model"
	"strconv"
	"time"
)

// ItinSugResponse ...
type DetailItinSugResponse struct {
	ItinCode     string                   `json:"itin_code"`
	Title        string                   `json:"title"`
	Content      string                   `json:"content"`
	Price        int64                    `json:"price"`
	DayPeriod    string                   `json:"day_period"`
	TotalVisited int32                    `json:"total_visited"`
	Img          string                   `json:"img"`
	StartDate    time.Time                `json:"start_date"`
	EndDate      time.Time                `json:"end_date"`
	Details      []map[string]interface{} `json:"detail"`
}

// Transform from itin suggetion model to itin suggetion response
func (r DetailItinSugResponse) Transform(i model.ItinSugEnt) DetailItinSugResponse {

	r.ItinCode = i.ItinCode
	r.Title = i.Title
	r.Content = i.Content
	r.Price = i.Price
	r.DayPeriod = strconv.Itoa(int(i.DayPeriod)) + "D" + strconv.Itoa(int(i.DayPeriod-1)) + "N"
	r.TotalVisited = 0
	r.Img = i.Img.String
	r.StartDate = i.StartDate
	r.EndDate = i.EndDate
	r.Details = i.Details

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
}

// Transform from itin suggetion model to itin suggetion response
func (r ItinSugResponse) Transform(i model.ItinSugEnt) ItinSugResponse {

	r.ItinCode = i.ItinCode
	r.Name = i.UserEnt.Name
	r.Title = i.Title
	r.CreatedDate = i.CreatedDate
	r.DayPeriod = strconv.Itoa(int(i.DayPeriod)) + "D" + strconv.Itoa(int(i.DayPeriod-1)) + "N"
	r.TotalVisited = 0
	r.Img = i.Img.String

	return r
}
