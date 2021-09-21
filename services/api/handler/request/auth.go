package request

type LoginReq struct {
	Username string `json:"username" validate:"required"` // can be an email or phone number
	Password string `json:"password" validate:"required"`
}

type AuthReq struct {
	Username string `json:"username" validate:"required"` // can be an email or phone number
}

type CheckTokenReq struct {
	Username string `json:"username" validate:"required"`
	Token    string `json:"token" validate:"required"`
}

type ChangePassForgotReq struct {
	Username       string `json:"username" validate:"required"`
	Password       string `json:"password" validate:"required"`
	RetypePassword string `json:"retype_password" validate:"required"`
}

// CheckPhoneReq ...
type CheckPhoneReq struct {
	Phone string `json:"phone" validate:"required"`
	Token string `json:"token" validate:"required"`
}

// PassOldReq ...
type PassOldReq struct {
	OldPassword string `json:"old_password" validate:"required"`
	Pass        string `json:"password" validate:"required"`
}

type ChangePassReq struct {
	Password       string `json:"password" validate:"required"`
	RetypePassword string `json:"retype_password" validate:"required"`
}
