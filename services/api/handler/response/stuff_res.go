package response

import (
	"panorama/services/api/model"

	"github.com/spf13/viper"
)

// MemberResponse ...
type StuffResponse struct {
	Code       		string `json:"code"`
	Name 	   		string `json:"name"`
	Image      		string  `json:"image"`
	Description		string `json:"description"`
	Price    		string `json:"price"`
	Type       		int32 `json:"type"`
}

// Transform from member model to member response
func (r StuffResponse) Transform(m model.StuffEnt) StuffResponse {

	r.Code = m.Code
	r.Name = m.Name
	
	if len(m.Image.String) > 0 {
		if IsUrl(m.Image.String) {
			r.Image = m.Image.String
		} else {
			r.Image = viper.GetString("aws.s3.public_url") + m.Image.String
		}

	} else {
		r.Image = ""
	}

	r.Description = m.Description
	r.Price = m.Price
	r.Type = m.Type

	return r
}

type StuffListResponse struct {
	Name 	   		string `json:"name"`
	Image      		string  `json:"image"`
	Description		string `json:"description"`
	Price    		string `json:"price"`
	CreatedDate     string `json:"created_date"`
}

// Transform from order model to itin order member response
func (r StuffListResponse) Transform(i model.StuffEnt) StuffListResponse {
	time := i.CreatedDate
	timeday := time.Format("January 02, 2006")


	r.Name = i.Name
	if len(i.Image.String) > 0 {
		if IsUrl(i.Image.String) {
			r.Image = i.Image.String
		} else {
			r.Image = viper.GetString("aws.s3.public_url") + i.Image.String
		}

	} else {
		r.Image = ""
	}
	r.Description = i.Description
	r.Price = i.Price
	r.CreatedDate = timeday

	return r
}

type StuffDetailResponse struct {
	Name 	   		string `json:"name"`
	Image      		string  `json:"image"`
	Description		string `json:"description"`
	Price    		string `json:"price"`
}

// Transform from order model to itin order member response
func (r StuffDetailResponse) Transform(i model.StuffEnt) StuffDetailResponse {

	r.Name = i.Name
	if len(i.Image.String) > 0 {
		if IsUrl(i.Image.String) {
			r.Image = i.Image.String
		} else {
			r.Image = viper.GetString("aws.s3.public_url") + i.Image.String
		}

	} else {
		r.Image = ""
	}
	r.Description = i.Description
	r.Price = i.Price

	return r
}