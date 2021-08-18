package request

// AddMemberReq ...
type AddNotifReq struct {
	MemberCode     string `json:"member_code" validate:"required"`
	Type           int32 `json:"type" validate:"required"`
	Title          string `json:"title" validate:"required"`
	Content        string `json:"content" validate:"required"`
	Link           string `json:"link"`
}