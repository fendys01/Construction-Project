package request

// AddMemberReq ...
type AddStuffReq struct {
	Name 			string `json:"name" validate:"required"`
	Image           string `json:"image" validate:"required"`
	Description     string `json:"description" validate:"required"`
	Price        	string `json:"price" validate:"required"`
	Type            int32  `json:"type"`
}