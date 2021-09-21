package request

import (
	"database/sql"
	"math/rand"
	"panorama/lib/utils"
	"panorama/services/api/model"
	"time"
)

type MemberItinReq struct {
	Title         string                   `json:"title" validate:"required"`
	EstPrice      int                      `json:"est_price"`
	MemberCode    string                   `json:"member_code"`
	StartDate     string                   `json:"start_date"`
	EndDate       string                   `json:"end_date"`
	Destination   string                   `json:"destination"`
	Details       []map[string]interface{} `json:"details" validate:"required"`
	Img           string                   `json:"img"`
	GroupChatCode string                   `json:"group_chat_code"`
	GroupMembers  []map[string]interface{} `json:"group_members"`
	OrderType     string                   `json:"order_type"`
}

func (req MemberItinReq) ToMemberItinEnt(isNew bool) (model.MemberItinEnt, error) {
	var code string

	if isNew {
		rand.Seed(time.Now().UnixNano())
		code, _ = utils.Generate(`MBIT-[a-z0-9]{6}`)
	}

	memberItin := model.MemberItinEnt{
		ItinCode:      code,
		Title:         req.Title,
		EstPrice:      sql.NullInt64{Int64: int64(req.EstPrice), Valid: true},
		Details:       req.Details,
		Img:           sql.NullString{String: req.Img, Valid: true},
		ChatGroupCode: req.GroupChatCode,
	}

	if req.Destination != "" {
		memberItin.Destination = req.Destination
	}

	if req.StartDate != "" && req.EndDate != "" {
		sDate, err := time.Parse("2006-01-02 15:04:05", req.StartDate)
		if err != nil {
			return model.MemberItinEnt{}, err
		}
		eDate, err := time.Parse("2006-01-02 15:04:05", req.EndDate)
		if err != nil {
			return model.MemberItinEnt{}, err
		}
		memberItin.StartDate = sDate
		memberItin.EndDate = eDate
	}

	return memberItin, nil
}
