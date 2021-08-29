package request

import (
	"panorama/services/api/model"
	"strconv"
)

// AddMemberReq ...
type AddStuffReq struct {
	Name 			string `json:"name" validate:"required"`
	Image           string `json:"image" validate:"required"`
	Description     string `json:"description" validate:"required"`
	Price        	string `json:"price" validate:"required"`
	Type            int32  `json:"type"`
}

type StuffReqUpdate struct {
	Name     		string `json:"name"`
	Image    		string `json:"email"`
	Description     string `json:"password"`
	Price    		string `json:"phone"`
	Type     		int32  `json:"type"`
	IsActive 		string `json:"is_active"`
}

// Transform MCUserReq to MCUserEnt
func (s StuffReqUpdate) Transform(m model.StuffEnt) model.StuffEnt {

	if len(s.Name) > 0 {
		m.Name = s.Name
	}

	if len(s.Image) > 0 {
		m.Image.String = s.Image
	}


	if len(s.Description) > 0 {
		m.Description = s.Description
	}

	if len(s.Price) > 0 {
		m.Price = s.Price
	}

	// if len(s.Type) > 0 {
	// 	m.Type = s.Type
	// }

	if len(s.IsActive) > 0 {
		a, _ := strconv.ParseBool(s.IsActive)
		m.IsActive = a
	}

	return m
}