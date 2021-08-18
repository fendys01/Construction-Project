package request

// NewChatGroupReq : request payload chat group
type NewChatGroupReq struct {
	MemberItinCode string `json:"member_itin_code"`
	Name           string `json:"name" validate:"required"`
	ChatGroupType  string `json:"chat_group_type" validate:"required"`
	ChatWithTc     bool   `json:"chat_with_tc"`
	// ChatGroupRelation []string `json:"chat_group_relation"`
}

// ChatGroupReq ...
type ChatGroupReq struct {
	ChatGroupCode string `json:"chat_group_code" validate:"required"`
}

// ChatGroupMessagesReq ...
type ChatGroupMessagesReq struct {
	ChatGroupCode string `json:"chat_group_code" validate:"required"`
	Message       string `json:"message" validate:"required"`
}
