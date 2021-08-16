package request

type LoginReq struct {
	Username string `json:"username" validate:"required"` // can be an email or phone number
	Password string `json:"password" validate:"required"`
}

type ForgotPassReq struct {
	Username string `json:"username" validate:"required"` // can be an email or phone number
}

type ForgotPassTokenReq struct {
	Username string `json:"username" validate:"required"`
	Token    string `json:"token" validate:"required"`
}

type ChangePassForgotReq struct {
	Username       string `json:"username" validate:"required"`
	Password       string `json:"password" validate:"required"`
	RetypePassword string `json:"retype_password" validate:"required"`
}

// RegisterPhoneValidationReq ...
type RegisterPhoneValidationReq struct {
	Phone string `json:"phone" validate:"required"`
	Token string `json:"token" validate:"required"`
}
