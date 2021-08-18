package request

import "panorama/services/api/model"

// AddMemberReq ...
type AddMemberReq struct {
	Username       string `json:"username" validate:"required"`
	Name           string `json:"name" validate:"required"`
	Email          string `json:"email" validate:"required,email"`
	Phone          string `json:"phone" validate:"required"`
	Img            string `json:"image"`
	Password       string `json:"password" validate:"required"`
	RetypePassword string `json:"retype_password" validate:"required"`
}

// UpdateMemberReq ...
type UpdateMemberReq struct {
	Username string `json:"username"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Img      string `json:"image"`
}

// Transform MCUserReq to MCUserEnt
func (u UpdateMemberReq) Transform(m model.MemberEnt) model.MemberEnt {

	if len(u.Username) > 0 {
		m.Username = u.Username
	}

	if len(u.Name) > 0 {
		m.Name = u.Name
	}

	if len(u.Email) > 0 {
		m.Email = u.Email
	}

	if len(u.Phone) > 0 {
		m.Phone = u.Phone
	}

	if len(u.Img) > 0 {
		m.Img.String = u.Img
	}

	return m
}

// passReq ...
type PassReq struct {
	OldPassword string `json:"old_password"`
	Pass        string `json:"password"`
}

type PassPhoneReq struct {
	Password       string `json:"password"`
	Phone          string `json:"phone"`
	RetypePassword string `json:"retype_password"`
}

type PassEmailReq struct {
	Phone string `json:"phone"`
	Email string `json:"email"`
}
