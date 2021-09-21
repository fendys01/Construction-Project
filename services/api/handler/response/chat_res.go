package response

import (
	"panorama/services/api/model"
	"time"
)

// ChatGroupRes ...
type ChatGroupRes struct {
	ChatGroupCode     string   `json:"chat_group_code"`
	Name              string   `json:"name"`
	CreatedBy         string   `json:"created_by"`
	ChatGroupType     string   `json:"chat_group_type"`
	ChatGroupRelation []string `json:"chat_group_relation"`
}

// Transform ChatGroupRes ...
func (r ChatGroupRes) Transform(m model.ChatGroupEnt) ChatGroupRes {
	r.ChatGroupCode = m.ChatGroupCode
	r.Name = m.Name
	r.CreatedBy = m.Member.Name
	r.ChatGroupType = m.ChatGroupType
	r.ChatGroupRelation = m.ChatGroupRelation

	return r
}

// ChatGroupMessageRes ...
type ChatGroupMessageRes struct {
	ID          int32     `json:"id"`
	Name        string    `json:"name"`
	Message     string    `json:"message"`
	Role        string    `json:"role"`
	IsRead      bool      `json:"is_read"`
	UserCode    string    `json:"user_code"`
	CreatedDate time.Time `json:"created_date"`
}

// Transform ChatGroupMessageRes ...
func (r ChatGroupMessageRes) Transform(m model.ChatMessagesEnt) ChatGroupMessageRes {
	r.ID = m.ID
	r.Name = m.Name
	r.Message = m.Message
	r.Role = m.Role
	r.IsRead = m.IsRead
	r.UserCode = m.UserCode
	r.CreatedDate = m.CreatedDate

	return r
}

// ChatGroupOrderRes ...
type ChatGroupOrderRes struct {
	ChatGroupCode            string    `json:"chat_group_code"`
	Name                     string    `json:"chat_group_name"`
	ChatGroupType            string    `json:"chat_group_type"`
	ChatGroupStatus          bool      `json:"chat_group_status"`
	ChatGroupTotal           int       `json:"chat_group_total"`
	ChatGroupLastMessage     string    `json:"chat_group_last_message"`
	ChatGroupUnreadTotal     int       `json:"chat_group_unread_total"`
	MemberCode               string    `json:"member_code"`
	MemberName               string    `json:"member_name"`
	MemberEmail              string    `json:"member_email"`
	MemberImg                string    `json:"member_img"`
	TcName                   string    `json:"tc_name"`
	TcCode                   string    `json:"tc_code"`
	ItinDestination          string    `json:"itin_destination"`
	ItinTitle                string    `json:"itin_title"`
	ItinCode                 string    `json:"itin_code"`
	ItinTripDayDuration      int       `json:"itin_trip_day_duration"`
	OrderCode                string    `json:"order_code"`
	OrderStatus              string    `json:"order_status"`
	OrderType                string    `json:"order_type"`
	OrderStatusDescription   string    `json:"order_status_description"`
	TCReplacementDescription string    `json:"tc_replacement_description"`
	CreatedDate              time.Time `json:"created_date"`
}

// Transform ChatGroupMessageRes ...
func (r ChatGroupOrderRes) Transform(m model.ChatGroupEnt) ChatGroupOrderRes {
	r.ChatGroupCode = m.ChatGroupCode
	r.Name = m.Name
	r.ChatGroupType = m.ChatGroupType
	r.ChatGroupStatus = m.Status
	r.ChatGroupTotal = int(m.ChatGroupTotal)
	r.ChatGroupLastMessage = m.ChatGroupLastMessage
	r.ChatGroupUnreadTotal = int(m.ChatGroupUnreadTotal)
	r.MemberCode = m.Member.MemberCode
	r.MemberName = m.Member.Name
	r.MemberEmail = m.Member.Email
	r.MemberImg = m.Member.Img.String
	r.TcCode = m.User.UserCode
	r.TcName = m.User.Name
	r.ItinDestination = m.MemberItin.Destination
	r.ItinTitle = m.MemberItin.Title
	r.ItinCode = m.MemberItin.ItinCode
	r.ItinTripDayDuration = int(m.MemberItin.DayPeriod)
	r.OrderCode = m.Order.OrderCode
	r.OrderStatus = m.Order.OrderStatus
	r.OrderType = m.Order.OrderType
	r.OrderStatusDescription = m.Order.OrderStatusDescription
	r.TCReplacementDescription = m.TCReplacementDescription
	r.CreatedDate = m.CreatedDate

	return r
}

