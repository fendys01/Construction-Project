package request

// NewChatGroupReq : request payload chat group
type NewChatGroupReq struct {
	MemberItinCode    string   `json:"member_itin_code"`
	Name              string   `json:"name" validate:"required"`
	ChatGroupType     string   `json:"chat_group_type" validate:"required"`
	TcCode            string   `json:"tc_code"`
	ChatGroupRelation []string `json:"chat_group_relation"`
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

// ChatGroupMessagesIsRead ...
type ChatGroupMessagesIsRead struct {
	ChatGroupCode string `json:"chat_group_code" validate:"required"`
	UserCode      string `json:"user_code" validate:"required"`
	Role          string `json:"role" validate:"required"`
}
