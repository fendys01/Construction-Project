package response

import (
	"panorama/services/api/model"
)

type ChatGroupRes struct {
	ChatGroupCode string `json:"chat_group_code"`
	Name          string `json:"name"`
	CreatedBy     string `json:"member_name"`
}

func (r ChatGroupRes) Transform(m model.ChatGroupEnt) ChatGroupRes {
	r.ChatGroupCode = m.ChatGroupCode
	r.Name = m.Name
	r.CreatedBy = m.Member.Name

	return r
}

type ChatGroupMessageRes struct {
	Name    string `json:"name"`
	Message string `json:"message"`
	Role    string `json:"role"`
}

func (r ChatGroupMessageRes) Transform(m model.ChatMessagesEnt) ChatGroupMessageRes {
	r.Name = m.Name
	r.Message = m.Message
	r.Role = m.Role

	return r
}
