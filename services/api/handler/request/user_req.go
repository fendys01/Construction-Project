package request

import (
	"panorama/services/api/model"
	"strconv"
)

type NewUserReq struct {
	Name     string `json:"name" validate:"required"`
	Email    string `json:"email" validate:"required"`
	Pass     string `json:"password" validate:"required"`
	Phone    string `json:"phone" validate:"required"`
	IsActive bool   `json:"is_active"`
	Role     string `json:"role" validate:"required"`
	Img      string `json:"img"`
}

type UserReqUpdate struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Pass     string `json:"password"`
	Phone    string `json:"phone"`
	IsActive string `json:"is_active"`
	Role     string `json:"role" `
	Img      string `json:"img"`
}

// Transform MCUserReq to MCUserEnt
func (u UserReqUpdate) Transform(m model.UserEnt) model.UserEnt {

	if len(u.Name) > 0 {
		m.Name = u.Name
	}

	if len(u.Email) > 0 {
		m.Email = u.Email
	}

	if len(u.Phone) > 0 {
		m.Phone = u.Phone
	}

	if len(u.IsActive) > 0 {
		a, _ := strconv.ParseBool(u.IsActive)
		m.IsActive = a
	}

	if len(u.Img) > 0 {
		m.Img.String = u.Img
	}

	return m
}