// ChatGroupHistoryRes ...
type ChatGroupHistoryRes struct {
	ChatGroupCode       string                `json:"chat_group_code"`
	Name                string                `json:"name"`
	ChatGroupType       string                `json:"chat_group_type"`
	TcDetail            ListUserGroupChat     `json:"tc"`
	CreatedBy           CreatedByGroupChat    `json:"createdby_member"`
	ItinMemberSimpleRes ItinMemberSimpleRes   `json:"itin_member"`
	OrderSimpleRes      OrderSimpleRes        `json:"order_detail"`
	ChatGroupMessageRes []ChatGroupMessageRes `json:"history_chat"`
	ListUser            []ListUserGroupChat   `json:"list_user"`
}

// Transform ChatGroupRes ...
func (r ChatGroupHistoryRes) Transform(m model.ChatGroupEnt) ChatGroupHistoryRes {
	r.ChatGroupCode = m.ChatGroupCode
	r.Name = m.Name
	r.ChatGroupType = m.ChatGroupType
	r.ItinMemberSimpleRes = r.ItinMemberSimpleRes.Transform(m.MemberItin)
	r.OrderSimpleRes = r.OrderSimpleRes.Transform(m.Order)

	r.CreatedBy = r.CreatedBy.Transform(m)

	var listResponse []ChatGroupMessageRes
	for _, g := range m.ChatMessagesEnt {
		if g.ID > 0 {
			var res ChatGroupMessageRes
			res = res.Transform(g)
			listResponse = append(listResponse, res)
		}
	}

	var listUSer []ListUserGroupChat
	var tc ListUserGroupChat

	for _, g := range m.ListUser {
		if len(g.Email) > 0 {
			var r ListUserGroupChat
			r = r.Transform(g)
			if g.Role == "tc" {
				tc = r
			} else {
				listUSer = append(listUSer, r)
			}
		}
	}

	r.ChatGroupMessageRes = listResponse
	r.ListUser = listUSer
	r.TcDetail = tc
	return r
}

// ListUserGroupChat ...
type ListUserGroupChat struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Image    string `json:"image"`
	Role     string `json:"role"`
	UserCode string `json:"user_code"`
}

// Transform ListUserGroupChat ...
func (r ListUserGroupChat) Transform(m model.ChatMessagesEnt) ListUserGroupChat {
	r.Name = m.Name
	r.Email = m.Email
	r.Image = m.Image.String
	r.Role = m.Role
	r.UserCode = m.UserCode

	return r
}

// CreatedByGroupChat ...
type CreatedByGroupChat struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Image    string `json:"image"`
	Role     string `json:"role"`
	UserCode string `json:"user_code"`
}

// Transform ListUserGroupChat ...
func (r CreatedByGroupChat) Transform(m model.ChatGroupEnt) CreatedByGroupChat {
	r.Name = m.Member.Name
	r.Email = m.Member.Email
	r.Image = m.Member.Img.String
	r.Role = "customer"
	r.UserCode = m.Member.MemberCode

	return r
}

// OrderSimpleRes ...
type OrderSimpleRes struct {
	OrderType string `json:"order_type"`
	OrderCode string `json:"order_code"`
}

// Transform from itin member model to itin member response
func (r OrderSimpleRes) Transform(i model.OrderEnt) OrderSimpleRes {

	r.OrderType = i.OrderType
	r.OrderCode = i.OrderCode

	return r
}
