package request

import "Contruction-Project/services/api/model"

// AddMemberReq ...
type AddMemberReq struct {
	Username       string `json:"username" validate:"required"`
	Name           string `json:"name" validate:"required"`
	Email          string `json:"email" validate:"required,email"`
	Phone          string `json:"phone" validate:"required"`
	Img            string `json:"img"`
	Password       string `json:"password" validate:"required"`
	RetypePassword string `json:"retype_password" validate:"required"`
}

// UpdateMemberReq ...
type UpdateMemberReq struct {
	Username string `json:"username"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
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

	return m
}

// passReq ...
type passReq struct {
	Pass string `json:"password"`
}
